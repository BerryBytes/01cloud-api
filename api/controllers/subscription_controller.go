package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var subscriptionInterface = models.NewSubscription()

// CreateSubscription godoc
// @Summary Create a new subscription
// @Description Create a new subscription with the input payload
// @Tags Subscription
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Subscription true "Create Subscription"
// @Success 201 {object} doc.Subscription
// @Router /subscription [post]
func (server *Server) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	subscription := models.Subscription{}
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	subscription.Prepare()
	err = subscription.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	// checks for duplicate entries
	if subscription.IsNameExists(server.DB, oid, subscription.Name) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("subscription name already exists "))
		return
	}

	if oid == 0 {
		subscription.UserID = uint64(user.ID)
	}
	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	if oid > 0 {
		organization, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}

	subscription.OrganizationID = uint64(oid)
	if oid > 0 && !subscription.CheckSubscriptionResourceLimit(server.DB, organization) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource limit exceeds "))
		return
	}
	subscriptionCreated, err := subscriptionInterface.Save(server.DB, subscription)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if subscription is created within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "create", "subscription", subscriptionCreated.Name, "")
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, subscriptionCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, subscriptionCreated)
		return
	}
	subscriptionResponse, err := CreateSubscriptionResponse(subscriptionCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    subscriptionResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) CreateSubscriptionForAdmin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	subscription := models.Subscription{}
	err = json.Unmarshal(body, &subscription)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	subscription.Prepare()
	err = subscription.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	// checks for duplicate entries
	if subscription.IsNameExists(server.DB, oid, subscription.Name) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("subscription name already exists "))
		return
	}

	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	if oid > 0 {
		organization, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}

	subscription.OrganizationID = uint64(oid)
	if oid > 0 && !subscription.CheckSubscriptionResourceLimit(server.DB, organization) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource limit exceeds "))
		return
	}
	subscriptionCreated, err := subscriptionInterface.Save(server.DB, subscription)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if subscription is created within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "create", "subscription", subscriptionCreated.Name, "")
	}

	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, subscriptionCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, subscriptionCreated)
		return
	}
	subscriptionResponse, err := CreateSubscriptionResponse(subscriptionCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    subscriptionResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetSubscriptions godoc
// @Summary Get Subscriptions
// @Description Get list of subscriptions
// @Tags Subscription
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.Subscription
// @Router /subscriptions [get]
func (server *Server) GetSubscriptions(w http.ResponseWriter, r *http.Request) {
	_, oid, _ := auth.ExtractTokenID(r)

	subscriptions, err := subscriptionInterface.FindAll(server.DB, uint64(oid))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, subscriptions)
		return
	}
	subscriptionResponseList := []doc.Subscription{}
	for _, item := range *subscriptions {
		subscriptionResponse, err := CreateSubscriptionResponse(&item)
		if err != nil {
			logrus.Error(err)
			continue
		}
		subscriptionResponseList = append(subscriptionResponseList, *subscriptionResponse)
	}
	resp := responses.Response{
		Data:    subscriptionResponseList,
		Success: 1,
		Message: "Success",
	}
	// server.Cache.Set(key, resp)
	responses.JSON(w, http.StatusOK, resp)
}

