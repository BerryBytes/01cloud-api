package controllers

import (
	"01cloud-api/api/models/doc"
	"01cloud-api/api/utils/prometheus"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"01cloud-api/api/notifications"

	smodel "github.com/berrybytes/01cloud-store/model"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/jinzhu/gorm"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/tidwall/gjson"
)

// CreateEnvironment godoc
// @Summary Create a new HelmEnvironment
// @Description Create a new HelmEnvironment with the input payload
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.HelmEnvironment true "Create HelmEnvironment"
// @Success 201 {object} doc.HelmEnvironment
// @Router /helm-environment [post]

var chartVersionInterface = models.NewChartVersion()

func (server *Server) CreateHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	helmEnvironment := models.HelmEnvironment{}
	err = json.Unmarshal(body, &helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	helmEnvironment.Prepare()
	err = helmEnvironment.Validate(true)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	application, err := applicationInterface.Find(server.DB, helmEnvironment.ApplicationID)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("invalid application"))
		return
	}

	if !application.Cluster.Active {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("cluster is disabled, please contact support team"))
		return
	}
	if helmEnvironment.IsNameExists(server.DB, uint(helmEnvironment.ApplicationID), helmEnvironment.Name) {
		responses.ERROR(w, http.StatusForbidden, errors.New("name already exists"))
		return
	}

	if uid != uint(application.Project.UserID) && !application.Project.User.IsAdmin {
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
			!authInterface.IsWriteAuthorizedProject(server.DB, uint64(uid), application.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to add helmEnvironment"))
			return
		}
	}
	helmEnvironment.Application = application
	chartVersion, err := chartVersionInterface.Find(server.DB, helmEnvironment.ChartVersionID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("latest plugin not available"))
		return
	}
	helmEnvironment.ChartVersion = chartVersion

	helmEnvironmentCreated, err := helmEnvironment.Save(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}
	//preventBuildStruct := struct {
	//	PreventDefaultBuild bool `json:"prevent_default_build"`
	//}{}
	//_ = json.Unmarshal(body, &preventBuildStruct)
	err = queue.Publish(constants.CreateHelmRelease, helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironment.Application.Name,
		ApplicationID:   int64(helmEnvironment.Application.ID),
		EnvironmentID:   int64(helmEnvironment.ID),
		EnvironmentName: helmEnvironment.Name,
		ProjectName:     helmEnvironment.Application.Project.Name,
		ProjectID:       int64(helmEnvironment.Application.Project.ID),
		OrganizationID:  int64(helmEnvironment.Application.Project.OrganizationId),
	}
	if helmEnvironment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = helmEnvironment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "created", fmt.Sprintf("helmEnvironment %s created", helmEnvironment.Name), "info", int64(uid))

	_, _ = server.SaveActivityWithJson("create", "environment", uid, application.Project, application, nil, helmEnvironmentCreated, "")
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, helmEnvironmentCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, helmEnvironmentCreated)
		return
	}
	helmEnvResponse, err := CreateHelmEnvironmentResponse(helmEnvironmentCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    helmEnvResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetEnvironmentsByEnvironment godoc
// @Summary Get HelmEnvironment by application
// @Description Get list helmEnvironments by application id.
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Application id"
// @Success 200 {array} doc.HelmEnvironment
// @Router /application/{id}/helmEnvironments [get]
func (server *Server) GetHelmEnvironmentsByApplication(w http.ResponseWriter, r *http.Request) {
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
	application, err := applicationInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if uid != uint(application.Project.UserID) && !application.Project.User.IsAdmin {
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
			!authInterface.IsWriteAuthorizedProject(server.DB, uint64(uid), application.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to get helmEnvironment"))
			return
		}
	}

	if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
		!authInterface.IsAuthorizedApplication(server.DB, uint64(uid), pid, application.ProjectID) &&
		application.Project.UserID != uint64(uid) && !user.IsAdmin {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("permission not granted"))
		return
	}
	helmEnvironment := models.HelmEnvironment{}
	helmEnvironments, err := helmEnvironment.FindAllByApplication(server.DB, pid, true, uint64(uid))
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	jsonData := []map[string]interface{}{}
	for _, env := range helmEnvironments {
		envResponse, _ := CreateHelmEnvironmentResponse(env)
		jsonEnv := map[string]interface{}{}
		envByte, _ := json.Marshal(envResponse)
		err = json.Unmarshal(envByte, &jsonEnv)
		if err != nil {
			log.Error(err)
		}
		state := "Stopped"
		if env.Active {
			_, state, _, err = server.StoreClient.PodState().GetEnvironmentState(helper.GetHelmNamespace(env))
			if err != nil {
				responses.JSON(w, http.StatusInternalServerError, err)
				return
			}
		}
		jsonEnv["status"] = state
		jsonData = append(jsonData, jsonEnv)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, jsonData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    jsonData,
		Success: 1,
		Message: "Success",
	})
}

