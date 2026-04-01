package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
)

// InstallAddOn godoc
// @Summary Install a new Addon
// @Description Create a new Addons with the input paylod
// @Tags Addons
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.Environment true "Create Addons"
// @Success 200 {object} doc.Environment
// @Router /environment/{id}/addons [post]

var queue helper.QueueDefInterface

func (server *Server) InstallAddOn(w http.ResponseWriter, r *http.Request) {
	parentEnvironment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !parentEnvironment.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment is stopped"))
		return
	}
	_, err = CheckPermission(server.DB, r, "write", parentEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	addonEnvironment := models.Environment{}
	err = json.Unmarshal(body, &addonEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	tempEnv := models.Environment{}
	err = server.DB.Model(&addonEnvironment).
		Take(&tempEnv, "parent_id = ? and plugin_version_id = ?", parentEnvironment.ID, addonEnvironment.PluginVersionID).Error
	if err == nil && tempEnv.ID > 0 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("addon is already installed"))
		return
	}
	addonEnvironment.PrepareAddon(parentEnvironment)
	err = addonEnvironment.Validate(false)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	plugin, err := pluginVersionInterface.Find(server.DB, addonEnvironment.PluginVersionID)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("addon not found"))
		return
	}
	addonEnvironment.PluginVersion = plugin
	var variablesMap = helper.GetDefaultVariable(plugin)
	if addonEnvironment.Variables.RawMessage != nil {
		var currentVariableMap map[string]interface{}
		err := json.Unmarshal(addonEnvironment.Variables.RawMessage, &currentVariableMap)
		if err != nil {
			log.Error(err)
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		helper.ReplaceCustomVariables(currentVariableMap, variablesMap)
	}
	jsonString, _ := json.Marshal(variablesMap)
	addonEnvironment.Variables.RawMessage = jsonString
	resource, err := resourceInterface.Find(server.DB, addonEnvironment.ResourceID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid resource"))
		return
	}

	if resource.Memory < addonEnvironment.PluginVersion.Plugin.MinMemory {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dmb memory per replica", addonEnvironment.PluginVersion.Plugin.MinMemory))
		return
	}
	if resource.Cores < addonEnvironment.PluginVersion.Plugin.MinCpu {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dm cpu per replica", addonEnvironment.PluginVersion.Plugin.MinCpu))
		return
	}
	addonEnvironment.Resource = resource
	if message, ok := addonEnvironment.IsValidResource(server.DB, parentEnvironment.ApplicationID, nil, resource, 1024); !ok {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New(message))
		return
	}
	data, err := addonEnvironment.Save(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("addon not installed"))
		return
	}
	if parentEnvironment.ServiceType == 1 {
		req := models.CIRequest{
			ImageRepoService:  parentEnvironment.Application.Cluster.ImageRegistry.Service,
			ImageRepoPassword: parentEnvironment.Application.Cluster.ImageRegistry.Password,
			ImageRepoUsername: parentEnvironment.Application.Cluster.ImageRegistry.UserName,
			ImageRepoProject:  parentEnvironment.Application.Cluster.ImageRegistry.ProjectName,
		}
		data.CiRequest = &req
		parentEnvironment.CiRequest = &req
	}
	data.PluginVersion.Plugin.SupportCi = false
	// this is safe as it is addons plugin
	envJ, _ := parentEnvironment.ToJson()
	dataJ, _ := data.ToJson()
	if parentEnvironment.ServiceType == 4 {
		packageName := parentEnvironment.Application.OperatorPackageName
		opReq := &models.OperatorRequest{}
		if parentEnvironment.Application.Cluster != nil {
			opReq, err = operatorRequestRepo.FindByNameAndClusterID(server.DB, parentEnvironment.Application.Cluster.ClusterRequestID, packageName)
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, errors.New("invalid service type"))
				return
			}
		}
		envJ["operators"] = map[string]interface{}{
			"namespace": helper.GetOperatorNamespace(opReq),
		}
		dataJ["operators"] = map[string]interface{}{
			"namespace": helper.GetOperatorNamespace(opReq),
		}
	}

	err = queue.Publish(constants.InstallAddOn, &map[string]interface{}{
		"environment": envJ,
		"add_on":      dataJ,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: parentEnvironment.Application.Name,
		ApplicationID:   int64(parentEnvironment.Application.ID),
		EnvironmentID:   int64(parentEnvironment.ID),
		EnvironmentName: parentEnvironment.Name,
		ProjectName:     parentEnvironment.Application.Project.Name,
		ProjectID:       int64(parentEnvironment.Application.Project.ID),
		OrganizationID:  int64(parentEnvironment.Application.Project.OrganizationId),
	}
	if parentEnvironment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = parentEnvironment.Application.Project.Organization.Name
	}
	uid, _, _ := auth.ExtractTokenID(r)
	notify(server.NotifyClient, notifyInfo, "plugin-install", fmt.Sprintf("environment %s plugin %s is being installed", parentEnvironment.Name, plugin.Plugin.Name), "info", int64(uid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message":     "Install addon triggered",
			"environment": parentEnvironment,
			"add_on":      data,
		})
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(parentEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	addonResponse, err := CreateEnvironmentResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"message":     "Install addon triggered",
			"environment": environmentResponse,
			"add_on":      addonResponse,
		},
		Success: 1,
		Message: "Success",
	})
}

