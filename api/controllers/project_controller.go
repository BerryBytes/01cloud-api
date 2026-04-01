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
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"
	"01cloud-api/api/utils/prometheus"

	"github.com/gorilla/mux"
)

func getErrorMessage(field, message string) error {
	return fmt.Errorf("project error: field = %s desc = %s", field, message)
}

var iproject = models.NewIProject()

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project with the input paylod
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Project true "Create Project"
// @Success 201 {object} doc.Project
// @Router /project [post]
func (server *Server) CreateProject(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, getErrorMessage("all", err.Error()))
		return
	}
	project := &models.Project{}

	err = json.Unmarshal(body, &project)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, getErrorMessage("all", err.Error()))
		return
	}
	project.Prepare()
	err = project.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, getErrorMessage("all", err.Error()))
		return
	}
	subscription, err := subscriptionInterface.Find(server.DB, project.SubscriptionID)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, getErrorMessage("subscription", "Invalid subscription"))
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, getErrorMessage("user", "User not found"))
		return
	}
	if subscription.Validity > 0 && subscription.Price == 0 {
		if user.UsedDemo {
			responses.ERROR(w, http.StatusNotFound, errors.New("demo project already created"))
			return
		} else {
			user.UsedDemo = true
			_, err = userInterface.UpdateAUser(server.DB, uint(project.UserID), user)
			if err != nil {
				responses.ERROR(w, http.StatusNotFound, err)
				return
			}
		}
	}

	project.Subscription = subscription
	project.User = user
	project.UserID = uint64(uid)
	project.OrganizationId = uint64(oid)
	if oid == 0 {
		if !paymentInterface.HasUserBalance(server.DB, uid) {
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
	ok := iproject.IsNameExists(server.DB, uid, project.Name, project)
	if ok {
		responses.ERROR(w, http.StatusUnprocessableEntity, getErrorMessage("name", "Name already exists"))
		return
	}

	var userVariables []map[string]string
	err = json.Unmarshal(project.Variables.RawMessage, &userVariables)
	if err == nil && !models.ValidateUserVariable(userVariables) {
		responses.ERROR(w, http.StatusUnprocessableEntity, getErrorMessage("variables", "You are not allowed to create duplicate variable"))
		return
	}

	if project.OrganizationId > 0 {
		org, err := orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, getErrorMessage("organization", err.Error()))
			return
		}
		project.Organization = org
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized to create project"))
			return
		}
	}
	projectList, err := iproject.FindUserProjectOnly(server.DB, uid, uint(project.OrganizationId))
	if err != nil {
		log.Error(err)
	}
	if err = helper.CheckUserQuota(user, "p", len(projectList)); err != nil {
		log.Error("quota error :: ", err)
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	projectCreated, err := iproject.Save(server.DB, project)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	notificationInfo := notifications.BasicInfoNotification{
		OrganizationID: int64(project.OrganizationId),
		ProjectID:      int64(project.ID),
		ProjectName:    project.Name,
	}
	if project.OrganizationId > 0 {
		notificationInfo.OrganizationName = project.Organization.Name
	}
	_, _ = server.SaveActivityWithJson("create", "project", uid, projectCreated, nil, nil, nil, "")
	notifyProject(server.NotifyClient, notificationInfo, "created", "created project "+project.Name, "info", int64(user.ID))
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, projectCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, projectCreated)
		return
	}
	projectResponse, err := CreateProjectResponse(projectCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	key := fmt.Sprintf("projects-list-%d-%d", uid, oid)
	server.Cache.Delete(key)
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    projectResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetProjects godoc
// @Summary Get Projects
// @Description Get list projects from token
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.Project
// @Router /projects [get]
func (server *Server) GetProjects(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	project := models.Project{}
	project.OrganizationId = uint64(oid)
	key := fmt.Sprintf("projects-list-%d-%d", uid, oid)
	var value interface{}
	if ok := server.Cache.Get(key, &value); ok {
		responses.JSON(w, http.StatusOK, value)
		return
	}
	projects, err := iproject.FindAllByUser(server.DB, uid, &project)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, projects)
		return
	}
	projectsResponseList := []doc.Project{}
	for _, item := range projects {
		projectsResponse, err := CreateProjectResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		projectsResponseList = append(projectsResponseList, *projectsResponse)
	}
	response := responses.Response{
		Data:    projectsResponseList,
		Success: 1,
		Message: "Success",
	}
	server.Cache.Set(key, response)
	responses.JSON(w, http.StatusOK, response)
}

