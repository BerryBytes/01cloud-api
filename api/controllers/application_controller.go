package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"01cloud-api/api/utils/helper"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var applicationInterface = models.NewApplication()

// CreateApplication godoc
// @Summary Create a new Application
// @Description Create a new Application with the input payload
// @Tags Application
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Application true "Create Application"
// @Success 201 {object} doc.Application
// @Router /application [post]
func (server *Server) CreateApplication(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	application := models.Application{}
	err = json.Unmarshal(body, &application)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if application.ClusterID <= 1 {
		region := &struct {
			Region string `json:"region"`
		}{}
		_ = json.Unmarshal(body, &region)
		regionId := region.Region
		cluster, err := clusterInterface.FindByRegion(server.DB, regionId, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("region not found"))
			return
		}
		application.ClusterID = uint64(cluster.ID)
		application.Cluster = cluster
	} else {
		cluster, err := clusterInterface.Find(server.DB, application.ClusterID)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("cluster not found"))
			return
		}
		application.Cluster = cluster
	}
	application.Prepare()
	err = application.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	application.Project.OrganizationId = uint64(oid)
	project, err := iproject.Find(server.DB, application.ProjectID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid project"))
		return
	}
	if oid == 0 {
		if !paymentInterface.HasUserBalance(server.DB, uint(project.UserID)) {
			responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to create environment due to remaining balance"))
			return
		}
	} else {
		org, err := orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if !paymentInterface.HasUserBalance(server.DB, uint(org.UserID)) {
			responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to create environment due to remaining balance"))
			return
		}
	}

	if application.GitUrl != "" {
		repo, err := server.GetRepository(uid, &application)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		application.GitRepository = repo
		application.GitRepoUrl = &repo.GitURL
	}

	if project.UserID != uint64(uid) &&
		!authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
		!authInterface.IsAuthorizedProject(server.DB, uint64(uid), application.ProjectID) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("permission not granted"))
		return
	}

	count := applicationInterface.FindAllApplicationCountByProject(server.DB, application.ProjectID, uint64(uid))
	if count >= int(project.Subscription.Apps) {
		responses.ERROR(w, http.StatusServiceUnavailable, errors.New("sorry you cannot create new app in this project, Quota limit exceeded"))
		return
	}
	application.Project = project
	if applicationInterface.IsNameExists(server.DB, application.ProjectID, application.Name) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("name already exists"))
		return
	}
	if application.ServiceType == 0 || application.ServiceType == 1 || application.ServiceType == 5 {
		plugin, err := pluginInterface.Find(server.DB, application.PluginID)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid plugin"))
			return
		}
		application.Plugin = plugin
	} else if application.ServiceType == 2 {
		pluginId := os.Getenv("DOCKER_PLUGIN_ID")
		if pluginId == "" {
			pluginId = "92"
		}
		number, err := strconv.ParseUint(pluginId, 10, 64)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid plugin id in config"))
			return
		}
		plugin, err := pluginInterface.Find(server.DB, number)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid plugin"))
			return
		}
		application.Plugin = plugin
		application.PluginID = number
		if application.ImageService == "dockerhub" {
			// check for person repo
			if len(application.ImageNamespace) == 0 || application.ImageNamespace == "personal" {
				_, username, err := server.GetRegistryUserFromUidName(uint64(uid), application.ImageService)
				if err != nil {
					responses.ERROR(w, http.StatusUnprocessableEntity, err)
					return
				}
				application.ImageNamespace = *username
			}
			application.ImageUrl = fmt.Sprintf("docker.io/%s/%s", application.ImageNamespace, application.ImageRepo)
		} else if application.ImageService == "aws_ecr" {
			repo, err := server.GetRegistryFromUidName(uint64(uid), application.ImageService, application.ImageNamespace, application.ImageRepo)
			if err != nil {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			application.ImageUrl = repo.Uri
		}
	} else if application.ServiceType == 3 {
		pluginId := os.Getenv("HELM_PLUGIN_ID")
		if pluginId == "" {
			pluginId = "95"
		}
		number, err := strconv.ParseUint(pluginId, 10, 64)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid plugin id in config"))
			return
		}
		plugin, err := pluginInterface.Find(server.DB, number)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid plugin"))
			return
		}
		application.Plugin = plugin
		application.PluginID = number
		if application.Cluster.OrganizationID == 0 {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("you cannot install helm chart on shared cluster"))
			return
		}
	} else {
		pluginId := os.Getenv("OPERATOR_PLUGIN_ID")
		if pluginId == "" {
			pluginId = "95"
		}
		number, err := strconv.ParseUint(pluginId, 10, 64)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid plugin id in config"))
			return
		}
		plugin, err := pluginInterface.Find(server.DB, number)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid plugin"))
			return
		}
		application.Plugin = plugin
		application.PluginID = number
		if application.Cluster.OrganizationID == 0 {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("you cannot install operator on shared cluster"))
			return
		}
	}
	application.OwnerId = uint64(uid)
	applicationCreated, err := applicationInterface.Save(server.DB, &application)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	_, _ = server.SaveActivityWithJson("create", "application", uid, project, &application, nil, nil, "")
	notificationInfo := notifications.BasicInfoNotification{
		OrganizationID:  int64(application.Project.OrganizationId),
		ProjectID:       int64(application.ProjectID),
		ProjectName:     application.Project.Name,
		ApplicationID:   int64(application.ID),
		ApplicationName: application.Name,
	}
	notifyApplication(server.NotifyClient, notificationInfo, "create", fmt.Sprintf("application %s created", application.Name), "info", int64(uid))
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, applicationCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, applicationCreated)
		return
	}
	applicationResponse, err := CreateApplicationResponse(applicationCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    applicationResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetApplicationsByProject godoc
// @Summary Get Application by project
// @Description Get list applications from by project id.
// @Tags Application
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Param region query string true "Region"
// @Param search query string true "Search"
// @Param plugin_id query int true "Plugin Id"
// @Success 200 {array} map[string]interface{}
// @Router /project/{id}/applications [get]
func (server *Server) GetApplicationsByProject(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// key := fmt.Sprintf("project-applications-%d", pid)
	// var value interface{}
	// if ok := server.Cache.Get(key, &value); ok {
	// 	responses.JSON(w, http.StatusOK, value)
	// 	return
	// }
	users, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	projectReceived, err := iproject.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if projectReceived.UserID != uint64(uid) && !users.IsAdmin &&
		!authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
		!authInterface.IsAuthorizedProject(server.DB, uint64(uid), pid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("permission not granted"))
		return
	}
	applications, err := applicationInterface.FindAllApplicationByProject(server.DB, pid, uint64(uid), true)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	filterSearch := r.FormValue("search")
	filterRegion := r.FormValue("region")
	filterRegionID := uint64(0)
	filterPluginId, _ := strconv.ParseUint(r.FormValue("plugin_id"), 10, 64)
	if len(filterRegion) > 0 {
		region, err := clusterInterface.FindByRegion(server.DB, filterRegion, oid)
		if err == nil {
			filterRegionID = uint64(region.ID)
		}
	}
	jsonData := []map[string]interface{}{}
	for _, app := range applications {
		jsonApp := map[string]interface{}{}
		if !checkAppFilter(app, filterSearch, filterPluginId, filterRegionID) {
			continue
		}
		if helper.IsV2(r) {
			appResponse, err := CreateApplicationResponse(app)
			if err != nil {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			appByte, _ := json.Marshal(appResponse)
			err = json.Unmarshal(appByte, &jsonApp)
			if err != nil {
				log.Error(err)
			}
		} else {
			jsonApp, _ = app.ToJson()
		}

		env := models.Environment{}
		helmenv := models.HelmEnvironment{}
		envList, _ := env.FindAllByApplication(server.DB, uint64(app.ID), true, uint64(uid))
		helmEnvList, _ := helmenv.FindAllByApplication(server.DB, uint64(app.ID), false, uint64(uid))
		jList := []map[string]interface{}{}
		for _, e := range envList {
			envResponse, _ := CreateEnvironmentResponse(e)
			je := map[string]interface{}{}
			envByte, _ := json.Marshal(envResponse)
			err = json.Unmarshal(envByte, &je)
			if err != nil {
				log.Error(err)
			}
			state := "Stopped"
			namespace := helper.GetNamespace(e)
			if e.ServiceType == 4 {
				packageName := e.Application.OperatorPackageName
				if e.Application.Cluster != nil {
					opReq, err := operatorRequestRepo.FindByNameAndClusterID(server.DB, e.Application.Cluster.ClusterRequestID, packageName)
					if err != nil {
						continue
					}
					if !opReq.GlobalOperator {
						namespace = fmt.Sprintf("%s-%d", opReq.PackageName, opReq.ID)
					}
				}
			}
			if e.Active {
				_, state, _, err = server.StoreClient.PodState().GetEnvironmentState(namespace)
				if err != nil {
					responses.JSON(w, http.StatusInternalServerError, err)
					return
				}
			}
			je["status"] = state
			jList = append(jList, je)
		}
		for _, e := range helmEnvList {
			envResponse, _ := CreateHelmEnvironmentResponse(e)
			je := map[string]interface{}{}
			envByte, _ := json.Marshal(envResponse)
			err = json.Unmarshal(envByte, &je)
			if err != nil {
				log.Error(err)
			}
			state := "Stopped"
			if e.Active {
				_, state, _, err = server.StoreClient.PodState().GetEnvironmentState(helper.GetHelmNamespace(e))
				if err != nil {
					responses.JSON(w, http.StatusInternalServerError, err)
					return
				}
			}
			je["status"] = state
			jList = append(jList, je)
		}
		jsonApp["env"] = jList
		jsonData = append(jsonData, jsonApp)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, jsonData)
		return
	}
	response := responses.Response{
		Data:    jsonData,
		Success: 1,
		Message: "Success",
	}
	// server.Cache.Set(key, response)
	responses.JSON(w, http.StatusOK, response)
}

