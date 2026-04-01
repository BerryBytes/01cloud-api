package controllers

import (
	"encoding/json"
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

var activityInterface = models.NewActivity()

// CreateActivity godoc
// @Summary Create Activity
// @Description Create Activity with input payload
// @Tags Activity
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Activity true "Create Activity"
// @Success 200 {object} doc.Activity
// @Router /activity [post]
func (server *Server) CreateActivity(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.Activity{}
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
	dataCreated, err := activityInterface.Save(server.DB, data)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
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
	activityResponse, err := CreateActivityResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    activityResponse,
		Success: 1,
		Message: "Success",
	})

}

func (server *Server) SaveActivityWithJson(action string, module string, userId uint, project *models.Project, application *models.Application, environment *models.Environment, helmEnvironment *models.HelmEnvironment, extras string) (*models.Activity, error) {
	data := models.Activity{ProjectID: project.ID, Module: module, Action: action, Active: true}
	if userId != 0 {
		data.UserID = userId
	}
	if application != nil {
		data.ApplicationID = application.ID
	}
	if environment != nil {
		data.EnvironmentID = environment.ID
	}
	if extras != "" {
		data.Extras = extras
	}
	var moduleName = ""
	if module == "project" {
		moduleName = project.Name
	}
	if module == "application" && application != nil {
		moduleName = application.Name
	}

	if module == "environment" && environment != nil {
		moduleName = environment.Name
	}
	if module == "environment" && helmEnvironment != nil {
		moduleName = helmEnvironment.Name
	}
	data.Remarks = moduleName
	err := data.Validate()
	if err != nil {
		return &models.Activity{}, err
	}
	dataCreated, err := activityInterface.Save(server.DB, data)
	return dataCreated, err
}

func (server *Server) SaveActivity(action string, module string, pid uint, uid uint, aid uint, eid uint, r *http.Request) (*models.Activity, error) {
	data := models.Activity{ProjectID: pid, Module: module, Action: action, Active: true}
	if uid != 0 {
		data.UserID = uid
	}
	if aid != 0 {
		data.ApplicationID = aid
	}
	if eid != 0 {
		data.EnvironmentID = eid
	}
	err := data.Validate()
	if err != nil {
		return &models.Activity{}, err
	}
	dataCreated, err := activityInterface.Save(server.DB, data)
	return dataCreated, err
}

func (server *Server) GetActivities(w http.ResponseWriter, r *http.Request) {
	datas, err := activityInterface.FindAll(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	activityResponseList := []doc.Activity{}
	for _, item := range *datas {
		activityResponse, err := CreateActivityResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		activityResponseList = append(activityResponseList, *activityResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    activityResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetProjectActivities godoc
// @Summary Get Project Activities
// @Description Get Project Activities
// @Tags Activity
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param limit query int true "limit"
// @Param page query int true "page"
// @Param action query string true "action"
// @Param module query string true "module"
// @Param datestart query string true "start date"
// @Param dateend query string true "end date"
// @Param userid query int true "User Id"
// @Param pid path int true "Get Project Activities"
// @Success 200 {object} []doc.Activity
// @Router /project/{pid}/activities [get]
func (server *Server) GetProjectActivities(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil || limit == 0 {
		limit = 100
	}
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil || page == 0 {
		page = 1
	}
	offset := (page - 1) * limit
	action := r.FormValue("action")
	module := r.FormValue("module")
	datestart := r.FormValue("datestart")
	dateend := r.FormValue("dateend")
	userid, _ := strconv.Atoi(r.FormValue("userid"))
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	datas, count, err := activityInterface.FindAllActivityByProject(server.DB, uint(pid), limit, offset, action, module, datestart, dateend, userid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"data":  datas,
			"count": count,
		})
		return
	}
	activityResponseList := []doc.Activity{}
	for _, item := range *datas {
		activityResponse, err := CreateActivityResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		activityResponseList = append(activityResponseList, *activityResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    activityResponseList,
		Success: 1,
		Message: "Success",
		Count:   count,
	})
}

// GetOrganizationActivities godoc
// @Summary Get Organization Activities
// @Description Get Organization Activities
// @Tags Activity
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param limit query int true "limit"
// @Param page query int true "page"
// @Param action query string true "action"
// @Param module query string true "module"
// @Param datestart query string true "start date"
// @Param dateend query string true "end date"
// @Param userid query int true "User Id"
// @Param oid path int true "Get Organization Activities"
// @Success 200 {array} doc.Activity
// @Router /organization/{oid}/activities [get]
func (server *Server) GetOrganizationActivities(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil || limit == 0 {
		limit = 100
	}
	page, err := strconv.Atoi(r.FormValue("page"))
	if err != nil || page == 0 {
		page = 1
	}
	offset := (page - 1) * limit
	action := r.FormValue("action")
	module := r.FormValue("module")
	datestart := r.FormValue("datestart")
	dateend := r.FormValue("dateend")
	userid, _ := strconv.Atoi(r.FormValue("userid"))
	oid, err := strconv.ParseUint(vars["oid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	datas, count, err := activityInterface.FindAllActivityByOrganization(server.DB, uint(oid), limit, offset, action, module, datestart, dateend, userid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"data":  datas,
			"count": count,
		})
		return
	}
	activityResponseList := []doc.Activity{}
	for _, item := range *datas {
		activityResponse, err := CreateActivityResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		activityResponseList = append(activityResponseList, *activityResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    activityResponseList,
		Success: 1,
		Message: "Success",
		Count:   count,
	})
}

// Get Activity godoc
// @Summary Get Activity
// @Description Get Activity
// @Tags Activity
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Activity"
// @Success 200 {object} doc.Activity
// @Router /activity/{id} [get]
func (server *Server) GetActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataReceived, err := activityInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	activityResponse, err := CreateActivityResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    activityResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateActivity godoc
// @Summary Update Activity
// @Description Update Activity
// @Tags Activity
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Update Activity"
// @Success 200 {object} doc.Activity
// @Router /activity/{id} [put]
func (server *Server) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data, err := activityInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.Activity{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate.ID = data.ID
	dataUpdated, err := activityInterface.Update(server.DB, dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	responses.JSON(w, http.StatusOK, dataUpdated)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	activityResponse, err := CreateActivityResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    activityResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteActivity godoc
// @Summary Delete Activity
// @Description Delete Activity
// @Tags Activity
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Delete Activity"
// @Success 200
// @Router /activity/{id} [delete]
func (server *Server) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	_, err = activityInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	_, err = activityInterface.Delete(server.DB, pid)
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

func (server *Server) SaveOrganizationActivityWithJson(uid, oid uint, action, module, remarks, extras string) (*models.Activity, error) {
	data := models.Activity{UserID: uid, OrganizationID: oid, ProjectID: 0, Module: module, Action: action, Remarks: remarks, Extras: extras, Active: true}
	err := data.Validate()
	if err != nil {
		return &models.Activity{}, err
	}
	dataCreated, err := activityInterface.Save(server.DB, data)
	return dataCreated, err
}

func CreateActivityResponse(activity *models.Activity) (*doc.Activity, error) {
	activityResponse := doc.Activity{}
	activityBytes, _ := json.Marshal(activity)
	err := json.Unmarshal(activityBytes, &activityResponse)
	if err != nil {
		return nil, err
	}
	return &activityResponse, nil
}