// GetEnvironmentsByApplicationForAdmin godoc
// @Summary Get HelmEnvironment by application
// @Description Get list helmEnvironments from by application id. Only available for admin.
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Application id"
// @Success 200 {array} doc.HelmEnvironment
// @Router /application/{id}/admin-env [get]
func (server *Server) GetHelmEnvironmentsByApplicationAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	helmEnvironment := models.HelmEnvironment{}
	helmEnvironments, err := helmEnvironment.FindAllByApplicationAdmin(server.DB, aid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, helmEnvironments)
		return
	}
	helmResponseList := []doc.HelmEnvironment{}
	for _, helm := range *helmEnvironments {
		helmResponse, err := CreateHelmEnvironmentResponse(&helm)
		if err != nil {
			log.Error(err)
		}
		helmResponseList = append(helmResponseList, *helmResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    helmResponseList,
		Success: 1,
		Message: "Success",
	})

}

// GetEnvironmentInsights godoc
// @Summary Get HelmEnvironment insight by id
// @Description Get HelmEnvironment insight by id
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} doc.InsightResponse
// @Router /helm-environment/{id}/insights [post]
func (server *Server) GetHelmEnvironmentInsights(w http.ResponseWriter, r *http.Request) {

	helmEnvironment, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	request := models.Insight{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	namespace := helper.GetHelmNamespace(helmEnvironment)
	request.Namespace = namespace
	res, err := prometheus.GetHelmInsight(request, helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
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

// GetEnvironmentOverview godoc
// @Summary Get HelmEnvironment overview by id
// @Description Get HelmEnvironment overview by id
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} doc.InsightResponse
// @Router /helmEnvironment/{id}/overview [post]
func (server *Server) GetHelmEnvironmentOverview(w http.ResponseWriter, r *http.Request) {

	env, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	request := models.Insight{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	namespace := helper.GetHelmNamespace(env)
	request.Namespace = namespace
	res, err := prometheus.GetHelmInsightOverview(request, env)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
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

// ReLaunchEnvironment godoc
// @Summary Rerun HelmEnvironment by id
// @Description Rerun HelmEnvironment by id
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} doc.SuccessResponse
// @Router /helmEnvironment/{id}/launch [get]
func (server *Server) ReLaunchHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	helmEnvironment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	helmEnvironment.Action = "Relaunching"
	err = queue.Publish(constants.UpgradeRelease, helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironment.Application.Name,
		ApplicationID:   int64(helmEnvironment.Application.ID),
		EnvironmentID:   int64(helmEnvironment.ID),
		EnvironmentName: helmEnvironment.Name,
		ProjectName:     helmEnvironment.Application.Project.Name,
		ProjectID:       int64(helmEnvironment.Application.Project.ID),
		OrganizationID:  int64(helmEnvironment.Application.Project.OrganizationId),
	}
	if helmEnvironment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = helmEnvironment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "relaunched", fmt.Sprintf("helmEnvironment %s relaunched", helmEnvironment.Name), "info", int64(user.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "HelmEnvironment Re-deploy triggered",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func CheckHelmEnvironmentPermission(db *gorm.DB, r *http.Request, role string, env *models.HelmEnvironment) (*models.User, error) {
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(db, uid)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if env.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if authInterface.IsAuthorizedOrganization(db, uid, oid) {
			return user, nil
		}
		if role == WRITE {
			if !authInterface.IsWriteAuthorizedHelmEnvironment(db, uint64(uid), env) {
				return nil, errors.New("you are not authorized to write this helmEnvironment")
			}
		} else if role == ADMIN {
			if !authInterface.IsAdminOfHelmEnvironment(db, uint64(uid), env) {
				return nil, errors.New("you are not authorized to update this helmEnvironment")
			}
		} else if role == READ {
			if !authInterface.IsAuthorizedHelmEnvironment(db, uint64(uid), env) {
				return nil, errors.New("you are not authorized to access this helmEnvironment")
			}
		}
	}
	return user, nil
}
func GetHelmEnvironmentFromRequest(db *gorm.DB, r *http.Request) (*models.HelmEnvironment, error) {
	vars := mux.Vars(r)
	environment := &models.HelmEnvironment{}
	eid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return environment, err
	}
	environmentReceived, err := environment.Find(db, int64(eid))
	if err != nil {
		return environment, err
	}
	return environmentReceived, nil
}

func (server *Server) GetHelmEnvironmentWithUserPermission(r *http.Request, mode string) (*models.HelmEnvironment, *models.User, int, error) {
	helmEnvironment, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		return nil, nil, http.StatusUnprocessableEntity, err
	}
	if !helmEnvironment.Active {
		return nil, nil, http.StatusForbidden, errors.New("HelmEnvironment is stopped ")
	}
	user, err := CheckHelmEnvironmentPermission(server.DB, r, mode, helmEnvironment)
	if err != nil {
		return nil, nil, http.StatusUnauthorized, err
	}

	return helmEnvironment, user, 0, nil
}

// StopEnvironment godoc
// @Summary Stop HelmEnvironment by id
// @Description Stop HelmEnvironment by id
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} doc.SuccessResponse
// @Router /helm-environment/{id}/stop [post]
func (server *Server) StopHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	helmEnvironmentReceived, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !helmEnvironmentReceived.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("helmEnvironment already stopped"))
		return
	}
	user, err := CheckHelmEnvironmentPermission(server.DB, r, WRITE, helmEnvironmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	helmEnvironmentReceived.Active = false
	envs := &models.HelmEnvironment{}
	_, err = envs.UpdateStatus(server.DB, uint64(helmEnvironmentReceived.ID), helmEnvironmentReceived.Active)
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	helmEnvironmentReceived.Action = "Stopping"
	err = queue.Publish(constants.ActiveDeactiveHelmEnvironment, helmEnvironmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
	}

	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironmentReceived.Application.Name,
		ApplicationID:   int64(helmEnvironmentReceived.Application.ID),
		EnvironmentID:   int64(helmEnvironmentReceived.ID),
		EnvironmentName: helmEnvironmentReceived.Name,
		ProjectName:     helmEnvironmentReceived.Application.Project.Name,
		ProjectID:       int64(helmEnvironmentReceived.Application.Project.ID),
		OrganizationID:  int64(helmEnvironmentReceived.Application.Project.OrganizationId),
	}

	if helmEnvironmentReceived.Application.Project.OrganizationId != 0 {
		notifyInfo.OrganizationName = helmEnvironmentReceived.Application.Project.Organization.Name
	}

	notify(server.NotifyClient, notifyInfo, "stopped", fmt.Sprintf("helmEnvironment %s stopped", helmEnvironmentReceived.Name), "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("stop", "helm-environment", user.ID, helmEnvironmentReceived.Application.Project, helmEnvironmentReceived.Application, nil, helmEnvironmentReceived, "")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "Sent command for stopping helmEnvironment",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// StartEnvironment godoc
// @Summary Start HelmEnvironment by id
// @Description Start HelmEnvironment by id
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} doc.SuccessResponse
// @Router /helm-environment/{id}/start [post]
func (server *Server) StartHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	helmEnvironmentReceived, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if helmEnvironmentReceived.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("helmEnvironment already started"))
		return
	}
	user, err := CheckHelmEnvironmentPermission(server.DB, r, WRITE, helmEnvironmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	helmEnvironmentReceived.Active = true
	envs := &models.HelmEnvironment{}
	_, err = envs.UpdateStatus(server.DB, uint64(helmEnvironmentReceived.ID), helmEnvironmentReceived.Active)
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	helmEnvironmentReceived.Action = "Starting"
	err = queue.Publish(constants.ActiveDeactiveHelmEnvironment, helmEnvironmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironmentReceived.Application.Name,
		ApplicationID:   int64(helmEnvironmentReceived.Application.ID),
		EnvironmentID:   int64(helmEnvironmentReceived.ID),
		EnvironmentName: helmEnvironmentReceived.Name,
		ProjectName:     helmEnvironmentReceived.Application.Project.Name,
		ProjectID:       int64(helmEnvironmentReceived.Application.Project.ID),
		OrganizationID:  int64(helmEnvironmentReceived.Application.Project.OrganizationId),
	}
	if helmEnvironmentReceived.Application.Project.OrganizationId != 0 {
		notifyInfo.OrganizationName = helmEnvironmentReceived.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "started", fmt.Sprintf("helmEnvironment %s started", helmEnvironmentReceived.Name), "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("start", "helm-environment", user.ID, helmEnvironmentReceived.Application.Project, helmEnvironmentReceived.Application, nil, helmEnvironmentReceived, "")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "Sent command for starting helmEnvironment",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// ScheduleStartStopHelmEnvironment godoc
// @Summary Schedule Start Stop Helm Environment
// @Description Schedule Start Stop Helm Environment
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Param type query string true "Type"
// @Success 200
// @Router /helm-environment/{id}/start-stop [post]
func (server *Server) ScheduleStartStopHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	scheduleType := r.URL.Query().Get("type")
	if scheduleType == "" {
		responses.ERROR(w, http.StatusNotFound, errors.New("required schedule type"))
		return
	}
	helmEnvironmentReceived, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if scheduleType == "stop" {
		if !helmEnvironmentReceived.Active {
			responses.ERROR(w, http.StatusForbidden, errors.New("helmEenvironment already stopped"))
			return
		}
		helmEnvironmentReceived.Active = false
		helmEnvironmentReceived.Action = "Stopping"
	} else if scheduleType == "start" {
		if !paymentInterface.HasUserBalance(server.DB, uint(helmEnvironmentReceived.Application.Project.UserID)) {
			responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to start the environment due to remaining balance"))
			return
		}
		if helmEnvironmentReceived.Active {
			responses.ERROR(w, http.StatusForbidden, errors.New("environment already started"))
			return
		}
		helmEnvironmentReceived.Active = true
		helmEnvironmentReceived.Action = "Starting"
	} else {
		responses.ERROR(w, http.StatusForbidden, errors.New("invalid schedule type"))
		return
	}
	envs := &models.HelmEnvironment{}
	_, err = envs.UpdateStatus(server.DB, uint64(helmEnvironmentReceived.ID), helmEnvironmentReceived.Active)
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	err = queue.Publish(constants.ActiveDeactiveHelmEnvironment, helmEnvironmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
	}

	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironmentReceived.Application.Name,
		ApplicationID:   int64(helmEnvironmentReceived.Application.ID),
		EnvironmentID:   int64(helmEnvironmentReceived.ID),
		EnvironmentName: helmEnvironmentReceived.Name,
		ProjectName:     helmEnvironmentReceived.Application.Project.Name,
		ProjectID:       int64(helmEnvironmentReceived.Application.Project.ID),
		OrganizationID:  int64(helmEnvironmentReceived.Application.Project.OrganizationId),
	}
	if helmEnvironmentReceived.Application.Project.OrganizationId != 0 {
		notifyInfo.OrganizationName = helmEnvironmentReceived.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "started", fmt.Sprintf("helmEnvironment %s started", helmEnvironmentReceived.Name), "info", 0)
	_, _ = server.SaveActivityWithJson("start", "helm-environment", 0, helmEnvironmentReceived.Application.Project, helmEnvironmentReceived.Application, nil, helmEnvironmentReceived, "")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "Sent command for starting helmEnvironment",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetEnvironment godoc
// @Summary Get HelmEnvironment by id
// @Description Get HelmEnvironment by id from token
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} doc.HelmEnvironment
// @Router /helm-environment/{id} [get]
func (server *Server) GetHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	helmEnvironmentReceived, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	_, err = CheckHelmEnvironmentPermission(server.DB, r, READ, helmEnvironmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	metadata, err := server.StoreClient.ReleaseInfo().GetData(fmt.Sprintf("%s-helm-metadata-%d", os.Getenv("GCLOUD_NAMESPACE"), helmEnvironmentReceived.ID))
	if err != nil {
		metadata = &smodel.K8sRelease{}
	}

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
				log.Error("invalid manifest", err)
			}
		}
		js := models.ToJson(manifestRespones)
		secrets := gjson.Get(js, "#(kind==Secret)#.data")
		envs := gjson.Get(js, "#(kind==Deployment).spec.template.spec.containers")
		envs_map := new([]map[string]interface{})
		secret_map := new([]map[string]interface{})
		_ = json.Unmarshal([]byte(envs.Raw), envs_map)
		_ = json.Unmarshal([]byte(secrets.Raw), secret_map)
		serviceDetail, err := server.StoreClient.Storage().GetStorageState(helper.GetHelmNamespace(helmEnvironmentReceived))
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		metadataJson = &map[string]interface{}{
			"name":      metadata.Name,
			"namespace": metadata.Namespace,
			"version":   metadata.Version,
			"info":      metadata.Info,
			"releaseId": metadata.ReleaseId,
			"service":   serviceDetail,
			"secret":    secret_map,
			"envs":      envs_map,
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"environment": helmEnvironmentReceived,
			"metadata":    metadataJson,
		})
		return
	}
	helmEnvResponse, err := CreateHelmEnvironmentResponse(helmEnvironmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"environment": helmEnvResponse,
			"metadata":    metadataJson,
		},
		Success: 1,
		Message: "Success",
	})
}

