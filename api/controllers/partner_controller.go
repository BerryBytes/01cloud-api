package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/mailer"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func (server *Server) CreateCampaignData(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("PARTNER-TOKEN") != os.Getenv("PARTNER-TOKEN") {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("invalid token"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
	}
	campaign := models.Campaign{}
	err = json.Unmarshal(body, &campaign)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	plugin, err := pluginInterface.Find(server.DB, uint64(campaign.PluginId))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	subscription, err := subscriptionInterface.Find(server.DB, uint64(campaign.SubscriptionId))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, getErrorMessage("subscription", "Invalid subscription"))
		return
	}
	var regions *[]string
	if os.Getenv("USE_DEFAULT_REGION") == "true" {
		regions, err = clusterInterface.FindAllRegions(server.DB, 0)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		regions, err = clusterInterface.FindAllRegions(server.DB, campaign.OrganizationId)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
	var cluster *models.Cluster
	if len(*regions) > 0 {
		cluster, err = clusterInterface.FindByRegion(server.DB, (*regions)[0], campaign.OrganizationId)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("region not found"))
			return
		}
	} else {
		responses.ERROR(w, http.StatusNotFound, errors.New("region not found"))
		return
	}
	pluginVersion, err := pluginVersionInterface.FindLatestPluginVersion(server.DB, uint64(campaign.PluginId))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	file, _ := os.ReadFile(pluginVersion.Url + "/versions.json")
	var res []map[string]interface{}
	var version postgres.Jsonb
	err = json.Unmarshal(file, &res)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if len(res) > 0 {
		if resValue, ok := res[0]["versions"]; ok {
			value := resValue.([]interface{})
			if len(value) > 0 {
				valueData := value[0].(map[string]interface{})
				jsonData, _ := json.Marshal(valueData)
				version = postgres.Jsonb{RawMessage: json.RawMessage(jsonData)}
			}
		}
	}
	password := auth.GeneratePassword()
	var userCreated *models.User
	isUser := true
	userCreated, err = userInterface.FindUserByEmail(server.DB, campaign.User.Email)
	if err != nil {
		isUser = false
		campaign.User.Password = password
		campaign.User.EmailVerified = true
		campaign.User.Reference = campaign.Reference
		userCreated, err = userInterface.SaveUser(server.DB, campaign.User)
		if err != nil {
			formattedError := formaterror.FormatError(err.Error())
			responses.ERROR(w, http.StatusInternalServerError, formattedError)
			return
		}
	}
	if campaign.OrganizationId == 0 && isUser {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("user already exists"))
		return
	}
	project := models.Project{
		Name:           campaign.Name + "-" + campaign.User.FirstName + "-" + strconv.FormatUint(uint64(userCreated.ID), 10),
		SubscriptionID: uint64(campaign.SubscriptionId),
		Subscription:   subscription,
		User:           userCreated,
		UserID:         uint64(userCreated.ID),
		Active:         true,
		OrganizationId: uint64(campaign.OrganizationId),
	}
	var organization *models.Organization
	if campaign.OrganizationId != 0 {
		organization, err = orgInterface.Find(server.DB, campaign.OrganizationId)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
			return
		}
		member := &models.OrganizationMembers{}
		member.UserID = uint64(userCreated.ID)
		member.OrganizationID = uint64(campaign.OrganizationId)
		if orgMember.CheckLimit(server.DB, organization.OrganizationPlan, campaign.OrganizationId) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("user limit exceeds"))
			return
		}
		if campaign.CampaignType == "new-user" {
			if orgMember.CheckMember(server.DB, member) {
				responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("user already exists"))
				return
			}
			err = orgMember.AddMember(server.DB, member)
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, err)
				return
			}
		} else {
			if !orgMember.CheckMember(server.DB, member) {
				err = orgMember.AddMember(server.DB, member)
				if err != nil {
					responses.ERROR(w, http.StatusInternalServerError, err)
					return
				}
			}
		}
		project.UserID = organization.UserID
	}
	projectCreated, err := iproject.Save(server.DB, &project)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	dataAuth := models.Authorization{}
	if campaign.OrganizationId != 0 {
		dataAuth = models.Authorization{
			Email:      campaign.User.Email,
			ProjectID:  uint64(projectCreated.ID),
			UserRoleID: 2,
			UserID:     uint64(userCreated.ID),
		}
		_, err = authInterface.Save(server.DB, &dataAuth)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
	application := models.Application{
		Name:      plugin.Name + "-application",
		ProjectID: uint64(projectCreated.ID),
		Project:   projectCreated,
		PluginID:  uint64(campaign.PluginId),
		Plugin:    plugin,
		OwnerId:   project.UserID,
	}
	application.ClusterID = uint64(cluster.ID)
	application.Cluster = cluster
	applicationCreated, err := applicationInterface.Save(server.DB, &application)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	environment := models.Environment{
		Name:            plugin.Name + "-env",
		ApplicationID:   uint64(applicationCreated.ID),
		Application:     applicationCreated,
		ResourceID:      1,
		Replicas:        1,
		PluginVersionID: uint64(pluginVersion.ID),
		PluginVersion:   pluginVersion,
		Version:         version,
		Active:          true,
	}
	var variablesMap = helper.GetDefaultVariable(pluginVersion)
	jsonString, _ := json.Marshal(variablesMap)
	environment.Variables.RawMessage = jsonString
	environmentCreated, err := environment.Save(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = server.publishForCICD(environmentCreated, &models.CICDOptions{})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		_, _ = environmentCreated.Delete(server.DB, uint64(environmentCreated.ID))
		return
	}
	var email notifications.SendEmailMessage
	if !isUser && campaign.OrganizationId != 0 {
		email = mailer.SendPartnerOrgNewUserEmail(userCreated.Email, uint64(environmentCreated.ID), password, organization.Name)
	} else if isUser {
		email = mailer.SendPartnerOrgExistingUserEmail(userCreated.Email, uint64(environmentCreated.ID), organization.Name)
	} else {
		email = mailer.SendPartnerNewUserEmail(userCreated.Email, uint64(environmentCreated.ID), password)
	}
	err = notifications.NotifyEmail(server.NotifyClient, &email)
	if err != nil {
		log.Error("error from notify email :: ", err)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetSubscriptionsForPartner(w http.ResponseWriter, r *http.Request) {
	oid, err := strconv.ParseUint(r.URL.Query().Get("org_id"), 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	subscription := models.Subscription{}
	_, err = orgInterface.Find(server.DB, uint(oid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	subscription.OrganizationID = uint64(oid)
	subscriptions, err := subscription.FindAllWithInActive(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	subscriptionResponseList := []doc.Subscription{}
	for _, item := range *subscriptions {
		subscriptionResponse, err := CreateSubscriptionResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		subscriptionResponseList = append(subscriptionResponseList, *subscriptionResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    subscriptionResponseList,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetPluginsForPartner(w http.ResponseWriter, r *http.Request) {
	oid, err := strconv.ParseUint(r.URL.Query().Get("org_id"), 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	organization, err := orgInterface.Find(server.DB, uint(oid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	pluginResponseList := []doc.Plugin{}
	for _, item := range organization.Plugins {
		pluginResponse, err := CreatePluginResponse(item)
		if err != nil {
			log.Error(err)
			continue
		}
		pluginResponseList = append(pluginResponseList, *pluginResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginResponseList,
		Success: 1,
		Message: "Success",
	})
}
