package controllers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/git"
	"01cloud-api/api/utils/helper"

	smodel "github.com/berrybytes/01cloud-store/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var ciConfigRepo = models.NewCiConfigRepo()

// ReRunCICD godoc
// @Summary ReRun CICD
// @Description ReRun the CI/CD process
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Id"
// @Success 200 {string} string
// @Router /environment/{id}/rerun-ci [get]
func (server *Server) ReRunCICD(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	//nodeName := r.FormValue("workflow_name")
	//if nodeName == "" {
	//	responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("Workflow name is required"))
	//	return
	//}
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
	if environmentReceived.GitBranch != "" {
		repo, err := server.GetRepository(uint(environmentReceived.Application.OwnerId), environmentReceived.Application)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environmentReceived.GitUrl = repo.CloneURL
		environmentReceived.GitRepository = repo
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to rerun this environment"))
			return
		}
	}
	buildCount, err := server.StoreClient.CIWOrkflow().GetWorkflowCount(helper.GetCiNamespace(environmentReceived), "", "")
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if buildCount >= int(environment.Application.Project.Subscription.CiBuild) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("CiBuild limit exceeded"))
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
		notify(server.NotifyClient, notifyInfo, "quota exceeded", fmt.Sprintf("project %s build limit exceeded", environment.Application.Project.Name), "info", int64(uid))

		return
	}

	err = server.publishForCICD(environmentReceived, &models.CICDOptions{
		ManualBuildAuthor: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "***  Running environment ***",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetWorkflows godoc
// @Summary Get Workflows
// @Description Get Workflows
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment Id"
// @Param page query int true "Page"
// @Param limit query int true "limit"
// @Success 200 {object} map[string]interface{}
// @Router /environment/{id}/workflow [get]
func (server *Server) GetWorkflows(w http.ResponseWriter, r *http.Request) {
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
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view workflow"))
			return
		}
	}

	status, err := server.StoreClient.CIWOrkflow().GetWorkflowDetail(helper.GetCiNamespace(environmentReceived), int(page), int(limit))
	if err != nil {
		status = []*smodel.WorkflowMetadata{{}}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"page":  page + 1,
			"limit": limit,
			"data":  status,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]interface{}{
			"page":  page + 1,
			"limit": limit,
			"data":  status,
		},
		Success: 1,
		Message: "Success",
	})
}

// GetWorkflowLog godoc
// @Summary Get Workflows Log
// @Description Get Workflows Log
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Workflows Logs"
// @Param workflow_name query string true "Workflow Name"
// @Success 200 {object} map[string]interface{}
// @Router /environment/{id}/workflow-log [get]
func (server *Server) GetWorkflowLog(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	workflowName := r.FormValue("workflow_name")
	if workflowName == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("workflow name is required"))
		return
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
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view this workflow"))
			return
		}
	}

	status, _ := server.StoreClient.CIWOrkflow().GetWorkflowLog(workflowName)
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

// StopCIBuild godoc
// @Summary Stop CI Build
// @Description Stop CI Build
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Stop CI Build"
// @Param workflow_name query string true "Workflow Name"
// @Success 200 {object} map[string]interface{}
// @Router /environment/{id}/stop-ci [get]
func (server *Server) StopCIBuild(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	workFlowName := r.FormValue("workflow_name")
	if workFlowName == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("workflow name is required"))
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
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if environmentReceived.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to rerun this environment"))
			return
		}
	}
	req := models.CIRequest{
		EnvironmentId: int64(environmentReceived.ID),
		Namespace:     helper.GetCiNamespace(environmentReceived),
		ConfigPath:    environmentReceived.Application.Cluster.ConfigPath,
		Name:          workFlowName,
	}
	mp := map[string]interface{}{
		"message": "Success",
	}
	if err := queue.Publish(constants.StopWorkflow, req); err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("failed to stop CI pipeline "+err.Error()))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, mp)
		return
	}

	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    mp,
		Success: 1,
		Message: "Success",
	})
}

// GetCiTriggerConfig godoc
// @Summary Get Ci Trigger Config
// @Description Get CI Trigger Config
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get ci-trigger config"
// @Success 200 {object} doc.CiConfig
// @Router /environment/{id}/ci-trigger [get]
func (server *Server) GetCiTriggerConfig(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}

	environment := &models.Environment{}
	environment, err = environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("environment not found"))
		return
	}

	if uid != uint(environment.Application.Project.UserID) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environment) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update this application"))
			return
		}
	}
	if environment.Application.ServiceType != 1 {
		responses.ERROR(w, http.StatusBadRequest, errors.New("feature only available for ci projects"))
		return
	}

	config, err := ciConfigRepo.FindByEnvironment(server.DB, environment.ID)

	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, config)
		return
	}
	ciConfigResponse, err := CreateCiConfigResponse(config)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    ciConfigResponse,
		Success: 1,
		Message: "Success",
	})
}