// UpdateEnvironment godoc
// @Summary Update a HelmEnvironment
// @Description Update a HelmEnvironment with the input payload
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Param body body doc.HelmEnvironment true "Update HelmEnvironment"
// @Success 200 {object} doc.HelmEnvironment
// @Router /helm-environment/{id} [put]
func (server *Server) UpdateHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	helmEnvironment, user, code, err := server.GetHelmEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	extras := ""
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	helmEnvironmentUpdate := models.HelmEnvironment{}
	err = json.Unmarshal(body, &helmEnvironmentUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	helmEnvironmentUpdate.Prepare()
	helmEnvironmentUpdate.Application = helmEnvironment.Application

	helmEnvironmentUpdate.ID = helmEnvironment.ID

	if helmEnvironmentUpdate.Name != "" {
		if helmEnvironmentUpdate.Name != helmEnvironment.Name && helmEnvironment.IsNameExists(server.DB, uint(helmEnvironment.ApplicationID), helmEnvironmentUpdate.Name) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("name already exists"))
			return
		}
		extras += fmt.Sprintf("name from '%s' to '%s' ", helmEnvironment.Name, helmEnvironmentUpdate.Name)
	}

	helmEnvironmentUpdate.ID = helmEnvironment.ID
	helmEnvironmentUpdated, err := helmEnvironmentUpdate.Update(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	env := &models.HelmEnvironment{}
	env, _ = env.Find(server.DB, int64(helmEnvironment.ID))

	env.Action = "Updating"
	err = queue.Publish(constants.UpgradeRelease, env)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}

	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironment.Application.Name,
		ApplicationID:   int64(helmEnvironment.Application.ID),
		EnvironmentID:   int64(helmEnvironment.ID),
		EnvironmentName: helmEnvironment.Name,
		ProjectName:     helmEnvironment.Application.Project.Name,
		ProjectID:       int64(helmEnvironment.Application.Project.ID),
		OrganizationID:  int64(helmEnvironment.Application.Project.OrganizationId),
	}
	if helmEnvironment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = helmEnvironment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "updated", fmt.Sprintf("helmEnvironment %s updated with %s", helmEnvironment.Name, extras), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, env.Application.Project, env.Application, nil, env, extras)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, helmEnvironmentUpdated)
		return
	}
	helmEnvResponse, err := CreateHelmEnvironmentResponse(helmEnvironmentUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    helmEnvResponse,
		Success: 1,
		Message: "Success",
	})
}

