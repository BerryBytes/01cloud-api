package controllers

import (
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"01cloud-api/api/registry"
	"01cloud-api/api/websocket"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/externalsecret"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/git"
	"01cloud-api/api/utils/helper"
	"01cloud-api/api/utils/prometheus"

	"01cloud-api/api/utils/logging"

	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/jpillora/go-tld"
	"github.com/tidwall/gjson"
)

const (
	READ  = "read"
	WRITE = "write"
	ADMIN = "admin"
)

var environmentInterface = models.NewEnvironment()
var paymentInterface = models.NewPayment()

// CreateEnvironment godoc
// @Summary Create a new Environment
// @Description Create a new Environment with the input payload
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Environment true "Create Environment"
// @Success 201 {object} doc.Environment
// @Router /environment [post]
func (server *Server) CreateEnvironment(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment := models.Environment{}
	err = json.Unmarshal(body, &environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment.Prepare()
	err = environment.Validate(true)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = logging.ValidateLoggingRequest(environment.ExternalLogging)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	application, err := applicationInterface.Find(server.DB, environment.ApplicationID)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("invalid application"))
		return
	}
	if oid == 0 {
		if !paymentInterface.HasUserBalance(server.DB, uint(application.Project.UserID)) {
			responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to create environment due to remaining balance"))
			return
		}
	} else {
		org, err := orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, err)
			return
		}
		if !paymentInterface.HasUserBalance(server.DB, uint(org.UserID)) {
			responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to create environment due to remaining balance"))
			return
		}
	}

	if !application.Cluster.Active {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("cluster is disabled, Please contact support team"))
		return
	}
	if environment.IsNameExists(server.DB, uint(environment.ApplicationID), environment.Name) {
		responses.ERROR(w, http.StatusForbidden, errors.New("name already exists"))
		return
	}

	if uid != uint(application.Project.UserID) && !application.Project.User.IsAdmin {
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
			!authInterface.IsWriteAuthorizedProject(server.DB, uint64(uid), application.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to add environment"))
			return
		}
	}
	environment.Application = application
	if environment.AutoScaler.RawMessage != nil {
		var autoScaler *models.AutoScaler
		err = json.Unmarshal(environment.AutoScaler.RawMessage, &autoScaler)
		if err != nil {
			log.Error(err)
		}
		if !helper.IsEmptyStruct(autoScaler.HorizontalPodAutoScaler) {
			environment.Replicas = uint16(autoScaler.HorizontalPodAutoScaler.MaxReplicas)
		} else {
			environment.Replicas = 1
		}
		autoScaler.AdvancedScheduling.ValidateAdvancedSheduling()
		res, _ := json.Marshal(autoScaler)
		environment.AutoScaler.RawMessage = res
	} else {
		environment.Replicas = 1
	}
	resource, err := resourceInterface.Find(server.DB, environment.ResourceID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid resource"))
		return
	}
	if environment.LoadBalancerID > 0 {
		loadbalancer, err := loadbalancerInterface.Find(server.DB, environment.LoadBalancerID)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid loabalancer"))
			return
		}
		if loadbalancer.ClusterID != environment.Application.ClusterID {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid cluster, load balancer and environment must be in same region"))
			return
		}
		environment.LoadBalancer = loadbalancer
	}
	if environment.ServiceType < 2 || environment.ServiceType == 5 {
		if resource.Memory < environment.Application.Plugin.MinMemory {
			responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dmb memory per replica", environment.Application.Plugin.MinMemory))
			return
		}
		if resource.Cores < environment.Application.Plugin.MinCpu {
			responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dm cpu per replica", environment.Application.Plugin.MinCpu))
			return
		}
	} else {
		environment.ImageUrl = application.ImageUrl
	}
	environment.Resource = nil

	if message, ok := environment.IsValidResource(server.DB, environment.ApplicationID, nil, resource, 2048); !ok {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New(message))
		return
	}
	environment.Resource = resource

	pluginVersion, err := pluginVersionInterface.FindLatestPluginVersion(server.DB, environment.Application.PluginID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("latest plugin not available"))
		return
	}
	environment.PluginVersion = pluginVersion
	environment.PluginVersionID = uint64(pluginVersion.ID)
	var variablesMap = helper.GetDefaultVariable(pluginVersion)
	if environment.Variables.RawMessage != nil {
		var currentVariableMap map[string]interface{}
		err := json.Unmarshal(environment.Variables.RawMessage, &currentVariableMap)
		if err != nil {
			log.Error(err)
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		helper.ReplaceCustomVariables(currentVariableMap, variablesMap)
	}

	jsonString, _ := json.Marshal(variablesMap)
	environment.Variables.RawMessage = jsonString
	if environment.ServiceType == 1 {
		repo, err := server.GetRepository(uint(application.OwnerId), application)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environment.GitUrl = repo.CloneURL
		environment.GitRepository = repo
		if environment.Scripts.RawMessage != nil {
			scripts := &models.Script{}
			err := json.Unmarshal(environment.Scripts.RawMessage, scripts)
			if err != nil {
				log.Error("script unmarshall error :: ", err)
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			dockerfile, err := server.GetDockerfile(&environment)
			if err == nil {
				scripts.Dockerfile = string(dockerfile)
				environment.Scripts.RawMessage = helper.GetBytes(scripts)
			}
		}
		imageName := fmt.Sprintf("%s/%s/%s", environment.Application.Cluster.ImageRegistry.Service, environment.Application.Cluster.ImageRegistry.ProjectName, helper.GetNamespace(&environment))
		tag, _, err := server.GetTag(uint(environment.Application.OwnerId), &environment)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		attributes := map[string]interface{}{
			"repository_name": imageName,
			"repository_tag":  tag,
		}
		attrs, _ := json.Marshal(attributes)
		environment.Attributes.RawMessage = attrs
	}
	if !prometheus.ValidateClusterResource(server.DB, &environment) {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("CPU, Memory or Storage may be full in this region, please contact support team"))
		return
	}

	var userVariables []map[string]string
	err = json.Unmarshal(environment.UserVariables.RawMessage, &userVariables)
	if err == nil && !models.ValidateUserVariable(userVariables) {
		responses.ERROR(w, http.StatusForbidden, errors.New("you are not allowed to create duplicate variable"))
		return
	}
	environmentCreated, err := environment.Save(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusUnprocessableEntity, formattedError)
		return
	}
	// prevent_default_build block
	preventBuildStruct := struct {
		PreventDefaultBuild bool `json:"prevent_default_build"`
	}{}
	_ = json.Unmarshal(body, &preventBuildStruct)
	if environment.GitBranch == "" || !preventBuildStruct.PreventDefaultBuild {
		environmentCreated.Action = "Creating"
		if externalsecret.IsExternalSecretEnable(environmentCreated) {
			externalSecretRequest, err := externalsecret.PrepareExternalSecretRequest(environmentCreated)
			if err != nil {
				_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			err = externalSecretRequest.ValidateExternalSecretRequest()
			if err != nil {
				_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			err = queue.Publish(constants.ExternalSecret, externalSecretRequest)
			if err != nil {
				_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
				responses.ERROR(w, http.StatusInternalServerError, err)
				return
			}
		} else {
			err = server.publishForCICD(environmentCreated, &models.CICDOptions{})
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, err)
				_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
				return
			}
		}
	}
	if environmentCreated.ExternalLogging.RawMessage != nil && environment.ExternalLogging.RawMessage != nil {
		loggingReq, err := logging.PrepareLoggingRequest(environmentCreated)
		if err != nil {
			_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		err = queue.Publish(constants.ExternalLogging, loggingReq)
		if err != nil {
			_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
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
	notify(server.NotifyClient, notifyInfo, "created", fmt.Sprintf("environment %s created", environment.Name), "info", int64(uid))

	_, _ = server.SaveActivityWithJson("create", "environment", uid, application.Project, application, environmentCreated, nil, "")
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, environmentCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, environmentCreated)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environmentCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetEnvironmentsByEnvironment godoc
// @Summary Get Environment by application
// @Description Get list environments by application id.
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Application id"
// @Success 200 {array} doc.Environment
// @Router /application/{id}/environments [get]
func (server *Server) GetEnvironmentsByApplication(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	application, err := applicationInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if uid != uint(application.Project.UserID) && !application.Project.User.IsAdmin {
		if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
			!authInterface.IsWriteAuthorizedProject(server.DB, uint64(uid), application.ProjectID) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to get environment"))
			return
		}
	}
	if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) &&
		!authInterface.IsAuthorizedApplication(server.DB, uint64(uid), pid, application.ProjectID) &&
		application.Project.UserID != uint64(uid) && !user.IsAdmin {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("permission not granted"))
		return
	}
	environment := models.Environment{}
	environments, err := environment.FindAllByApplication(server.DB, pid, true, uint64(uid))
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	jsonData := []map[string]interface{}{}
	for _, env := range environments {
		jsonEnv := map[string]interface{}{}
		if helper.IsV2(r) {
			envResponse, err := CreateEnvironmentResponse(env)
			if err != nil {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			appByte, _ := json.Marshal(envResponse)
			err = json.Unmarshal(appByte, &jsonEnv)
			if err != nil {
				log.Error(err)
			}
		} else {
			jsonEnv, _ = env.ToJson()
		}
		jsonEnv["addons"], err = environment.FindAllAddons(server.DB, uint64(env.ID))
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		if env.GitBranch != "" {
			metadata, err := server.StoreClient.ReleaseInfo().GetData(fmt.Sprintf("%s-metadata-%d", os.Getenv("GCLOUD_NAMESPACE"), env.ID))
			if err != nil {
				log.Error(err)
			}
			if metadata != nil {
				jsonEnv["last_deployed"] = metadata.Info.LastDeployed
			}
		}
		namsespace := helper.GetNamespace(env)
		if environment.ServiceType == 4 {
			packageName := environment.Application.OperatorPackageName
			if environment.Application.Cluster != nil {
				opReq, _ := operatorRequestRepo.FindByNameAndClusterID(server.DB, environment.Application.Cluster.ClusterRequestID, packageName)
				namsespace = helper.GetOperatorNamespace(opReq)
			}
		}
		_, state, _, err := server.StoreClient.PodState().GetEnvironmentState(namsespace)
		if err != nil {
			state = "Pending"
			log.Errorf("environment state %d :: %v ", env.ID, err)
		}
		jsonEnv["status"] = state
		jsonData = append(jsonData, jsonEnv)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, jsonData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    jsonData,
		Success: 1,
		Message: "Success",
	})
}

