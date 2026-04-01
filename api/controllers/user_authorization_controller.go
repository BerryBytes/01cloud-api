package controllers

import (
	"01cloud-api/api/mailer"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"01cloud-api/api/auth"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var authInterface = models.NewAuthorizationRepo()

// CreateAuthorization godoc
// @Summary Create Authorization
// @Description Create Authorization
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 201 {object} doc.Authorization
// @Param pid path int true "Project Id"
// @Router /project/{pid}/user [post]
// @Param eid path int true "Environment Id"
// @Router /environment/{eid}/user [post]
// @Router /authorization [post]
func (server *Server) CreateAuthorization(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := &models.Authorization{}
	err = json.Unmarshal(body, data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err == nil {
		if pid > 0 {
			data.ProjectID = pid
		}
	}
	eid, _ := strconv.ParseUint(vars["eid"], 10, 64)
	var env *models.Environment
	var project *models.Project
	if eid > 0 {
		data.EnvironmentID = eid
		enve := models.Environment{}
		env, err = enve.Find(server.DB, eid)
		if err == nil {
			data.ApplicationID = env.ApplicationID
			data.ProjectID = env.Application.ProjectID
		}
	}

	if env == nil {
		proj := models.Project{}
		proj.OrganizationId = uint64(oid)
		project, err = iproject.Find(server.DB, data.ProjectID)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		if uint64(uid) != project.UserID && !authInterface.IsAdminOfProject(server.DB, uint64(uid), data.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to share this project"))
			return
		}
	} else {
		project = env.Application.Project
		if uint64(uid) != project.UserID && !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), env) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to share this environment"))
			return
		}
	}

	data.Prepare()
	err = data.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	group := &models.Group{}
	user := &models.User{}
	if data.Email != "" {
		user, err = userInterface.FindUserByEmail(server.DB, data.Email)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("user with email \""+data.Email+"\" not found"))
			return
		}
		if oid > 0 {
			org := &models.OrganizationMembers{}
			org.OrganizationID = uint64(oid)
			org.UserID = uint64(user.ID)
			if !orgMember.CheckMember(server.DB, org) {
				responses.ERROR(w, http.StatusNotFound, errors.New("user is not a member of this organization "))
				return
			}
		}
		data.UserID = uint64(user.ID)
	}
	if data.GroupID != 0 {
		group, err = grepo.Find(server.DB, data.GroupID, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
	}
	if data.UserID == project.UserID {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("user is already owner"))
		return
	}
	if authInterface.IsUserExists(server.DB, data) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("user/group already exist"))
		return
	}
	dataCreated, err := authInterface.Save(server.DB, data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	rol, _ := userroleInterface.Find(server.DB, data.UserRoleID)
	name := group.Name
	if data.Email != "" {
		name = user.FirstName + " " + user.LastName
		shareEmail := mailer.SendShareEmail(data.Email, data.ProjectID, project.Name, rol.Name, data.EnvironmentID != 0)
		err = notifications.NotifyEmail(server.NotifyClient, &shareEmail)
		if err != nil {
			logrus.Error("error from notify email :: ", err)
		}
	}
	_, _ = server.SaveActivityWithJson("share", "project", uid, project, nil, env, nil, fmt.Sprintf("to %s as %s role", name, rol.Name))
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	authorizationResponse, err := CreateAuthorizationResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    authorizationResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetAuthorizations godoc
// @Summary Get Authorizations
// @Description Get Authorizations
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.Authorization
// @Router /authorizations [get]
func (server *Server) GetAuthorizations(w http.ResponseWriter, r *http.Request) {
	datas, err := authInterface.FindAll(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	authorizationResponseList := []doc.Authorization{}
	for _, item := range *datas {
		authorizationResponse, err := CreateAuthorizationResponse(&item)
		if err != nil {
			logrus.Error(err)
			continue
		}
		authorizationResponseList = append(authorizationResponseList, *authorizationResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    authorizationResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetUsersInProject godoc
// @Summary Get Users In Project
// @Description Get Users In Project
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Project Id"
// @Success 200 {array} doc.Authorization
// @Router /project/{pid}/users [get]
func (server *Server) GetUsersInProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	datas, err := authInterface.FindAllInProject(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	authorizationResponseList := []doc.Authorization{}
	for _, item := range *datas {
		authorizationResponse, err := CreateAuthorizationResponse(&item)
		if err != nil {
			logrus.Error(err)
			continue
		}
		authorizationResponseList = append(authorizationResponseList, *authorizationResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    authorizationResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetUsersInEnv godoc
// @Summary Get Users In Env
// @Description Get Users In Environment
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param eid path int true "Environment Id"
// @Success 200 {array} doc.Authorization
// @Router /environment/{eid}/users [get]
func (server *Server) GetUsersInEnv(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	eid, err := strconv.ParseUint(vars["eid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	datas, err := authInterface.FindAllInEnv(server.DB, eid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	authorizationResponseList := []doc.Authorization{}
	for _, item := range *datas {
		authorizationResponse, err := CreateAuthorizationResponse(&item)
		if err != nil {
			logrus.Error(err)
			continue
		}
		authorizationResponseList = append(authorizationResponseList, *authorizationResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    authorizationResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetAuthMod godoc
// @Summary Get Auth Mod
// @Description Get Authorization Model
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param model path string true "Model"
// @Param id path int true "Get Auth Model"
// @Success 200 {object} doc.UserRole
// @Router /role/{model}/{id} [get]
func (server *Server) GetAuthMod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	model := vars["model"]
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("user not found"))
		return
	}
	if authInterface.IsAuthorizedOrganization(server.DB, uid, oid) {
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, &map[string]interface{}{"ID": 0, "name": "Org Admin"})
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    &map[string]interface{}{"ID": 0, "name": "Org Admin"},
			Success: 1,
			Message: "Success",
		})
		return
	}
	if model == "project" {
		data := models.Project{}
		data.OrganizationId = uint64(oid)
		project, err := iproject.Find(server.DB, id)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if project.UserID == uint64(uid) {
			responses.JSON(w, http.StatusOK, &map[string]interface{}{"ID": 0, "name": "Owner"})
			return
		}
		au, err := authInterface.GetRoleProject(server.DB, uint64(uid), id)
		if err != nil {
			if user.IsAdmin {
				responses.JSON(w, http.StatusOK, responses.Response{
					Data:    &map[string]interface{}{"ID": 0, "name": "IsAdmin"},
					Success: 1,
					Message: "Success",
				})
				return
			}
			responses.ERROR(w, http.StatusForbidden, err)
			return
		}
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, au.UserRole)
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    au.UserRole,
			Success: 1,
			Message: "Success",
		})
		return
	} else if model == "application" {
		app, err := applicationInterface.Find(server.DB, id)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if app.Project.UserID == uint64(uid) {
			responses.JSON(w, http.StatusOK, &map[string]interface{}{"ID": 0, "name": "Owner"})
			return
		}
		au, err := authInterface.GetRoleApp(server.DB, uint64(uid), app.ProjectID, id)
		if err != nil {
			if user.IsAdmin {
				responses.JSON(w, http.StatusOK, responses.Response{
					Data:    &map[string]interface{}{"ID": 0, "name": "IsAdmin"},
					Success: 1,
					Message: "Success",
				})
				return
			}
			responses.ERROR(w, http.StatusForbidden, err)
			return
		}
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, au.UserRole)
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    au.UserRole,
			Success: 1,
			Message: "Success",
		})
		return
	} else if model == "environment" {
		data := models.Environment{}
		env, err := data.FindWithProApp(server.DB, id)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if env.Application.Project.UserID == uint64(uid) {
			responses.JSON(w, http.StatusOK, &map[string]interface{}{"ID": 0, "name": "Owner"})
			return
		}
		au, err := authInterface.GetRoleEnv(server.DB, uint64(uid), env.Application.ProjectID, env.ApplicationID, id)
		if err != nil {
			if user.IsAdmin {
				responses.JSON(w, http.StatusOK, responses.Response{
					Data:    &map[string]interface{}{"id": 0, "name": "IsAdmin"},
					Success: 1,
					Message: "Success",
				})
				return
			}
			responses.ERROR(w, http.StatusForbidden, err)
			return
		}
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, au.UserRole)
			return
		}
		userroleResponse, err := CreateUserRoleResponse(au.UserRole)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    userroleResponse,
			Success: 1,
			Message: "Success",
		})
		return
	}
	responses.ERROR(w, http.StatusNotFound, errors.New("model not found"))
}