func checkAppFilter(app *models.Application, filterSearch string, filterPluginId, filterRegionID uint64) bool {
	if !strings.Contains(app.Name, filterSearch) {
		return false
	}
	if filterPluginId > 0 && app.PluginID != filterPluginId {
		return false
	}
	if filterRegionID > 0 && app.ClusterID != filterRegionID {
		return false
	}
	return true
}

// GetApplicationsByProjectForAdmin godoc
// @Summary Get Application by project
// @Description Get list applications from by project id. Only available for admin.
// @Tags Application
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Success 200 {array} doc.Application
// @Router /project/{id}/admin-app [get]
func (server *Server) GetApplicationsByProjectForAdmin(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	applications, err := applicationInterface.FindAllApplicationByProjectForAdmin(server.DB, pid, uint64(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, applications)
		return
	}
	applicationResponseList := []doc.Application{}
	for _, application := range *applications {
		applicationResponse, err := CreateApplicationResponse(&application)
		if err != nil {
			log.Error(err)
		}
		applicationResponseList = append(applicationResponseList, *applicationResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    applicationResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetApplication godoc
// @Summary Get Application by id
// @Description Get Application by id from token
// @Tags Application
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Application id"
// @Success 200 {object} doc.Application
// @Router /application/{id} [get]
func (server *Server) GetApplication(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	applicationReceived, err := applicationInterface.Find(server.DB, aid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !applicationReceived.Project.Active {
		responses.ERROR(w, http.StatusNotFound, errors.New("project is deactivated"))
		return
	}
	if applicationReceived.GitUrl != "" {
		repo, err := server.GetRepository(uint(applicationReceived.OwnerId), applicationReceived)
		if err == nil {
			applicationReceived.GitRepository = repo
		}
	}
	if uid == uint(applicationReceived.Project.UserID) || user.IsAdmin ||
		authInterface.IsAuthorizedOrganization(server.DB, uid, oid) ||
		authInterface.IsAuthorizedApplication(server.DB, uint64(uid), aid, applicationReceived.ProjectID) {
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, applicationReceived)
			return
		}
		applicationResponse, err := CreateApplicationResponse(applicationReceived)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return

		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    applicationResponse,
			Success: 1,
			Message: "Success",
		})
	} else {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view this application"))
	}
}