// GetEnvironmentsByApplicationForAdmin godoc
// @Summary Get Environment by application
// @Description Get list environments from by application id. Only available for admin.
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Application id"
// @Success 200 {array} doc.Environment
// @Router /application/{id}/admin-env [get]
func (server *Server) GetEnvironmentsByApplicationAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment := models.Environment{}
	environments, err := environment.FindAllByApplicationAdmin(server.DB, aid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, environments)
		return
	}
	environmentResponseList := []doc.Environment{}
	for _, environment := range *environments {
		environmentResponse, err := CreateEnvironmentResponse(&environment)
		if err != nil {
			log.Error(err)
		}
		environmentResponseList = append(environmentResponseList, *environmentResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    environmentResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetEnvironmentVariables godoc
// @Summary Get environment variables by id
// @Description Get Environment variables by environment id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} doc.VariableResponse
// @Router /environment/{id}/variables [get]
func (server *Server) GetEnvironmentVariables(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment := models.Environment{}
	evars, err := environment.Find(server.DB, pid)

	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	//systemVariables := &map[string]interface{}{}
	//userVariables := []map[string]interface{}{}
	//_ = json.Unmarshal([]byte(evars.Variables), systemVariables)
	//_ = json.Unmarshal([]byte(evars.UserVariables), &userVariables)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"system_variables": evars.Variables,
			"user_variables":   evars.UserVariables,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"system_variables": evars.Variables,
			"user_variables":   evars.UserVariables,
		},
		Success: 1,
		Message: "Success",
	})
}

