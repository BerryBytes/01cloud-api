package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var operatorRepo = models.NewOperatorRepo()
var operatorRequestRepo = models.NewOperatorRequest()

// GetOperators godoc
// @Summary Get Operators
// @Description Get list of Operators
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param page query int true "Page"
// @Param limit query int true "Limit"
// @Success 200 {array} map[string]interface{}
// @Router /Operators [get]
func (server *Server) GetOperators(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseUint(r.FormValue("page"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.ParseUint(r.FormValue("limit"), 10, 64)
	if limit < 1 {
		limit = 1000
	}
	data, err := operatorRepo.FindAllOperator(server.DB, true, limit, page)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data["operators"])
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    data["operators"],
		Success: 1,
		Message: "Success",
	})
}

// GetOperator godoc
// @Summary Get operator
// @Description Get operator
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param packageName query string true "Package Name"
// @Success 200 {object} doc.Operator
// @Router /operator [get]
func (server *Server) GetOperator(w http.ResponseWriter, r *http.Request) {
	packageName := r.FormValue("packageName")
	if packageName == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("required package name"))
		return
	}
	data, err := operatorRepo.FindOperator(server.DB, packageName)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data)
		return
	}
	operatorResponse, err := CreateOperatorResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    operatorResponse,
		Success: 1,
		Message: "Success",
	})
}

// InstallOperator godoc
// @Summary Install Operator
// @Description Install Operator
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.OperatorRequest true "Operator Request"
// @Success 200 {object} doc.Operator
// @Router /cluster/{id}/operator/install [post]
func (server *Server) InstallOperator(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//clusterRequest := models.ClusterRequest{}
	clusterRequest, err := clusterRequestInterface.FindWithDns(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uint64(oid) != clusterRequest.OrganizationID {
		responses.ERROR(w, http.StatusNotFound, errors.New("you dont have access to change this"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, getErrorMessage("all", err.Error()))
		return
	}
	operatorReq := models.OperatorRequest{}
	err = json.Unmarshal(body, &operatorReq)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, getErrorMessage("all", err.Error()))
		return
	}
	if operatorReq.PackageName == "" {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("operator package name is empty"))
		return
	}
	response, err := http.Get(constants.OperatorBaseUrl + "operator?packageName=" + operatorReq.PackageName)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to fetch operator"))
		return
	}

	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to fetch operator"))
		return
	}
	var detail map[string]interface{}
	err = json.Unmarshal(responseData, &detail)
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to fetch operator"))
		return
	}
	mDetail, _ := json.Marshal(detail)
	operatorReq.OperatorDetails.RawMessage = mDetail
	operatorReq.ClusterRequestID = cid
	operatorReq.OrganizationID = uint64(oid)
	operatorReq.GlobalOperator = detail["operator"].(map[string]interface{})["globalOperator"].(bool)
	d, _ := operatorRequestRepo.FindByNameAndClusterID(server.DB, operatorReq.ClusterRequestID, operatorReq.PackageName)
	if d != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("package already installed in cluster"))
		return
	}

	or, err := operatorRequestRepo.Save(server.DB, operatorReq)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, getErrorMessage("all", err.Error()))
		return
	}
	request, err := operatorRequestRepo.Find(server.DB, uint64(or.ID))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, getErrorMessage("all", err.Error()))
		return
	}
	request.Cluster = request.ClusterRequest.Cluster
	//request.OperatorDetails.RawMessage = nil
	err = queue.Publish(constants.InstallOperator, request)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("error installing operator"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "Operator install initiated")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// UnInstallOperator godoc
// @Summary UnInstall Operator
// @Description UnInstall Operator
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Cluster Id"
// @Param oid path int true "Operator Id"
// @Success 200 {string} string
// @Router /cluster/{id}/operator/{oid} [delete]
func (server *Server) UnInstallOperator(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	operatorId, err := strconv.ParseUint(vars["oid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := clusterRequestInterface.FindWithDns(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uint64(oid) != data.OrganizationID {
		responses.ERROR(w, http.StatusNotFound, errors.New("you don't have access to change this"))
		return
	}

	operator, err := operatorRequestRepo.Find(server.DB, operatorId)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("operator not available"))
		return
	}
	//operator.OperatorDetails.RawMessage = nil
	err = queue.Publish(constants.UnInstallOperator, operator)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("error uninstalling operator"))
		return
	}
	_, _ = operatorRequestRepo.Delete(server.DB, operatorId)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "Operator uninstall initiated for "+operator.PackageName)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// UnInstallOperator godoc
