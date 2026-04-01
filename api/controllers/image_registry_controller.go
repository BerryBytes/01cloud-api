package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"01cloud-api/api/auth"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var imageRegistryInterface = models.NewImageRegistry()

// CreateImageRegistry godoc
// @Summary Create Image Registry
// @Description Create Image Registry
// @Tags ImageRegistry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.ImageRegistry true "Create Image Registry"
// @Success 201 {object} doc.ImageRegistry
// @Router /registry [post]
func (server *Server) CreateImageRegistry(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.ImageRegistry{}
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

	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
	// checks for duplicate registry entries
	if imageRegistryInterface.IsNameExists(server.DB, oid, data.Name) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("registry name already exists "))
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

	if oid > 0 {
		_, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}

	data.OrganizationID = uint64(oid)

	if data.Provider == constants.GCP {
		var credentialsMap map[string]interface{}
		err := json.Unmarshal(data.Credentials.RawMessage, &credentialsMap)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid credentials"))
			return
		}
		credentials := helper.ReadFileInString(credentialsMap["gcp"].(map[string]interface{})["google_app_cred"].(string))
		if credentials == "" {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid credentials"))
			return
		}
		var serviceAccount map[string]interface{}
		err = json.Unmarshal([]byte(credentials), &serviceAccount)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid credentials"))
			return
		}
		data.Service = "gcr.io"
		data.UserName = "_json_key"
		data.Password = credentials
		data.ProjectName = serviceAccount["project_id"].(string)
	}
	if data.Provider == constants.CUSTOM {
		var credentialsMap map[string]interface{}
		err := json.Unmarshal(data.Credentials.RawMessage, &credentialsMap)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid credentials"))
			return
		}
		credentialsValue := credentialsMap["credentials"].(map[string]interface{})
		jsonbody, err := json.Marshal(credentialsValue)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		credentials := models.Credentials{}
		if err := json.Unmarshal(jsonbody, &credentials); err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		data.Service = credentials.Server
		data.UserName = credentials.UserName
		data.Password = credentials.Password
		data.ProjectName = credentials.ProjectName
	}
	dataCreated, err := imageRegistryInterface.Save(server.DB, &data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if registry is created within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "create", "registry", data.Name, "")
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	registryResponse, err := CreateRegistryDetailResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    registryResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetImageRegistrys godoc
// @Summary Get Image Registries
// @Description Get Image Registries
// @Tags ImageRegistry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 201 {array} doc.ImageRegistry
// @Router /registries [get]
func (server *Server) GetImageRegistrys(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
	_, err = userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	data := models.ImageRegistry{}
	data.OrganizationID = uint64(oid)
	datas, err := imageRegistryInterface.FindAll(server.DB, &data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	registryResponseList := []doc.ImageRegistryDetails{}
	for _, registry := range *datas {
		dnsResponse, err := CreateRegistryDetailResponse(&registry)
		if err != nil {
			log.Error(err)
		}
		registryResponseList = append(registryResponseList, *dnsResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    registryResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetRegistryConfig godoc
// @Summary Get Registry Config
// @Description Get Registry Config
// @Tags ImageRegistry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} map[string]interface{}
// @Router /registry-config [get]
func (server *Server) GetRegistryConfig(w http.ResponseWriter, r *http.Request) {
	file, _ := os.ReadFile("/data/public/registry.schema.json")
	res := []map[string]interface{}{}
	err := json.Unmarshal(file, &res)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, res)
}
func (server *Server) PostRegistryConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if _, err = os.Stat("/data/registry"); err != nil {
		err = os.Mkdir("/data/registry", 0600)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
	err = os.WriteFile("/data/public/registry.schema.json", body, 0777)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "Successfully Updated")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetRegion godoc
// @Summary Get Region
// @Description Get Aws Region
// @Tags ImageRegistry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /registry/aws-region [get]
func (server *Server) GetRegion(w http.ResponseWriter, r *http.Request) {
	file, _ := os.ReadFile("/data/registry/aws-region.json")
	var data map[string]interface{}
	err := json.Unmarshal(file, &data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, data)
}

func (server *Server) GetImageRegistrysForAdmin(w http.ResponseWriter, r *http.Request) {
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

	data := models.ImageRegistry{}
	data.OrganizationID = uint64(oid)
	datas, err := imageRegistryInterface.FindAllWithInactive(server.DB, &data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	registryResponseList := []doc.ImageRegistryDetails{}
	for _, registry := range *datas {
		dnsResponse, err := CreateRegistryDetailResponse(&registry)
		if err != nil {
			log.Error(err)
		}
		registryResponseList = append(registryResponseList, *dnsResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    registryResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetImageRegistry godoc
// @Summary Get Image Registry
// @Description Get Image Registry
// @Tags ImageRegistry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Image Registry"
// @Success 200 {object} doc.ImageRegistry
// @Router /registry/{id} [get]
func (server *Server) GetImageRegistry(w http.ResponseWriter, r *http.Request) {
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
	dataReceived, err := imageRegistryInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	registryResponse, err := CreateRegistryDetailResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    registryResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateImageRegistry godoc
// @Summary Update Image Registry
// @Description Update Image Registry
// @Tags ImageRegistry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Update Image Registry"
// @Param body body doc.ImageRegistry true "Update Image Registry"
// @Success 200 {object} doc.ImageRegistry
// @Router /registry/{id} [put]
func (server *Server) UpdateImageRegistry(w http.ResponseWriter, r *http.Request) {
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
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	data := models.ImageRegistry{}
	err = server.DB.Model(models.ImageRegistry{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("ImageRegistry not found "))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.ImageRegistry{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// checks for duplicate registry entries
	if len(dataUpdate.Name) > 0 && data.Name != dataUpdate.Name {
		if imageRegistryInterface.IsNameExists(server.DB, oid, dataUpdate.Name) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("registry name already exists "))
			return
		}
	}

	if oid > 0 {
		_, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}
	if dataUpdate.Credentials.RawMessage != nil {
		if dataUpdate.Provider == constants.GCP {
			var credentialsMap map[string]interface{}
			err := json.Unmarshal(dataUpdate.Credentials.RawMessage, &credentialsMap)
			if err != nil {
				responses.ERROR(w, http.StatusNotFound, errors.New("invalid credentials"))
				return
			}
			credentials := helper.ReadFileInString(credentialsMap["gcp"].(map[string]interface{})["google_app_cred"].(string))
			if credentials == "" {
				responses.ERROR(w, http.StatusNotFound, errors.New("invalid credentials"))
				return
			}
			var serviceAccount map[string]interface{}
			err = json.Unmarshal([]byte(credentials), &serviceAccount)
			if err != nil {
				responses.ERROR(w, http.StatusNotFound, errors.New("invalid credentials"))
				return
			}
			dataUpdate.Service = "gcr.io"
			dataUpdate.UserName = "_json_key"
			dataUpdate.Password = credentials
			dataUpdate.ProjectName = serviceAccount["project_id"].(string)
		}
		if dataUpdate.Provider == constants.CUSTOM {
			var credentialsMap map[string]interface{}
			err := json.Unmarshal(dataUpdate.Credentials.RawMessage, &credentialsMap)
			if err != nil {
				responses.ERROR(w, http.StatusNotFound, errors.New("invalid credentials"))
				return
			}
			credentialsValue := credentialsMap["credentials"].(map[string]interface{})
			jsonbody, err := json.Marshal(credentialsValue)
			if err != nil {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}

			credentials := models.Credentials{}
			if err := json.Unmarshal(jsonbody, &credentials); err != nil {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			dataUpdate.Service = credentials.Server
			dataUpdate.UserName = credentials.UserName
			dataUpdate.Password = credentials.Password
			dataUpdate.ProjectName = credentials.ProjectName
		}
	}
	dataUpdate.ID = data.ID
	dataUpdated, err := imageRegistryInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if registry is Updated within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "update", "registry", dataUpdated.Name, "")
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	registryResponse, err := CreateRegistryDetailResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    registryResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteImageRegistry godoc
// @Summary Delete Image Registry
// @Description Delete Image Registry
// @Tags ImageRegistry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Delete Image Registry"
// @Success 200
// @Router /registry/{id} [delete]
func (server *Server) DeleteImageRegistry(w http.ResponseWriter, r *http.Request) {
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

	data := models.ImageRegistry{}
	err = server.DB.Model(models.ImageRegistry{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	count := 0
	err = server.DB.Model(models.Cluster{}).Where("image_registry_id = ?", pid).Count(&count).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	if count > 0 {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("this image registry is used by some active clusters "))
		return
	}
	_, err = imageRegistryInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// Checks if registry is Deleted within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "delete", "registry", data.Name, "")
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

func CreateRegistryResponse(registry *models.ImageRegistry) (*doc.ImageRegistry, error) {
	registryResponse := doc.ImageRegistry{}
	registryBytes, _ := json.Marshal(registry)
	err := json.Unmarshal(registryBytes, &registryResponse)
	if err != nil {
		return nil, err
	}
	return &registryResponse, nil
}

func CreateRegistryDetailResponse(registry *models.ImageRegistry) (*doc.ImageRegistryDetails, error) {
	registryResponse := doc.ImageRegistryDetails{}
	registryBytes, _ := json.Marshal(registry)
	err := json.Unmarshal(registryBytes, &registryResponse)
	if err != nil {
		return nil, err
	}
	return &registryResponse, nil
}