// GetEnvironmentInsights godoc
// @Summary Get Environment insight by id
// @Description Get Environment insight by id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} doc.InsightResponse
// @Router /environment/{id}/insights [post]
func (server *Server) GetEnvironmentInsights(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment := models.Environment{}
	env, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	request := models.Insight{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	namespace := helper.GetNamespace(env)
	request.Namespace = namespace
	res, err := prometheus.GetInsight(request, &environment)
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

// GetEnvironmentOverview godoc
// @Summary Get Environment overview by id
// @Description Get Environment overview by id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} doc.InsightResponse
// @Router /environment/{id}/overview [post]
func (server *Server) GetEnvironmentOverview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment := models.Environment{}
	env, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	request := models.Insight{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	namespace := helper.GetNamespace(env)
	request.Namespace = namespace
	res, err := prometheus.GetInsightOverview(request, &environment)
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

// GetHpaGraph godoc
// @Summary Get Environment hpa insight by id
// @Description Get Environment insight by id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} doc.HpaInsightResponse
// @Router /environment/{id}/hpa-insight [post]
func (server *Server) GetHpaGraph(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environment := models.Environment{}
	env, err := environment.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	request := models.Insight{}
	err = json.Unmarshal(body, &request)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	namespace := helper.GetNamespace(env)
	request.Namespace = namespace
	res, err := prometheus.GetHPAGraph(request, &environment)
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

// ReLaunchEnvironment godoc
// @Summary Rerun Environment by id
// @Description Rerun Environment by id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} doc.SuccessResponse
// @Router /environment/{id}/launch [get]
func (server *Server) ReLaunchEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	environment.Action = "Relaunching"
	err = server.publishForCICD(environment, &models.CICDOptions{
		ManualBuildAuthor: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
	})
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
	notify(server.NotifyClient, notifyInfo, "relaunched", fmt.Sprintf("environment %s relaunched", environment.Name), "info", int64(user.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": "Environment Re-deploy triggered",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func GetEnvironmentFromRequest(db *gorm.DB, r *http.Request) (*models.Environment, error) {
	vars := mux.Vars(r)
	environment := &models.Environment{}
	eid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		return environment, err
	}
	environmentReceived, err := environment.Find(db, eid)
	if err != nil {
		return environment, err
	}
	return environmentReceived, nil
}

func CheckPermission(db *gorm.DB, r *http.Request, role string, env *models.Environment) (*models.User, error) {
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(db, uid)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if env.Application.Project.UserID != uint64(uid) && !user.IsAdmin {
		if authInterface.IsAuthorizedOrganization(db, uid, oid) {
			return user, nil
		}
		if role == WRITE {
			if !authInterface.IsWriteAuthorizedEnvironment(db, uint64(uid), env) {
				return nil, errors.New("you are not authorized to write this environment")
			}
		} else if role == ADMIN {
			if !authInterface.IsAdminOfEnvironment(db, uint64(uid), env) {
				return nil, errors.New("you are not authorized to update this environment")
			}
		} else if role == READ {
			if !authInterface.IsAuthorizedEnvironment(db, uint64(uid), env) {
				return nil, errors.New("you are not authorized to access this environment")
			}
		}
	}
	return user, nil
}

func (server *Server) GetEnvironmentWithUserPermission(r *http.Request, mode string) (*models.Environment, *models.User, int, error) {
	environment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		return nil, nil, http.StatusUnprocessableEntity, err
	}
	if !environment.Active {
		return nil, nil, http.StatusForbidden, errors.New("environment is stopped ")
	}
	user, err := CheckPermission(server.DB, r, mode, environment)
	if err != nil {
		return nil, nil, http.StatusUnauthorized, err
	}

	return environment, user, 0, nil
}

// StopEnvironment godoc
// @Summary Stop Environment by id
// @Description Stop Environment by id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} doc.SuccessResponse
// @Router /environment/{id}/stop [post]
func (server *Server) StopEnvironment(w http.ResponseWriter, r *http.Request) {
	environmentReceived, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !environmentReceived.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment already stopped"))
		return
	}
	user, err := CheckPermission(server.DB, r, WRITE, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	environmentReceived.Active = false
	envs := &models.Environment{}
	_, err = envs.UpdateStatus(server.DB, uint64(environmentReceived.ID), environmentReceived.Active)
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	environmentReceived.Action = "Stopping"
	err = queue.Publish(constants.ActiveDeactiveEnvironment, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
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
	notify(server.NotifyClient, notifyInfo, "stopped", fmt.Sprintf("environment %s stopped", environmentReceived.Name), "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("stop", "environment", user.ID, environmentReceived.Application.Project, environmentReceived.Application, environmentReceived, nil, "")
	responses.JSON(w, http.StatusOK, &map[string]interface{}{
		"message": "Sent command for stopping environment",
	})

}

// StartEnvironment godoc
// @Summary Start Environment by id
// @Description Start Environment by id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} doc.SuccessResponse
// @Router /environment/{id}/start [post]
func (server *Server) StartEnvironment(w http.ResponseWriter, r *http.Request) {
	environmentReceived, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !paymentInterface.HasUserBalance(server.DB, uint(environmentReceived.Application.Project.UserID)) {
		responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to start environment due to remaining balance"))
		return
	}
	if environmentReceived.Active {
		responses.ERROR(w, http.StatusForbidden, errors.New("environment already started"))
		return
	}
	user, err := CheckPermission(server.DB, r, WRITE, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	environmentReceived.Active = true
	envs := &models.Environment{}
	_, err = envs.UpdateStatus(server.DB, uint64(environmentReceived.ID), environmentReceived.Active)
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	environmentReceived.Action = "Starting"
	err = queue.Publish(constants.ActiveDeactiveEnvironment, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
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
	notify(server.NotifyClient, notifyInfo, "started", fmt.Sprintf("environment %s started", environmentReceived.Name), "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("start", "environment", user.ID, environmentReceived.Application.Project, environmentReceived.Application, environmentReceived, nil, "")
	responses.JSON(w, http.StatusOK, &map[string]interface{}{
		"message": "Sent command for starting environment",
	})
}

func notify(conn *grpc.ClientConn, info notifications.BasicInfoNotification, action, body, types string, triggeredBy int64) {
	_, err := PublishNotificationBase(conn, &notifications.Notification{
		ApplicationName:         info.ApplicationName,
		ApplicationID:           info.ApplicationID,
		EnvironmentName:         info.EnvironmentName,
		EnvironmentID:           info.EnvironmentID,
		EnvironmentResourceName: info.EnvironmentResourceName,
		EnvironmentResourceID:   info.EnvironmentResourceID,
		ProjectName:             info.ProjectName,
		ProjectID:               info.ProjectID,
		OrganizationName:        info.OrganizationName,
		OrganizationID:          info.OrganizationID,
		TriggeredBy:             triggeredBy,
		Scope:                   "environment",
		Action:                  action,
		Type:                    types,
		Body:                    body,
		SendBy:                  "api",
	})
	if err != nil {
		log.Error(err)
	}
}

// GetEnvironment godoc
// @Summary Get Environment by id
// @Description Get Environment by id from token
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} doc.Environment
// @Router /environment/{id} [get]
func (server *Server) GetEnvironment(w http.ResponseWriter, r *http.Request) {
	environmentReceived, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	_, err = CheckPermission(server.DB, r, READ, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	if !environmentReceived.Application.Project.Active {
		responses.ERROR(w, http.StatusNotFound, errors.New("project is deactivated"))
		return
	}
	metadata, err := server.StoreClient.ReleaseInfo().GetData(fmt.Sprintf("%s-metadata-%d", os.Getenv("GCLOUD_NAMESPACE"), environmentReceived.ID))
	if err != nil {
		log.Error(err)
	}
	var overview []map[string]interface{}
	if environmentReceived.Application.Cluster != nil {
		overview = append(overview, map[string]interface{}{
			"name":  "Region",
			"value": environmentReceived.Application.Cluster.Region,
		})
	}
	if environmentReceived.ServiceType == 1 {
		wf, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(environmentReceived))
		if err == nil {
			workflow := helper.ConvertWorkflowType(wf)
			if workflow != nil {
				environmentReceived.CiRequest = &workflow.CIRequest
				environmentReceived.RepositoryImage = &workflow.CIRequest.RepositoryImage
			}
		}
		overview = append(overview, map[string]interface{}{
			"name":  "Git Url",
			"value": environmentReceived.GitUrl,
			"type":  "url",
		})
		overview = append(overview, map[string]interface{}{
			"name":  "Git Branch",
			"value": environmentReceived.GitBranch,
			"type":  "text",
		})
		subDir := "/"
		envScripts := map[string]interface{}{}
		err = json.Unmarshal(environmentReceived.Scripts.RawMessage, &envScripts)
		if err == nil {
			if s, ok := envScripts["sub_dir"].(string); ok {
				sc := strings.TrimLeft(s, "/")
				subDir += sc
			}
		}
		overview = append(overview, map[string]interface{}{
			"name":  "Sub Directory",
			"value": subDir,
			"type":  "text",
		})
	}
	if environmentReceived.ServiceType == 2 {
		overview = append(overview, map[string]interface{}{
			"name":  "Image Url",
			"value": environmentReceived.ImageUrl,
		})
		overview = append(overview, map[string]interface{}{
			"name":  "Image Tag",
			"value": environmentReceived.ImageTag,
		})
	}
	file, err := os.ReadFile(environmentReceived.PluginVersion.Url + "/config.json")
	if err == nil {
		var config = map[string]interface{}{}
		err = json.Unmarshal(file, &config)
		if err != nil {
			log.Error("config unmarshall error :: ", err)
		}
		var cname string
		metadata, err := server.StoreClient.ReleaseInfo().GetData(fmt.Sprintf("%s-metadata-%d", os.Getenv("GCLOUD_NAMESPACE"), environmentReceived.ID))
		if err != nil || metadata == nil {
			namespace := helper.GetNamespace(environmentReceived)
			domain := strings.TrimRight(environmentReceived.Application.Cluster.DNS.BaseDomain, ".")
			if domain != "" {
				cname = fmt.Sprintf("%s.%s", namespace, domain)
			}
		} else {
			cname = metadata.CName
		}
		var variables map[string]interface{}
		err = json.Unmarshal(environmentReceived.Variables.RawMessage, &variables)
		if err != nil {
			responses.JSON(w, http.StatusInternalServerError, err)
			return
		}
		data, err := helper.GetValues(config, "overview", "data")
		if err == nil {
			for _, newData := range helper.ConvertInterfaceToMapArray(data) {
				if newData["value_from"] == "cname" {
					newData["value"] = "https://" + cname + newData["value"].(string)
				} else if newData["value_from"] == "system_variable" {
					data, _ := helper.GetValues(variables, strings.Split(newData["value"].(string), ".")...)
					newData["value"] = data
				}
				delete(newData, "value_from")
				overview = append(overview, newData)
			}
		}
	}
	var metadataJson = &map[string]interface{}{}
	if metadata != nil {
		overview = append(overview, map[string]interface{}{
			"name":  "Last Deployed",
			"value": metadata.Info.LastDeployed,
			"type":  "time",
		})
		var manifests = strings.Split(metadata.Manifest, "---")
		var manifestRespones []models.Manifest
		for _, v := range manifests {
			manifestJson := &models.Manifest{}
			err := yaml.Unmarshal([]byte(v), manifestJson)
			if err == nil && (manifestJson.Kind == "Secret" || manifestJson.Kind == "Deployment") {
				manifestRespones = append(manifestRespones, *manifestJson)
			}
			if err != nil {
				log.Error("invalid manifest")
			}
		}
		js := models.ToJson(manifestRespones)
		secrets := gjson.Get(js, "#(kind==Secret)#.data")
		envs := gjson.Get(js, "#(kind==Deployment).spec.template.spec.containers")
		envs_map := new([]map[string]interface{})
		secret_map := new([]map[string]interface{})
		_ = json.Unmarshal([]byte(envs.Raw), envs_map)
		_ = json.Unmarshal([]byte(secrets.Raw), secret_map)
		metadataJson = &map[string]interface{}{
			"name":      metadata.Name,
			"namespace": metadata.Namespace,
			"version":   metadata.Version,
			"info":      metadata.Info,
			"releaseId": metadata.ReleaseId,
			//"cname":     metadata.CName,
			"secret": secret_map,
			"envs":   envs_map,
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"environment": environmentReceived,
			"metadata":    metadataJson,
			"overview":    overview,
		})
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"environment": environmentResponse,
			"metadata":    metadataJson,
			"overview":    overview,
		},
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetRevisionEnvironment(w http.ResponseWriter, r *http.Request) {
	environmentReceived, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	_, err = CheckPermission(server.DB, r, READ, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	metadata, err := server.StoreClient.ReleaseInfo().GetRevisonData(helper.GetNamespace(environmentReceived))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if err != nil {
		log.Error("error occured in get revison history :: ", err)
		responses.JSON(w, http.StatusOK, models.RevisionHistory{})
		return
	}
	RevisionList := []map[string]interface{}{}
	for _, data := range metadata.Release {
		jsonRev := map[string]interface{}{}
		jsonRev["username"] = environmentReceived.Application.Project.User.FirstName
		jsonRev["image_tag"] = data.Tag
		jsonRev["name"] = data.Name
		jsonRev["namespace"] = data.Namespace
		jsonRev["description"] = data.Description
		jsonRev["version"] = data.Version
		jsonRev["first_deployed"] = data.FirstDeployed
		jsonRev["last_deployed"] = data.LastDeployed
		jsonRev["status"] = data.Status
		RevisionList = append(RevisionList, jsonRev)
	}
	myList := []map[string]interface{}{}
	for i := (len(RevisionList) - 1); i >= 0; i-- {
		myList = append(myList, RevisionList[i])
	}
	log.Debug("Deployments::", myList)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"deployments": myList,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"deployments": myList,
		},
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) RevisionEnvironment(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, &map[string]interface{}{
		"message": "Publish envirnoment for deployment list",
	})
}

// UpdateEnvironment godoc
// @Summary Update a Environment
// @Description Update a Environment with the input payload
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.Environment true "Update Environment"
// @Success 200 {object} doc.Environment
// @Router /environment/{id} [put]
func (server *Server) UpdateEnvironment(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	extras := ""
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	environmentUpdate := models.Environment{}
	err = json.Unmarshal(body, &environmentUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	environmentUpdate.Prepare()
	environmentUpdate.Application = environment.Application
	if environmentUpdate.ResourceID != 0 {
		resource, err := resourceInterface.Find(server.DB, environmentUpdate.ResourceID)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid resource"))
			return
		}
		environmentUpdate.Resource = resource
		extras += fmt.Sprintf("resource to %d milli core and %d MB RAM ", resource.Cores, resource.Memory)
	} else {
		environmentUpdate.ResourceID = environment.ResourceID
		resource, err := resourceInterface.Find(server.DB, environment.ResourceID)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid resource"))
			return
		}
		environmentUpdate.Resource = resource
		if environmentUpdate.ServiceType < 2 {
			if environmentUpdate.Resource.Memory < environmentUpdate.Application.Plugin.MinMemory {
				responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dmb memory per replica", environment.Application.Plugin.MinMemory))
				return
			}
			if environmentUpdate.Resource.Cores < environmentUpdate.Application.Plugin.MinCpu {
				responses.ERROR(w, http.StatusUnprocessableEntity, fmt.Errorf("invalid resource, this plugin requires minimum %dm cpu per replica", environment.Application.Plugin.MinCpu))
				return
			}
		}
	}

	environmentUpdate.ID = environment.ID
	if environmentUpdate.AutoScaler.RawMessage != nil {
		var value models.AutoScaler
		err = json.Unmarshal(environmentUpdate.AutoScaler.RawMessage, &value)
		if err != nil {
			log.Error(err)
		}
		if !helper.IsEmptyStruct(value.HorizontalPodAutoScaler) {
			environmentUpdate.Replicas = uint16(value.HorizontalPodAutoScaler.MaxReplicas)
		} else {
			environmentUpdate.Replicas = 1
		}
	} else {
		environmentUpdate.Replicas = environment.Replicas
	}
	if environmentUpdate.Scripts.RawMessage != nil {
		err := environmentUpdate.UpdatePrepareScript(environment.Scripts)
		if err != nil {
			log.Error("script unmarshall error :: ", err)
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	}

	if message, ok := environment.IsValidResource(server.DB, environment.ApplicationID, &environmentUpdate, environmentUpdate.Resource, 0); !ok {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New(message))
		return
	}
	if environmentUpdate.Name != "" {
		if environmentUpdate.Name != environment.Name && environment.IsNameExists(server.DB, uint(environment.ApplicationID), environmentUpdate.Name) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("name already exists"))
			return
		}
		extras += fmt.Sprintf("name from '%s' to '%s' ", environment.Name, environmentUpdate.Name)
	}
	if !prometheus.ValidateClusterResource(server.DB, &environmentUpdate) {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("CPU, Memory or Storage may be full in this region, please contact support team"))
		return
	}
	environmentUpdate.ID = environment.ID
	environmentUpdated, err := environmentUpdate.Update(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	env := &models.Environment{}
	env, _ = env.Find(server.DB, uint64(environment.ID))
	if env.ServiceType == 1 {
		wf, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(env))
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, errors.New("CI workflow not available"))
			return
		}
		workflow := helper.ConvertWorkflowType(wf)
		if workflow != nil {
			env.CiRequest = &workflow.CIRequest
			env.RepositoryImage = &workflow.CIRequest.RepositoryImage
		} else {
			responses.ERROR(w, http.StatusInternalServerError, errors.New("CI workflow not available"))
			return
		}
		if env.GitBranch != "" {
			repo, err := server.GetRepository(uint(env.Application.OwnerId), env.Application)
			if err != nil {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			env.GitUrl = repo.CloneURL
			env.GitRepository = repo
		}
	}
	env.Action = "Updating"
	if env.ServiceType == 2 {
		gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), env.Application.ImageService)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		uenv, _ := env.ToJson()
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
		uenv["secrets"] = secrets
		err = queue.Publish(constants.UpgradeRelease, uenv)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else if env.ServiceType == 4 {
		envJ, _ := env.ToJson()
		packageName := env.Application.OperatorPackageName
		opReq := &models.OperatorRequest{}
		if env.Application.Cluster != nil {
			opReq, err = operatorRequestRepo.FindByNameAndClusterID(server.DB, env.Application.Cluster.ClusterRequestID, packageName)
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, errors.New("invalid service type"))
				return
			}
		}
		envJ["operators"] = map[string]interface{}{
			"package_name":    helper.GetOperatorNamespace(opReq),
			"global_operator": opReq.GlobalOperator,
		}
		err = queue.Publish(constants.UpdateOperatorApp, envJ)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		err = queue.Publish(constants.UpgradeRelease, env)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
	if environment.Replicas != env.Replicas {
		extras += fmt.Sprintf("replicas from %d to %d ", environment.Replicas, env.Replicas)
	}
	//if environment.Attributes != env.Attributes {
	//	extras += "environment variables "
	//}
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
	notify(server.NotifyClient, notifyInfo, "updated", fmt.Sprintf("environment %s updated with %s", environment.Name, extras), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, env.Application.Project, env.Application, env, nil, extras)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, environmentUpdated)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environmentUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