// @Summary Operator Status
// @Description Operator Status
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Cluster Id"
// @Success 200 {string} string
// @Router /cluster/{id}/operator/status [get]
func (server *Server) StatusOperator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := clusterRequestInterface.FindWithDns(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	operatorReq := models.OperatorRequest{}
	operatorReq.ClusterRequestID = cid
	operators, err := operatorRequestRepo.FindAll(server.DB, operatorReq)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	namespaces := []string{"operators"}
	for _, o := range *operators {
		if ns := helper.GetOperatorNamespace(&o); ns != "operators" {
			namespaces = append(namespaces, ns)
		}
	}
	d, err := json.Marshal(map[string]interface{}{"data": namespaces})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	operatorReq.OperatorDetails.RawMessage = d
	operatorReq.Cluster = data.Cluster

	operatorReq.Cluster.ClusterRequestID = cid
	err = queue.Publish(constants.StatusOperator, operatorReq)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "Operator status will be sent via websocket")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// ReInstallOperator godoc
// @Summary ReInstall Operator
// @Description ReInstalling Operator
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Cluster Id"
// @Param oid path int true "Operator Id"
// @Success 200 {string} string
// @Router /cluster/{id}/operator/{oid}/reinstall [post]
func (server *Server) ReInstallOperator(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	operatorId, err := strconv.ParseUint(vars["oid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := clusterRequestInterface.FindWithDns(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uint64(oid) != data.OrganizationID {
		responses.ERROR(w, http.StatusNotFound, errors.New("you don't have access to change this"))
		return
	}
	operator, err := operatorRequestRepo.Find(server.DB, operatorId)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("operator not available"))
		return
	}
	err = queue.Publish(constants.ReInstallOperator, operator)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("error reinstalling operator"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "Operator re-install initiated")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})

}

// GetOrganizationOperators godoc
// @Summary Get Organization Operators
// @Description Get Organization Operators
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Cluster Id"
// @Success 200 {array} doc.OperatorRequest
// @Router /cluster/{id}/operators [Get]
func (server *Server) GetOrganizationOperators(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.OperatorRequest{}
	data.ClusterRequestID = cid
	datas, err := operatorRequestRepo.FindAll(server.DB, data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	operatorRequestList := []doc.OperatorRequest{}
	for _, item := range *datas {
		activityResponse, err := CreateOperatorRequestResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		operatorRequestList = append(operatorRequestList, *activityResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    operatorRequestList,
		Success: 1,
		Message: "Success",
	})
}

// GetOperatorActivityLog godoc
// @Summary Get Operator Activity Log
// @Description Get Operator Activity Log
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Cluster Id"
// @Param oid path int true "Operator Id"
// @Param page query int true "Page"
// @Param limit query int true "Limit"
// @Success 200 {array} map[string]interface{}
// @Router /cluster/{id}/operator/{oid}/activity-log [Get]
func (server *Server) GetOperatorActivityLog(w http.ResponseWriter, r *http.Request) {
	_, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	operatorId, err := strconv.ParseUint(vars["oid"], 10, 64)
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
	data, err := clusterRequestInterface.FindWithDns(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uint64(oid) != data.OrganizationID {
		responses.ERROR(w, http.StatusNotFound, errors.New("you don't have access to change this"))
		return
	}
	operator, err := operatorRequestRepo.Find(server.DB, operatorId)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("operator not available"))
		return
	}
	namespace := helper.GetOperatorNamespace(operator)
	if namespace == "operators" {
		namespace = fmt.Sprintf("operators-%d", operator.ID)
	}
	status, err := server.StoreClient.Activity().GetActivityLog(namespace, "", page, limit)
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

// SyncAllOperator godoc
// @Summary Sync All Operator
// @Description Sync All Operator
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Operator  true "Operator"
// @Success 200 {string} string
// @Router /admin/operator/sync-all [post]
func (server *Server) SyncAllOperator(w http.ResponseWriter, r *http.Request) {
	type Operators struct {
		Data []models.Operator `json:"operators"`
	}
	opList := &Operators{}
	response, err := http.Get(constants.OperatorBaseUrl + "operators")
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to fetch operators"))
		return
	}
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	err = json.Unmarshal(responseData, opList)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	for _, j := range opList.Data {
		opReq, err := getOperators(j.PackageName)
		if err != nil {
			continue
		}
		err = operatorRepo.SyncOperator(server.DB, &opReq.Data)
		if err != nil {
			continue
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "Synced Successfully")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// SyncOperator godoc
// @Summary Sync Operator
// @Description Sync Operator
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Operator  true "Operator Request"
// @Param packageName query string true "Package Name"
// @Success 200 {string} string
// @Router /admin/operator/sync [post]
func (server *Server) SyncOperator(w http.ResponseWriter, r *http.Request) {
	packageName := r.FormValue("packageName")
	response, err := http.Get(constants.OperatorBaseUrl + "operator?packageName=" + packageName)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to fetch operator"))
		return
	}
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	opReq := &models.OperatorReq{}
	err = json.Unmarshal(responseData, opReq)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = operatorRepo.SyncOperator(server.DB, &opReq.Data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "updated data successfully")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// EnableDisableOperator godoc
// @Summary Enable Disable Operator
// @Description Enable or Disable Operator
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param packageName query string true "Package Name"
// @Success 200 {string} string
// @Router /admin/operator/enable-disable [post]
func (server *Server) EnableDisableOperator(w http.ResponseWriter, r *http.Request) {
	packageName := r.FormValue("packageName")
	if packageName == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("packageName required"))
		return
	}
	isActive, err := strconv.ParseBool(r.FormValue("isActive"))
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("isActive required"))
		return
	}
	err = operatorRepo.EnableDisable(server.DB, packageName, isActive)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "isActive field updated successfully")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetOperatorsForAdmin godoc