// RedeployEnvironment godoc
// @Summary Redeploy HelmEnvironment by id
// @Description Redeploy HelmEnvironment by id
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Param body body doc.RedeployRequest true "Redeploy helmEnvironment"
// @Success 200 {object} doc.SuccessResponse
// @Router /helmEnvironment/{id}/re-deploy [post]
func (server *Server) RedeployHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	helmEnvironment, user, code, err := server.GetHelmEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input := map[string]interface{}{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	rawVersion, err := json.Marshal(input["version"])
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid version"))
		return
	}

	helmEnvironment.Version.RawMessage = rawVersion
	if values, ok := input["values"]; ok {
		helmEnvironment.Values = values.(string)
	}

	_, err = helmEnvironment.Update(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to recreate helmEnvironment"))
		return
	}
	err = queue.Publish(constants.UpgradeHelmRelease, helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironment.Application.Name,
		ApplicationID:   int64(helmEnvironment.Application.ID),
		EnvironmentID:   int64(helmEnvironment.ID),
		EnvironmentName: helmEnvironment.Name,
		ProjectName:     helmEnvironment.Application.Project.Name,
		ProjectID:       int64(helmEnvironment.Application.Project.ID),
		OrganizationID:  int64(helmEnvironment.Application.Project.OrganizationId),
	}
	if helmEnvironment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = helmEnvironment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "redeployed", fmt.Sprintf("helmEnvironment %s redeployed", helmEnvironment.Name), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, helmEnvironment.Application.Project, helmEnvironment.Application, nil, helmEnvironment, "re-deployed")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, helmEnvironment)
		return
	}
	helmEnvResponse, err := CreateHelmEnvironmentResponse(helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    helmEnvResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteEnvironment godoc
// @Summary Delete a HelmEnvironment
// @Description Delete a HelmEnvironment with the input payload
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 204 {object} doc.HelmEnvironment
// @Router /helm-environment/{id} [delete]
func (server *Server) DeleteHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	helmEnvironment, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user, err := CheckHelmEnvironmentPermission(server.DB, r, ADMIN, helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	helmEnvironment.Action = "Deleting"
	err = queue.Publish(constants.DestroyRelease, helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironment.Application.Name,
		ApplicationID:   int64(helmEnvironment.Application.ID),
		EnvironmentID:   int64(helmEnvironment.ID),
		EnvironmentName: helmEnvironment.Name,
		ProjectName:     helmEnvironment.Application.Project.Name,
		ProjectID:       int64(helmEnvironment.Application.Project.ID),
		OrganizationID:  int64(helmEnvironment.Application.Project.OrganizationId),
	}
	if helmEnvironment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = helmEnvironment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "deleted", fmt.Sprintf("helmEnvironment %s deleted", helmEnvironment.Name), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("delete", "environment", user.ID, helmEnvironment.Application.Project, helmEnvironment.Application, nil, helmEnvironment, "")
	time.Sleep(time.Duration(time.Second))
	_, err = helmEnvironment.Delete(server.DB, uint64(helmEnvironment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// RollbackEnvironment godoc
// @Summary Rollback to the provided deployment
// @Description Rollback your HelmEnvironment to by choosing any previously build images and pass that information in body
// @Tags Rollback
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Param body body doc.RepositoryImage true "Deployment Rollback"
// @Success 200 {object} doc.HelmEnvironment
// @Router /helm-environment/{id}/rollback [post]
func (server *Server) RollbackHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	helmEnvironment, user, code, err := server.GetHelmEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	image := models.RepositoryImage{}
	err = json.Unmarshal(body, &image)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	_, err = helmEnvironment.Update(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to rollback helmEnvironment "))
		return
	}
	helmEnvironment.Action = "Rolling back"
	err = queue.Publish(constants.UpgradeHelmRelease, helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: helmEnvironment.Application.Name,
		ApplicationID:   int64(helmEnvironment.Application.ID),
		EnvironmentID:   int64(helmEnvironment.ID),
		EnvironmentName: helmEnvironment.Name,
		ProjectName:     helmEnvironment.Application.Project.Name,
		ProjectID:       int64(helmEnvironment.Application.Project.ID),
		OrganizationID:  int64(helmEnvironment.Application.Project.OrganizationId),
	}
	if helmEnvironment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = helmEnvironment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "rollback", fmt.Sprintf("helmEnvironment %s rollback", helmEnvironment.Name), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, helmEnvironment.Application.Project, helmEnvironment.Application, nil, helmEnvironment, "rollback")

	responses.JSON(w, http.StatusOK, helmEnvironment)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, helmEnvironment)
		return
	}
	helmEnvResponse, err := CreateHelmEnvironmentResponse(helmEnvironment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    helmEnvResponse,
		Success: 1,
		Message: "Success",
	})
}

func CreateHelmEnvironmentResponse(helmEnvironment *models.HelmEnvironment) (*doc.HelmEnvironment, error) {
	helmEnvironmentResponse := doc.HelmEnvironment{}
	helmEnvironmentBytes, _ := json.Marshal(helmEnvironment)
	err := json.Unmarshal(helmEnvironmentBytes, &helmEnvironmentResponse)
	if err != nil {
		return nil, err
	}
	return &helmEnvironmentResponse, nil

}