// RedeployEnvironment godoc
// @Summary Redeploy Environment by id
// @Description Redeploy Environment by id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.RedeployRequest true "Redeploy environment"
// @Success 200 {object} doc.SuccessResponse
// @Router /environment/{id}/re-deploy [post]
func (server *Server) RedeployEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input := map[string]interface{}{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	currentPluginVersion, err := pluginVersionInterface.Find(server.DB, environment.PluginVersionID)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("current plugin not available"))
		return
	}
	if currentPluginVersion.Upgradable {
		pluginVersion, err := pluginVersionInterface.FindLatestPluginVersion(server.DB, environment.Application.PluginID)
		if err != nil {
			log.Error("latest plugin not available")
			responses.ERROR(w, http.StatusNotFound, errors.New("latest plugin not available"))
			return
		}
		environment.PluginVersionID = uint64(pluginVersion.ID)
	}

	rawVersion, err := json.Marshal(input["version"])
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid version"))
		return
	}
	otherRawVersion, err := json.Marshal(input["other_version"])
	if err != nil {
		log.Error(err)
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid deployment versions"))
		return
	}
	environment.Version.RawMessage = rawVersion
	environment.OtherVersion.RawMessage = otherRawVersion

	if environment.ServiceType == 1 {
		wf, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(environment))
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, errors.New("workflow not available"))
			return
		}
		workflow := helper.ConvertWorkflowType(wf)
		if workflow != nil {
			if v, ok := workflow.Workflow["status"]; ok {
				status := v.(map[string]interface{})["phase"].(string)
				if status == "" || strings.ToLower(status) == "running" {
					responses.ERROR(w, http.StatusForbidden, errors.New("you cannot recreate the environment at this time. CI build is already running"))
					return
				}
			} else {
				responses.ERROR(w, http.StatusInternalServerError, errors.New("CI workflow not available"))
				return
			}

			environment.CiRequest = &workflow.CIRequest
			environment.RepositoryImage = &workflow.CIRequest.RepositoryImage
		} else {
			responses.ERROR(w, http.StatusInternalServerError, errors.New("CI workflow not available"))
			return
		}
		if environment.GitBranch != "" {
			repo, err := server.GetRepository(uint(environment.Application.OwnerId), environment.Application)
			if err != nil {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			environment.GitUrl = repo.CloneURL
			environment.GitRepository = repo
		}
	}
	_, err = environment.Update(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to recreate environment"))
		return
	}
	if environment.ServiceType == 1 {
		err = server.publishForCICD(environment, &models.CICDOptions{
			ManualBuildAuthor: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		})
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		env := environment
		req := models.CIRequest{
			ImageRepoService:  environment.Application.Cluster.ImageRegistry.Service,
			ImageRepoPassword: environment.Application.Cluster.ImageRegistry.Password,
			ImageRepoUsername: environment.Application.Cluster.ImageRegistry.Name,
			ImageRepoProject:  environment.Application.Cluster.ImageRegistry.ProjectName,
		}
		env.CiRequest = &req
		env.Action = "Redeploying"
		err = queue.Publish(constants.UpgradeRelease, env)
	}
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
	notify(server.NotifyClient, notifyInfo, "redeployed", fmt.Sprintf("environment %s redeployed", environment.Name), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, environment.Application.Project, environment.Application, environment, nil, "re-deployed")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, environment)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