// UnInstallAddOn godoc
// @Summary Uninstall Addons
// @Description Uninstall Addons with the addon id
// @Tags Addons
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param aid path int true "Addon id"
// @Success 200 {object} object
// @Router /environment/{id}/addons/{aid} [delete]
func (server *Server) UnInstallAddOn(w http.ResponseWriter, r *http.Request) {
	parentEnvironment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !parentEnvironment.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment already stopped"))
		return
	}
	_, err = CheckPermission(server.DB, r, "write", parentEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["aid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	plg := models.Environment{}
	addOnEnvironment, err := plg.FindAddon(server.DB, aid, uint64(parentEnvironment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("addon not found"))
		return
	}
	addonName := addOnEnvironment.PluginVersion.Plugin.Name
	_, err = addOnEnvironment.Delete(server.DB, uint64(addOnEnvironment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("addon not found"))
		return
	}
	namespace := helper.GetAddOnRelease(parentEnvironment, addOnEnvironment)
	for _, data := range parentEnvironment.Storage {
		if strings.Contains(data.Name, namespace) {
			_, err = storageInterface.Delete(server.DB, uint64(data.ID))
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, err)
				return
			}
		}
	}

	err = queue.Publish(constants.UnInstallAddOn, &map[string]interface{}{
		"environment": parentEnvironment,
		"add_on":      addOnEnvironment,
	})

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: parentEnvironment.Application.Name,
		ApplicationID:   int64(parentEnvironment.Application.ID),
		EnvironmentID:   int64(parentEnvironment.ID),
		EnvironmentName: parentEnvironment.Name,
		ProjectName:     parentEnvironment.Application.Project.Name,
		ProjectID:       int64(parentEnvironment.Application.Project.ID),
		OrganizationID:  int64(parentEnvironment.Application.Project.OrganizationId),
	}
	if parentEnvironment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = parentEnvironment.Application.Project.Organization.Name
	}
	uid, _, _ := auth.ExtractTokenID(r)
	notify(server.NotifyClient, notifyInfo, "plugin-uninstall", fmt.Sprintf("environment %s plugin %s is being uninstalled", parentEnvironment.Name, addonName), "info", int64(uid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "Uninstall addon triggered",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetAddOnState godoc
// @Summary Get Addons state
// @Description Get state of Addons installed with the input id
// @Tags Addons
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} object
// @Router /environment/{id}/addons-status [get]
func (server *Server) GetAddOnState(w http.ResponseWriter, r *http.Request) {
	parentEnvironment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	_, err = CheckPermission(server.DB, r, "read", parentEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	podStatus, err := server.StoreClient.PodState().GetAddOnState(fmt.Sprintf("%s-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), parentEnvironment.ApplicationID, parentEnvironment.ID))
	if err != nil {
		responses.JSON(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, podStatus)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    podStatus,
		Success: 1,
		Message: "Success",
	})
}

// GetAddonEnvironments godoc
// @Summary Get all Addons
// @Description Get all Addons with the input environment id
// @Tags Addons
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {array} doc.Environment
// @Router /environment/{id}/addons [get]
func (server *Server) GetAddonEnvironments(w http.ResponseWriter, r *http.Request) {
	parentEnvironment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !parentEnvironment.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment already stopped"))
		return
	}
	_, err = CheckPermission(server.DB, r, "read", parentEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	data, err := parentEnvironment.FindAllAddons(server.DB, uint64(parentEnvironment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	datas := []map[string]interface{}{}
	for _, v := range data {
		js, _ := v.ToJson()
		value := map[string]interface{}{}
		if v.PluginVersion.Plugin.ServiceDetail.RawMessage != nil {
			err = json.Unmarshal(v.PluginVersion.Plugin.ServiceDetail.RawMessage, &value)
			if err != nil {
				value = map[string]interface{}{}
			}
			if d, ok := value["service_type"]; ok && d == constants.ServiceTypeExternal {
				value["service_external_url"] = "/" + v.PluginVersion.Plugin.Name
			}
			if v.ExternalURL &&
				parentEnvironment.Application != nil &&
				parentEnvironment.Application.Cluster != nil &&
				parentEnvironment.Application.Cluster.DNS != nil {
				v.Parent = parentEnvironment
				addonNamespace := helper.GetAddonNamespace(v)
				baseDomain := parentEnvironment.Application.Cluster.DNS.BaseDomain
				baseDomain = strings.TrimSuffix(baseDomain, ".")
				if err == nil {
					value["external_url"] = fmt.Sprintf("%s.%s", addonNamespace, baseDomain)
				}
			}
			js["service_detail"] = value
		}
		datas = append(datas, js)
	}
	envs := []doc.Environment{}
	for _, env := range datas {
		envByte, _ := json.Marshal(env)
		envData := doc.Environment{}
		err = json.Unmarshal(envByte, &envData)
		if err != nil {
			log.Error(err)
		}
		envs = append(envs, envData)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"addons": datas,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"addons": envs,
		},
		Success: 1,
		Message: "Success",
	})
}

// GetAddonEnvironment godoc
// @Summary Get all Addons
// @Description Get all Addons with the input environment id
// @Tags Addons
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param aid path int true "Addon id"
// @Success 200 {object} doc.Environment
// @Router /environment/{id}/addons/{aid} [get]
func (server *Server) GetAddonEnvironment(w http.ResponseWriter, r *http.Request) {
	parentEnvironment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !parentEnvironment.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment already stopped"))
		return
	}
	_, err = CheckPermission(server.DB, r, "read", parentEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["aid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	plg := models.Environment{}
	data, err := plg.FindAddon(server.DB, aid, uint64(parentEnvironment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	metadata, err := server.StoreClient.ReleaseInfo().GetData(GetAddOnReleaseId(uint(data.ParentID), data.PluginVersion.Plugin.ID))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	var addon models.Plugin
	value, err := GetServiceDetails(data.PluginVersion.Url)
	if err != nil {
		log.Error(err)
		value = map[string]interface{}{}
	}
	master_enabled := false
	if data.Variables.RawMessage != nil {
		var currentVariableMap map[string]interface{}
		err := json.Unmarshal(data.Variables.RawMessage, &currentVariableMap)
		if err == nil {
			if _, ok := currentVariableMap["Replication"]; ok {
				if d, ok := currentVariableMap["Replication"].(map[string]interface{})["enabled"]; ok && d.(bool) {
					master_enabled = true
				}
			}
		}
	}

	if d, ok := value["master_slave"]; ok && d == true && master_enabled {
		value["slave_url"] = fmt.Sprintf("%s.%s.svc.cluster.local", GetServiceName(uint(data.ParentID), data.PluginVersion.Plugin.ID, data.PluginVersion.Plugin.Name, "slave"), helper.GetNamespace(parentEnvironment))
	}
	value["master_url"] = fmt.Sprintf("%s.%s.svc.cluster.local", GetServiceName(uint(data.ParentID), data.PluginVersion.Plugin.ID, data.PluginVersion.Plugin.Name, ""), helper.GetNamespace(parentEnvironment))
	if d, ok := value["service_type"]; ok && d == constants.ServiceTypeExternal {
		value["service_external_url"] = "/" + data.PluginVersion.Plugin.Name
	}
	if data.ExternalURL {
		data.Parent = parentEnvironment
		addonNamespace := helper.GetAddonNamespace(data)
		baseDomain := parentEnvironment.Application.Cluster.DNS.BaseDomain
		baseDomain = strings.TrimSuffix(baseDomain, ".")
		if err == nil {
			value["external_url"] = fmt.Sprintf("%s.%s", addonNamespace, baseDomain)
		}
	}
	delete(value, "master_slave")
	valueBytes, _ := json.Marshal(value)
	data.PluginVersion.Plugin.ServiceDetail.RawMessage = valueBytes
	var metadataJson = &map[string]interface{}{}
	if metadata != nil {
		var manifests = strings.Split(metadata.Manifest, "---")
		var manifestRespones []models.Manifest
		for _, v := range manifests {
			manifestJson := &models.Manifest{}
			err := yaml.Unmarshal([]byte(v), manifestJson)
			if err == nil && (manifestJson.Kind == "Secret" || manifestJson.Kind == "Deployment") {
				manifestRespones = append(manifestRespones, *manifestJson)
			}
			if err != nil {
				log.Error("invalid manifest")
			}
		}
		js := models.ToJson(manifestRespones)
		secrets := gjson.Get(js, "#(kind==Secret)#.data")
		envs := gjson.Get(js, "#(kind==Deployment).spec.template.spec.containers")
		envs_map := new([]map[string]interface{})
		secret_map := new([]map[string]interface{})
		_ = json.Unmarshal([]byte(envs.Raw), envs_map)
		_ = json.Unmarshal([]byte(secrets.Raw), secret_map)
		metadataJson = &map[string]interface{}{
			"name":      metadata.Name,
			"namespace": metadata.Namespace,
			"version":   metadata.Version,
			"info":      metadata.Info,
			"releaseId": metadata.ReleaseId,
			//"cname":     metadata.CName,
			"secret": secret_map,
			"envs":   envs_map,
			"addon":  addon,
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"environment": data,
			"metadata":    metadataJson,
		})
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"environment": environmentResponse,
			"metadata":    metadataJson,
		},
		Success: 1,
		Message: "Success",
	})
}

// UpdateAddonsEnvironment godoc
// @Summary Update a Addon
// @Description Update a Addon with the input paylod
// @Tags Addons
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param aid path int true "Addon id"
// @Param body body doc.Environment true "Update Addons"
// @Success 200 {object} doc.Environment
// @Router /environment/{id}/addons{aid} [put]
func (server *Server) UpdateAddonsEnvironment(w http.ResponseWriter, r *http.Request) {

	parentEnvironment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !parentEnvironment.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment already stopped"))
		return
	}
	_, err = CheckPermission(server.DB, r, "write", parentEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	extras := ""
	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["aid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	plg := models.Environment{}
	originalAddOnEnvironment, err := plg.FindAddon(server.DB, aid, uint64(parentEnvironment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("addon not found"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	updateAddOnEnvironment := models.Environment{}
	err = json.Unmarshal(body, &updateAddOnEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	updateAddOnEnvironment.PrepareAddon(parentEnvironment)
	if updateAddOnEnvironment.PluginVersionID > 0 {
		plugin, err := pluginVersionInterface.Find(server.DB, updateAddOnEnvironment.PluginVersionID)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("addon not found"))
			return
		}
		updateAddOnEnvironment.PluginVersion = plugin
	} else {
		updateAddOnEnvironment.PluginVersion = originalAddOnEnvironment.PluginVersion
	}
	if updateAddOnEnvironment.ResourceID != 0 {
		resource, err := resourceInterface.Find(server.DB, updateAddOnEnvironment.ResourceID)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid resource"))
			return
		}
		updateAddOnEnvironment.Resource = resource
		extras += fmt.Sprintf("resource to %d milli core and %d MB RAM ", resource.Cores, resource.Memory)
	}
	if updateAddOnEnvironment.Resource.Memory < updateAddOnEnvironment.PluginVersion.Plugin.MinMemory {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dmb memory per replica", updateAddOnEnvironment.PluginVersion.Plugin.MinMemory))
		return
	}
	if updateAddOnEnvironment.Resource.Cores < updateAddOnEnvironment.PluginVersion.Plugin.MinCpu {
		responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dm cpu per replica", updateAddOnEnvironment.PluginVersion.Plugin.MinCpu))
		return
	}
	if message, ok := originalAddOnEnvironment.IsValidResource(server.DB, parentEnvironment.ApplicationID, &updateAddOnEnvironment, updateAddOnEnvironment.Resource, 0); !ok {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New(message))
		return
	}

	updateAddOnEnvironment.ID = originalAddOnEnvironment.ID
	if updateAddOnEnvironment.AutoScaler.RawMessage != nil {
		var value models.AutoScaler
		err = json.Unmarshal(updateAddOnEnvironment.AutoScaler.RawMessage, &value)
		if err != nil {
			log.Error(err)
		}
		if !helper.IsEmptyStruct(value.HorizontalPodAutoScaler) {
			updateAddOnEnvironment.Replicas = uint16(value.HorizontalPodAutoScaler.MaxReplicas)
		} else {
			updateAddOnEnvironment.Replicas = 1
		}
	} else {
		if updateAddOnEnvironment.Replicas == 0 {
			updateAddOnEnvironment.Replicas = originalAddOnEnvironment.Replicas
		}
	}

	environmentUpdated, err := updateAddOnEnvironment.Update(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	env := &models.Environment{}
	env, _ = env.Find(server.DB, uint64(updateAddOnEnvironment.ID))

	plugin, err := pluginVersionInterface.Find(server.DB, env.PluginVersionID)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("addon not found"))
		return
	}
	env.PluginVersion = plugin
	req := models.CIRequest{
		ImageRepoService:  parentEnvironment.Application.Cluster.ImageRegistry.Service,
		ImageRepoPassword: parentEnvironment.Application.Cluster.ImageRegistry.Password,
		ImageRepoUsername: parentEnvironment.Application.Cluster.ImageRegistry.UserName,
		ImageRepoProject:  parentEnvironment.Application.Cluster.ImageRegistry.ProjectName,
	}
	env.CiRequest = &req
	env.PluginVersion.Plugin.SupportCi = false
	err = queue.Publish(constants.InstallAddOn, map[string]interface{}{
		"environment": parentEnvironment,
		"add_on":      env,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if updateAddOnEnvironment.Replicas != originalAddOnEnvironment.Replicas {
		extras += fmt.Sprintf("replicas from %d to %d ", originalAddOnEnvironment.Replicas, updateAddOnEnvironment.Replicas)
	}
	log.Debug(extras)
	//if environment.Attributes != env.Attributes {
	//	extras += "environment variables "
	//}
	//_, _ = server.SaveActivityWithJson("update", "environment", uid, parentEnvironment., env.Application, env, extras)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, environmentUpdated)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environmentUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateAddonsEnvironment godoc
// @Summary Update a Addon
// @Description Update a Addon with the input paylod
// @Tags Addons
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param aid path int true "Addon id"
// @Param body body doc.Environment true "Update Addons"
// @Success 200 {object} doc.Environment
// @Router /environment/{id}/addons{aid}/external-url [put]
func (server *Server) UpdateAddonsEnvironmentExternalURL(w http.ResponseWriter, r *http.Request) {
	parentEnvironment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !parentEnvironment.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment already stopped"))
		return
	}
	_, err = CheckPermission(server.DB, r, "write", parentEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["aid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	plg := models.Environment{}
	originalAddOnEnvironment, err := plg.FindAddon(server.DB, aid, uint64(parentEnvironment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("addon not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	externalURL := models.AddonExternalURLRequest{}
	err = json.Unmarshal(body, &externalURL)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	originalAddOnEnvironment.ExternalURL = externalURL.ExternalURL
	environmentUpdated, err := originalAddOnEnvironment.Update(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	env := &models.Environment{}
	env, _ = env.Find(server.DB, uint64(originalAddOnEnvironment.ID))
	err = queue.Publish(constants.AddonExternalURL, map[string]interface{}{
		"environment": parentEnvironment,
		"add_on":      env,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environmentUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

func GetServiceName(envId uint, pluginId uint, pluginName string, masterSlave string) string {
	if masterSlave != "" {
		return fmt.Sprintf("%s-addon-%d-%d-%s-%s", os.Getenv("GCLOUD_NAMESPACE"), envId, pluginId, pluginName, masterSlave)
	}
	return fmt.Sprintf("%s-addon-%d-%d-%s", os.Getenv("GCLOUD_NAMESPACE"), envId, pluginId, pluginName)
}
func GetAddOnReleaseId(envId uint, pluginId uint) string {
	return fmt.Sprintf("%s-addon-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), envId, pluginId)
}
func GetServiceDetails(url string) (map[string]interface{}, error) {
	file, _ := os.ReadFile(url + "/config.json")
	res := make(map[string]interface{})
	err := json.Unmarshal([]byte(file), &res)
	if err != nil {
		log.Error("unmarshall err :: ", err)
	}
	data := map[string]interface{}{}
	if service, ok := res["service_detail"]; ok {
		data = service.(map[string]interface{})
	}
	return data, err
}
