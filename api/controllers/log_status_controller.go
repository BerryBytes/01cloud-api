package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
)

// GetEnvironmentActivityLog godoc
// @Summary Get Environment Activity Log
// @Description Get Environment Activity Log
// @Tags LogStatus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Param page query int true "Page"
// @Param limit query int true "Limit"
// @Success  200 {array} map[string]interface{}
// @Router /environment/{id}/activity-log [get]
func (server *Server) GetEnvironmentActivityLog(w http.ResponseWriter, r *http.Request) {
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
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	organization := models.Organization{}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if organization.CheckRole(server.DB, uint64(uid), environmentReceived.Application.Project.OrganizationId) {
			if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
				return
			}
		}
	}
	namespace := helper.GetNamespace(environmentReceived)
	if environmentReceived.ServiceType == 4 {
		packageName := environment.Application.OperatorPackageName
		if environment.Application.Cluster != nil {
			opReq, err := operatorRequestRepo.FindByNameAndClusterID(server.DB, environment.Application.Cluster.ClusterRequestID, packageName)
			if err == nil {
				namespace = helper.GetOperatorNamespace(opReq)
			}
		}
	}
	status, err := server.StoreClient.Activity().GetActivityLog(namespace, "", page, limit)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if status == nil {
		responses.JSON(w, http.StatusOK, []map[string]interface{}{})
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

// GetEnvironmentStatus godoc
// @Summary Get Environment Status
// @Description Get Environment Status
// @Tags LogStatus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Success  200 {array} map[string]interface{}
// @Router /environment/{id}/status [get]
func (server *Server) GetEnvironmentStatus(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
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
		if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
			return
		}
	}
	namespace := helper.GetNamespace(environmentReceived)
	if environmentReceived.ServiceType == 4 {
		namespace, err = server.NonGlobalOperatorNs(environmentReceived)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
	}

	status, err := server.StoreClient.PodState().GetPodStatus(namespace)

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

// GetEnvironmentState godoc
// @Summary Get Environment State
// @Description Get Environment State
// @Tags LogStatus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Success  200 {array} map[string]interface{}
// @Router /environment/{id}/state [get]
func (server *Server) GetEnvironmentState(w http.ResponseWriter, r *http.Request) {
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
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	organization := models.Organization{}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if organization.CheckRole(server.DB, uint64(uid), environmentReceived.Application.Project.OrganizationId) {
			if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
				return
			}
		}
	}
	var cname string
	var releaseId string
	var master_url string
	var external_url string
	namsespace := fmt.Sprintf("%s-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), environmentReceived.ApplicationID, pid)
	if environment.ServiceType == 4 {
		packageName := environment.Application.OperatorPackageName
		if environment.Application.Cluster != nil {
			opReq, _ := operatorRequestRepo.FindByNameAndClusterID(server.DB, environment.Application.Cluster.ClusterRequestID, packageName)
			namsespace = helper.GetOperatorNamespace(opReq)
		}
	}
	podStatus, state, ready, err := server.StoreClient.PodState().GetEnvironmentState(namsespace)
	if err != nil {
		responses.JSON(w, http.StatusInternalServerError, err)
	}
	certificateInfos, err := server.StoreClient.Certificate().GetCertificateDetail(namsespace)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if state == "" {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"state":        "Pending",
			"cname":        "",
			"external_url": "",
			"master_url":   "",
			"ready":        false,
			"pod_status":   []string{},
			"release_id":   releaseId,
			"certificates": certificateInfos,
		})
		return
	}
	if state == "Running" {
		ready = true
	}
	if ready {
		metadata, err := server.StoreClient.ReleaseInfo().GetData(fmt.Sprintf("%s-metadata-%d", os.Getenv("GCLOUD_NAMESPACE"), pid))
		if environment.ServiceType == 5 {
			master_url = fmt.Sprintf("%s.%s.svc.cluster.local", helper.GetReleaseName(environmentReceived), helper.GetNamespace(environmentReceived))
			if environment.ExternalURL {
				namespace := helper.GetNamespace(environmentReceived)
				domain := strings.TrimRight(environmentReceived.Application.Cluster.DNS.BaseDomain, ".")
				if domain != "" {
					external_url = fmt.Sprintf("%s.%s", namespace, domain)
				}
			}
		} else if err != nil || metadata == nil {
			namespace := helper.GetNamespace(environmentReceived)
			domain := strings.TrimRight(environmentReceived.Application.Cluster.DNS.BaseDomain, ".")
			if domain != "" {
				cname = fmt.Sprintf("%s.%s", namespace, domain)
			}
		} else {
			cname = metadata.CName
			releaseId = metadata.ReleaseId
		}
	}
	if state == "Completed" {
		state = "Running"
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"state":        state,
			"cname":        cname,
			"external_url": external_url,
			"master_url":   master_url,
			"ready":        ready,
			"pod_status":   podStatus,
			"release_id":   releaseId,
			"certificates": certificateInfos,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"state":        state,
			"cname":        cname,
			"external_url": external_url,
			"master_url":   master_url,
			"ready":        ready,
			"pod_status":   podStatus,
			"release_id":   releaseId,
			"certificates": certificateInfos,
		},
		Success: 1,
		Message: "Success",
	})
}

