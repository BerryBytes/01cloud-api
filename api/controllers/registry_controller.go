package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/mailer"
	"01cloud-api/api/models"
	"01cloud-api/api/notifications"
	"01cloud-api/api/registry"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/helper"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

type Service struct {
	Service string `json:"service_name"`
}

type RegistryRepoSetting struct {
	Service string `json:"service_name"`
	Push    bool   `json:"push"`
}

// ConnectToRegistry godoc
// @Summary Connect to registry Provider
// @Description Connect user to registry provider
// @Tags Registry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.ExternalLogin true "Service Data"
// @Success 201 {object} doc.GitUser
// @Router /registry/connect [post]
func (server *Server) ConnectToRegistry(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	_, err = userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	externalLogin := models.ExternalLogin{}

	err = json.Unmarshal(body, &externalLogin)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	usr, err := registry.RegistryLogin(externalLogin)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	oldUser, _ := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), externalLogin.Service)
	usr.UserID = uint64(uid)
	if oldUser != nil {
		usr.ID = oldUser.ID
		oldUser, err = gitUserInterface.Update(server.DB, usr)
	} else {
		oldUser, err = gitUserInterface.Save(server.DB, usr)
	}
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uint(oldUser.UserID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user Not Found"))
		return
	}
	linkUnlinkEmail := mailer.SendLinkUnlinkEmail(user.Email, oldUser.ServiceName, true)
	err = notifications.NotifyEmail(server.NotifyClient, &linkUnlinkEmail)
	if err != nil {
		logrus.Error("error from notify email :: ", err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"user": oldUser,
		})
		return
	}
	resourceResponse, err := CreateGitUserResponse(oldUser)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"user": resourceResponse,
		},
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetRegistryUser(r *http.Request) (registry.RegistryInterface, *string, error) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		return nil, nil, err
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}
	var serviceMap Service
	err = json.Unmarshal(body, &serviceMap)
	if err != nil {
		return nil, nil, err
	}
	return server.GetRegistryUserFromUidName(uint64(uid), serviceMap.Service)
}

func (server *Server) GetRegistryUserFromUidName(uid uint64, serviceName string) (registry.RegistryInterface, *string, error) {
	var registryProvider registry.RegistryInterface
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		return nil, nil, err
	}
	switch serviceName {
	case "dockerhub":
		registryProvider, err = registry.NewDockerHubRegistrty(context.Background(), gtUsr.ServiceUserName, gtUsr.AccessToken)
		if err != nil {
			return nil, nil, err
		}
	}
	return registryProvider, &gtUsr.ServiceUserName, nil
}

// CreateWebhookForRegistryRepo godoc
// @Summary Dynamically Create Webhook for the Repo
// @Description Dynamically Create Webhook for the Repo
// @Tags Registry
// @Accept  json
// @Produce  json
// @Param namespace path string true "registry namespace"
// @Param repo path string true "repo name"
// @Security ApiKeyAuth
// @Param body body RegistryRepoSetting true "request"
// @Success 200 {object} string
// @Router /registry/{eid}/settings [post]
func (server *Server) SaveSettingForRegistryRepo(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	environmentReceived, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	metadata, err := server.StoreClient.ReleaseInfo().GetData(fmt.Sprintf("%s-metadata-%d", os.Getenv("GCLOUD_NAMESPACE"), environmentReceived.ID))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	var imageURL, namespace, repoWithTag, repo string

	if metadata != nil {
		var manifestRespones []models.Manifest
		js := models.ToJson(manifestRespones)

		image := gjson.Get(js, "#(kind==Deployment).spec.template.spec.containers.image")
		imageURL = image.String()
	}

	splittedImageURL := strings.Split(imageURL, "/")

	if len(splittedImageURL) == 0 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to split words"))
		return
	}

	namespace = splittedImageURL[1]
	repoWithTag = splittedImageURL[2]

	splittedImageRepo := strings.Split(repoWithTag, ":")

	if len(splittedImageURL) == 0 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to split words"))
		return
	}

	repo = splittedImageRepo[0]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var setting RegistryRepoSetting
	err = json.Unmarshal(body, &setting)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	client, _, err := server.GetRegistryUserFromUidName(uint64(uid), setting.Service)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	ctx := context.Background()

	if setting.Push {
		webhookURL := fmt.Sprintf("%v/webhook/image/%v/%d/%d", os.Getenv("API_URL"), setting.Service, environmentReceived.ID, uid)
		err = client.CreateWebhook(ctx, namespace, repo, namespace, webhookURL)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
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

// GetRegistryOrganization godoc
// @Summary Get all the Organization of the User
// @Description Get all the organization belongs to registry user
// @Tags Registry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body Service true "Service"
// @Success 201 {object} []registry.Organization
// @Router /registry/organizations [post]
func (server *Server) GetRegistryOrganizations(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	vars := mux.Vars(r)
	serviceName := vars["service"]
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("not authorized to registry, please connect to registry"))
		return
	}
	client, err := registry.NewRegistry(*gtUsr)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	ctx := context.Background()

	organizations, err := client.GetOrganizations(ctx, 100)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to get organizations"))
		return
	}

	resData := []map[string]string{}
	for _, b := range organizations {
		resData = append(resData, map[string]string{
			"name":  b.Name,
			"value": b.Name,
		})
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, resData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resData,
		Success: 1,
		Message: "Success",
	})
}