// @Summary Get Operator for Admin
// @Description  Get Operator for Admin
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param page query int true "Page"
// @Param limit query int true "Limit"
// @Success 200 {string} string
// @Router /admin/operators [get]
func (server *Server) GetOperatorsForAdmin(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseUint(r.FormValue("page"), 10, 64)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.ParseUint(r.FormValue("limit"), 10, 64)
	if limit < 1 {
		limit = 1000
	}
	data, err := operatorRepo.FindAllOperator(server.DB, false, limit, page)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"data":  data["operators"],
			"total": data["count"],
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]interface{}{
			"data":  data["operators"],
			"total": data["count"],
		},
		Success: 1,
		Message: "Success",
	})
}

// UpdateOperator godoc
// @Summary Update Operator
// @Description  Update Operator
// @Tags Operators
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {string} string
// @Router /admin/operator/{packageName} [put]
func (server *Server) UpdateOperator(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	packageName := vars["packageName"]
	if packageName == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("packageName required"))
		return
	}
	data, err := operatorRepo.FindOperator(server.DB, packageName)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := &models.Operator{}
	err = json.Unmarshal(body, dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate.PackageName = data.PackageName
	err = operatorRepo.UpdateOperator(server.DB, dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "operator updated successfully")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func getOperators(packageName string) (*models.OperatorReq, error) {
	opReq := &models.OperatorReq{}
	responsess, err := http.Get(constants.OperatorBaseUrl + "operator?packageName=" + packageName)
	if err != nil {
		return nil, err
	}
	responseDatas, err := io.ReadAll(responsess.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = json.Unmarshal(responseDatas, opReq)
	if err != nil {
		return nil, err
	}
	return opReq, nil
}

func CreateOperatorResponse(operator *models.Operator) (*doc.Operator, error) {
	operatorResponse := doc.Operator{}
	operatorBytes, _ := json.Marshal(operator)
	err := json.Unmarshal(operatorBytes, &operatorResponse)
	if err != nil {
		return nil, err
	}
	return &operatorResponse, nil
}

func CreateOperatorRequestResponse(operator *models.OperatorRequest) (*doc.OperatorRequest, error) {
	operatorResponse := doc.OperatorRequest{}
	operatorBytes, _ := json.Marshal(operator)
	err := json.Unmarshal(operatorBytes, &operatorResponse)
	if err != nil {
		return nil, err
	}
	return &operatorResponse, nil
}