// SetEnvironmentVariables godoc
// @Summary Set environment variables
// @Description Set environment variables in environment by id
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.VariableResponse true "Set environment variables"
// @Success 200 {object} doc.SuccessResponse
// @Router /environment/{id}/variables [post]
func (server *Server) SetEnvironmentVariables(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	input := map[string]interface{}{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	environmentUpdate := &models.Environment{}
	environmentUpdate.Variables = environment.Variables
	if sv, ok := input["system_variables"].(map[string]interface{}); ok {
		if len(sv) > 0 {
			systemInput, err := json.Marshal(sv)
			if err != nil {
				responses.ERROR(w, http.StatusUnprocessableEntity, err)
				return
			}
			environmentUpdate.Variables.RawMessage = systemInput
		}
	}
	log.Debug("system variables :: ", string(environmentUpdate.Variables.RawMessage))
	userInput, err := json.Marshal(input["user_variables"])
	if environment.ServiceType > 0 {
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		var userVariables []map[string]string
		_ = json.Unmarshal(userInput, &userVariables)
		if !models.ValidateUserVariable(userVariables) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("you are not allowed to create duplicate variable"))
			return
		}
		if len(userInput) > 0 {
			environmentUpdate.UserVariables.RawMessage = userInput
		}
	}
	if ai, ok := input["apply_immediately"]; ok {
		applyImmediately := ai.(bool)
		environmentUpdate.ApplyImmediately = applyImmediately
	} else {
		environmentUpdate.ApplyImmediately = environment.ApplyImmediately
	}
	environmentUpdate.ID = environment.ID
	_, err = environmentUpdate.UpdateVariables(server.DB)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	env, err := server.PreparePublishEnvironment(environmentUpdate.ID)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	env.Action = "Updating Variables"
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
	notify(server.NotifyClient, notifyInfo, "updated", fmt.Sprintf("environment %s variable updated", environment.Name), "info", int64(user.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, input)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    input,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) SetDNS(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	input := map[string]interface{}{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	systemInput, err := json.Marshal(input["system_variables"])
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	client := map[string]interface{}{}
	err = json.Unmarshal(systemInput, &client)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = environment.CheckDublicateDNS(server.DB, client["clients_fqdn"].(string))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	err = server.CheckDNS(*environment, client, user.ID)
	if err != nil {
		responses.ERROR(w, http.StatusAccepted, err)
		return
	}
	domainInput, err := server.SetSecondaryDNS(client)
	if err != nil {
		log.Error(err)
	}
	err = server.SaveDNS(*environment, domainInput, user.ID)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, input)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    input,
		Success: 1,
		Message: "Success",
	})
}

