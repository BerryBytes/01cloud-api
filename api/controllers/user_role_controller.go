package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var userroleInterface = models.NewUserRole()

// CreateUserRole godoc
// @Summary Create User Role
// @Description Create User Role
// @Tags UserRole
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.UserRole true "Create User Role"
// @Success 200 {object} doc.UserRole
// @Router /role [post]
func (server *Server) CreateUserRole(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.UserRole{}
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
	dataCreated, err := userroleInterface.Save(server.DB, &data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	userRoleResponse, err := CreateUserRoleResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    userRoleResponse,
		Success: 1,
		Message: "Success",
	})

}

// GetUserRoles godoc
// @Summary Get User Roles
// @Description Get User Roles
// @Tags UserRole
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.UserRole
// @Router /roles [get]
func (server *Server) GetUserRoles(w http.ResponseWriter, r *http.Request) {
	datas, err := userroleInterface.FindAll(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	userRoleResponse := []doc.UserRole{}
	for _, item := range *datas {
		response, err := CreateUserRoleResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		userRoleResponse = append(userRoleResponse, *response)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userRoleResponse,
		Success: 1,
		Message: "Success",
	})

}

// GetUserRole godoc
// @Summary Get User Role
// @Description Get User Role
// @Tags UserRole
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get User Role"
// @Success 200 {object} doc.UserRole
// @Router /role/{id} [get]
func (server *Server) GetUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := userroleInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	userRoleResponse, err := CreateUserRoleResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userRoleResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateUserRole godoc
// @Summary Update User Role
// @Description Update User Role
// @Tags UserRole
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Update User Role"
// @Param body body doc.UserRole true "Update User Role"
// @Success 200 {object} doc.UserRole
// @Router /role/{id} [put]
func (server *Server) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := userroleInterface.Find(server.DB, pid)
	//err = server.DB.Model(models.UserRole{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("UserRole not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.UserRole{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate.Prepare()
	//err = dataUpdate.Validate()
	//if err != nil {
	//	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	//	return
	//}
	dataUpdate.ID = data.ID
	dataUpdated, err := userroleInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	userRoleResponse, err := CreateUserRoleResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    userRoleResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteUserRole godoc
// @Summary Delete User Role
// @Description Delete User Role
// @Tags UserRole
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Delete User Role"
// @Success 200
// @Router /role/{id} [put]
func (server *Server) DeleteUserRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	//err = server.DB.Model(models.UserRole{}).Where("id = ?", pid).Take(&data).Error
	_, err = userroleInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}
	_, err = userroleInterface.Delete(server.DB, pid)
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

func CreateUserRoleResponse(userRole *models.UserRole) (*doc.UserRole, error) {
	userRoleResponse := doc.UserRole{}
	userRoleBytes, _ := json.Marshal(userRole)
	err := json.Unmarshal(userRoleBytes, &userRoleResponse)
	if err != nil {
		return nil, err
	}
	return &userRoleResponse, nil
}
