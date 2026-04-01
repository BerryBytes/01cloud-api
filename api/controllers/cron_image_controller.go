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

var cronImageInterface = models.NewCronImage()

func (server *Server) CreateCronImage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.CronImage{}
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
	dataCreated, err := cronImageInterface.Save(server.DB, &data)
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
	cronimageResponse, err := CreateCronImageResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    cronimageResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetCronImages(w http.ResponseWriter, r *http.Request) {
	datas, err := cronImageInterface.FindAll(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	cronimageResponseList := []doc.CronImage{}
	for _, cronimage := range *datas {
		cronimageResponse, err := CreateCronImageResponse(&cronimage)
		if err != nil {
			log.Error(err)
		}
		cronimageResponseList = append(cronimageResponseList, *cronimageResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    cronimageResponseList,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetCronImagesForAdmin(w http.ResponseWriter, r *http.Request) {
	datas, err := cronImageInterface.FindAllWithInactive(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	cronimageResponseList := []doc.CronImage{}
	for _, cronimage := range *datas {
		cronimageResponse, err := CreateCronImageResponse(&cronimage)
		if err != nil {
			log.Error(err)
		}
		cronimageResponseList = append(cronimageResponseList, *cronimageResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    cronimageResponseList,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetCronImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := cronImageInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	cronimageResponse, err := CreateCronImageResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    cronimageResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) UpdateCronImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.CronImage{}
	err = server.DB.Model(models.CronImage{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("CronImage not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.CronImage{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate.ID = data.ID
	dataUpdated, err := cronImageInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	cronimageResponse, err := CreateCronImageResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    cronimageResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) DeleteCronImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.CronImage{}
	err = server.DB.Model(models.CronImage{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}
	_, err = cronImageInterface.Delete(server.DB, pid)
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

func CreateCronImageResponse(cronImage *models.CronImage) (*doc.CronImage, error) {
	cronImageResponse := doc.CronImage{}
	cronImageBytes, _ := json.Marshal(cronImage)
	err := json.Unmarshal(cronImageBytes, &cronImageResponse)
	if err != nil {
		return nil, err
	}
	return &cronImageResponse, nil
}