// TriggerCi godoc
// @Summary Trigger CI
// @Description Trigger CI
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Id"
// @Param body body doc.CiConfig true "Trigger CI"
// @Success 200 {object} doc.CiConfig
// @Router /environment/{id}/ci-trigger [post]
func (server *Server) TriggerCi(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	ciConfig := models.CiConfig{}
	ciConfig.EnvironmentID = uint(pid)
	err = json.Unmarshal(body, &ciConfig)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	environment := &models.Environment{}
	environment, err = environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New(" environment not found "))
		return
	}
	if !environment.Active {
		responses.ERROR(w, http.StatusBadRequest, errors.New(" environment is in stopped state "))
		return
	}

	if uid != uint(environment.Application.Project.UserID) && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environment) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New(" you are not authorized to update this application "))
			return
		}
	}
	if environment.Application.ServiceType != 1 {
		responses.ERROR(w, http.StatusBadRequest, errors.New(" feature only available for ci projects "))
		return
	}
	repo, err := server.GetRepository(uint(environment.Application.OwnerId), environment.Application)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment.GitUrl = repo.CloneURL
	environment.GitRepository = repo
	var hook *git.Hook
	var response *http.Response
	config, _ := ciConfigRepo.FindByEnvironment(server.DB, environment.ID)
	if config == nil || len(config.HookId) == 0 {
		hook, response, err = server.CreateWebhook(environment.Application.OwnerId, *environment, ciConfig)
	} else if len(config.HookId) > 0 {
		hook, response, err = server.UpdateWebhook(environment.Application.OwnerId, *environment, ciConfig, config.HookId)
	}

	if err != nil {
		statusCode := 500
		if response != nil {
			statusCode = response.StatusCode
		}
		responses.ERROR(w, statusCode, err)
		return
	}
	if hook != nil {
		ciConfig.HookId = hook.ID
	} else {
		ciConfig.HookId = ""
	}
	if config == nil {
		ciConfig.Validate()
		config, err = ciConfigRepo.Save(server.DB, &ciConfig)
	} else {
		ciConfig.ID = config.ID
		ciConfig.EnvironmentID = config.EnvironmentID
		config, err = ciConfigRepo.Update(server.DB, &ciConfig)
	}
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, config)
		return
	}
	ciConfigResponse, err := CreateCiConfigResponse(config)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    ciConfigResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) publishForCICD(environmentCreated *models.Environment, options *models.CICDOptions) error {
	var uid uint = uint(environmentCreated.Application.OwnerId)
	if environmentCreated.ServiceType == 0 || environmentCreated.ServiceType == 5 {
		return queue.Publish(constants.CreateRelease, environmentCreated)
	} else if environmentCreated.ServiceType == 2 {
		gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), environmentCreated.Application.ImageService)
		if err != nil {
			return err
		}
		env, _ := environmentCreated.ToJson()
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
		return queue.Publish(constants.CreateRelease, env)
	} else if environmentCreated.ServiceType == 1 {
		if environmentCreated.GitRepository == nil {
			repo, err := server.GetRepository(uint(environmentCreated.Application.OwnerId), environmentCreated.Application)
			if err != nil {
				return err
			}
			environmentCreated.GitRepository = repo
		}
		tag, gtUsr, err := server.GetTag(uid, environmentCreated)
		if err != nil {
			return err
		}
		gitI, err := git.NewGit(*gtUsr)
		if err != nil {
			gitI, err = git.ExpiredToken(*gtUsr, server.DB)
			if err != nil {
				return err
			}
		}
		tagValue := tag.SHA
		if options.TagName != "" {
			tagValue = options.TagName
		}
		imageName := fmt.Sprintf("%s/%s/%s", environmentCreated.Application.Cluster.ImageRegistry.Service, environmentCreated.Application.Cluster.ImageRegistry.ProjectName, helper.GetNamespace(environmentCreated))
		if environmentCreated.Application.Cluster.ImageRegistry.Provider == constants.EKS {
			imageName = helper.GetNamespace(environmentCreated)
		}
		repoInfo := models.RepositoryImage{
			Name:          imageName,
			Repository:    helper.GetNamespace(environmentCreated),
			Tag:           tagValue,
			CommitMessage: tag,
		}
		var versionMap map[string]interface{}
		err = json.Unmarshal(environmentCreated.Version.RawMessage, &versionMap)
		if err != nil {
			log.Error(err)
			return err
		}
		author := tag.Author
		if options.ManualBuildAuthor != "" {
			author = options.ManualBuildAuthor
		}
		scripts := &models.Script{}
		if environmentCreated.Scripts.RawMessage != nil {
			_ = json.Unmarshal(environmentCreated.Scripts.RawMessage, scripts)
		}
		if scripts.Dockerfile == "" {
			dockerfile, err := server.GetDockerfile(environmentCreated)
			if err == nil {
				scripts.Dockerfile = string(dockerfile)
				environmentCreated.Scripts.RawMessage = helper.GetBytes(scripts)
			}
			err = environmentCreated.UpdateEnvironmentScript(server.DB)
			if err != nil {
				log.Error("environment update script error :: ", err)
			}
		}
		gitUrl := environmentCreated.GitUrl
		ciFiles := gitI.GetWorkflowContent(environmentCreated)
		//if environmentCreated.Application.GitService == "gitlab" && gtUsr.IsOauth {
		//	gitUrl = strings.Replace(gitUrl, "gitlab", "oauth2:"+gtUsr.AccessToken + "@gitlab", 1)
		//}

		scripts.Dockerfile = strings.ReplaceAll(scripts.Dockerfile, "$BASE_IMAGE", versionMap["repo"].(string))
		scripts.Dockerfile = strings.ReplaceAll(scripts.Dockerfile, "$BASE_TAG", versionMap["tag"].(string))
		scripts.Dockerfile = base64.StdEncoding.EncodeToString([]byte(scripts.Dockerfile))
		runScript := ""
		if len(scripts.CIVariables) > 0 {
			ciVariablesBytes, _ := json.Marshal(scripts.CIVariables)
			runScript = string(ciVariablesBytes)
		}
		req := models.CIRequest{
			EnvironmentId:     int64(environmentCreated.ID),
			Namespace:         helper.GetCiNamespace(environmentCreated),
			ConfigPath:        environmentCreated.Application.Cluster.ConfigPath,
			ImageRepoService:  environmentCreated.Application.Cluster.ImageRegistry.Service,
			ImageRepoPassword: environmentCreated.Application.Cluster.ImageRegistry.Password,
			ImageRepoUsername: environmentCreated.Application.Cluster.ImageRegistry.UserName,
			ImageRepoProject:  environmentCreated.Application.Cluster.ImageRegistry.ProjectName,
			ImageRepoProvider: environmentCreated.Application.Cluster.ImageRegistry.Provider,
			Name:              options.WorkflowName,
			RepositoryImage:   repoInfo,
			GitUrl:            gitUrl,
			GitAccessToken:    gtUsr.AccessToken,
			GitUserName:       gtUsr.ServiceUserName,
			SubDirectory:      scripts.SubDir,
			GitBranch:         environmentCreated.GitBranch,
			CommitMessage:     tag.Message,
			Author:            author,
			PluginUrl:         environmentCreated.PluginVersion.Url,
			BaseImage:         versionMap["repo"].(string),
			BaseTag:           versionMap["tag"].(string),
			BuildScript:       scripts.Dockerfile,
			RunScript:         runScript,
			CISteps:           scripts.CiSteps,
			CloneEnv:          environmentCreated.CloneEnvironment,
			CiFiles: &map[string]interface{}{
				"ci_files": &ciFiles,
			},
		}
		return queue.Publish(constants.BuildWorkflow, req)
	} else if environmentCreated.ServiceType == 4 {
		env, _ := environmentCreated.ToJson()
		packageName := environmentCreated.Application.OperatorPackageName
		if environmentCreated.Application.Cluster != nil {
			opReq, _ := operatorRequestRepo.FindByNameAndClusterID(server.DB, environmentCreated.Application.Cluster.ClusterRequestID, packageName)
			env["operators"] = map[string]interface{}{
				"id":              opReq.ID,
				"package_name":    helper.GetOperatorNamespace(opReq),
				"global_operator": opReq.GlobalOperator,
			}
		}
		return queue.Publish(constants.InstallOperatorApp, env)
	} else {
		return errors.New("invalid service type")
	}
}

