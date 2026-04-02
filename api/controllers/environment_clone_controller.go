package controllers

import (
	"01cloud-api/api/auth"
	pb "01cloud-api/api/backup/proto"
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/externalsecret"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"
	"01cloud-api/api/utils/logging"
	"01cloud-api/api/utils/prometheus"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// CloneEnvironment godoc
// @Summary Clone Environment
// @Description Clone Environment with the input payload
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Environment true "Clone Environment"
// @Success 201 {object} doc.Environment
// @Router /environment/{id}/clone [post]
func (server *Server) CloneEnvironment(w http.ResponseWriter, r *http.Request) {
	_, oid, _ := auth.ExtractTokenID(r)
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	oldEnvName := environment.Name
	oldEnvId := environment.ID
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.CloneEnvironment{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment.Name = environment.Name + "-clone"
	if data.Name != "" {
		environment.Name = data.Name
	}
	environment.Prepare()
	storages := environment.Storage
	environment.Storage = nil
	err = environment.Validate(true)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = logging.ValidateLoggingRequest(environment.ExternalLogging)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	application, err := applicationInterface.Find(server.DB, environment.ApplicationID)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("invalid application"))
		return
	}

	if !application.Cluster.Active {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("cluster is disabled, Please contact support team"))
		return
	}
	if environment.IsNameExists(server.DB, uint(environment.ApplicationID), environment.Name) {
		responses.ERROR(w, http.StatusForbidden, errors.New("name already exists"))
		return
	}

	if user.ID != uint(application.Project.UserID) && !application.Project.User.IsAdmin {
		if !authInterface.IsAuthorizedOrganization(server.DB, user.ID, oid) &&
			!authInterface.IsWriteAuthorizedProject(server.DB, uint64(user.ID), application.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to add environment"))
			return
		}
	}
	environment.Application = application
	if environment.AutoScaler.RawMessage != nil {
		var value models.AutoScaler
		err = json.Unmarshal(environment.AutoScaler.RawMessage, &value)
		if err != nil {
			log.Error(err)
		}
		if !helper.IsEmptyStruct(value.HorizontalPodAutoScaler) {
			environment.Replicas = uint16(value.HorizontalPodAutoScaler.MaxReplicas)
		} else {
			environment.Replicas = 1
		}
	} else {
		environment.Replicas = 1
	}
	resource, err := resourceInterface.Find(server.DB, environment.ResourceID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid resource"))
		return
	}
	if environment.LoadBalancerID > 0 {
		loadBalancer, err := loadbalancerInterface.Find(server.DB, environment.LoadBalancerID)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid load balancer"))
			return
		}
		if loadBalancer.ClusterID != environment.Application.ClusterID {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid cluster, load balancer and environment must be in same region"))
			return
		}
		environment.LoadBalancer = loadBalancer
	}
	if environment.ServiceType < 2 {
		if resource.Memory < environment.Application.Plugin.MinMemory {
			responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dmb memory per replica", environment.Application.Plugin.MinMemory))
			return
		}
		if resource.Cores < environment.Application.Plugin.MinCpu {
			responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dm cpu per replica", environment.Application.Plugin.MinCpu))
			return
		}
	} else {
		environment.ImageUrl = application.ImageUrl
	}
	environment.Resource = nil

	if message, ok := environment.IsValidResource(server.DB, environment.ApplicationID, nil, resource, 0); !ok {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New(message))
		return
	}
	environment.Resource = resource

	pluginVersion, err := pluginVersionInterface.FindLatestPluginVersion(server.DB, environment.Application.PluginID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("latest plugin not available"))
		return
	}
	environment.PluginVersion = pluginVersion
	environment.PluginVersionID = uint64(pluginVersion.ID)
	var variablesMap = helper.GetDefaultVariable(pluginVersion)
	var currentVariableMap map[string]interface{}
	err = json.Unmarshal(environment.Variables.RawMessage, &currentVariableMap)
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	helper.ReplaceCustomVariables(currentVariableMap, variablesMap)
	variablesMap["clients_fqdn"] = ""
	jsonString, _ := json.Marshal(variablesMap)
	environment.Variables.RawMessage = jsonString
	if environment.GitBranch != "" {
		repo, err := server.GetRepository(uint(application.OwnerId), application)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environment.GitUrl = repo.CloneURL
		environment.GitRepository = repo
	}
	if !prometheus.ValidateClusterResource(server.DB, environment) {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("CPU, Memory or Storage may be full in this region, please contact support team"))
		return
	}
	var userVariables []map[string]string
	err = json.Unmarshal(environment.UserVariables.RawMessage, &userVariables)
	if err == nil && !models.ValidateUserVariable(userVariables) {
		responses.ERROR(w, http.StatusForbidden, errors.New("you are not allowed to create duplicate variable"))
		return
	}
	environmentCreated, err := environment.Save(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}
	if len(storages) != 0 {
		for _, storage := range storages {
			storage.ID = 0
			storage.EnvironmentID = environmentCreated.ID
			st, err := storageInterface.SaveStorage(server.DB, *storage)
			if err != nil {
				log.Error("Clone stoage create error")
				continue
			}
			environmentCreated.Storage = append(environmentCreated.Storage, st)
		}
	}
	environmentCreated.Action = "Cloning"
	data.EnvId = int64(environmentCreated.ID)
	data.OldEnvId = int64(oldEnvId)
	environmentCreated.CloneEnvironment = &data
	if externalsecret.IsExternalSecretEnable(environmentCreated) {
		externalSecretRequest, err := externalsecret.PrepareExternalSecretRequest(environmentCreated)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		err = externalSecretRequest.ValidateExternalSecretRequest()
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		err = queue.Publish(constants.ExternalSecret, externalSecretRequest)
		if err != nil {
			_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		err = server.publishForCICD(environmentCreated, &models.CICDOptions{})
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
			return
		}
	}