// DeleteEnvironment godoc
// @Summary Delete a Environment
// @Description Delete a Environment with the input payload
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 204 {object} doc.Environment
// @Router /environment/{id} [delete]
func (server *Server) DeleteEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user, err := CheckPermission(server.DB, r, ADMIN, environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	environment.Action = "Deleting"
	if environment.Application.ServiceType == 4 && environment.OperatorPayload.RawMessage != nil {
		env, _ := environment.ToJson()
		packageName := environment.Application.OperatorPackageName
		if environment.Application.Cluster != nil {
			opReq, err := operatorRequestRepo.FindByNameAndClusterID(server.DB, environment.Application.Cluster.ClusterRequestID, packageName)
			if err == nil {
				env["operators"] = map[string]interface{}{
					"package_name":    helper.GetOperatorNamespace(opReq),
					"global_operator": opReq.GlobalOperator,
				}
			}
		}
		err = queue.Publish(constants.UnInstallOperatorApp, env)
	} else {
		err = queue.Publish(constants.DestroyRelease, environment)
	}
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if environment.ServiceType == 1 {
		if environment.GitBranch != "" {
			repo, err := server.GetRepository(uint(environment.Application.OwnerId), environment.Application)
			if err == nil {
				environment.GitUrl = repo.CloneURL
				environment.GitRepository = repo
				conf, err := ciConfigRepo.FindByEnvironment(server.DB, environment.ID)
				if err == nil {
					if len(conf.HookId) > 0 {
						err = server.DeleteWebhook(environment, conf.HookId, r)
						if err != nil {
							log.Error(err)
						}
					}
				}
			}
		}
	}
	go helper.SendAlertDeleteRequest(environment.ID)
	go server.DeleteScanReport(environment)
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
	notify(server.NotifyClient, notifyInfo, "deleted", fmt.Sprintf("environment %s deleted", environment.Name), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("delete", "environment", user.ID, environment.Application.Project, environment.Application, environment, nil, "")
	time.Sleep(time.Duration(time.Second))
	_, err = environment.Delete(server.DB, uint64(environment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// RollbackEnvironment godoc
// @Summary Rollback to the provided deployment
// @Description Rollback your Environment to by choosing any previously build images and pass that information in body
// @Tags Rollback
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.RepositoryImage true "Deployment Rollback"
// @Success 200 {object} doc.Environment
// @Router /environment/{id}/rollback [post]
func (server *Server) RollbackEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
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
	if environment.ServiceType < 1 {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("rollback not available for template environment "))
		return
	}
	wf, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(environment))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("workflow not available"))
		return
	}
	workflow := helper.ConvertWorkflowType(wf)
	if workflow != nil {
		if v, ok := workflow.Workflow["status"]; ok {
			status := v.(map[string]interface{})["phase"].(string)
			if status == "" || strings.ToLower(status) == "running" {
				responses.ERROR(w, http.StatusForbidden, errors.New("you cannot rollback the environment at this time. CI build is already running "))
				return
			}
		} else {
			responses.ERROR(w, http.StatusInternalServerError, errors.New("rollback not available "))
			return
		}
		image.Repository = workflow.CIRequest.RepositoryImage.Repository
		workflow.CIRequest.RepositoryImage = image
		environment.CiRequest = &workflow.CIRequest
		environment.RepositoryImage = &image
	} else {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("no image for rollback "))
		return
	}
	attributes := map[string]interface{}{
		"repository_name": image.Name,
		"repository_tag":  image.Tag,
	}
	attrs, _ := json.Marshal(attributes)
	environment.Attributes.RawMessage = attrs
	_, err = environment.Update(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to rollback environment "))
		return
	}
	environment.Action = "Rolling back"
	err = queue.Publish(constants.UpgradeRelease, environment)
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
	notify(server.NotifyClient, notifyInfo, "rollback", fmt.Sprintf("environment %s rollback", environment.Name), "info", int64(user.ID))

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, environment.Application.Project, environment.Application, environment, nil, "rollback")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, environment)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}
func (server *Server) SetSecondaryDNS(client map[string]interface{}) ([]byte, error) {
	clients_fqdn := client["clients_fqdn"].(string)
	if strings.Contains(clients_fqdn, "www.") {
		domain := strings.Replace(clients_fqdn, "www.", "", 4)
		client["clients_fqdn_secondary"] = domain
	} else {
		url, _ := tld.Parse(clients_fqdn)
		if url.Subdomain == "" && url.Domain != "" {
			client["clients_fqdn_secondary"] = fmt.Sprintf("www.%s", clients_fqdn)
		}
	}
	variables, err := json.Marshal(client)
	if err != nil {
		return nil, err
	}
	return variables, nil
}
func (server *Server) CheckDNS(environment models.Environment, client map[string]interface{}, userId uint) error {
	clients_fqdn := client["clients_fqdn"].(string)
	err := ResolveIP(clients_fqdn)
	if err == nil {
		return nil
	}
	client["clients_fqdn_status"] = "pending"
	client["clients_fqdn_temp"] = clients_fqdn
	client["clients_fqdn"] = ""
	variables, err := json.Marshal(client)
	if err != nil {
		log.Error(err)
	}
	_, err = server.VariablesUpdate(environment, variables)
	if err != nil {
		log.Error(err)
	}
	go server.CheckRoutine(environment, client, userId)
	return errors.New(" domain currently not available, retrying again, check back in few minutes")
}

func ResolveIP(domain string) error {
	_, err := net.ResolveIPAddr("ip4:icmp", domain)
	if err != nil {
		return err
	}
	//TODO check for ip match with loadbalancer
	return nil
}

func (server *Server) CheckRoutine(environment models.Environment, client map[string]interface{}, userId uint) {

	wsConn, mutex := websocket.WebsocketConn(fmt.Sprintf("env-%d", environment.ID))
	clients_fqdn := client["clients_fqdn_temp"].(string)
	for i := 0; ; i++ {
		websocket.EmitMessage("pending", helper.GetNamespace(&environment), "env", wsConn, mutex)
		time.Sleep(1 * time.Minute)

		err := ResolveIP(clients_fqdn)
		if err == nil {
			client["clients_fqdn"] = clients_fqdn
			client["clients_fqdn_status"] = "processing"
			systemInput, _ := server.SetSecondaryDNS(client)
			_ = server.SaveDNS(environment, systemInput, userId)
			websocket.EmitMessage("processing", helper.GetNamespace(&environment), "env", wsConn, mutex)
			return
		}
		if i >= 3 {
			log.Error(err)
			break
		}
	}
	client["clients_fqdn_status"] = "failed"
	variables, err := json.Marshal(client)
	if err != nil {
		log.Error(err)
	}
	_, err = server.VariablesUpdate(environment, variables)
	if err != nil {
		log.Error(err)
	}
	websocket.EmitMessage("failed", helper.GetNamespace(&environment), "env", wsConn, mutex)
}