// GetPodList godoc
// @Summary Get Pod List
// @Description Get Pod List
// @Tags LogStatus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Success  200 {array} map[string]interface{}
// @Router /environment/{id}/pods [get]
func (server *Server) GetPodList(w http.ResponseWriter, r *http.Request) {
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
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
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

// FetchLogs godoc
// @Summary Fetch Logs
// @Description Fetch Logs
// @Tags LogStatus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param fetch_time query int true "Fetch Time"
// @Param no_of_lines query int true "Number of Lines"
// @Param id path int true "Environment Id"
// @Success  201 {string} string
// @Router /environment/{id}/fetch-logs [get]
func (server *Server) FetchLogs(w http.ResponseWriter, r *http.Request) {
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
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
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
	err = queue.Publish(constants.FetchLogs, payload)
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
	responses.JSON(w, http.StatusCreated, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// FetchEnvironmentState godoc
// @Summary Fetch Environment State
// @Description Fetch Environment State
// @Tags LogStatus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param action query string true "Action"
// @Param id path int true "Environment Id"
// @Success  201 {string} string
// @Router /environment/{id}/fetch-state [get]
func (server *Server) FetchEnvironmentState(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
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
	env, _ := environment.ToJson()
	if environment.ServiceType == 4 {
		packageName := environment.Application.OperatorPackageName
		// opReq := models.OperatorRequest{}
		if environment.Application.Cluster != nil {
			opReq, _ := operatorRequestRepo.FindByNameAndClusterID(server.DB, environment.Application.Cluster.ClusterRequestID, packageName)
			env["operators"] = map[string]interface{}{
				"namespace": helper.GetOperatorNamespace(opReq),
			}
		}
	}
	err = queue.Publish(constants.UpdateEnvironmentState, env)
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
	responses.JSON(w, http.StatusCreated, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// FetchEnvironmentPackageStatus godoc
// @Summary Fetch Environment Package Status
// @Description Fetch Environment Package Status
// @Tags LogStatus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Success  201 {string} string
// @Router /environment/{id}/package-status [get]
func (server *Server) FetchEnvironmentPackageStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	err = queue.Publish(constants.CheckPackageStatus, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]string{
			"message": "Fetching package status",
		})
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// FetchOperatorEnvironmentService godoc
// @Summary Fetch Operator Environment Service
// @Description Fetch Operator Environment Service
// @Tags LogStatus
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Success  201 {string} string
// @Router /environment/{id}/operator-service [get]
func (server *Server) FetchOperatorEnvironmentService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, uint64(pid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	_, err = CheckPermission(server.DB, r, READ, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	if environmentReceived.ServiceType == 4 {
		envJ, _ := environmentReceived.ToJson()
		packageName := environmentReceived.Application.OperatorPackageName
		opReq := &models.OperatorRequest{}
		if environmentReceived.Application.Cluster != nil {
			opReq, err = operatorRequestRepo.FindByNameAndClusterID(server.DB, environmentReceived.Application.Cluster.ClusterRequestID, packageName)
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, errors.New("invalid service type"))
				return
			}
		}
		envJ["operators"] = map[string]interface{}{
			"namespace": helper.GetOperatorNamespace(opReq),
		}
		err = queue.Publish(constants.FetchOperatorService, envJ)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("invalid operator envirnoment"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]string{
			"message": "*** Fetching operator environment service ***",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
