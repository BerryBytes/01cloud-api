package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/externalsecret"
	"01cloud-api/api/utils/helper"
	"errors"
	"fmt"
	"io"
	"strconv"

	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func (server *Server) RetryExternalSecret(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	retryRequest := &externalsecret.RetryExternalSecretRequest{}
	err = json.Unmarshal(body, retryRequest)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	rawBytes, err := retryRequest.PrepareExternalSecretUpdate(environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment.ExternalSecret.RawMessage = rawBytes
	envUpdate, err := environment.UpdateExternalSecret(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	externalSecretRequest, err := externalsecret.PrepareExternalSecretRequest(envUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = externalSecretRequest.ValidateExternalSecretRequest()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = queue.Publish(constants.ExternalSecret, externalSecretRequest)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) SyncExternalSecret(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	if environment.ExternalSecret.RawMessage == nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("external secret is not enabled"))
		return
	}
	externalSecretRequest, err := externalsecret.PrepareExternalSecretRequest(environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = externalSecretRequest.ValidateExternalSecretRequest()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = queue.Publish(constants.ExternalSecretSync, externalSecretRequest)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) ExternalSecretStatus(input interface{}) error {
	data := &externalsecret.ExternalSecretResponse{}
	inputBytes := input.([]byte)
	err := json.Unmarshal(inputBytes, data)
	if err != nil {
		return err
	}
	if !data.Success {
		return fmt.Errorf("error on external secret :: %v", data.StatusMessage)
	}
	environment := models.Environment{}
	env, err := environment.Find(server.DB, uint64(data.EnvironmentId))
	if err != nil {
		log.Error("error fetch environment:: ", err)
		return err
	}
	err = server.publishForCICD(env, &models.CICDOptions{})
	if err != nil {
		_, _ = env.Delete(server.DB, uint64(data.EnvironmentId))
		return err
	}
	return nil
}

func (server *Server) ExternalSecretSyncStatus(input interface{}) error {
	data := &externalsecret.ExternalSecretResponse{}
	inputBytes := input.([]byte)
	err := json.Unmarshal(inputBytes, data)
	if err != nil {
		return err
	}
	if !data.Success {
		return fmt.Errorf("error on external secret sync :: %v", data.StatusMessage)
	}
	env := &models.Environment{}
	env, err = env.Find(server.DB, uint64(data.EnvironmentId))
	if err != nil {
		log.Error("error fetch environment:: ", err)
		return err
	}
	env.Action = "Updating"
	if env.ServiceType == 1 {
		wf, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(env))
		if err != nil {
			log.Error("error fetch ci workflow:: ", err)
			return err
		}
		workflow := helper.ConvertWorkflowType(wf)
		if workflow == nil {
			return errors.New("latest CI workflow not found")
		}
		env.CiRequest = &workflow.CIRequest
		env.RepositoryImage = &workflow.CIRequest.RepositoryImage
		if env.GitBranch == "" {
			return errors.New("git branch is requred for non-template")
		}
		repo, err := server.GetRepository(uint(env.Application.OwnerId), env.Application)
		if err != nil {
			return err
		}
		env.GitUrl = repo.CloneURL
		env.GitRepository = repo
	} else if env.ServiceType == 2 {
		gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, env.Application.OwnerId, env.Application.ImageService)
		if err != nil {
			return err
		}
		env, _ := env.ToJson()
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
	}
	err = queue.Publish(constants.UpgradeRelease, env)
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) GetExternalSecretActivityLog(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
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
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view logs"))
			return
		}
	}
	namespace := helper.GetNamespace(environmentReceived)
	if environmentReceived.ServiceType == 4 {
		namespace, err = server.NonGlobalOperatorNs(environmentReceived)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
	}
	status, err := server.StoreClient.Activity().GetActivityLog(namespace, "ExternalSecret", page, limit)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if status == nil {
		responses.ERROR(w, http.StatusNoContent, errors.New("no activity log available"))
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
