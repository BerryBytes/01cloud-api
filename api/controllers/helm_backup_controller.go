package controllers

import (
	"01cloud-api/api/backup"
	pb "01cloud-api/api/backup/proto"
	"01cloud-api/api/notifications"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/responses"

	"github.com/gorilla/mux"
)

// CreateBackup godoc
// @Summary trigger create backup command for 01cloud-backup
// @Description trigger create backup command for 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Param body body doc.Backup true "Create Backup"
// @Success 200 {object} string
// @Router /helm-environment/{id}/backup [post]
func (server *Server) CreateHelmBackup(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetHelmEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	namespace := helper.GetHelmNamespace(environment)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := &backup.Backup{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.TriggeredBy = int64(user.ID)
	data.TriggeredByName = user.FirstName + " " + user.LastName
	data.Namespace = namespace
	data.EnvironmentType = constants.TagHelm
	mes := &pb.BackupMessage{}
	data.ToGrpc(mes)

	c := pb.NewBackupClient(server.BackupClient)

	gr, err := c.CreateBackup(context.Background(), &pb.CreateBackupRequest{Backup: mes,
		Path: environment.Application.Cluster.ConfigPath})
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: environment.Application.Name,
		ApplicationID:   int64(environment.Application.ID),
		EnvironmentID:   int64(environment.ID),
		EnvironmentName: environment.Name,
		ProjectName:     environment.Application.Project.Name,
		ProjectID:       int64(environment.Application.Project.ID),
		OrganizationID:  int64(environment.Application.Project.OrganizationId),
	}
	if environment.Application.Project.OrganizationId != 0 {
		notifyInfo.OrganizationName = environment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "backup-init", fmt.Sprintf("environment %s backup started", environment.Name), "info", int64(user.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, map[string]interface{}{
			"success": gr.GetOk(),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// ListBackup godoc
// @Summary listing backup command for 01cloud-backup
// @Description listing backup command for 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} string
// @Router /helm-environment/{id}/backup [get]
func (server *Server) ListHelmBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetHelmEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	namespace := helper.GetHelmNamespace(environment)
	c := pb.NewBackupClient(server.BackupClient)
	gr, err := c.ListBackup(context.Background(), &pb.ListBackupRequest{
		Namespace: namespace,
		Path:      environment.Application.Cluster.ConfigPath,
		Limit:     int64(environment.Application.Project.Subscription.Backups),
	})
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"success": gr.GetOk(),
			"data":    gr.GetBackups(),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]interface{}{
			"success": gr.GetOk(),
			"data":    gr.GetBackups(),
		},
		Success: 1,
		Message: "Success",
	})
}