//func (server *Server) GetProjectByUser(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	userId, err := strconv.ParseUint(vars["id"], 10, 64)
//	if err != nil {
//		responses.ERROR(w, http.StatusBadRequest, err)
//		return
//	}
//	project := models.Project{}
//	projects, err := project.FindAllByUser(server.DB, uint(userId))
//	if err != nil {
//		responses.ERROR(w, http.StatusInternalServerError, err)
//		return
//	}
//	responses.JSON(w, http.StatusOK, projects)
//}

// GetProjectOfUserOnly godoc
// @Summary Get Projects by user
// @Description Get list projects from by user id. Only available for admin user
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User id"
// @Success 200 {array} doc.Project
// @Router /user/{id}/projects [get]
func (server *Server) GetProjectOfUserOnly(w http.ResponseWriter, r *http.Request) {
	_, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	userId, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	project := models.Project{}
	project.OrganizationId = uint64(oid)
	projects, err := iproject.FindUserProjectOnly(server.DB, uint(userId), uint(project.OrganizationId))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, projects)
		return
	}
	resp := []map[string]interface{}{}
	for _, item := range projects {
		projectsResponse, err := CreateProjectResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		count := applicationInterface.FindAllApplicationCountByProject(server.DB, uint64(projectsResponse.ID), uint64(projectsResponse.UserID))
		js, _ := helper.ToJson(projectsResponse)
		js["apps_count"] = count
		resp = append(resp, js)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resp,
		Success: 1,
		Message: "Success",
	})
}

// GetProject godoc
// @Summary Get Project by id
// @Description Get Project by id from token
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Success 200 {object} doc.Project
// @Router /project/{id} [get]
func (server *Server) GetProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	// key := fmt.Sprintf("project-%d", pid)
	// var value interface{}
	// if ok := server.Cache.Get(key, &value); ok {
	// 	responses.JSON(w, http.StatusOK, value)
	// 	return
	// }
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	projectReceived, err := iproject.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uid == uint(projectReceived.UserID) || user.IsAdmin || authInterface.IsAuthorizedOrganization(server.DB, uid, oid) ||
		authInterface.IsAuthorizedProject(server.DB, uint64(uid), pid) {
		count := applicationInterface.FindAllApplicationCountByProject(server.DB, uint64(projectReceived.ID), uint64(projectReceived.UserID))
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusCreated, projectReceived)
			return
		}
		projectResponse, err := CreateProjectResponse(projectReceived)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return

		}
		js, _ := helper.ToJson(projectResponse)
		js["apps_count"] = count
		response := responses.Response{
			Data:    js,
			Success: 1,
			Message: "Success",
		}
		// server.Cache.Set(key, response)
		responses.JSON(w, http.StatusOK, response)
	} else {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
	}
}

func (server *Server) GetProjectByOrganizationForAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	oid, err := strconv.ParseUint(vars["oid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataList, err := iproject.FindProject(server.DB, uint(oid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataList)
		return
	}
	projectsResponseList := []doc.Project{}
	for _, item := range *dataList {
		projectsResponse, err := CreateProjectResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		projectsResponseList = append(projectsResponseList, *projectsResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    projectsResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetGlobalVariables godoc
// @Summary Get variables by project id
// @Description Get Global Project Variables by project id from token
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Project id"
// @Success 200 {object} doc.Project
// @Router /project/{pid}/variables [get]
func (server *Server) GetGlobalVariables(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
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
	projectReceived, err := iproject.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uid == uint(projectReceived.UserID) || user.IsAdmin || authInterface.IsAuthorizedOrganization(server.DB, uid, oid) ||
		authInterface.IsAuthorizedProject(server.DB, uint64(uid), pid) {
		var variables []map[string]interface{}
		err := json.Unmarshal(projectReceived.Variables.RawMessage, &variables)
		if err != nil {
			responses.JSON(w, http.StatusOK, []map[string]interface{}{})
			return
		}
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, variables)
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    variables,
			Success: 1,
			Message: "Success",
		})
	} else {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
	}
}