// GetAuthorization godoc
// @Summary Get Authorization
// @Description Get Authorization
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Authorization"
// @Success 200 {object} doc.Authorization
// @Router /authorization/{id} [get]
func (server *Server) GetAuthorization(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := authInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	authorizationResponse, err := CreateAuthorizationResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    authorizationResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateAuthorization godoc
// @Summary Update Authorization
// @Description Update Authorization
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Project Id"
// @Param eid path int true "User Id"
// @Param id path int true "Update Authorization"
// @Param body body doc.Authorization true "Update Authorization"
// @Success 200 {object} doc.Authorization
// @Router /project/{pid}/user/{id} [put]
// @Router /environment/{eid}/user/{id} [put]
// @Router /authorization/{id} [put]
func (server *Server) UpdateAuthorization(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Authorization{}
	err = server.DB.
		Model(models.Authorization{}).
		Preload("User").
		Preload("Project").
		//Preload("Application").
		Preload("Environment").
		Preload("Environment.Application").
		Preload("UserRole").
		//Preload("Group").
		Where("id = ?", pid).
		Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("authorization not found"))
		return
	}
	if data.Project.UserID != uint64(uid) {
		if data.Environment != nil && !authInterface.IsAdminOfEnvironment(server.DB, uint64(uid), data.Environment) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorizad to update this user"))
			return
		} else if !authInterface.IsAdminOfProject(server.DB, uint64(uid), data.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorizad to update this"))
			return
		}
		if uint64(uid) == data.UserID {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorizad to update your own role"))
			return
		}
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.Authorization{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	//dataUpdate.Prepare()
	//err = dataUpdate.Validate()
	//if err != nil {
	//	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	//	return
	//}
	dataUpdate.UpdatedAt = time.Now()
	dataUpdate.ID = data.ID
	dataUpdate.Email = data.Email
	dataUpdated, err := authInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if data.Email != "" && dataUpdate.UserRoleID != data.UserRoleID {
		rol, _ := userroleInterface.Find(server.DB, dataUpdate.UserRoleID)
		updateEmail := mailer.SendUpdateEmail(data.Email, data.ProjectID, data.Project.Name, rol.Name, data.EnvironmentID != 0, data.UserRole.Name)
		err = notifications.NotifyEmail(server.NotifyClient, &updateEmail)
		if err != nil {
			logrus.Error("error from notify email :: ", err)
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	authorizationResponse, err := CreateAuthorizationResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    authorizationResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteAuthorization godoc
// @Summary Delete Authorization
// @Description Delete Authorization
// @Tags Authorization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Project Id"
// @Param eid path int true "User Id"
// @Param id path int true "Delete Authorization"
// @Success 200
// @Router /project/{pid}/user/{id} [delete]
// @Router /environment/{eid}/user/{id} [delete]
// @Router /authorization/{id} [delete]
func (server *Server) DeleteAuthorization(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Authorization{}
	err = server.DB.
		Model(models.Authorization{}).
		Preload("User").
		Preload("Project").
		//Preload("Application").
		Preload("Environment").
		Preload("Environment.Application").
		Preload("Group").
		Where("id = ?", pid).
		Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("authorization not found"))
		return
	}
	if data.Project.UserID != uint64(uid) && data.UserID != uint64(uid) {
		if data.Environment != nil && !authInterface.IsAdminOfEnvironment(server.DB, uint64(uid), data.Environment) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorizad to remove this user"))
			return
		} else if !authInterface.IsAdminOfProject(server.DB, uint64(uid), data.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorizad to remove this user"))
			return
		}
	}
	user := data.User
	group := data.Group
	_, err = authInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if data.GroupID > 0 {
		_, _ = server.SaveActivityWithJson("unshare", "project", uid, data.Project, data.Application, data.Environment, nil, fmt.Sprintf("to %s", group.Name))
		//	_, _ = server.SaveActivityWithJson("delete", "user", uid, data.Project, nil, nil, "")
	} else {
		_, _ = server.SaveActivityWithJson("unshare", "project", uid, data.Project, data.Application, data.Environment, nil, fmt.Sprintf("to %s %s", user.FirstName, user.LastName))

		//	_, _ = server.SaveActivityWithJson("delete", "user", uid, data.Project, nil, nil, "")
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

func CreateAuthorizationResponse(authorization *models.Authorization) (*doc.Authorization, error) {
	authorizationResponse := doc.Authorization{}
	authorizationBytes, _ := json.Marshal(authorization)
	err := json.Unmarshal(authorizationBytes, &authorizationResponse)
	if err != nil {
		return nil, err
	}
	return &authorizationResponse, nil
}