// WebhookTrigger godoc
// @Summary Webhook Trigger
// @Description Trigger Webhook
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param eid path int true "eId"
// @Param userid path int true "User Id"
// @Success 200 {string} string
// @Router /webhook/{eid}/{userId} [post]
func (server *Server) WebhookTrigger(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["eid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, err := strconv.ParseUint(vars["userId"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	payload := map[string]interface{}{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment := models.Environment{}
	environmentReceived, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	tag := ""
	branch := ""
	if !environmentReceived.Active {
		responses.ERROR(w, http.StatusBadRequest, errors.New("environment is in stopped state"))
		return
	}
	if environmentReceived.GitBranch != "" {
		repo, err := server.GetRepository(uint(environmentReceived.Application.OwnerId), environmentReceived.Application)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environmentReceived.GitUrl = repo.CloneURL
		environmentReceived.GitRepository = repo
		tag, branch, err = git.ParseWebHookPayload(payload, *repo, environmentReceived)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
		}
		if branch == "" {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid branch"))
			return
		}
		if branch != environment.GitBranch {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid payload"))
			return
		}
	}
	user, err := userInterface.FindUserByID(server.DB, uint(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if environmentReceived.Application.Project.UserID != uid && !user.IsAdmin {
		if !authInterface.IsWriteAuthorizedEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to rerun this environment"))
			return
		}
	}

	err = server.publishForCICD(environmentReceived, &models.CICDOptions{
		TagName: tag,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "***  Running environment ***",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetMetrics godoc
// @Summary Get Metrics
// @Description Get Metrics
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Metrics"
// @Success 200 {object} map[string]interface{}
// @Router /environment/{id}/ci-metrics [get]
func (server *Server) GetMetrics(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
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
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view workflow"))
			return
		}
	}
	status, err := server.StoreClient.CIMetric().GetCiMetrics(helper.GetCiNamespace(environmentReceived))
	if err != nil || status == nil {
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    map[string]interface{}{},
			Success: 1,
			Message: "Success",
		})
		return
	}
	resp := map[string]interface{}{
		"aborted":   status["aborted"],
		"not_build": status["not_build"],
		"success":   status["success"],
		"unstable":  status["unstable"],
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, resp)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resp,
		Success: 1,
		Message: "Success",
	})
}

