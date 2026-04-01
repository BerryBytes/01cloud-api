package controllers

import (
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// ScheduleEnvironment godoc
// @Summary Schedule Environment
// @Description Schedule Environment
// @Tags Schedule
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path string true "Environment Id"
// @Param body body models.RequestSchedule true "Schedule Environment"
// @Success 200 {string} string
// @Router /environment/{id}/schedule [post]
func (server *Server) ScheduleEnvironment(w http.ResponseWriter, r *http.Request) {
	environmentReceived, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := CheckPermission(server.DB, r, WRITE, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input := &models.RequestSchedule{}
	err = json.Unmarshal(body, input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input.CreatedAt = time.Now()
	err = input.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environmentReceived.Schedules.RawMessage = inputBytes
	_, err = environmentReceived.UpdateScheudle(server.DB, environmentReceived.ID)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	err = queue.Publish(constants.ScheduleEnvironment, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: environmentReceived.Application.Name,
		ApplicationID:   int64(environmentReceived.Application.ID),
		EnvironmentID:   int64(environmentReceived.ID),
		EnvironmentName: environmentReceived.Name,
		ProjectName:     environmentReceived.Application.Project.Name,
		ProjectID:       int64(environmentReceived.Application.Project.ID),
		OrganizationID:  int64(environmentReceived.Application.Project.OrganizationId),
	}
	if environmentReceived.Application.Project.OrganizationId != 0 {
		notifyInfo.OrganizationName = environmentReceived.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "schedule", fmt.Sprintf("environment %s scheduled", environmentReceived.Name), "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("schedule", "environment", user.ID, environmentReceived.Application.Project, environmentReceived.Application, environmentReceived, nil, "")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "Sent command for schedule environment",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetScheduleEnvironmentLogs godoc
// @Summary Get Schedule Environment Logs
// @Description Get Schedule Environment Logs
// @Tags Schedule
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path string true "Environment Id"
// @Param page query int true "Page"
// @Param limit query int true "Limit"
// @Success 200 {array} map[string]interface{}
// @Router /environment/{id}/schedule/logs [get]
func (server *Server) GetScheduleEnvironmentLogs(w http.ResponseWriter, r *http.Request) {
	environmentReceived, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
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
	namespace := helper.GetNamespace(environmentReceived)
	data, err := server.StoreClient.CronJob().GetCronJobLogs(namespace, namespace+"-schedule", int(page), int(limit))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if data == nil {
		responses.JSON(w, http.StatusOK, []map[string]interface{}{})
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    data,
		Success: 1,
		Message: "Success",
	})
}

// ScheduleHelmEnvironment godoc
// @Summary Schedule Helm Environment
// @Description Schedule Helm Environment
// @Tags Schedule
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path string true "Environment Id"
// @Param body body models.RequestSchedule true "Schedule Helm Environment"
// @Success 200 {string} string
// @Router /helm-environment/{id}/schedule [post]
func (server *Server) ScheduleHelmEnvironment(w http.ResponseWriter, r *http.Request) {
	environmentReceived, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := CheckHelmEnvironmentPermission(server.DB, r, WRITE, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input := &models.RequestSchedule{}
	err = json.Unmarshal(body, input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input.CreatedAt = time.Now()
	err = input.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	inputBytes, err := json.Marshal(input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environmentReceived.Schedules.RawMessage = inputBytes
	_, err = environmentReceived.UpdateHelmScheudle(server.DB, environmentReceived.ID)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	err = queue.Publish(constants.ScheduleHelmEnvironment, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: environmentReceived.Application.Name,
		ApplicationID:   int64(environmentReceived.Application.ID),
		EnvironmentID:   int64(environmentReceived.ID),
		EnvironmentName: environmentReceived.Name,
		ProjectName:     environmentReceived.Application.Project.Name,
		ProjectID:       int64(environmentReceived.Application.Project.ID),
		OrganizationID:  int64(environmentReceived.Application.Project.OrganizationId),
	}
	if environmentReceived.Application.Project.OrganizationId != 0 {
		notifyInfo.OrganizationName = environmentReceived.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "schedule", fmt.Sprintf("environment %s scheduled", environmentReceived.Name), "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("schedule", "environment", user.ID, environmentReceived.Application.Project, environmentReceived.Application, nil, environmentReceived, "")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "Sent command for schedule environment",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetScheduleHelmEnvironment godoc
// @Summary Get Schedule Helm Environment
// @Description Get Schedule Helm Environment
// @Tags Schedule
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param page query int true "Page"
// @Param limit query int true "Limit"
// @Param id path string true "Environment Id"
// @Success 200 {array} map[string]interface{}
// @Router /helm-environment/{id}/schedule/logs [get]
func (server *Server) GetScheduleHelmEnvironmentLogs(w http.ResponseWriter, r *http.Request) {
	environmentReceived, err := GetHelmEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
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
	namespace := helper.GetHelmNamespace(environmentReceived)
	data, err := server.StoreClient.CronJob().GetCronJobLogs(namespace, namespace+"-schedule", int(page), int(limit))
	if err == nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if data == nil {
		responses.JSON(w, http.StatusOK, []map[string]interface{}{})
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    data,
		Success: 1,
		Message: "Success",
	})
}
