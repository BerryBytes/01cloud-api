package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var initContainerInterface = models.NewInitContainer()

// CreateInitContainer godoc
// @Summary Create InitContainer
// @Description Create InitContainer
// @Tags InitContainer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Param body body doc.InitContainer true "Create InitContainer"
// @Success 200
// @Router /environment/{id}/init-container [post]
func (server *Server) CreateInitContainer(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	InitContainer := &models.InitContainer{}
	err = json.Unmarshal(body, InitContainer)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	InitContainer.Prepare()
	err = InitContainer.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
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

	if !environmentReceived.Active {
		responses.ERROR(w, http.StatusBadRequest, errors.New("environment is in stopped state"))
		return
	}

	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to create cron job"))
			return
		}
	}

	if environmentReceived.ServiceType < 1 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("not supported for this plugin"))
		return
	}

	if InitContainer.Image == "" {
		data, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetNamespace(environmentReceived))
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("image not found"))
			return
		}
		if data == nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("image not found"))
			return
		}
		img := fmt.Sprintf("%s:%s", data.CIRequest.RepositoryImage.Name, data.CIRequest.RepositoryImage.Tag)
		InitContainer.Image = img
	}

	//_, state, ready := datastore_helper.GetEnvironmentState(server.FirestoreClient, fmt.Sprintf("%s-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), environmentReceived.ApplicationID, pid))
	//if state == "" {
	//	responses.ERROR(w, http.StatusTooEarly, errors.New("No data available"))
	//	return
	//}
	//if !ready {
	//	responses.ERROR(w, http.StatusTooEarly, errors.New("Environment is not ready"))
	//	return
	//}
	data, err := initContainerInterface.SaveInitContainer(server.DB, pid, InitContainer)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	err = queue.Publish(constants.CreateInitContainer, &map[string]interface{}{
		"environment":    environmentReceived,
		"init_container": data,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "***  InitContainer Requested ***",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// UpdateInitContainer godoc
// @Summary Update InitContainer
// @Description Update InitContainer
// @Tags InitContainer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Param cid path int true "cid"
// @Param body body doc.InitContainer true "Update InitContainer"
// @Success 200
// @Router /environment/{id}/init-container/{cid} [put]
func (server *Server) UpdateInitContainer(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	cid, err := strconv.ParseUint(vars["cid"], 10, 64)
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

	if !environmentReceived.Active {
		responses.ERROR(w, http.StatusBadRequest, errors.New("environment is in stopped state"))
		return
	}
	if environmentReceived.ServiceType < 1 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("not supported for this plugin"))
		return
	}

	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to create init container"))
			return
		}
	}

	data := models.InitContainer{}
	err = server.DB.Model(models.InitContainer{}).Where("id = ? and environment_id = ?", cid, pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("InitContainer not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	InitContainerUpdate := &models.InitContainer{}

	err = json.Unmarshal(body, InitContainerUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if InitContainerUpdate.EnvironmentID == 0 {
		InitContainerUpdate.EnvironmentID = data.EnvironmentID
	}
	if InitContainerUpdate.Name != "" && InitContainerUpdate.Name != data.Name {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("name cannot be changed"))
		return
	}
	if InitContainerUpdate.Image == "" {
		data, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetNamespace(environmentReceived))
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("image not found"))
			return
		}
		img := fmt.Sprintf("%s:%s", data.CIRequest.RepositoryImage.Name, data.CIRequest.RepositoryImage.Tag)
		InitContainerUpdate.Image = img
	}
	InitContainerUpdate.ID = data.ID
	cu, err := initContainerInterface.Update(server.DB, InitContainerUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	err = queue.Publish(constants.UpdateInitContainer, &map[string]interface{}{
		"environment":    environmentReceived,
		"init_container": cu,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "***  InitContainer Update Requested ***",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetInitContainers godoc
// @Summary Get InitContainers
// @Description Get InitContainers
// @Tags InitContainer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Success  200 {array} doc.InitContainer
// @Router /environment/{id}/init-container [get]
func (server *Server) GetInitContainers(w http.ResponseWriter, r *http.Request) {
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
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to create cron job"))
			return
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, environmentReceived.InitContainers)
		return
	}
	initcontainerResponseList := []doc.InitContainer{}
	for _, initcontainer := range environmentReceived.InitContainers {
		initcontainerResponse, err := CreateInitContainerResponse(initcontainer)
		if err != nil {
			log.Error(err)
		}
		initcontainerResponseList = append(initcontainerResponseList, *initcontainerResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    initcontainerResponseList,
		Success: 1,
		Message: "Success",
	})

}

// GetInitContainer godoc
// @Summary Get InitContainer
// @Description Get InitContainer
// @Tags InitContainer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Param cid path int true "Get InitContainer"
// @Success  200 {object} doc.InitContainer
// @Router /environment/{id}/init-container/{cid} [get]
func (server *Server) GetInitContainer(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	cid, err := strconv.ParseUint(vars["cid"], 10, 64)
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
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to create cron job"))
			return
		}
	}

	dataReceived, err := initContainerInterface.Find(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	initcontainerResponse, err := CreateInitContainerResponse(dataReceived)
	if err != nil {
		log.Error(err)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    initcontainerResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteInitContainer godoc
// @Summary Delete InitContainer
// @Description Delete InitContainer
// @Tags InitContainer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Param cid path int true "Delete InitContainer"
// @Success  200
// @Router /environment/{id}/init-container/{cid} [delete]
func (server *Server) DeleteInitContainer(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	cid, err := strconv.ParseUint(vars["cid"], 10, 64)
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
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to create cron job"))
			return
		}
	}

	data := models.InitContainer{}
	err = server.DB.Model(models.InitContainer{}).Where("id = ? and environment_id = ?", cid, pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("InitContainer not found"))
		return
	}
	err = server.DB.Model(&data).Delete(data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = queue.Publish(constants.DeleteInitContainer, &map[string]interface{}{
		"environment":    environmentReceived,
		"init_container": data,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", cid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func CreateInitContainerResponse(initContainer *models.InitContainer) (*doc.InitContainer, error) {
	initContainerResponse := doc.InitContainer{}
	initContainerBytes, _ := json.Marshal(initContainer)
	err := json.Unmarshal(initContainerBytes, &initContainerResponse)
	if err != nil {
		return nil, err
	}
	return &initContainerResponse, nil
}
