package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
)

// GetHelmEnvironmentActivityLog godoc
// @Summary Get Helm Environment Activity Log
// @Description Get Helm Environment Activity Log
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param page query int true "Page"
// @Param id path int true "Helm Environment Id"
// @Param limit query int true "Query"
// @Success 200 {array} map[string]interface{}
// @Router /helm-environment/{id}/activity-log [get]
func (server *Server) GetHelmEnvironmentActivityLog(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	page, err := strconv.ParseUint(r.FormValue("page"), 10, 64)
	if err != nil || page < 1 {
		page = 0
	} else {
		page = page - 1
	}
	limit, err := strconv.ParseUint(r.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 100
	}
	environment := models.HelmEnvironment{}
	environmentReceived, err := environment.Find(server.DB, int64(pid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsAuthorizedHelmEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
			return
		}
	}
	status, err := server.StoreClient.Activity().GetActivityLog(helper.GetHelmNamespace(environmentReceived), "", page, limit)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if status == nil {
		responses.ERROR(w, http.StatusNoContent, errors.New("no data available"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, status)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    status,
		Success: 1,
		Message: "Success",
	})
}

// GetHelmEnvironmentStatus godoc
// @Summary Get Helm Environment Status
// @Description Get Helm Environment Status
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Helm Environment Id"
// @Success 200 {array} map[string]interface{}
// @Router /helm-environment/{id}/status [get]
func (server *Server) GetHelmEnvironmentStatus(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	environment := models.HelmEnvironment{}
	environmentReceived, err := environment.Find(server.DB, int64(pid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsAuthorizedHelmEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
			return
		}
	}

	status, err := server.StoreClient.PodState().GetPodStatus(helper.GetHelmNamespace(environmentReceived))

	if err != nil {
		responses.ERROR(w, http.StatusNoContent, err)
		return
	}

	if status == nil {
		responses.ERROR(w, http.StatusNoContent, errors.New("no data available"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, status)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    status,
		Success: 1,
		Message: "Success",
	})
}

// GetHelmEnvironmentState godoc
// @Summary Get Helm Environment State
// @Description Get Helm Environment State
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Helm Environment Id"
// @Success 200 {object} map[string]interface{}
// @Router /helm-environment/{id}/state [get]
func (server *Server) GetHelmEnvironmentState(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	environment := models.HelmEnvironment{}
	environmentReceived, err := environment.Find(server.DB, int64(pid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !environmentReceived.Active {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"state":        "Stopped",
			"ready":        true,
			"pod_status":   []string{},
			"release_id":   "",
			"certificates": []string{},
		})
		return
	}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsAuthorizedHelmEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
			return
		}
	}
	//var cname string
	var releaseId string
	podStatus, state, ready, err := server.StoreClient.PodState().GetEnvironmentState(fmt.Sprintf("%s-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), environmentReceived.ApplicationID, pid))
	if err != nil {
		responses.JSON(w, http.StatusInternalServerError, err)
		return
	}
	if state == "" {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"state":      "Pending",
			"ready":      false,
			"pod_status": []string{},
			"release_id": releaseId,
		})
		return
	}

	if ready {
		metadata, err := server.StoreClient.ReleaseInfo().GetData(fmt.Sprintf("%s-helm-metadata-%d", os.Getenv("GCLOUD_NAMESPACE"), pid))
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		if metadata != nil {
			//cname = metadata.CName
			releaseId = metadata.ReleaseId
		}
	}
	if state == "Completed" {
		state = "Running"
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"state":      state,
			"ready":      ready,
			"pod_status": podStatus,
			"release_id": releaseId,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"state":      state,
			"ready":      ready,
			"pod_status": podStatus,
			"release_id": releaseId,
		},
		Success: 1,
		Message: "Success",
	})
}

// GetHelmPodList godoc
// @Summary Get Helm Pod List
// @Description Get Helm Pod List
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Helm Environment Id"
// @Success 200 {array} map[string]interface{}
// @Router /helm-environment/{id}/pods [get]
func (server *Server) GetHelmPodList(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	environment := models.HelmEnvironment{}
	environmentReceived, err := environment.Find(server.DB, int64(pid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsAuthorizedHelmEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
			return
		}
	}
	podStatus, _, _, err := server.StoreClient.PodState().GetEnvironmentState(fmt.Sprintf("%s-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), environmentReceived.ApplicationID, pid))
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

// FetchHelmLogs godoc
// @Summary Fetch Helm Logs
// @Description Fetch Helm Logs
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Helm Environment Id"
// @Param fetch_time query int true "Fetch Time"
// @Param no_of_lines query int true "Number of Lines"
// @Success 200
// @Router /helm-environment/{id}/fetch-logs [get]
func (server *Server) FetchHelmLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fetchTime, err := strconv.ParseInt(r.FormValue("fetch_time"), 10, 64)
	if err != nil {
		fetchTime = 0
	}

	noOfLines, err := strconv.ParseInt(r.FormValue("no_of_lines"), 10, 64)
	if err != nil {
		noOfLines = 0
	}

	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	environment := models.HelmEnvironment{}
	environmentReceived, err := environment.Find(server.DB, int64(pid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	payload, err := environmentReceived.ToJson()
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	payload["fetch_time"] = fetchTime
	payload["no_of_lines"] = noOfLines
	err = queue.Publish(constants.FetchHelmLogs, payload)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]string{
			"message": "** Fetching logs ***",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// FetchHelmEnvironmentState godoc
// @Summary Fetch Helm Environment State
// @Description Fetch Helm Environment State
// @Tags HelmEnvironment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Helm Environment Id"
// @Param action query string true "Action"
// @Success 200
// @Router /helm-environment/{id}/fetch-state [get]
func (server *Server) FetchHelmEnvironmentState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	environment := models.HelmEnvironment{}
	environmentReceived, err := environment.Find(server.DB, int64(pid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	defaultAction := r.FormValue("action")
	if defaultAction != "" {
		environmentReceived.Action = defaultAction
	}
	//if environmentReceived.Active == false{
	//	responses.ERROR(w, http.StatusForbidden, errors.New("environment is stopped"))
	//	return
	//}
	err = queue.Publish(constants.UpdateHelmEnvironmentState, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]string{
			"message": "*** Updating environment state ***",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
