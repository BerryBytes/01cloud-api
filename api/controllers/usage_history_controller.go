package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// ActiveDeactiveProject godoc
// @Summary ActiveDeactiveProject a Project
// @Description Operation a Project with the input payload
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Param body body doc.Project true "ActiveDeactiveProject Project"
// @Success 200 {object} doc.Project
// @Router /project/{id} [post]
func (server *Server) ActiveDeactiveProject(w http.ResponseWriter, r *http.Request) {
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
	project, err := iproject.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("project not found"))
		return
	}
	if oid == 0 {
		if !paymentInterface.HasUserBalance(server.DB, uint(project.UserID)) {
			responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to perform action due to remaining balance"))
			return
		}
	} else {
		org, err := orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if !paymentInterface.HasUserBalance(server.DB, uint(org.UserID)) {
			responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to perform action due to remaining balance"))
			return
		}
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

	useHistory := &models.UsageHistory{}
	err = json.Unmarshal(body, useHistory)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	action := "deactivated"
	activityAction := "Deactivated"
	if useHistory.IsOperational {
		action = "activated"
		activityAction = "Activated"
	}
	if project.Active == useHistory.IsOperational {
		responses.ERROR(w, http.StatusInternalServerError, fmt.Errorf("project is already %v", action))
		return
	}
	useHistory.UserId = user.ID
	useHistory.ProjectId = project.ID
	err = server.projectActivation(useHistory.ProjectId, useHistory.UserId, useHistory.IsOperational)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	err = useHistory.ProjectUsageHistory(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ProjectName: project.Name,
		ProjectID:   int64(project.ID),
	}
	if project.OrganizationId != 0 {
		notifyInfo.OrganizationID = int64(project.Organization.ID)
		notifyInfo.OrganizationName = project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "update", fmt.Sprintf("project %s %s", project.Name, action), "info", int64(uid))
	err = notifications.NotifyEmail(server.NotifyClient, &notifications.SendEmailMessage{
		User:    []string{project.User.Email},
		Subject: "Project " + project.Name + " " + action,
		Body:    fmt.Sprintf("Your %s project is successfully %s.", project.Name, action),
	})
	if err != nil {
		log.Error("error from notify email :: ", err)
	}
	_, _ = server.SaveActivityWithJson(activityAction, "project", 0, project, nil, nil, nil, "")
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// ActiveDeactiveProject godoc
// @Summary ActiveDeactiveProject a Project
// @Description Operation a Project with the input payload
// @Tags Project
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Param body body doc.Project true "ActiveDeactiveProject Project"
// @Success 200 {object} doc.Project
// @Router /project/{id} [post]
func (server *Server) DeactiveProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	project, err := iproject.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("project not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	useHistory := &models.UsageHistory{}
	err = json.Unmarshal(body, useHistory)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	action := "deactivated"
	if useHistory.IsOperational {
		action = "activated"
	}
	if project.Active == useHistory.IsOperational {
		responses.ERROR(w, http.StatusInternalServerError, fmt.Errorf("project is already %v", action))
		return
	}
	useHistory.UserId = uint(project.UserID)
	useHistory.ProjectId = project.ID
	err = server.projectActivation(useHistory.ProjectId, useHistory.UserId, useHistory.IsOperational)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	err = useHistory.ProjectUsageHistory(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ProjectName: project.Name,
		ProjectID:   int64(project.ID),
	}
	if project.OrganizationId != 0 {
		notifyInfo.OrganizationID = int64(project.Organization.ID)
		notifyInfo.OrganizationName = project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "update", fmt.Sprintf("project %s %s", project.Name, action), "info", int64(project.UserID))
	err = notifications.NotifyEmail(server.NotifyClient, &notifications.SendEmailMessage{
		User:    []string{project.User.Email},
		Subject: "Project " + project.Name + " " + action,
		Body:    fmt.Sprintf("Project %s is %s.", project.Name, action),
	})
	if err != nil {
		log.Error("error from notify email :: ", err)
	}
	_, _ = server.SaveActivityWithJson("update", "project", 0, project, nil, nil, nil, "")
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) projectActivation(projectId, userId uint, isOperational bool) error {
	project := &models.Project{}
	project.ID = projectId
	env := &models.Environment{}
	err := iproject.ChangeIsActive(server.DB, project, isOperational)
	if err != nil {
		return err
	}
	apps, err := applicationInterface.FindAllApplicationByProjectForAdmin(server.DB, uint64(projectId), uint64(userId))
	if err == nil {
		for _, app := range *apps {
			err = applicationInterface.ChangeIsActive(server.DB, isOperational, app.ID)
			if err != nil {
				continue
			}
			if isOperational {
				continue
			}
			envs, err := env.FindAllByApplication(server.DB, uint64(app.ID), true, uint64(userId))
			if err == nil {
				for _, env := range envs {
					err = env.ChangeIsActive(server.DB, isOperational)
					if err == nil {
						env.Active = isOperational
						err := queue.Publish(constants.ActiveDeactiveEnvironment, env)
						if err != nil {
							log.Error(err)
						}
					}
				}
			}
		}
	}
	return nil
}

func (server *Server) deleteAllProjectResource(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	project, err := iproject.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("project not found"))
		return
	}
	applications, err := applicationInterface.FindAllApplicationByInactiveProject(server.DB, uint64(project.ID))
	if err != nil {
		log.Error("error quering application by project :: ", err)
		return
	}
	for _, application := range *applications {
		env := &models.Environment{}
		environments, err := env.FindAllByApplicationAdmin(server.DB, uint64(application.ID))
		if err != nil {
			log.Error("error quering application by project :: ", err)
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
		log.Error("error occurs deleting project :: ", err)
		return
	}
	userGotten, err := userInterface.FindUserByID(server.DB, uint(project.UserID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = userGotten.ChangeActive(server.DB, false)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	server.ActivateDeactiveOrganization(userGotten.ID, false)
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