// UpdateProject godoc
// @Summary Update a Project
// @Description Update a Project with the input payload
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Param body body doc.Project true "Update Project"
// @Success 200 {object} doc.Project
// @Router /project/{id} [put]
func (server *Server) UpdateProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
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
	project := models.Project{}
	err = server.DB.Model(models.Project{}).
		Preload("User").
		Preload("Subscription").
		Where("id = ?", pid).
		Where("organization_id = ?", oid).
		Take(&project).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("project not found"))
		return
	}
	if uid != uint(project.UserID) && !user.IsAdmin {
		if oid > 0 {
			if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
				return
			}
		} else {
			if !authInterface.IsWriteAuthorizedProject(server.DB, uint64(uid), pid) {
				responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
				return
			}
		}
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	projectUpdate := models.Project{}
	err = json.Unmarshal(body, &projectUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	projectUpdate.Prepare()
	projectUpdate.Subscription = project.Subscription
	user.Password = ""
	projectUpdate.User = user

	if projectUpdate.Variables.RawMessage != nil {
		var userVariables []map[string]string
		err = json.Unmarshal(projectUpdate.Variables.RawMessage, &userVariables)
		if err == nil && !models.ValidateUserVariable(userVariables) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("you are not allowed to create duplicate variable"))
			return
		}
	}
	if projectUpdate.Name != "" {
		if projectUpdate.Name != project.Name && iproject.IsNameExists(server.DB, uid, projectUpdate.Name, &project) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("name already exists"))
			return
		}
	}

	extras := ""
	if project.Name != projectUpdate.Name && projectUpdate.Name != "" {
		extras += fmt.Sprintf("name from '%s' to '%s' ", project.Name, projectUpdate.Name)
	}
	if project.Image != projectUpdate.Image && projectUpdate.Image != "" {
		extras += "successfully updated project icon"
	}
	if project.SubscriptionID != projectUpdate.SubscriptionID && projectUpdate.SubscriptionID != 0 {
		subscription, err := subscriptionInterface.Find(server.DB, projectUpdate.SubscriptionID)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid subscription"))
			return
		}
		extras += fmt.Sprintf("subscription from '%s' to '%s' ", project.Subscription.Name, subscription.Name)
	}

	projectUpdate.ID = project.ID
	projectUpdate.OrganizationId = project.OrganizationId
	err = projectUpdate.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	projectUpdated, err := iproject.Update(server.DB, &projectUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	notificationInfo := notifications.BasicInfoNotification{
		OrganizationID: int64(project.OrganizationId),
		ProjectID:      int64(project.ID),
		ProjectName:    project.Name,
	}
	if project.OrganizationId > 0 {
		org, err := orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, getErrorMessage("organization", err.Error()))
			return
		}
		notificationInfo.OrganizationName = org.Name
	}
	notifyProject(server.NotifyClient, notificationInfo, "updated", extras, "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("update", "project", uid, &projectUpdate, nil, nil, nil, extras)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, projectUpdated)
		return
	}
	projectResponse, err := CreateProjectResponse(projectUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	key := fmt.Sprintf("projects-list-%d-%d", uid, oid)
	server.Cache.Delete(key)
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    projectResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteProject godoc
// @Summary Delete a Project
// @Description Delete a Project with the input payload
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Success 204 {object} doc.Project
// @Router /project/{id} [delete]
func (server *Server) DeleteProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	project := models.Project{}
	err = server.DB.Model(models.Project{}).
		Where("id = ?", pid).
		Where("organization_id = ?", oid).
		Take(&project).Error

	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}
	if uid != uint(project.UserID) {
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
			!authInterface.IsAdminOfProject(server.DB, uint64(uid), pid) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not unauthorized to delete this project"))
			return
		}
	}
	count := applicationInterface.FindAllApplicationCountByProject(server.DB, pid, uint64(uid))
	if count > 0 {
		responses.ERROR(w, http.StatusPreconditionFailed, errors.New("cannot delete non-empty project"))
		return
	}
	if project.Image != "" {
		path := strings.ReplaceAll(project.Image, os.Getenv("BASE_URL"), ".")
		_ = os.Remove(path)
	}
	datas, err := loadbalancerInterface.FindAllByProject(server.DB, pid)
	if err == nil {
		for _, lb := range *datas {
			_ = queue.Publish(constants.DestroyLoadBalancer, lb)
			_, _ = loadbalancerInterface.Delete(server.DB, uint64(lb.ID))
		}
	}
	notificationInfo := notifications.BasicInfoNotification{
		OrganizationID: int64(project.OrganizationId),
		ProjectID:      int64(project.ID),
		ProjectName:    project.Name,
	}
	if project.OrganizationId > 0 {
		org, err := orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, getErrorMessage("organization", err.Error()))
			return
		}
		notificationInfo.OrganizationName = org.Name
	}
	notifyProject(server.NotifyClient, notificationInfo, "deleted", fmt.Sprintf("project %s deleted", project.Name), "info", int64(uid))
	_, _ = server.SaveActivityWithJson("delete", "project", uid, &project, nil, nil, nil, "")
	time.Sleep(time.Duration(time.Second))
	_, err = iproject.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	key := fmt.Sprintf("projects-list-%d-%d", uid, oid)
	server.Cache.Delete(key)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetResourceUsed godoc
