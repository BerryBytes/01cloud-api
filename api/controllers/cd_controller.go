package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// Webhook godoc
// @Summary Listen to Webhook from various registry providers
// @Description Listen to Webhook from various registry providers
// @Tags Registry
// @Accept  json
// @Produce  json
// @Param service path string true "registry service"
// @Param eid path int true "environment id"
// @Param userId path int true "user id"
// @Security ApiKeyAuth
// @Param body body string true "Webhook Data"
// @Success 200 {object} string
// @Router /webhook/image/{service}/{eid}/{userId} [post]
func (server *Server) Webhook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, err := strconv.ParseUint(vars["userId"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	userId, err := GetTokenID(r, server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	pid, err := strconv.ParseUint(vars["eid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uint(userId))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	environment := &models.Environment{}
	environment, err = environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if !environment.Active {
		responses.ERROR(w, http.StatusBadRequest, errors.New("environment is in stopped state"))
		return
	}
	if uint(uid) != uint(environment.Application.Project.UserID) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environment) {
			responses.ERROR(w, http.StatusBadRequest, errors.New("you are not authorized to deploy"))
			return
		}
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	image := models.RepositoryImage{}
	err = json.Unmarshal(body, &image)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	env, _ := environment.ToJson()
	if environment.ServiceType == 2 {
		err = image.Validate(environment)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environment.ImageTag = image.Tag
		gtUsr, err := server.VerifyRepoAndTags(environment)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		provider := gtUsr.ServiceName
		if provider == "aws_ecr" {
			provider = "aws"
		}
		secrets := map[string]interface{}{
			"provider":            provider,
			"access_key_id":       gtUsr.AccessToken,
			"secret_access_key":   gtUsr.SecretKey,
			"aws_region":          gtUsr.Region,
			"docker_username":     gtUsr.ServiceUserName,
			"docker_token":        gtUsr.AccessToken,
			"gcr_service_account": gtUsr.AccessToken,
		}
		env["secrets"] = secrets
	} else {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("only for container registry"))
		return
	}
	_, err = environment.Update(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	env["image_tag"] = environment.ImageTag
	environment.Action = "Deploying"
	err = queue.Publish(constants.UpgradeRelease, env)
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
	if environment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = environment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "update", fmt.Sprintf("environment %s update", environment.Name), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, environment.Application.Project, environment.Application, environment, nil, "tag updated")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]bool{
			"success": true,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})

}

func (server *Server) GetCDStrategyConfig(w http.ResponseWriter, r *http.Request) {
	file, _ := os.ReadFile("/data/public/deployment.schema.json")
	res := make(map[string]interface{})
	err := json.Unmarshal(file, &res)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, res)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    res,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) DeployImage(pid uint64, uid uint, url, tag string) error {
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		return errors.New("user not found")
	}
	environment := &models.Environment{}
	environment, err = environment.Find(server.DB, pid)
	if err != nil {
		return errors.New("environment not found")
	}
	if !environment.Active {
		return errors.New("environment is in stopped state")
	}
	if uid != uint(environment.Application.Project.UserID) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environment) {
			return errors.New("you are not authorized to deploy")
		}
	}

	if environment.Application.ServiceType != 2 {
		return errors.New("deployment from Registry not available for template environment")
	}
	wf, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(environment))
	if err != nil {
		return errors.New("image not found")
	}
	workflow := helper.ConvertWorkflowType(wf)
	if workflow != nil {
		if v, ok := workflow.Workflow["Status"]; ok {
			status := v.(map[string]interface{})["Phase"].(string)
			if status == "" || strings.ToLower(status) == "running" {
				return errors.New("you cannot deploy the environment at this time. CI build is already running ")
			}
		} else {
			return errors.New("deployment from Registry not available. ")
		}
		image := models.RepositoryImage{
			Tag:  tag,
			Name: url,
		}
		image.Repository = workflow.CIRequest.RepositoryImage.Repository
		workflow.CIRequest.RepositoryImage = image
		environment.CiRequest = &workflow.CIRequest
		environment.RepositoryImage = &image
	} else {
		return errors.New("no image for deployment")
	}
	_, err = environment.Update(server.DB)
	if err != nil {
		return errors.New("unable to deploy")
	}
	environment.Action = "Deploying"
	err = queue.Publish(constants.UpgradeRelease, environment)
	if err != nil {
		return err
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
	if environment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = environment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, "deployment", fmt.Sprintf("environment %s deploying", environment.Name), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("update", "environment", uid, environment.Application.Project, environment.Application, environment, nil, "deploying")

	return nil
}
func GetTokenID(r *http.Request, db *gorm.DB) (uint, error) {
	token, err := GetToken(r)
	if err != nil {
		return 0, err
	}
	id, err := GetIDFromToken(db, token)
	if err != nil {
		return 0, err
	}
	return id, nil
}
func GetToken(r *http.Request) (string, error) {
	var token string
	if r.Header.Get("X-TOKEN") != "" {
		token = r.Header.Get("X-TOKEN")
	} else if r.URL.Query().Get("token") != "" {
		token = r.URL.Query().Get("token")
	} else {
		var tokenValue map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&tokenValue)
		if err != nil {
			return "", err
		}
		if t, ok := tokenValue["token"]; ok {
			token = t.(string)
		}
	}
	if token == "" {
		return "", errors.New("token is required")
	}
	return token, nil
}
func GetIDFromToken(db *gorm.DB, token string) (uint, error) {
	tokenReceived, err := tokenInterface.GetToken(db, token)
	if err != nil {
		return 0, errors.New("unauthorized")
	}
	if tokenReceived.ExpiryDate != nil {
		t := time.Now()
		if t.After(*tokenReceived.ExpiryDate) {
			return 0, errors.New("token expired")
		}
	}
	return uint(tokenReceived.UserId), nil
}
