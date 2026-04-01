package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models/doc"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var grepo = models.NewGroupRepo()

// CreateGroup godoc
// @Summary Create a new Group
// @Description Create a new Group with the input paylod
// @Tags Group
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Group true "Create Group"
// @Success 201 {object} doc.Group
// @Router /groups [post]
func (server *Server) CreateGroup(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := &models.Group{}
	err = json.Unmarshal(body, data)
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
	dataList, err := grepo.FindAll(server.DB, uint64(oid), "")
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	for _, j := range *dataList {
		if j.Name == data.Name {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("group name already exists"))
			return
		}
	}
	data.OrganizationID = uint64(oid)
	dataCreated, err := grepo.Save(server.DB, data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if group is created within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "create", "group", dataCreated.Name, "")
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	groupResponse, err := CreateGroupResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    groupResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetGroups godoc
// @Summary Get Groups
// @Description Get list Group from token
// @Tags Group
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param query query string true "Query"
// @Success 200 {array} doc.Group
// @Router /groups [get]
func (server *Server) GetGroups(w http.ResponseWriter, r *http.Request) {
	_, oid, _ := auth.ExtractTokenID(r)
	query := r.FormValue("query")
	dataList, err := grepo.FindAll(server.DB, uint64(oid), query)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataList)
		return
	}
	groupResponseList := []doc.Group{}
	for _, group := range *dataList {
		groupResponse, err := CreateGroupResponse(&group)
		if err != nil {
			log.Error(err)
		}
		groupResponseList = append(groupResponseList, *groupResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    groupResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetGroup godoc
// @Summary Get a Group
// @Description Get a Group with the input id
// @Tags Group
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param gid path int true "Group id"
// @Success 200 {object} doc.Group
// @Router /groups/{gid} [get]
func (server *Server) GetGroup(w http.ResponseWriter, r *http.Request) {
	_, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	gid, err := strconv.ParseUint(vars["gid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := grepo.Find(server.DB, gid, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data)
		return
	}
	groupResponse, err := CreateGroupResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    groupResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateGroup godoc
// @Summary Update a Group
// @Description Update a  Group with the input paylod
// @Tags Group
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param gid path int true "Group id"
// @Param body body doc.Group true "Update Group"
// @Success 200 {object} doc.Group
// @Router /groups/{gid} [put]
func (server *Server) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	gid, err := strconv.ParseUint(vars["gid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := grepo.Find(server.DB, gid, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("group not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := &models.Group{}
	err = json.Unmarshal(body, dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataList, err := grepo.FindAll(server.DB, uint64(oid), "")
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	for _, j := range *dataList {
		if j.Name == dataUpdate.Name {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("duplicate group name please enter another name"))
			return
		}
	}
	dataUpdate.ID = data.ID
	dataUpdated, err := grepo.Update(server.DB, dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if group is update within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "update", "group", dataUpdated.Name, "")
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	groupResponse, err := CreateGroupResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    groupResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteGroup godoc
// @Summary Delete a Group
// @Description Delete a Group with the group id
// @Tags Group
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param gid path int true "Group id"
// @Success 200 {object} doc.Group
// @Router /groups/{gid} [delete]
func (server *Server) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	gid, err := strconv.ParseUint(vars["gid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := grepo.Find(server.DB, gid, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("unauthorized"))
		return
	}
	if len(data.Members) > 0 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("members should be removed first to delete group"))
		return
	}

	_, err = grepo.Delete(server.DB, int64(gid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// Checks if group is delete within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "delete", "group", data.Name, "")
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", gid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// AddMembersToGroup godoc
// @Summary Add Members to a Group
// @Description Add a new Member to a Group with the input email
// @Tags Group
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param gid path int true "Group id"
// @Param body body doc.User true "Pass only email"
// @Success 200 {object} object
// @Router /groups/{gid}/members [post]
func (server *Server) AddMembersToGroup(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	gid, err := strconv.ParseUint(vars["gid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := grepo.Find(server.DB, gid, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("group not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var user models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dbUser, err := userInterface.FindUserByEmail(server.DB, user.Email)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	organization, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	err = grepo.AddMember(server.DB, dbUser, data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if organization.UserID != uint64(dbUser.ID) {
		org := &models.OrganizationMembers{}
		org.UserID = uint64(dbUser.ID)
		org.OrganizationID = uint64(oid)
		_ = orgMember.AddMember(server.DB, org)
	}
	// Checks if member is added into group within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "create", "member", user.Email, "to "+data.Name)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": fmt.Sprintf("%q Added", user.Email),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// DeleteMembersFromGroup godoc
// @Summary Delete Members From a Group
// @Description Delete member from  Group with the payload
// @Tags Group
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param gid path int true "Group id"
// @Param body body doc.User true "Pass only email Group"
// @Success 204 {string} No content
// @Router /groups/{gid}/members [delete]
func (server *Server) DeleteMembersFromGroup(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	gid, err := strconv.ParseUint(vars["gid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := grepo.Find(server.DB, gid, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("group not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var user *models.User
	err = json.Unmarshal(body, &user)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user, err = userInterface.FindUserByEmail(server.DB, user.Email)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = grepo.DeleteMember(server.DB, user, data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	// Checks if member is removed from group within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "delete", "member", user.Email, "from "+data.Name)
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", gid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func CreateGroupResponse(group *models.Group) (*doc.Group, error) {
	groupResponse := doc.Group{}
	dnsBytes, _ := json.Marshal(group)
	err := json.Unmarshal(dnsBytes, &groupResponse)
	if err != nil {
		return nil, err
	}
	return &groupResponse, nil

}