func (server *Server) GetSubscriptionsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, _ := strconv.ParseUint(vars["id"], 10, 64)
	private := r.URL.Query().Get("private")
	var subscriptions *[]models.Subscription
	var err error
	if private == "true" {
		subscriptions, err = subscriptionInterface.FindByUserID(server.DB, uint64(uid))
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		subscriptions, err = subscriptionInterface.FindAllByUserID(server.DB, uint64(uid))
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
	subscriptionResponseList := []doc.Subscription{}
	for _, item := range *subscriptions {
		subscriptionResponse, err := CreateSubscriptionResponse(&item)
		if err != nil {
			logrus.Error(err)
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

// GetSubscriptionsForAdmin godoc
// @Summary Get Subscriptions by organization
// @Description Get list of subscriptions by organization
// @Tags Subscription
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.Subscription
// @Router /admin/subscriptions [get]
func (server *Server) GetSubscriptionsForAdmin(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	// key := fmt.Sprintf("subscriptions-admin-%d-%d", uid, oid)
	// var value interface{}
	// if ok := server.Cache.Get(key, &value); ok {
	// 	responses.JSON(w, http.StatusOK, value)
	// 	return
	// }
	subscription := models.Subscription{}
	subscription.OrganizationID = uint64(oid)
	subscriptions, err := subscription.FindAllWithInActive(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, subscriptions)
		return
	}
	subscriptionResponseList := []doc.Subscription{}
	for _, item := range *subscriptions {
		subscriptionResponse, err := CreateSubscriptionResponse(&item)
		if err != nil {
			logrus.Error(err)
			continue
		}
		subscriptionResponseList = append(subscriptionResponseList, *subscriptionResponse)
	}
	// server.Cache.Set(key, subscriptionResponseList)
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    subscriptionResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetSubscription godoc
// @Summary Get Subscription by id
// @Description Get Subscription by id from token
// @Tags Subscription
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Subscription id"
// @Success 200 {object} doc.Subscription
// @Router /subscription/{id} [get]
func (server *Server) GetSubscription(w http.ResponseWriter, r *http.Request) {
	_, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// key := fmt.Sprintf("subscription-%d", pid)
	// var value interface{}
	// if ok := server.Cache.Get(key, &value); ok {
	// 	responses.JSON(w, http.StatusOK, value)
	// 	return
	// }
	subscription := models.Subscription{}
	subscription.OrganizationID = uint64(oid)
	subscriptionReceived, err := subscriptionInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, subscriptionReceived)
		return
	}
	subscriptionResponse, err := CreateSubscriptionResponse(subscriptionReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// server.Cache.Set(key, subscriptionResponse)
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    subscriptionResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateSubscription godoc
// @Summary Update a Subscription
// @Description Update a Subscription with the input payload
// @Tags Subscription
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Subscription id"
// @Param body body doc.Subscription true "Update Subscription"
// @Success 200 {object} doc.Subscription
// @Router /subscription/{id} [put]
func (server *Server) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	subscription := models.Subscription{}
	err = server.DB.Model(models.Subscription{}).
		Where("id = ?", pid).
		Where("organization_id = ?", oid).
		Take(&subscription).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("subscription not found "))
		return
	}
	if !user.IsAdmin && oid == 0 {
		if subscription.UserID != uint64(user.ID) {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	if oid > 0 {
		organization, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	subscriptionUpdate := models.Subscription{}
	err = json.Unmarshal(body, &subscriptionUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	// checks for duplicate entries
	if len(subscriptionUpdate.Name) > 0 && subscription.Name != subscriptionUpdate.Name {
		if subscription.IsNameExists(server.DB, oid, subscriptionUpdate.Name) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("subscription name already exists "))
			return
		}
	}

	subscriptionUpdate.ID = subscription.ID
	if oid > 0 && !subscriptionUpdate.CheckSubscriptionResourceLimit(server.DB, organization) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("resource limit exceeds "))
		return
	}
	subscriptionUpdated, err := subscriptionInterface.Update(server.DB, subscriptionUpdate)

	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if subscription is Updated within Orgination level then we have to add to audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "update", "subscription", subscriptionUpdated.Name, "")
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, subscriptionUpdated)
		return
	}
	subscriptionResponse, err := CreateSubscriptionResponse(subscriptionUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    subscriptionResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteSubscription godoc
// @Summary Delete a Subscription
// @Description Delete a Subscription with the input payload
// @Tags Subscription
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Subscription id"
// @Success 204 {object} doc.Subscription
// @Router /subscription/{id} [delete]
func (server *Server) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	// Is a valid subscription id given to us?
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	// Is this user authenticated?
	uid, oid, _ := auth.ExtractTokenID(r)
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	// Check if the subscription exist
	subscription := models.Subscription{}
	err = server.DB.Model(models.Subscription{}).
		Where("id = ?", pid).
		Where("organization_id = ?", oid).
		Take(&subscription).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !user.IsAdmin && oid == 0 {
		if subscription.UserID != uint64(user.ID) {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
	}
	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}

	count := 0
	err = server.DB.Model(models.Project{}).Where("subscription_id = ?", pid).Count(&count).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	if count > 0 {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("this subscription is used by some active project "))
		return
	}
	_, err = subscriptionInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// Checks if subscription is Updated within Orgination level then we have to add to audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "delete", "subscription", subscription.Name, "")
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

func CreateSubscriptionResponse(subscription *models.Subscription) (*doc.Subscription, error) {
	subscriptionResponse := doc.Subscription{}
	subscriptionBytes, _ := json.Marshal(subscription)
	err := json.Unmarshal(subscriptionBytes, &subscriptionResponse)
	if err != nil {
		return nil, err
	}
	return &subscriptionResponse, nil

}

// func (server *Server) NewSubscriptionInterface(r *http.Request) models.SubscriptionInterface {
// 	test := r.URL.Query().Get("subscription")
// 	if test != "" {
// 		testdata := server.MockInterface.(models.SubscriptionInterface)
// 		return testdata
// 	}
// 	data := models.NewSubscription(server.DB)
// 	return data
// }
