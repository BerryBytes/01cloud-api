package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/helper"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var dashboard = models.NewDashboard()

// Home godoc
// @Summary Home
// @Description Home
// @Tags Home
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {string} string
// @Router / [get]
func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "Welcome To This 01Cloud API")

}

// Dashboard godoc
// @Summary Get Admin Dashboard
// @Description Get Admin Dashboard
// @Tags Home
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /admin/dashboard [get]
func (server *Server) Dashboard(w http.ResponseWriter, r *http.Request) {
	count, err := dashboard.Dashboard(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, count)

}

// SidebarApi godoc
// @Summary Sidebar Api
// @Description Get Sidebar
// @Tags Home
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} map[string]interface{}
// @Router /sidebar [get]
func (server *Server) SidebarApi(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	project := models.Project{}
	project.OrganizationId = uint64(oid)
	projects, err := iproject.FindAllByUser(server.DB, uint(uid), &project)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	var response []map[string]interface{}
	for _, p := range projects {
		projectResponse, _ := CreateProjectResponse(&p)
		projectByte, _ := json.Marshal(projectResponse)
		proj := map[string]interface{}{}
		err = json.Unmarshal(projectByte, &proj)
		if err != nil {
			log.Error(err)
		}
		apps, err := applicationInterface.FindAllApplicationByProject(server.DB, uint64(p.ID), uint64(uid), false)
		var applications []map[string]interface{}
		for _, app := range apps {
			envModel := models.Environment{}
			helmEnvModel := models.HelmEnvironment{}
			appResponse, _ := CreateApplicationResponse(app)
			appByte, _ := json.Marshal(appResponse)
			appJson := map[string]interface{}{}
			err = json.Unmarshal(appByte, &appJson)
			if err != nil {
				log.Error(err)
			}
			envs, err := envModel.FindAllByApplication(server.DB, uint64(app.ID), true, uint64(uid))
			if err == nil {
				envData := []doc.Environment{}
				for _, env := range envs {
					envResponse, _ := CreateEnvironmentResponse(env)
					envData = append(envData, *envResponse)
				}

				appJson["environments"] = envData
			}
			helmEnvs, err := helmEnvModel.FindAllByApplication(server.DB, uint64(app.ID), true, uint64(uid))
			if err == nil {
				helmData := []doc.HelmEnvironment{}
				for _, helm := range helmEnvs {
					helmResponse, _ := CreateHelmEnvironmentResponse(helm)
					helmData = append(helmData, *helmResponse)
				}
				appJson["helm_environments"] = helmData
			}
			applications = append(applications, appJson)
		}
		if err == nil {
			proj["applications"] = applications
		}
		response = append(response, proj)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, response)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    response,
		Success: 1,
		Message: "Success",
	})
}

// UploadFile godoc
// @Summary Upload File
// @Description Upload File
// @Tags Home
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param file formData file true  "This upload file"
// @Param file_name query string true "File Name"
// @Param file_type query string true "File Type"
// @Success 200 {object} map[string]string{}
// @Router /upload [post]
func (server *Server) UploadFile(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	file, handler, err := r.FormFile("file")
	fileName := r.FormValue("file_name")
	fileType := r.FormValue("file_type")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer file.Close()

	if fileName == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("file name is required"))
		return
	}
	if fileType == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("file type is required"))
		return
	}

	path := fmt.Sprintf("%s/%s", "/data/uploads", fileType)
	var extension = filepath.Ext(handler.Filename)
	err = os.MkdirAll(path, 0777)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	fn := fmt.Sprintf("%s%s", uuid.New().String(), extension)
	fullFp := fmt.Sprintf("%s/%s", path, fn)
	f, err := os.OpenFile(fullFp, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	contentType, err := mimetype.DetectFile(fullFp)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if fileType != "json" {
		if strings.HasPrefix(fileType, "image") && !strings.HasPrefix(contentType.String(), "image") {
			_ = os.Remove(fullFp)
			responses.ERROR(w, http.StatusBadRequest, errors.New("uploaded file is not a valid image. Only JPG, PNG and GIF files are allowed"))
			return
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]string{
			"message": "file uploaded successfully",
			"path":    fullFp,
			"url":     fmt.Sprintf("%s/%s/%s/%s", os.Getenv("BASE_URL"), "uploads", fileType, fn),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]string{
			"message": "file uploaded successfully",
			"path":    fullFp,
			"url":     fmt.Sprintf("%s/%s/%s/%s", os.Getenv("BASE_URL"), "uploads", fileType, fn),
		},
		Success: 1,
		Message: "Success",
	})
}

// UploadFileToBucket godoc
// @Summary Upload File To Bucket
// @Description Upload File To Bucket
// @Tags Home
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param file formData file true  "Upload File to Bucket"
// @Success 200 {object} map[string]string{}
// @Router /upload-gcs [post]
func (server *Server) UploadFileToBucket(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	file, handler, err := r.FormFile("file")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	defer func() {
		_ = file.Close()
	}()

	var extension = filepath.Ext(handler.Filename)
	fn := fmt.Sprintf("images/%s%s", uuid.New().String(), extension)
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()
	//_ = storage_helper.CreateBucket(server.StorageClient, os.Getenv("GCLOUD_PROJECT"), "uploads")
	_, err = server.StorageClient.Bucket("zerone-uploads").Attrs(ctx)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("bucket not exists"))
		return
	}
	wc := server.StorageClient.Bucket("zerone-uploads").Object(fn).NewWriter(ctx)

	if _, err = io.Copy(wc, file); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if err := wc.Close(); err != nil {
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]string{
			"message": "file uploaded successfully",
			"url":     fmt.Sprintf("https://storage.googleapis.com/%s/%s", "zerone-uploads", fn),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]string{
			"message": "file uploaded successfully",
			"url":     fmt.Sprintf("https://storage.googleapis.com/%s/%s", "zerone-uploads", fn),
		},
		Success: 1,
		Message: "Success",
	})
}