func (server *Server) VariablesUpdate(environment models.Environment, systemInput []byte) (*models.Environment, error) {
	if len(systemInput) > 0 {
		environment.Variables.RawMessage = systemInput
	}
	_, err := environment.UpdateVariables(server.DB)
	if err != nil {
		return nil, err
	}
	return &environment, nil
}

func (server *Server) SaveDNS(environment models.Environment, systemInput []byte, userId uint) error {
	var err error
	_, err = server.VariablesUpdate(environment, systemInput)
	if err != nil {
		return err
	}
	env, err := server.PreparePublishEnvironment(environment.ID)
	if err != nil {
		return err
	}
	env.Action = "Updating Custom Domain"
	err = queue.Publish(constants.UpgradeRelease, env)
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
	notify(server.NotifyClient, notifyInfo, "updated", fmt.Sprintf("environment %s variable updated", environment.Name), "info", int64(userId))
	return nil
}

func (server *Server) NonGlobalOperatorNs(input *models.Environment) (string, error) {
	var ns string
	if input.Application.Cluster != nil {
		opReq, err := operatorRequestRepo.FindByNameAndClusterID(server.DB, input.Application.Cluster.ClusterRequestID, input.Application.OperatorPackageName)
		if err != nil {
			log.Info("operator not found")
			return "", err
		}
		if !opReq.GlobalOperator {
			ns = fmt.Sprintf("%s-%d", opReq.PackageName, opReq.ID)
		}
	}
	return ns, nil
}

// EnableDisableFileManager godoc
// @Summary Enable Disable FileManager
// @Description Enabale and Disable FileManager
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param status query string true "Status"
// @Success 200 {string} string
// @Router /environment/{id}/enable-disable-filemanager [post]
func (server *Server) EnableDisableFileManager(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	status := r.URL.Query().Get("status")
	fileManagerStatus, err := strconv.ParseBool(status)
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("status must be specified"))
		return
	}
	if (fileManagerStatus && environment.FileManagerEnabled != nil) || (!fileManagerStatus && environment.FileManagerEnabled == nil) {
		responses.ERROR(w, http.StatusBadRequest, errors.New("curent status and previous status can't be same"))
		return
	}
	environment.FileManagerEnabled = nil

	if fileManagerStatus {
		currentTime := time.Now()
		environment.FileManagerEnabled = &currentTime
	}
	_, err = environmentInterface.UpdateFileManagerStatus(server.DB, environment)
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	err = queue.Publish(constants.EnableDisableFileManager, environment)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
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
	var notifyFilemanager string
	if environment.FileManagerEnabled == nil {
		notifyFilemanager = "disable"
	} else {
		notifyFilemanager = "enable"
	}
	if environment.Application.Project.Organization != nil {
		notifyInfo.OrganizationName = environment.Application.Project.Organization.Name
	}
	notify(server.NotifyClient, notifyInfo, notifyFilemanager, fmt.Sprintf("environment %s filemanager "+notifyFilemanager+"d", environment.Name), "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, environment.Application.Project, environment.Application, environment, nil, "with filemanager "+notifyFilemanager+"d")
	responses.JSON(w, http.StatusOK, &map[string]interface{}{
		"message": "Sent command for file manager",
	})
}

// ScheduleStartStopEnvironment godoc
// @Summary Schedule Start Stop Environment
// @Description Schedule Start and Stop of Environment
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param type query string true "Type"
// @Success 200 {string} string
// @Router /environment/{id}/start-stop [post]
func (server *Server) ScheduleStartStopEnvironment(w http.ResponseWriter, r *http.Request) {
	log.Info("schedule env processed")
	scheduleType := r.URL.Query().Get("type")
	log.Debug("schedule-type :: ", scheduleType)
	if scheduleType == "" {
		responses.ERROR(w, http.StatusNotFound, errors.New("required schedule type"))
		return
	}
	action := ""
	environmentReceived, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	log.Debug("environment received :: ", environmentReceived)
	if scheduleType == "stop" {
		log.Info("stop initiated")
		if !environmentReceived.Active {
			responses.ERROR(w, http.StatusForbidden, errors.New("environment already stopped"))
			return
		}
		environmentReceived.Active = false
		environmentReceived.Action = "Stopping"
		action = scheduleType + "ped"
	} else if scheduleType == "start" {
		if !paymentInterface.HasUserBalance(server.DB, uint(environmentReceived.Application.Project.UserID)) {
			responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to start environment due to remaining balance"))
			return
		}
		if environmentReceived.Active {
			responses.ERROR(w, http.StatusForbidden, errors.New("environment already started"))
			return
		}
		environmentReceived.Active = true
		environmentReceived.Action = "Starting"
		action = scheduleType + "ed"
	} else {
		responses.ERROR(w, http.StatusForbidden, errors.New("invalid schedule type"))
		return
	}
	envs := &models.Environment{}
	_, err = envs.UpdateStatus(server.DB, uint64(environmentReceived.ID), environmentReceived.Active)
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	err = queue.Publish(constants.ActiveDeactiveEnvironment, environmentReceived)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
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
	notify(server.NotifyClient, notifyInfo, scheduleType, fmt.Sprintf("environment %s %s", environmentReceived.Name, action), "info", 0)
	_, _ = server.SaveActivityWithJson(scheduleType, "environment", 0, environmentReceived.Application.Project, environmentReceived.Application, environmentReceived, nil, "")
	responses.JSON(w, http.StatusOK, &map[string]interface{}{
		"message": "Sent command for " + scheduleType + " environment",
	})
}

func (server *Server) ChangeBranchEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	git := models.GitRepository{}
	err = json.Unmarshal(body, &git)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if environment.ServiceType == 1 {
		err = git.Validate(environment)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environment.GitBranch = git.Branch
		_, err = server.VerifyRepoAndBranch(environment)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		repo, err := server.GetRepository(uint(environment.Application.OwnerId), environment.Application)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environment.GitUrl = repo.CloneURL
		environment.GitRepository = repo
	} else {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("only for git provider"))
		return
	}
	_, err = environment.Update(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	environment.Action = "Updating"
	err = server.publishForCICD(environment, &models.CICDOptions{})
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

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, environment.Application.Project, environment.Application, environment, nil, "update")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, environment)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) ChangeTagEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
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
	environment.Action = "Updating"
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

	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, environment.Application.Project, environment.Application, environment, nil, "update")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, environment)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) VerifyRepoAndBranch(input *models.Environment) (*models.GitUser, error) {
	service := input.Application.GitService
	repoId := input.Application.GitUrl
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, input.Application.OwnerId, service)
	if err != nil {
		return nil, err
	}
	client, err := git.NewGit(*gtUsr)
	if err != nil {
		return nil, err
	}
	_, err = client.GetRepoDetails(repoId)
	if err != nil {
		return nil, err
	}
	branches, err := client.GetBranches(repoId)
	if err != nil {
		return nil, err
	}
	for _, branch := range branches {
		if branch.Value == input.GitBranch {
			return gtUsr, nil
		}
	}
	return nil, errors.New("invalid git repo and branch")
}