// GetRegistryRepos godoc
// @Summary Get Repos of the namespace
// @Description Get info of current registry user like namespace
// @Tags Registry
// @Accept  json
// @Produce  json
// @Param namespace path string true "registry namespace"
// @Security ApiKeyAuth
// @Param body body Service true "Service"
// @Param namespace query string true "Namespace"
// @Success 201 {object} []string
// @Router /registry/repos/{namespace} [post]
func (server *Server) GetRegistryRepos(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	vars := mux.Vars(r)
	namespace := r.URL.Query().Get("namespace")
	serviceName := vars["service"]
	if namespace == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("namespace is required"))
		return
	}
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("not authorized to registry, please connect to registry"))
		return
	}
	client, err := registry.NewRegistry(*gtUsr)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// client, username, err := server.GetRegistryUser(r)
	// if err != nil {
	// 	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	// 	return
	// }
	if namespace == "personal" {
		namespace = gtUsr.ServiceUserName
	}
	ctx := context.Background()
	repos, err := client.GetRepositories(ctx, namespace)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to get repositories"))
		return
	}

	resData := []map[string]string{}
	for _, b := range repos {
		resData = append(resData, map[string]string{
			"name":  b.Name,
			"value": b.Name,
		})
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, resData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resData,
		Success: 1,
		Message: "Success",
	})
}

// GetRegistryDetails godoc
// @Summary Get Details of Registry Repo
// @Description Get Details Info of the Registry Repo
// @Tags Registry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param namespace path string true "registry namespace"
// @Param repo path string true "repo name"
// @Param body body Service true "Service"
// @Success 201 {object} registry.Repository
// @Router /registry/repo/{namespace}/{repo} [post]
func (server *Server) GetRegistryDetails(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	vars := mux.Vars(r)
	namespace := r.URL.Query().Get("namespace")
	repo := r.URL.Query().Get("repo")
	serviceName := vars["service"]
	if namespace == "" && repo == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("both namespace and repo is required"))
		return
	}
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("not authorized to registry, please connect to registry"))
		return
	}
	if namespace == "personal" {
		namespace = gtUsr.ServiceUserName
	}
	client, err := registry.NewRegistry(*gtUsr)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	ctx := context.Background()
	repoDetails, err := client.GetRegistry(ctx, namespace, repo)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to get user details"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, repoDetails)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    repoDetails,
		Success: 1,
		Message: "Success",
	})
}

// GetRegistryRepoTags godoc
// @Summary Get Tags of the Repo
// @Description Get Details Info of the Registry Repo
// @Tags Registry
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param namespace path string true "registry namespace"
// @Param repo path string true "repo name"
// @Param uid query int true "User Id"
// @Param body body Service true "Service"
// @Success 201 {object} []map[string]interface{}
// @Router /registry/repo/{namespace}/{repo}/tags [post]
func (server *Server) GetRegistryRepoTags(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	vars := mux.Vars(r)
	namespace := r.URL.Query().Get("namespace")
	repo := r.URL.Query().Get("repo")
	serviceName := vars["service"]
	limit := 100
	user := r.URL.Query().Get("uid")
	if user != "" {
		userid, err := strconv.ParseUint(user, 10, 64)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		uid = uint(userid)
	}
	if namespace == "" && repo == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("both namespace and repo is required"))
		return
	}
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("not authorized to registry, please connect to registry"))
		return
	}
	client, err := registry.NewRegistry(*gtUsr)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if namespace == "personal" {
		namespace = gtUsr.ServiceUserName
	}
	ctx := context.Background()

	tags, err := client.GetTags(ctx, namespace, repo, limit)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to get tags on registry"))
		return
	}
	resData := []map[string]string{}
	for _, b := range tags {
		resData = append(resData, map[string]string{
			"name":  b.Name,
			"value": b.Name,
		})
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, resData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resData,
		Success: 1,
		Message: "Success",
	})

}

func (server *Server) GetRegistryFromUidName(uid uint64, serviceName, namespace, repo string) (*registry.Repository, error) {
	var registryProvider registry.RegistryInterface
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		return nil, err
	}
	repoData := &registry.Repository{}
	switch serviceName {
	case "aws_ecr":
		registryProvider, err = registry.NewEcr(*gtUsr)
		if err != nil {
			return nil, err
		}
		repoData, err = registryProvider.GetRegistry(context.Background(), namespace, repo)
		if err != nil {
			return nil, err
		}
	}
	return repoData, nil
}
