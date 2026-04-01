package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
)

// CreateCronJob godoc
// @Summary Create a new CronJob
// @Description Create a new CronJob with the input paylod
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.CronJob true "Create CronJob"
// @Success 201 {object} doc.CronJob
// @Router /environment/{id}/cronjob [post]

var cronJobInterface = models.NewCronJob()

func (server *Server) CreateCronJob(w http.ResponseWriter, r *http.Request) {
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
	cronJob := &models.CronJob{}
	err = json.Unmarshal(body, cronJob)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	cronJob.Prepare()
	err = cronJob.Validate()
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

	if environmentReceived.Application.ServiceType < 1 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("not supported for this plugin"))
		return
	}

	if cronJob.Image == "" {
		if environmentReceived.Application.ServiceType == 1 {
			data, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(environmentReceived))
			if err != nil {
				responses.ERROR(w, http.StatusNotFound, errors.New("image not found"))
				return
			}
			if data == nil {
				responses.ERROR(w, http.StatusNotFound, errors.New("image not found"))
				return
			}
			img := fmt.Sprintf("%s:%s", data.CIRequest.RepositoryImage.Name, data.CIRequest.RepositoryImage.Tag)
			cronJob.Image = img

		} else if environmentReceived.Application.ServiceType == 2 {
			img := fmt.Sprintf("%s:%s", environmentReceived.ImageUrl, environmentReceived.ImageTag)
			cronJob.Image = img
		} else {
			responses.ERROR(w, http.StatusNotFound, errors.New("not supported for this image"))
			return
		}
	}
	err = cronJob.VerifyResource(environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	_, state, _, err := server.StoreClient.PodState().GetEnvironmentState(fmt.Sprintf("%s-%d-%d", os.Getenv("GCLOUD_NAMESPACE"), environmentReceived.ApplicationID, pid))
	if err != nil {
		responses.JSON(w, http.StatusNotFound, err)
		return
	}
	if state == "" {
		responses.ERROR(w, http.StatusTooEarly, errors.New("invalid environment state"))
		return
	}
	cronJob.UserID = uint64(uid)
	data, err := cronJobInterface.SaveCronJob(server.DB, cronJob, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	err = queue.Publish(constants.CreateCronJob, &map[string]interface{}{
		"environment": environmentReceived,
		"cron_job":    data,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "CronJob Requested",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// UpdateCronJob godoc
// @Summary update a CronJob
// @Description update a CronJob with the input paylod
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param cid path int true "CronJob id"
// @Param body body doc.CronJob true "Update CronJob"
// @Success 200 {object} doc.CronJob
// @Router /environment/{id}/cronjob/{cid} [put]
func (server *Server) UpdateCronJob(w http.ResponseWriter, r *http.Request) {
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
	if environmentReceived.Application.ServiceType < 1 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("not supported for this Plugin"))
		return
	}

	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update cron job"))
			return
		}
	}
	data, err := cronJobInterface.Find(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cron job not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	cronJobUpdate := &models.CronJob{}
	err = json.Unmarshal(body, cronJobUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if cronJobUpdate.Name != "" && cronJobUpdate.Name != data.Name {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if cronJobUpdate.Image == "" {
		data, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(environmentReceived))
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("image not found"))
			return
		}
		img := fmt.Sprintf("%s:%s", data.CIRequest.RepositoryImage.Name, data.CIRequest.RepositoryImage.Tag)
		cronJobUpdate.Image = img
	}
	cronJobUpdate.ID = data.ID
	cu, err := cronJobInterface.Update(server.DB, cronJobUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}

	err = queue.Publish(constants.UpdateCronJob, &map[string]interface{}{
		"environment": environmentReceived,
		"cron_job":    cu,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "CronJob Update Requested",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetCronJobs godoc
// @Summary Get all CronJobs
// @Description Get array of CronJobs with the environment id
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {array} doc.CronJob
// @Router /environment/{id}/cronjob [get]
func (server *Server) GetCronJobs(w http.ResponseWriter, r *http.Request) {
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
		if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view cron job"))
			return
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, environmentReceived.CronJob)
		return
	}
	cronjobResponseList := []doc.CronJob{}
	for _, cronJob := range environmentReceived.CronJob {
		cronjobResponse, err := CreateCronJobResponse(cronJob)
		if err != nil {
			log.Error(err)
		}
		cronjobResponseList = append(cronjobResponseList, *cronjobResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    cronjobResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetCronJob godoc
// @Summary Get CronJob
// @Description Get a CronJob with the cronjob id
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param cid path int true "CronJob id"
// @Success 200 {object} doc.CronJob
// @Router /environment/{id}/cronjob/{cid} [get]
func (server *Server) GetCronJob(w http.ResponseWriter, r *http.Request) {
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
		if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view cron job"))
			return
		}
	}

	dataReceived, err := cronJobInterface.Find(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	cronjobResponse, err := CreateCronJobResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    cronjobResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteCronJob godoc
// @Summary Delete CronJob by id
// @Description Delete a CronJob with CronJob id
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param cid path int true "CronJob id"
// @Success 200 {object} doc.CronJob
// @Router /environment/{id}/cronjob/{cid} [delete]
func (server *Server) DeleteCronJob(w http.ResponseWriter, r *http.Request) {
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
		if !authInterface.IsAdminOfEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to delete cron job"))
			return
		}
	}

	data, err := cronJobInterface.Find(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cron job not found"))
		return
	}
	err = server.DB.Model(&data).Delete(data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = queue.Publish(constants.DeleteCronJob, &map[string]interface{}{
		"environment": environmentReceived,
		"cron_job":    data,
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

// RequestCronJobStatus godoc
// @Summary Request CronJob Status
// @Description Request CronJob status for fetching the status
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} object
// @Router /environment/{id}/cronjob-fetch [get]
func (server *Server) RequestCronJobStatus(w http.ResponseWriter, r *http.Request) {
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
		if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to request cron job"))
			return
		}
	}

	err = queue.Publish(constants.FetchCronJob, &map[string]interface{}{
		"environment": environmentReceived,
		"cron_job":    &models.CronJob{},
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "CronJob Fetch Requested",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetCronJobStatus godoc
// @Summary Get CronJob Status
// @Description Get CronJob status after fetching the status
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} object
// @Router /environment/{id}/cronjob-status [get]
func (server *Server) GetCronJobStatus(w http.ResponseWriter, r *http.Request) {
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
		if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to get cron job status"))
			return
		}
	}
	status, err := server.StoreClient.CronJob().GetCronJobStatus(helper.GetNamespace(environmentReceived))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, status)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    status,
		Success: 1,
		Message: "Success",
	})
}

