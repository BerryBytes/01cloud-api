package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models/doc"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var resourceInterface = models.NewResource()

// CreateResource godoc
// @Summary Create a new resource
// @Description Create a new resource with the input payload
// @Tags Resource
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Resource true "Create Resource"
// @Success 201 {object} doc.Resource
// @Router /resource [post]
func (server *Server) CreateResource(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.Resource{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.Prepare()
	err = data.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	// checks for duplicate entries
	if resourceInterface.IsNameExists(server.DB, oid, data.Name) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource name already exists "))
		return
	}

	// checks for duplicate resources
	if data.IsResourceExist(server.DB, oid, data.Cores, data.Memory) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource already exists "))
		return
	}
	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	if oid > 0 {
		organization, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}

	data.OrganizationID = uint64(oid)
	if oid > 0 && !resourceInterface.CheckResourceLimit(server.DB, organization, &data) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource limit exceeds "))
		return
	}
	dataCreated, err := resourceInterface.Save(server.DB, &data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if resource is Updated within Orgination level then we have to add to audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "create", "resource", data.Name, "")
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	resourceResponse, err := CreateResourceResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    resourceResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetResources godoc
// @Summary Get Resources
// @Description Get list of resources
// @Tags Resource
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.Resource
// @Router /resources [get]
func (server *Server) GetResources(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	// key := fmt.Sprintf("resources-%d-%d", uid, oid)
	// var value interface{}
	// if ok := server.Cache.Get(key, &value); ok {
	// 	responses.JSON(w, http.StatusOK, value)
	// 	return
	// }
	_, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.Resource{}
	data.OrganizationID = uint64(oid)
	datas, err := resourceInterface.FindAll(server.DB, &data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	resourceResponseList := []doc.Resource{}
	for _, item := range *datas {
		resourceResponse, err := CreateResourceResponse(&item)
		if err != nil {
			logrus.Error(err)
			continue
		}
		resourceResponseList = append(resourceResponseList, *resourceResponse)
	}
	resp := responses.Response{
		Data:    resourceResponseList,
		Success: 1,
		Message: "Success",
	}
	// server.Cache.Set(key, resp)
	responses.JSON(w, http.StatusOK, resp)
}

// GetResourcesForAdmin godoc
// @Summary Get Resources by organization
// @Description Get list of resources by organization
// @Tags Resource
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.Resource
// @Router /admin/resources [get]
func (server *Server) GetResourcesForAdmin(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}
	// key := fmt.Sprintf("resources-admin-%d-%d", uid, oid)
	// var value interface{}
	// if ok := server.Cache.Get(key, &value); ok {
	// 	responses.JSON(w, http.StatusOK, value)
	// 	return
	// }
	data := models.Resource{}
	data.OrganizationID = uint64(oid)
	datas, err := resourceInterface.FindAllWithInactive(server.DB, &data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	resourceResponseList := []doc.Resource{}
	for _, item := range *datas {
		resourceResponse, err := CreateResourceResponse(&item)
		if err != nil {
			logrus.Error(err)
			continue
		}
		resourceResponseList = append(resourceResponseList, *resourceResponse)
	}
	// server.Cache.Set(key, resourceResponseList)
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resourceResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetResource godoc
// @Summary Get Resource by id
// @Description Get Resource by id from token
// @Tags Resource
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Resource id"
// @Success 200 {object} doc.Resource
// @Router /resource/{id} [get]
func (server *Server) GetResource(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// key := fmt.Sprintf("resource-%d", pid)
	// var value interface{}
	// if ok := server.Cache.Get(key, &value); ok {
	// 	responses.JSON(w, http.StatusOK, value)
	// 	return
	// }
	dataReceived, err := resourceInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	resourceResponse, err := CreateResourceResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// server.Cache.Set(key, resourceResponse)
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    resourceResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateResource godoc
// @Summary Update a Resource
// @Description Update a Resource with the input payload
// @Tags Resource
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Resource id"
// @Param body body doc.Resource true "Update Resource"
// @Success 200 {object} doc.Resource
// @Router /resource/{id} [put]
func (server *Server) UpdateResource(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	data, err := resourceInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("resource not found "))
		return
	}
	dataUpdate := models.Resource{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// checks for duplicate entries
	if len(dataUpdate.Name) > 0 && data.Name != dataUpdate.Name {
		if resourceInterface.IsNameExists(server.DB, oid, dataUpdate.Name) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource name already exists "))
			return
		}
	}
	// checks for duplicate resources
	if dataUpdate.Cores != data.Cores || dataUpdate.Memory != data.Memory {
		if data.IsResourceExist(server.DB, oid, dataUpdate.Cores, dataUpdate.Memory) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource already exists "))
			return
		}
	}

	if oid > 0 {
		organization, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}

	dataUpdate.ID = data.ID
	if oid > 0 && !resourceInterface.CheckResourceLimit(server.DB, organization, &dataUpdate) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource limit exceeds "))
		return
	}
	dataUpdated, err := resourceInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if resources is Updated within Orgination level then we have to add to audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "update", "resource", dataUpdated.Name, "")
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	resourceResponse, err := CreateResourceResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resourceResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteResource godoc
// @Summary Delete a Resource
// @Description Delete a Resource with the input payload
// @Tags Resource
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Resource id"
// @Success 204 {object} doc.Resource
// @Router /resource/{id} [delete]
func (server *Server) DeleteResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}
	resourceData, err := resourceInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	count := 0
	err = server.DB.Model(models.Environment{}).Where("resource_id = ?", pid).Count(&count).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	if count > 0 {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("this resource is used by some active environment "))
		return
	}
	_, err = resourceInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// Checks if resources is Deleted within Orgination level then we have to add to audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "delete", "resource", resourceData.Name, "")
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

func CreateResourceResponse(resource *models.Resource) (*doc.Resource, error) {
	resourceResponse := doc.Resource{}
	resourceBytes, _ := json.Marshal(resource)
	err := json.Unmarshal(resourceBytes, &resourceResponse)
	if err != nil {
		return nil, err
	}
	return &resourceResponse, nil
}