// @Summary Get resource used by a Project
// @Description Get resource used by project
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Success 200 {object} doc.ResourceUsedResponse
// @Router /project/{id}/resource [get]
func (server *Server) GetResourceUsed(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	key := fmt.Sprintf("project-resource-used-%d", pid)
	var value interface{}
	if ok := server.Cache.Get(key, &value); ok {
		responses.JSON(w, http.StatusOK, value)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	project, err := iproject.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if project.UserID != uint64(uid) && !user.IsAdmin &&
		!authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
		!authInterface.IsAuthorizedProject(server.DB, uint64(uid), pid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("permission not granted"))
		return
	}
	count := applicationInterface.FindAllApplicationCountByProject(server.DB, pid, uint64(uid))
	environment := models.Environment{}
	usage, _ := environment.GetUsedResource(server.DB, pid, false)
	totalTransmitValue, totalReceiveValue, err := server.GetDataTransfer(pid)
	if err != nil {
		log.Error("error occured getting data transfer rate :: ", err)
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	totalCiBuild, err := server.GetTotalCiBuild(pid)
	if err != nil {
		log.Error("error occured getting ci build :: ", err)
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	usage.Apps = count
	usage.DataTransfer = &map[string]interface{}{
		"data_transfer": map[string]interface{}{
			"receive":  totalReceiveValue,
			"transmit": totalTransmitValue,
		},
	}
	usage.TotalCiBuild = uint64(totalCiBuild)

	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, usage)
		return
	}
	response := responses.Response{
		Data:    usage,
		Success: 1,
		Message: "Success",
	}
	server.Cache.Set(key, response)
	responses.JSON(w, http.StatusOK, response)
}