if environmentCreated.ExternalLogging.RawMessage != nil && len(environmentCreated.ExternalLogging.RawMessage) > 0 {
		loggingReq, err := logging.PrepareLoggingRequest(environmentCreated)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		err = queue.Publish(constants.ExternalLogging, loggingReq)
		if err != nil {
			_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}

	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: environment.Application.Name,
		ApplicationID:   int64(environment.Application.ID),
		EnvironmentID:   int64(environment.ID),
		EnvironmentName: environment.Name,
		ProjectName:     environment.Application.Project.Name,
		ProjectID:       int64(environment.Application.Project.ID),
		OrganizationID:  int64(environment.Application.Project.OrganizationId),
	}
	if environment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = environment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "create", fmt.Sprintf("environment %s cloned from %s", environment.Name, oldEnvName), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("create", "environment", user.ID, application.Project, application, environmentCreated, nil, "")
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, environmentCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, environmentCreated)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environmentCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) CloneEnvironmentFunctionSet(cloneData interface{}) error {
	input := models.CloneEnvironment{}
	dat := cloneData.([]byte)
	err := json.Unmarshal(dat, &input)
	if err != nil {
		log.Error(err)
		return err
	}
	log.Debug("Data :: ", input)
	oldEnv, err := (&models.Environment{}).Find(server.DB, uint64(input.OldEnvId))
	if err != nil {
		return err
	}
	newEnv, err := (&models.Environment{}).Find(server.DB, uint64(input.EnvId))
	if err != nil {
		return err
	}
	if input.IsUserPermission {
		err := server.CloneUserPermission(oldEnv, newEnv)
		if err != nil {
			return err
		}
	}
	if input.IsCIConfig {
		err := server.CloneCiConfig(oldEnv, newEnv)
		if err != nil {
			return err
		}
	}
	if input.IsBackupSetting {
		err := server.CloneBackupSettings(oldEnv, newEnv)
		if err != nil {
			return err
		}
	}
	if input.IsCronJob {
		err := server.CloneCronJob(oldEnv, newEnv)
		if err != nil {
			return err
		}
	}
	if input.IsAddon {
		err := server.CloneAddon(oldEnv, newEnv)
		if err != nil {
			return err
		}
	}
	if input.IsScheduler {
		err := server.CloneScheduler(oldEnv, newEnv)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
	This function is responsible for transferring old environment permission to cloned
	environment
*/

func (server *Server) CloneUserPermission(oldEnv, newEnv *models.Environment) error {
	authorizations, err := authInterface.FindAllInEnv(server.DB, uint64(oldEnv.ID))
	if err != nil {
		return err
	}
	log.Debug("Data Auths :: ", authorizations)
	for _, data := range *authorizations {
		data.Prepare()
		data.EnvironmentID = uint64(newEnv.ID)
		_, err := authInterface.Save(server.DB, &data)
		if err != nil {
			continue
		}
	}
	return nil
}

func (server *Server) CloneCronJob(oldEnv, newEnv *models.Environment) error {
	for _, cronJob := range oldEnv.CronJob {
		cronJob.Prepare()
		cronJob.ID = 0
		data, err := cronJobInterface.SaveCronJob(server.DB, cronJob, uint64(newEnv.ID))
		if err != nil {
			continue
		}
		err = queue.Publish(constants.CreateCronJob, &map[string]interface{}{
			"environment": newEnv,
			"cron_job":    data,
		})
		if err != nil {
			continue
		}
	}
	return nil
}

func (server *Server) CloneCiConfig(oldEnv, newEnv *models.Environment) error {
	CiConfig, err := ciConfigRepo.FindByEnvironment(server.DB, oldEnv.ID)
	if err == nil {
		CiConfig.Prepare()
		CiConfig.EnvironmentID = newEnv.ID
		CiConfig.Environment = nil
		_, err = ciConfigRepo.Save(server.DB, CiConfig)
		if err != nil {
			return err
		}
		return nil
	}
	return err
}

func (server *Server) CloneBackupSettings(oldEnv, newEnv *models.Environment) error {
	namespace := helper.GetNamespace(oldEnv)
	c := pb.NewBackupClient(server.BackupClient)
	gr, err := c.GetSettingBackup(context.Background(), &pb.GetSettingBackupRequest{
		Namespace: namespace,
		Path:      oldEnv.Application.Cluster.ConfigPath,
	})
	if err != nil {
		return err
	}
	settings := gr.JsonSetting
	cloneNamespace := helper.GetNamespace(newEnv)
	_, err = c.SaveSettingBackup(context.Background(), &pb.SaveSettingBackupRequest{
		Namespace:   cloneNamespace,
		JsonSetting: settings,
		Path:        newEnv.Application.Cluster.ConfigPath,
	})
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) CloneAddon(oldEnv, newEnv *models.Environment) error {
	addons, err := newEnv.FindAllAddons(server.DB, uint64(oldEnv.ID))
	if err != nil {
		return err
	}
	for _, data := range addons {
		data.ID = 0
		data.ParentID = uint64(newEnv.ID)
		envAddon, err := data.Save(server.DB)
		if err != nil {
			return err
		}
		envJ, _ := newEnv.ToJson()
		envAddonJ, _ := envAddon.ToJson()
		err = queue.Publish(constants.InstallAddOn, &map[string]interface{}{
			"environment": envJ,
			"add_on":      envAddonJ,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (server *Server) CloneScheduler(oldEnv, newEnv *models.Environment) error {
	newEnv.Schedules = oldEnv.Schedules
	_, err := newEnv.UpdateScheudle(server.DB, newEnv.ID)
	if err != nil {
		return err
	}
	err = queue.Publish(constants.ScheduleEnvironment, newEnv)
	if err != nil {
		return err
	}
	return nil
}