// UpdateApplication godoc
// @Summary Update a Application
// @Description Update a Application with the input payload
// @Tags Application
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Application id"
// @Param body body doc.Application true "Update Application"
// @Success 200 {object} doc.Application
// @Router /application/{id} [put]
func (server *Server) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not founnd"))
		return
	}
	application, err := applicationInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("application not found"))
		return
	}
	if uid != uint(application.Project.UserID) && !user.IsAdmin {
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
			!authInterface.IsWriteAuthorizedApplication(server.DB, uint64(uid), pid, application.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	applicationUpdate := models.Application{}
	err = json.Unmarshal(body, &applicationUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	applicationUpdate.Prepare()
	//err = applicationUpdate.Validate()
	//if err != nil {
	//	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	//	return
	//}
	if applicationUpdate.Name != "" {
		if applicationUpdate.Name != application.Name && applicationInterface.IsNameExists(server.DB, application.ProjectID, applicationUpdate.Name) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("name already exists"))
			return
		}
	}
	applicationUpdate.ID = application.ID
	applicationUpdated, err := applicationInterface.Update(server.DB, &applicationUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	_, _ = server.SaveActivityWithJson("update", "application", uid, application.Project, &applicationUpdate, nil, nil, "")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, applicationUpdated)
		return
	}
	applicationResponse, err := CreateApplicationResponse(applicationUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    applicationResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetAvailableResource godoc
// @Summary Get available resource in application
// @Description Get available resource in application
// @Tags Application
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Application id"
// @Success 200 {object} doc.AvailableResourceResponse
// @Router /application/{id}/available-resource [get]
func (server *Server) GetAvailableResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	app, err := applicationInterface.Find(server.DB, aid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("application not found"))
		return
	}
	if uid != uint(app.Project.UserID) && !user.IsAdmin {
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
			!authInterface.IsAuthorizedApplication(server.DB, uint64(uid), aid, app.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized, You are not authorized to use this application"))
			return
		}
	}
	res := app.GetAvailableResource(server.DB, app.ProjectID)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, res)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    res,
		Success: 1,
		Message: "Success",
	})

}