// GetCIPipelineLogs godoc
// @Summary Get CI Pipeline Logs
// @Description Get CI Pipeline Logs
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get CI Pipeline Logs"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Router /environment/{id}/{pipeline}/{task}/{step} [get]
func (server *Server) GetCIPipelineStepLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	pipeline := vars["pipeline"]
	task := vars["task"]
	step := vars["step"]
	err = models.ValidatePipeline(pipeline, task, step)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	env := models.Environment{}
	environment, err := env.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	data, err := server.StoreClient.CIWOrkflow().GetWorkflow(pipeline)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("pipeline not found"))
		return
	}
	jsonbody, err := json.Marshal(data)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("invalid pipeline data"))
		return
	}
	meta := models.WorkflowMetadata{}
	if err := json.Unmarshal(jsonbody, &meta); err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("invalid pipeline data"))
		return
	}
	log.Debugf("Data Pipeline ::%v ", meta.Pipeline)
	isTask := meta.Pipeline.IsTaskAvailable(task)
	if !isTask {
		responses.ERROR(w, http.StatusNotFound, fmt.Errorf("%s task not found in %s pipeline", task, pipeline))
		return
	}
	isStep := meta.Pipeline.IsStepAvailable(task, step)
	if !isStep {
		responses.ERROR(w, http.StatusNotFound, fmt.Errorf("%s step not found in %s task", step, task))
		return
	}
	namespace := helper.GetCiNamespace(environment)
	pipelineLogFilePATH := os.Getenv("PIPELINE_LOG_PATH")
	if pipelineLogFilePATH == "" {
		pipelineLogFilePATH = "/data/logs/pipeline"
	}
	logFilePath := fmt.Sprintf("%s/%s/%s/%s/%s.log", pipelineLogFilePATH, namespace, pipeline, task, step)
	w.Header().Set("Content-Type", "text/plain")
	file, err := os.Open(logFilePath)
	if err != nil {
		emptyLog := []byte("unavailable logs")
		_, _ = w.Write(emptyLog)
		return
	}
	defer file.Close()
	_, err = io.Copy(w, file)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("failed to serve pipeline step log file"))
		return
	}
}