func (server *Server) GetDataTransferUsed(w http.ResponseWriter, r *http.Request) {
	projects, err := iproject.FindAllProjects(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	for _, project := range *projects {
		bandwidthUsed, err := server.getBandwidth(uint64(project.ID))
		if err != nil {
			continue
		}
		mp := map[string]interface{}{
			"data_transfer": fmt.Sprint(*bandwidthUsed),
		}
		usage, err := json.Marshal(mp)
		if err != nil {
			continue
		}
		project.UsageQuota.RawMessage = usage
		_, err = iproject.UpdateUsageQuota(server.DB, &project)
		if err != nil {
			continue
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "Data transfer updated successfully")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func notifyProject(conn *grpc.ClientConn, info notifications.BasicInfoNotification, _, body, types string, triggeredBy int64) {
	_, err := PublishNotificationBase(conn, &notifications.Notification{
		ApplicationName:         info.ApplicationName,
		ApplicationID:           info.ApplicationID,
		EnvironmentName:         info.EnvironmentName,
		EnvironmentID:           info.EnvironmentID,
		EnvironmentResourceName: info.EnvironmentResourceName,
		ProjectName:             info.ProjectName,
		ProjectID:               info.ProjectID,
		EnvironmentResourceID:   info.EnvironmentResourceID,
		OrganizationName:        info.OrganizationName,
		OrganizationID:          info.OrganizationID,
		TriggeredBy:             triggeredBy,
		Type:                    types,
		Body:                    body,
		Scope:                   "project",
		SendBy:                  "api",
	})
	if err != nil {
		log.Error(err)
	}
}

func (server *Server) GetDataTransfer(pid uint64) (float64, float64, error) {
	totalTransmitValue, totalReceiveValue := 0.0, 0.0
	data := &models.Environment{}
	envs, err := data.FindEnvironmentsByProject(server.DB, pid, "", "")
	if err != nil {
		return totalTransmitValue, totalReceiveValue, err
	}
	log.Debugf("environment length :: %d", len(envs))
	for _, env := range envs {
		transmitValue, receiveValue, err := prometheus.GetDataTransfer(env)
		if err != nil {
			return totalTransmitValue, totalReceiveValue, err
		}
		totalTransmitValue += transmitValue
		totalReceiveValue += receiveValue
		log.Debugf("receive rate :: %f transmit rate :: %f ", totalReceiveValue, totalTransmitValue)
	}
	return totalTransmitValue, totalReceiveValue, nil
}

func (server *Server) GetTotalCiBuild(pid uint64) (int, error) {
	totalCiBuild := 0
	data := &models.Environment{}
	envs, err := data.FindEnvironmentsByProject(server.DB, pid, "", "")
	if err != nil {
		return totalCiBuild, err
	}
	log.Debugf("environment length :: %d", len(envs))
	for _, env := range envs {
		if env.ServiceType == 1 {
			countCiBuild, err := server.StoreClient.CIWOrkflow().GetWorkflowCount(helper.GetCiNamespace(env), "", "")
			if err != nil {
				return totalCiBuild, err
			}
			totalCiBuild += countCiBuild
			log.Debugf("cicount ::%d", totalCiBuild)
		}
	}
	return totalCiBuild, nil
}

func (server *Server) getBandwidth(pid uint64) (*float64, error) {
	totalTransmitValue, totalReceiveValue, err := server.GetDataTransfer(pid)
	if err != nil {
		return nil, err
	}
	bandwidthUsed := ((totalTransmitValue + totalReceiveValue) / (1024 * 1024))
	return &bandwidthUsed, nil
}

func CreateProjectResponse(project *models.Project) (*doc.Project, error) {
	projectResponse := doc.Project{}
	projectBytes, _ := json.Marshal(project)
	err := json.Unmarshal(projectBytes, &projectResponse)
	if err != nil {
		return nil, err
	}
	return &projectResponse, nil

}

func (server *Server) TerminateProject(w http.ResponseWriter, r *http.Request) {
	config, err := helper.ReadAdminConfigFile()
	if err != nil {
		log.Error("Error occurs reading config file :: ", err)
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if config.ProjectThresholdDays == 0 {
		config.ProjectThresholdDays = 21
	}
	users, err := userInterface.FindAllUsers(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	for _, user := range *users {
		projects, err := iproject.FindAllDeactivatedProjectsByThresholdDaysByUser(server.DB, user.ID, config.ProjectThresholdDays)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		for _, project := range *projects {
			applications, err := applicationInterface.FindAllApplicationByInactiveProject(server.DB, uint64(project.ID))
			if err != nil {
				log.Error("Error quering application by project :: ", err)
				continue
			}
			for _, application := range *applications {
				env := &models.Environment{}
				environments, err := env.FindAllByApplicationAdmin(server.DB, uint64(application.ID))
				if err != nil {
					log.Error("Error quering application by project :: ", err)
					continue
				}
				server.DeleteEnvironments(environments, r)
				time.Sleep(time.Second)
				_, err = applicationInterface.Delete(server.DB, uint64(application.ID))
				if err != nil {
					log.Error("error while deleting application", err)
					continue
				}
			}
			_, err = iproject.Delete(server.DB, uint64(project.ID))
			if err != nil {
				log.Error("Error occurs deleting project :: ", err)
				continue
			}
		}
		mailer.SendProjectTerminationEmail(user.Email, len(*projects), config.ProjectThresholdDays)
	}

	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) DeleteEnvironments(environments *[]models.Environment, r *http.Request) {
	for _, environment := range *environments {
		environment.Action = "Deleting"
		if environment.Application.ServiceType == 4 && environment.OperatorPayload.RawMessage != nil {
			env, _ := environment.ToJson()
			packageName := environment.Application.OperatorPackageName
			if environment.Application.Cluster != nil {
				opReq, err := operatorRequestRepo.FindByNameAndClusterID(server.DB, environment.Application.Cluster.ClusterRequestID, packageName)
				if err == nil {
					env["operators"] = map[string]interface{}{
						"package_name":    helper.GetOperatorNamespace(opReq),
						"global_operator": opReq.GlobalOperator,
					}
				}
			}
			err := queue.Publish(constants.UnInstallOperatorApp, env)
			if err != nil {
				log.Error("Error occurs while publishing message to UninstallOperatorApp ::", err)
			}
		} else {
			err := queue.Publish(constants.DestroyRelease, environment)
			if err != nil {
				log.Error("Error occurs while publishing message to DestroyRelease ::", err)
			}
		}

		if environment.ServiceType == 1 {
			if environment.GitBranch != "" {
				repo, err := server.GetRepository(uint(environment.Application.OwnerId), environment.Application)
				if err == nil {
					environment.GitUrl = repo.CloneURL
					environment.GitRepository = repo
					conf, err := ciConfigRepo.FindByEnvironment(server.DB, environment.ID)
					if err == nil {
						if len(conf.HookId) > 0 {
							err = server.DeleteWebhook(&environment, conf.HookId, r)
							if err != nil {
								log.Error("Error occurs while deleting webhook ::", err)
							}
						}
					} else {
						log.Error("error while finding env :: ", err)
					}
				}
			}
		}
		_, err := environment.Delete(server.DB, uint64(environment.ID))
		if err != nil {
			return
		}
	}
}

func (server *Server) CheckProjectValidity(w http.ResponseWriter, r *http.Request) {

	if r.URL.Query().Get("server-token") != os.Getenv("API_SECRET") {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
		return
	}
	projects, err := iproject.FindAllProjects(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	for _, project := range *projects {
		if project.Subscription.Validity > 0 && project.Subscription.Price == 0 {
			t := time.Now()
			difference := t.Sub(project.CreatedAt)
			totalDays := int(difference.Hours() / 24)
			if totalDays > int(project.Subscription.Validity) {
				err = server.projectActivation(project.ID, uint(project.UserID), false)
				if err != nil {
					responses.ERROR(w, http.StatusInternalServerError, err)
					return
				}
				notifyInfo := notifications.BasicInfoNotification{
					ProjectName: project.Name,
					ProjectID:   int64(project.ID),
				}
				notify(server.NotifyClient, notifyInfo, "update", fmt.Sprintf("project %s %s", project.Name, "deactivated"), "info", int64(project.UserID))
				err = notifications.NotifyEmail(server.NotifyClient, &notifications.SendEmailMessage{
					User:    []string{project.User.Email},
					Subject: "Project " + project.Name + " " + "deactivated",
					Body:    fmt.Sprintf("Project %s is %s.", project.Name, "deactivated"),
				})
				if err != nil {
					log.Error("error from notify email :: ", err)
				}
				_, _ = server.SaveActivityWithJson("update", "project", 0, &project, nil, nil, nil, "")
			}
		}
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