// DeleteApplication godoc
// @Summary Delete a Application
// @Description Delete a Application with the input payload
// @Tags Application
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Application id"
// @Success 204 {object} doc.Application
// @Router /application/{id} [delete]
func (server *Server) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	app, err := applicationInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("application not found"))
		return
	}
	if uid != uint(app.Project.UserID) && !user.IsAdmin {
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
			!authInterface.IsAdminOfApplication(server.DB, uint64(uid), pid, app.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized, You are not authorized to delete this application"))
			return
		}
	}

	environment := models.Environment{}
	count := environment.FindAllEnvironmentCountByApplication(server.DB, pid)
	if count > 0 {
		responses.ERROR(w, http.StatusPreconditionFailed, errors.New("cannot delete non-empty application"))
		return
	}
	notificationInfo := notifications.BasicInfoNotification{
		OrganizationID:  int64(app.Project.OrganizationId),
		ProjectID:       int64(app.ProjectID),
		ProjectName:     app.Project.Name,
		ApplicationID:   int64(app.ID),
		ApplicationName: app.Name,
	}
	notifyApplication(server.NotifyClient, notificationInfo, "deleted", fmt.Sprintf("application %s deleted", app.Name), "info", int64(uid))
	_, _ = server.SaveActivityWithJson("delete", "application", uid, app.Project, app, nil, nil, "")
	time.Sleep(time.Duration(time.Second))
	_, err = applicationInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", pid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func notifyApplication(conn *grpc.ClientConn, info notifications.BasicInfoNotification, action, body, types string, triggeredBy int64) {
	_, err := PublishNotificationBase(conn, &notifications.Notification{
		ApplicationName:  info.ApplicationName,
		ApplicationID:    info.ApplicationID,
		ProjectName:      info.ProjectName,
		ProjectID:        info.ProjectID,
		OrganizationName: info.OrganizationName,
		OrganizationID:   info.OrganizationID,
		TriggeredBy:      triggeredBy,
		Type:             types,
		Body:             body,
		Scope:            "application",
		SendBy:           "api",
	})
	if err != nil {
		log.Error(err)
	}
}

func CreateApplicationResponse(application *models.Application) (*doc.Application, error) {
	applicationResponse := doc.Application{}
	applicationBytes, _ := json.Marshal(application)
	err := json.Unmarshal(applicationBytes, &applicationResponse)
	if err != nil {
		return nil, err
	}
	return &applicationResponse, nil

}