// CreateJobNow godoc
// @Summary Create a new Job
// @Description Create a new Job with the input paylod
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param cid path int true "CronJob id"
// @Param body body doc.CronJob true "Create Job"
// @Success 201 {object} doc.CronJob
// @Router /environment/{id}/cronjob/{cid}/run [post]
func (server *Server) CreateJobNow(w http.ResponseWriter, r *http.Request) {
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
	if environmentReceived.Application.ServiceType < 1 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("not supported for this plugin"))
		return
	}

	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to create cron job"))
			return
		}
	}

	data, err := cronJobInterface.Find(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cron job not found"))
		return
	}

	err = queue.Publish(constants.RunJobNow, &map[string]interface{}{
		"environment": environmentReceived,
		"cron_job":    data,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "CronJob Update Requested",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetCronJobsLogs godoc
// @Summary Get CronJob Logs
// @Description Get CronJob logs
// @Tags CronJob
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param cid path int true "CronJob id"
// @Param page query int true "Page"
// @Param limit query int true "limit"
// @Success 200 {object} object
// @Router /environment/{id}/cronjob/{cid}/logs [get]
func (server *Server) GetCronJobsLogs(w http.ResponseWriter, r *http.Request) {
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
	if environmentReceived.Application.ServiceType < 1 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("not supported for this plugin"))
		return
	}

	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to get cron job logs"))
			return
		}
	}
	data, err := cronJobInterface.Find(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cron job not found"))
		return
	}
	status, err := server.StoreClient.CronJob().GetCronJobLogs(helper.GetNamespace(environmentReceived), slug.Make(data.Name), int(page), int(limit))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, status)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    status,
		Success: 1,
		Message: "Success",
	})

}

func CreateCronJobResponse(cronJob *models.CronJob) (*doc.CronJob, error) {
	cronjobResponse := doc.CronJob{}
	cronjobBytes, _ := json.Marshal(cronJob)
	err := json.Unmarshal(cronjobBytes, &cronjobResponse)
	if err != nil {
		return nil, err
	}
	return &cronjobResponse, nil

}