// DeleteWorkflow godoc
// @Summary Delete Workflow
// @Description Delete Workflow
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "id"
// @Param wid path int true "Workflow Id"
// @Success 200
// @Router /environment/{id}/workflow/{wid} [delete]
func (server *Server) DeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	if _, ok := vars["wid"]; !ok {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	wid := vars["wid"]
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
		if !authInterface.IsAdminOfEnvironment(server.DB, uint64(uid), environmentReceived) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to delete workflow"))
			return
		}
	}

	err = server.StoreClient.CIWOrkflow().DeleteWorkflow(helper.GetCiNamespace(environmentReceived), wid)
	if err != nil {
		responses.ERROR(w, http.StatusNoContent, errors.New("no data available"))
		return
	}

	w.Header().Set("Entity", wid)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetBuildImages godoc
// @Summary Get Build Images
// @Description Get your previously build images of your Environment for rollback deployment
// @Tags Rollback
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param page query int true "page"
// @Param limit query int true "limit"
// @Success 200 {object} object
// @Router /environment/{id}/build-images [get]
func (server *Server) GetBuildImages(w http.ResponseWriter, r *http.Request) {
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
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view workflow"))
			return
		}
	}
	status, err := server.StoreClient.CIWOrkflow().GetBuildImages(helper.GetCiNamespace(environmentReceived), int(page), int(limit))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if status == nil {
		responses.ERROR(w, http.StatusNoContent, errors.New("no data available"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"page":  page + 1,
			"limit": limit,
			"data":  status,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]interface{}{
			"page":  page + 1,
			"limit": limit,
			"data":  status,
		},
		Success: 1,
		Message: "Success",
	})
}

// TestNotification godoc
// @Summary Test Notification
// @Description Test Notification
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param source path string true "Test Notification"
// @Success 200 {string} string
// @Router /environment/{id}/test/{source} [get]
func (server *Server) TestNotification(w http.ResponseWriter, r *http.Request) {
	environment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !environment.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment is stopped"))
		return
	}
	_, err = CheckPermission(server.DB, r, "write", environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	vars := mux.Vars(r)
	if d, ok := vars["source"]; !ok && d != "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("source not defined "))
		return
	}
	server.testNotification(environment.ID, vars["source"])
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"message": "Test message sent to " + vars["source"],
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func CreateCiConfigResponse(ciConfig *models.CiConfig) (*doc.CiConfig, error) {
	ciConfigResponse := doc.CiConfig{}
	ciConfigBytes, _ := json.Marshal(ciConfig)
	err := json.Unmarshal(ciConfigBytes, &ciConfigResponse)
	if err != nil {
		return nil, err
	}
	return &ciConfigResponse, nil
}

// AfterClone godoc
// @Summary ReRun CICD
// @Description ReRun the CI/CD process
// @Tags CI
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Id"
// @Success 200 {string} string
// @Router /environment/{id}/after-clone [post]
func (server *Server) AfterClone(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	var mapFiles map[string]interface{} = map[string]interface{}{}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if len(body) != 0 {
		if string(body[0]) == "{" {
			err = json.Unmarshal(body, &mapFiles)
			if err != nil {
				log.Error("can't unmarshall ci_files:: ", err)
			}
		}
	}
	mapFiles["cloned"] = true
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
	wf, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(environmentReceived))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("CI workflow not available"))
		return
	}
	workflow := helper.ConvertWorkflowType(wf)
	if workflow != nil && workflow.CloneEnv != nil {
		environmentReceived.CloneEnvironment = workflow.CloneEnv
	}
	if environmentReceived.GitBranch != "" {
		repo, err := server.GetRepository(uint(environmentReceived.Application.OwnerId), environmentReceived.Application)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environmentReceived.GitUrl = repo.CloneURL
		environmentReceived.GitRepository = repo
	}
	user, err := userInterface.FindUserByID(server.DB, uint(environmentReceived.Application.Project.UserID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	buildCount, err := server.StoreClient.CIWOrkflow().GetWorkflowCount(helper.GetCiNamespace(environmentReceived), "", "")
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if buildCount >= int(environment.Application.Project.Subscription.CiBuild) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("CiBuild limit exceeded"))
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
		notify(server.NotifyClient, notifyInfo, "quota exceeded", fmt.Sprintf("project %s build limit exceeded", environment.Application.Project.Name), "info", int64(user.ID))

		return
	}

	err = server.publishForCICD(environmentReceived, &models.CICDOptions{
		ManualBuildAuthor: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "***  Running environment ***",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
