package controllers

import (
	"01cloud-api/api/backup"
	pb "01cloud-api/api/backup/proto"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
	"context"
	"encoding/json"
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
// @Param id path int true "Environment id"
// @Param body body doc.Backup true "Create Backup"
// @Success 200 {object} string
// @Router /environment/{id}/backup [post]
func (server *Server) CreateBackup(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	namespace := helper.GetNamespace(environment)
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
	mes := &pb.BackupMessage{}
	data.ToGrpc(mes)

	c := pb.NewBackupClient(server.BackupClient)

	gr, err := c.CreateBackup(context.Background(), &pb.CreateBackupRequest{Backup: mes,
		Path: environment.Application.Cluster.ConfigPath})
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, map[string]interface{}{
			"success": gr.GetOk(),
		})
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
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
// @Param id path int true "Environment id"
// @Success 200 {object} string
// @Router /environment/{id}/backup [get]
func (server *Server) ListBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	namespace := helper.GetNamespace(environment)
	c := pb.NewBackupClient(server.BackupClient)
	gr, err := c.ListBackup(context.Background(), &pb.ListBackupRequest{
		Namespace: namespace,
		Path:      environment.Application.Cluster.ConfigPath,
		Limit:     int64(environment.Application.Project.Subscription.Backups),
	})
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
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
			"data": gr.GetBackups(),
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
// @Param id path int true "Environment id"
// @Param bid path int true "Backup id"
// @Success 200 {object} string
// @Router /environment/{id}/backup/{bid}/preserve [post]
func (server *Server) PreserveBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
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
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
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

func (server *Server) AddLabelBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
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
	label := []*pb.Label{}
	for _, data := range data.Label {
		l := &pb.Label{}
		data.ToGrpc(l)
		label = append(label, l)
	}
	gr, err := c.AddLabelBackup(context.Background(), &pb.AddLabelRequest{Id: int64(bid), Label: label,
		Path: environment.Application.Cluster.ConfigPath})
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
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
// @Param id path int true "Environment id"
// @Param bid path int true "Backup id"
// @Success 200 {object} string
// @Router /environment/{id}/backup/{bid}/restore [post]
func (server *Server) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
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
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	environment.Action = "Restoring"
	err = queue.Publish(constants.UpdateEnvironmentState, environment)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
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

// RestoreBackupList godoc
// @Summary view restore list of backup instance in 01cloud-backup
// @Description view restore list of backup instance in 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} string
// @Router /environment/{id}/restores [get]
func (server *Server) RestoreBackupList(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	c := pb.NewBackupClient(server.BackupClient)
	namespace := helper.GetNamespace(environment)
	gr, err := c.RestoreListBackup(context.Background(), &pb.RestoreListRequest{Namespace: namespace,
		Path: environment.Application.Cluster.ConfigPath})
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
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
// @Param id path int true "Environment id"
// @Param bid path int true "Backup id"
// @Success 200 {object} string
// @Router /environment/{id}/backup/{bid} [delete]
func (server *Server) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, ADMIN)
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
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
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
// @Param id path int true "Environment id"
// @Success 200 {object} string
// @Router /environment/{id}/backup/setting [get]
func (server *Server) GetSettingBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	namespace := helper.GetNamespace(environment)
	c := pb.NewBackupClient(server.BackupClient)
	gr, err := c.GetSettingBackup(context.Background(), &pb.GetSettingBackupRequest{
		Namespace: namespace,
		Path:      environment.Application.Cluster.ConfigPath,
	})
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
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
// @Param id path int true "Environment id"
// @Success 200 {object} string
// @Router /environment/{id}/backup/setting [post]
func (server *Server) SaveSettingBackup(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	namespace := helper.GetNamespace(environment)
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
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
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

// ScheduleBackup godoc
// @Summary trigger schedule backup command for 01cloud-backup
// @Description trigger schedule backup command for 01cloud-backup
// @Tags Backup
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.Backup true "schedule Backup"
// @Success 200 {object} string
// @Router /environment/{id}/schedule-backup [post]
func (server *Server) ScheduleBackup(w http.ResponseWriter, r *http.Request) {
	environment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	namespace := helper.GetNamespace(environment)
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
	data.Name = namespace + "-backup-schedule"
	data.Namespace = namespace
	mes := &pb.BackupMessage{}
	data.ToGrpc(mes)

	c := pb.NewBackupClient(server.BackupClient)
	gr, err := c.CreateBackup(context.Background(), &pb.CreateBackupRequest{Backup: mes,
		Path: environment.Application.Cluster.ConfigPath})
	if err != nil || !gr.GetOk() {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