func (server *Server) VerifyRepoAndTags(input *models.Environment) (*models.GitUser, error) {
	service := input.Application.ImageService
	namespace := input.Application.ImageNamespace
	repo := input.Application.ImageRepo
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, input.Application.OwnerId, service)
	if err != nil {
		return nil, err
	}
	client, err := registry.NewRegistry(*gtUsr)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if namespace == "personal" {
		namespace = gtUsr.ServiceUserName
	}
	_, err = client.GetRegistry(ctx, namespace, repo)
	if err != nil {
		return nil, err
	}
	tags, err := client.GetTags(ctx, namespace, repo, 100)
	if err != nil {
		return nil, err
	}
	if len(tags) > 0 {
		for _, tag := range tags {
			if tag.Name == input.ImageTag {
				return gtUsr, nil
			}
		}
	}
	return nil, errors.New("invalid registry repo and tag")
}

func CreateEnvironmentResponse(environment *models.Environment) (*doc.Environment, error) {
	environmentesponse := doc.Environment{}
	environmentBytes, _ := json.Marshal(environment)
	err := json.Unmarshal(environmentBytes, &environmentesponse)
	if err != nil {
		return nil, err
	}
	if environmentesponse.ExternalSecret != nil {
		externalSecret, _ := helper.ToJson(environmentesponse.ExternalSecret)
		delete(externalSecret, "aws_credential")
		delete(externalSecret, "gcp_credential")
		delete(externalSecret, "vault_credential")
		environmentesponse.ExternalSecret = externalSecret
	}
	if environmentesponse.ExternalLogging != nil {
		externalLogger, _ := helper.ToJson(environmentesponse.ExternalLogging)

		if _, ok := externalLogger["aws_credential"]; ok {
			delete(externalLogger["aws_credential"].(map[string]interface{}), "access_key")
			delete(externalLogger["aws_credential"].(map[string]interface{}), "access_key")
		}
		if _, ok := externalLogger["elastic_credential"]; ok {
			delete(externalLogger["elastic_credential"].(map[string]interface{}), "password")
			delete(externalLogger["elastic_credential"].(map[string]interface{}), "host")
		}
		if _, ok := externalLogger["gcp_credential"]; ok {
			delete(externalLogger["gcp_credential"].(map[string]interface{}), "key_file")
		}
		environmentesponse.ExternalLogging = externalLogger
	}
	return &environmentesponse, nil

}

func CreateHemEnvironmentResponse(environment *models.HelmEnvironment) (*doc.Environment, error) {
	environmentesponse := doc.Environment{}
	environmentBytes, _ := json.Marshal(environment)
	err := json.Unmarshal(environmentBytes, &environmentesponse)
	if err != nil {
		return nil, err
	}
	return &environmentesponse, nil
}

func (server *Server) PreparePublishEnvironment(id uint) (*models.Environment, error) {
	env := &models.Environment{}
	env, _ = env.Find(server.DB, uint64(id))
	if env.ServiceType == 1 {
		wf, err := server.StoreClient.CIWOrkflow().GetLatestWorkflow(helper.GetCiNamespace(env))
		if err != nil {
			return nil, err

		}
		workflow := helper.ConvertWorkflowType(wf)
		if workflow == nil {
			return nil, errors.New("ci workflow not available")
		}
		env.CiRequest = &workflow.CIRequest
		env.RepositoryImage = &workflow.CIRequest.RepositoryImage
		if env.GitBranch != "" {
			repo, err := server.GetRepository(uint(env.Application.OwnerId), env.Application)
			if err == nil {
				env.GitUrl = repo.CloneURL
				env.GitRepository = repo
			}
		}
	} else if env.ServiceType == 2 {
		gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(env.Application.OwnerId), env.Application.ImageService)
		if err != nil {
			return nil, err
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
	} else if env.ServiceType == 4 {
		packageName := env.Application.OperatorPackageName
		opReq := &models.OperatorRequest{GlobalOperator: true}
		if env.Application.Cluster != nil {
			opReq, _ = operatorRequestRepo.FindByNameAndClusterID(server.DB, env.Application.Cluster.ClusterRequestID, packageName)
			env, _ := env.ToJson()
			env["operators"] = map[string]interface{}{
				"id":              opReq.ID,
				"package_name":    helper.GetOperatorNamespace(opReq),
				"global_operator": opReq.GlobalOperator,
			}
		}
	}
	return env, nil
}

// UpdateEnvironment godoc
// @Summary Update Environment IP Whitelist
// @Description Update a Environment with the input payload for ip whitelisting
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body models.WhiteListedIP true "Update Environment ip whitelisting"
// @Success 200 {object} doc.Environment
// @Router /environment/{id}/ip-whitelist [put]
func (server *Server) UpdateEnvironmentWhiteListedIP(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	extras := ""
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	whiteListedIps := &models.WhitelistedIP{}
	err = json.Unmarshal(body, &whiteListedIps)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	whiteListedIps.PrepareWhiteListedIP()
	if err := whiteListedIps.ValidateWhiteListedIP(); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	env := &models.Environment{}
	env, _ = env.Find(server.DB, uint64(environment.ID))
	updatedEnv, err := env.UpdateWhiteListedIPs(server.DB, uint64(environment.ID), whiteListedIps.WhitelistedIPs)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := env.Find(server.DB, uint64(environment.ID))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	err = queue.Publish(constants.WhitelistedIPS, data)
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
	notify(server.NotifyClient, notifyInfo, "updated", fmt.Sprintf("environment %s updated with %s", environment.Name, extras), "info", int64(user.ID))
	_, _ = server.SaveActivityWithJson("update", "environment", user.ID, env.Application.Project, env.Application, env, nil, extras)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, updatedEnv)
		return
	}
	environmentResponse, err := CreateEnvironmentResponse(updatedEnv)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    environmentResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateEnvironment godoc
// @Summary Update Environment ExternalURL
// @Description Update a Environment with the input payload for external url
// @Tags Environment
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body models.ExternalURLRequest true "Update Environment external url"
// @Success 200 {object} doc.Environment
// @Router /environment/{id}/external-url [put]
func (server *Server) UpdateEnvironmentExternalURL(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}

	// extras := ""
	if environment.ServiceType == 5 {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		externalURL := models.ExternalURLRequest{}
		err = json.Unmarshal(body, &externalURL)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		environment.ExternalURL = externalURL.ExternalURL
		_, err = environment.Update(server.DB)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			responses.ERROR(w, http.StatusInternalServerError, formattedError)
			return
		}

		env := &models.Environment{}
		env, _ = env.Find(server.DB, uint64(environment.ID))
		err = queue.Publish(constants.EnvironmentExternalURL, env)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}

		fmt.Println("_____Updated external url___", env.ExternalURL)
		environmentResponse, err := CreateEnvironmentResponse(env)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}

		fmt.Println("_____Updated external url resp___", environmentResponse.ExternalURL)
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    environmentResponse,
			Success: 1,
			Message: "Success",
		})
	}

}