// PreserveBackup godoc
// @Summary preserve this backup instance in 01cloud-backup
// @Description preserve this backup instance in 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Param bid path int true "Backup id"
// @Success 200 {object} string
// @Router /helm-environment/{id}/backup/{bid}/preserve [post]
func (server *Server) PreserveHelmBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetHelmEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	vars := mux.Vars(r)
	bid, err := strconv.ParseUint(vars["bid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := &backup.Backup{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	c := pb.NewBackupClient(server.BackupClient)

	gr, err := c.PreserveBackup(context.Background(), &pb.PreserveBackupRequest{Id: int64(bid), Preserve: data.Preserved,
		Path: environment.Application.Cluster.ConfigPath})
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"success": gr.GetOk(),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// RestoreBackup godoc
// @Summary restore this backup instance in 01cloud-backup
// @Description restore this backup instance in 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Param bid path int true "Backup id"
// @Success 200 {object} string
// @Router /helm-environment/{id}/backup/{bid}/restore [post]
func (server *Server) RestoreHelmBackup(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetHelmEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	vars := mux.Vars(r)
	bid, err := strconv.ParseUint(vars["bid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	c := pb.NewBackupClient(server.BackupClient)

	gr, err := c.RestoreBackup(context.Background(), &pb.RestoreBackupRequest{
		Id:              int64(bid),
		Path:            environment.Application.Cluster.ConfigPath,
		TriggeredByName: user.FirstName + " " + user.LastName,
	})
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}
	environment.Action = "Restoring"
	err = queue.Publish(constants.UpdateHelmEnvironmentState, environment)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	notifyInfo := notifications.BasicInfoNotification{
		ApplicationName: environment.Application.Name,
		ApplicationID:   int64(environment.Application.ID),
		EnvironmentID:   int64(environment.ID),
		EnvironmentName: environment.Name,
		ProjectName:     environment.Application.Project.Name,
		ProjectID:       int64(environment.Application.Project.ID),
		OrganizationID:  int64(environment.Application.Project.OrganizationId),
	}
	if environment.Application.Project.OrganizationId != 0 {
		notifyInfo.OrganizationName = environment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "restore-init", fmt.Sprintf("environment %s restoring", environment.Name), "info", int64(user.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"success": gr.GetOk(),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// RestoreBackupList godoc
// @Summary view restore list of backup instance in 01cloud-backup
// @Description view restore list of backup instance in 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} string
// @Router /helm-environment/{id}/restores [get]
func (server *Server) RestoreHelmBackupList(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetHelmEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	c := pb.NewBackupClient(server.BackupClient)
	namespace := helper.GetHelmNamespace(environment)
	gr, err := c.RestoreListBackup(context.Background(), &pb.RestoreListRequest{Namespace: namespace,
		Path: environment.Application.Cluster.ConfigPath})
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"success":  gr.GetOk(),
			"restores": gr.GetBackups(),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]interface{}{
			"success":  gr.GetOk(),
			"restores": gr.GetBackups(),
		},
		Success: 1,
		Message: "Success",
	})
}

// DeleteBackup godoc
// @Summary delete this backup instance from 01cloud-backup
// @Description delete this backup instance from 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Param bid path int true "Backup id"
// @Success 200 {object} string
// @Router /helm-environment/{id}/backup/{bid} [delete]
func (server *Server) DeleteHelmBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetHelmEnvironmentWithUserPermission(r, ADMIN)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	vars := mux.Vars(r)
	bid, err := strconv.ParseUint(vars["bid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	c := pb.NewBackupClient(server.BackupClient)

	gr, err := c.DeleteBackup(context.Background(), &pb.DeleteBackupRequest{Id: int64(bid),
		Path: environment.Application.Cluster.ConfigPath})
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"success": gr.GetOk(),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetSettingBackup godoc
// @Summary get the backup setting for this environment in 01cloud-backup
// @Description get the backup setting for this environment in 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} string
// @Router /helm-environment/{id}/backup/setting [get]
func (server *Server) GetSettingHelmBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetHelmEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	namespace := helper.GetHelmNamespace(environment)
	c := pb.NewBackupClient(server.BackupClient)
	gr, err := c.GetSettingBackup(context.Background(), &pb.GetSettingBackupRequest{
		Namespace: namespace,
		Path:      environment.Application.Cluster.ConfigPath,
	})
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"success":  gr.GetOk(),
			"settings": gr.GetJsonSetting(),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]interface{}{
			"success":  gr.GetOk(),
			"settings": gr.GetJsonSetting(),
		},
		Success: 1,
		Message: "Success",
	})
}

// SaveSettingBackup godoc
// @Summary get the backup setting for this environment in 01cloud-backup
// @Description get the backup setting for this environment in 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "HelmEnvironment id"
// @Success 200 {object} string
// @Router /helm-environment/{id}/backup/setting [post]
func (server *Server) SaveSettingHelmBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetHelmEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	namespace := helper.GetHelmNamespace(environment)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var settings string
	err = json.Unmarshal(body, &settings)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	c := pb.NewBackupClient(server.BackupClient)

	gr, err := c.SaveSettingBackup(context.Background(), &pb.SaveSettingBackupRequest{
		Namespace:   namespace,
		JsonSetting: settings,
		Path:        environment.Application.Cluster.ConfigPath,
	})
	if err != nil {
		responses.ERROR(w, http.StatusServiceUnavailable, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"success": gr.GetOk(),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
