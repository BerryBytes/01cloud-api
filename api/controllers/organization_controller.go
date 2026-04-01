package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"01cloud-api/api/auth"
	"01cloud-api/api/mailer"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
)

var orgInterface = models.NewOrganization()
var orgMember = models.NewOrganizationMember()

// CreateOrganization godoc
// @Summary Create a new Organization
// @Description Create a new Organization with the input paylod
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Organization true "Create Organization"
// @Success 201 {object} doc.Organization
// @Router /organization [post]
func (server *Server) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.Organization{}
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
	data.UserID = uint64(uid)
	if !paymentInterface.HasUserBalance(server.DB, uid) {
		responses.ERROR(w, http.StatusPaymentRequired, errors.New("unable to create organization due to remaining balance"))
		return
	}
	dataList, err := orgInterface.FindAll(server.DB, uid, true, "")
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, getErrorMessage("user", "User not found"))
		return
	}
	count, err := orgInterface.CountOrginationByUser(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if err = helper.CheckUserQuota(user, "o", count); err != nil {
		log.Error("quota error :: ", err)
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	for _, j := range *dataList {
		if strings.EqualFold(j.Name, data.Name) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("organization already exists"))
			return
		}
	}
	dataCreated, err := orgInterface.Save(server.DB, &data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	plan, _ := orgPlanInterface.Find(server.DB, data.OrganizationPlanID)
	defMem := uint32(2048)
	if defMem > plan.Memory {
		defMem = plan.Memory
	}
	defCore := uint32(2000)
	if defCore > plan.Cores {
		defCore = plan.Cores
	}
	sub := &models.Subscription{
		Name:           "Default",
		Apps:           10,
		DiskSpace:      10240,
		Memory:         defMem,
		Cores:          defCore,
		Price:          0,
		CronJob:        5,
		DataTransfer:   10240,
		OrganizationID: uint64(dataCreated.ID),
		Active:         true,
	}
	sub.Prepare()
	_, _ = subscriptionInterface.Save(server.DB, *sub)
	defMemory := uint32(512)
	if defMemory > plan.Memory {
		defMemory = plan.Memory
	}
	defCores := uint32(500)
	if defCores > plan.Cores {
		defCores = plan.Cores
	}
	res := &models.Resource{
		Name:           "Default",
		Cores:          uint64(defCores),
		Memory:         uint64(defMemory),
		OrganizationID: uint64(dataCreated.ID),
	}
	res.Prepare()
	_, _ = resourceInterface.Save(server.DB, res)
	_, _ = server.SaveOrganizationActivityWithJson(uid, dataCreated.ID, "create", "organization", dataCreated.Name, "")
	sessionId, _ := auth.ExtractSessionID(r)
	token, err := auth.CreateToken(uid, dataCreated.ID, sessionId)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]interface{}{
			"organization": dataCreated,
			"token":        token,
		})
		return
	}
	organizationResponse, err := CreateOrganizationResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data: &map[string]interface{}{
			"organization": organizationResponse,
			"token":        token,
		},
		Success: 1,
		Message: "Success",
	})
}

// GetOrganizationsList godoc
// @Summary Get Organizations
// @Description Get list Organization from token
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param query query string true "Query"
// @Success 200 {array} doc.Organization
// @Router /organizations [get]
func (server *Server) GetOrganizationsList(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	query := r.FormValue("query")
	key := fmt.Sprintf("organizations-list-%d", uid)
	var value interface{}
	if ok := server.Cache.Get(key, &value); ok {
		responses.JSON(w, http.StatusOK, value)
		return
	}
	data := models.Organization{}
	dataList, err := orgInterface.FindAll(server.DB, uid, true, query)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	jsonOrg, _ := data.ToJson()
	helper.DeleteKey(jsonOrg, "UpdatedAt", "DeletedAt")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataList)
		return
	}
	orgResponseList := []doc.Organization{}
	for _, item := range *dataList {
		orgResponse, err := CreateOrganizationResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		orgResponseList = append(orgResponseList, *orgResponse)
	}
	response := responses.Response{
		Data:    orgResponseList,
		Success: 1,
		Message: "Success",
	}
	server.Cache.Set(key, &response)
	responses.JSON(w, http.StatusOK, response)
}

// GetOrganizationsListForAdmin godoc
// @Summary Get Organizations List For Admin
// @Description Get Organizations List For Admin
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param size query int true "Size"
// @Param page query int true "Page"
// @Success 200 {array} doc.Organization
// @Router /admin/organizations [get]
func (server *Server) GetOrganizationsListForAdmin(w http.ResponseWriter, r *http.Request) {
	size, _ := strconv.ParseUint(r.URL.Query().Get("size"), 10, 32)
	page, _ := strconv.ParseUint(r.URL.Query().Get("page"), 10, 32)
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	query := r.FormValue("query")
	data := models.Organization{}
	dataList, count, err := data.FindAllForAdmin(server.DB, size, page, query)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	jsonOrg, _ := data.ToJson()
	helper.DeleteKey(jsonOrg, "UpdatedAt", "DeletedAt")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"count": count,
			"data":  dataList,
		})
		return
	}
	orgResponseList := []doc.Organization{}
	for _, item := range *dataList {
		orgResponse, err := CreateOrganizationResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		orgResponseList = append(orgResponseList, *orgResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    orgResponseList,
		Count:   int(count),
		Success: 1,
		Message: "Success",
	})
}

// GetOrganization godoc
// @Summary Get Organizations
// @Description Get Organization from token
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} doc.Organization
// @Router /organization [get]
func (server *Server) GetOrganization(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	data, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	member := []*models.OrganizationMembers{}
	for _, m := range data.Members {
		if !m.User.Active {
			continue
		}
		member = append(member, m)
	}
	data.Members = member
	if !data.Verify(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data)
		return
	}
	organizationResponse, err := CreateOrganizationResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    organizationResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetOrganizationForAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	oid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	data, err := orgInterface.Find(server.DB, uint(oid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data)
		return
	}
	organizationResponse, err := CreateOrganizationResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    organizationResponse,
		Success: 1,
		Message: "Success",
	})

}

// UpdateOrganization godoc
// @Summary Update a Organization
// @Description Update a  Organization with the input payloaad
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Organization true "Update Organization"
// @Success 200 {object} doc.Organization
// @Router /organization [put]
func (server *Server) UpdateOrganization(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	data, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.Organization{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !authInterface.IsAuthorizedOrganization(server.DB, uid, oid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized to update organization"))
		return
	}
	if dataUpdate.OrganizationPlanID != 0 && dataUpdate.OrganizationPlanID != data.OrganizationPlanID && data.UserID != uint64(uid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("owner can only change subscription plan"))
		return
	}
	dataUpdate.ID = data.ID
	dataUpdated, err := orgInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	_, _ = server.SaveOrganizationActivityWithJson(uid, dataUpdated.ID, "update", "organization", dataUpdated.Name, "")
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	organizationResponse, err := CreateOrganizationResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	key := fmt.Sprintf("organizations-list-%d", uid)
	server.Cache.Delete(key)
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    organizationResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteOrganization godoc
// @Summary Delete a Organization
// @Description Delete a Organization with the Organization id
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 204 {string} No content
// @Router /organization [delete]
func (server *Server) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	data, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("unauthorized"))
		return
	}
	if data.UserID != uint64(uid) {
		responses.ERROR(w, http.StatusNotFound, errors.New("user should be owner to delete Organization"))
		return
	}
	if !data.Verify(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	prj := &models.Project{}
	prj.OrganizationId = uint64(data.ID)
	projects, err := iproject.FindAll(server.DB, prj)
	if err == nil && len(*projects) > 0 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("projects should be removed to delete Organization"))
		return
	}
	clusterRequest, err := clusterRequestInterface.FindAllByOrganization(server.DB, int64(data.ID))
	if err == nil && len(*clusterRequest) > 0 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("cluster request should be deleted  to delete organization"))
		return
	}
	clusters, err := clusterInterface.FindAllClusterWithOrganizationId(server.DB, uint64(data.ID))
	if err == nil && len(*clusters) > 0 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("cluster should be deleted  to delete organization"))
		return
	}
	for _, m := range data.Members {
		_ = orgMember.DeleteMember(server.DB, m.UserID, uint64(oid))
	}
	_, err = orgInterface.Delete(server.DB, int64(oid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "delete", "organization", data.Name, "")
	w.Header().Set("Entity", fmt.Sprintf("%d", oid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	key := fmt.Sprintf("organizations-list-%d", uid)
	server.Cache.Delete(key)
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// AddMembersToOrganization godoc
// @Summary Add Members to a Organization
// @Description Add a new Member to a Organization with the input email
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body models.RequestObject true "Pass only email"
// @Success 200 {object} object
// @Router /organization/members [post]
func (server *Server) AddMembersToOrganization(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	_, err = orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var reqBody models.RequestObject
	err = json.Unmarshal(body, &reqBody)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dbUser, err := userInterface.FindUserByEmail(server.DB, reqBody.Email)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	organization, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	if organization.User.Email == reqBody.Email {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("user is already owner"))
		return
	}
	member := &models.OrganizationMembers{}
	member.UserID = uint64(dbUser.ID)
	member.OrganizationID = uint64(oid)
	member.UserRole = reqBody.Role
	if orgMember.CheckLimit(server.DB, organization.OrganizationPlan, oid) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("user limit exceeds"))
		return
	}
	err = orgMember.AddMember(server.DB, member)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	org, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	var roleName string
	if reqBody.Role == 1 {
		roleName = "admin"
	} else {
		roleName = "member"
	}
	// Checks if member is added into group within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "share", "organization", org.Name, "to "+dbUser.FirstName+" "+dbUser.LastName+" as "+roleName)
	}
	shareOrgEmail := mailer.SendShareOrganization(reqBody.Email, org.Name, roleName, int(oid), dbUser.FirstName)
	err = notifications.NotifyEmail(server.NotifyClient, &shareOrgEmail)
	if err != nil {
		log.Error("error from notify email :: ", err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": fmt.Sprintf("%q Added", dbUser.Email),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: fmt.Sprintf("%q Added", dbUser.Email),
	})
}

// UpdateMembersFromOrganization godoc
// @Summary Update Members From a Organization
// @Description Delete member from  Organization with the payload
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body models.RequestObject true "Member "
// @Success 200 {string} email
// @Router /organization/members [put]
func (server *Server) UpdateMembersFromOrganization(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
	org, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var ruser models.RequestObject
	err = json.Unmarshal(body, &ruser)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	user, err := userInterface.FindUserByEmail(server.DB, ruser.Email)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if orgMember.IsRoleExist(server.DB, oid, user.ID, ruser.Role) {
		responses.ERROR(w, http.StatusBadRequest, errors.New("user role already exist"))
		return
	}
	err = orgMember.UpdateMember(server.DB, uint64(user.ID), uint64(oid), ruser.Role)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	var roleName, oldRoleName string
	if ruser.Role == 1 {
		roleName = "admin"
		oldRoleName = "member"
	} else {
		roleName = "member"
		oldRoleName = "admin"

	}
	// Checks if member is updated into group within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "share", "organization", org.Name, "to "+user.FirstName+" "+user.LastName+" as "+roleName+" from "+oldRoleName)
	}
	shareOrgEmail := mailer.SendUpdateOrganization(ruser.Email, org.Name, roleName, int(oid), user.FirstName)
	err = notifications.NotifyEmail(server.NotifyClient, &shareOrgEmail)
	if err != nil {
		log.Error("notify email error :: ", err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"message": fmt.Sprintf("Role updated for %s as %v", ruser.Email, ruser.Role),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: fmt.Sprintf("Role updated for %s as %v", ruser.Email, ruser.Role),
	})
}

// DeleteMembersFromOrganization godoc
// @Summary Delete Members From a Organization
// @Description Delete member from  Organization with the payload
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body models.RequestObject true "Member"
// @Success 204 {string} No content
// @Router /organization/members [delete]
func (server *Server) DeleteMembersFromOrganization(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	org, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
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
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = orgMember.DeleteMember(server.DB, uint64(user.ID), uint64(oid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	gp, err := grepo.FindAll(server.DB, uint64(oid), "")
	if err == nil {
		for _, d := range *gp {
			_ = grepo.DeleteMember(server.DB, user, &d)
		}
	}
	_, err = authInterface.DeleteOrganizationMember(server.DB, uint64(user.ID), uint64(oid))
	if err != nil {
		log.Error(err)
	}
	// Checks if member is updated into group within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "unshare", "organization", org.Name, "to "+user.FirstName+" "+user.LastName)
	}
	// _, err = iproject.DeleteProjectOwner(server.DB, uint64(user.ID), uint64(oid))
	// if err != nil {
	// 	log.Error(err)
	// }
	server.ActiveDeactiveProjects(uid, oid, false)
	w.Header().Set("Entity", fmt.Sprintf("%d", oid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// SwitchOrganization godoc
// @Summary Switch to another Organization
// @Description Switch from  Organization to other organization
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param oid path int true "Organization id"
// @Success 200 {object} object
// @Router /organization/{oid}/switch [get]
func (server *Server) SwitchOrganization(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	orgId, err := strconv.ParseUint(vars["oid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	sessionId, _ := auth.ExtractSessionID(r)
	if orgId == 0 {
		token, _ := auth.CreateToken(uid, uint(orgId), sessionId)
		responses.JSON(w, http.StatusCreated, responses.Response{
			Data: &map[string]interface{}{
				"organization": doc.Organization{},
				"token":        token,
			},
			Success: 1,
			Message: "Success",
		})
		return
	}

	data, err := orgInterface.Find(server.DB, uint(orgId))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	if !data.Verify(server.DB, uint64(uid), orgId) {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	token, err := auth.CreateToken(uid, uint(orgId), sessionId)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	organizationResponse, err := CreateOrganizationResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data: &map[string]interface{}{
			"organization": organizationResponse,
			"token":        token,
		},
		Success: 1,
		Message: "Success",
	})
}

// AddPluginToOrganization godoc
// @Summary Add plugin to  Organization
// @Description Add Plugin to the Organization
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 200 {object} object
// @Router /organization/plugin/{id} [get]
func (server *Server) AddPluginToOrganization(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	data, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	if !data.Verify(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}

	vars := mux.Vars(r)
	pluginId, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	plugin, err := pluginInterface.Find(server.DB, pluginId)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	_, err = data.AddPlugin(server.DB, plugin)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": "Plugin added to the Organization",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Plugin added to the Organization",
	})
}

// AddMultiplePluginToOrganization godoc
// @Summary Add plugins to  Organization
// @Description Add Plugins to the Organization
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} object
// @Router /organization/add-plugins [post]
func (server *Server) AddMultiplePluginToOrganization(w http.ResponseWriter, r *http.Request) {
	server.AddRemovePlugin(w, r, true)
}

// RemoveMultiplePluginFromOrganization godoc
// @Summary remove plugins from  Organization
// @Description remove Plugins from the Organization
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} object
// @Router /organization/remove-plugins [post]
func (server *Server) RemoveMultiplePluginFromOrganization(w http.ResponseWriter, r *http.Request) {
	server.AddRemovePlugin(w, r, false)
}

func (server *Server) AddRemovePlugin(w http.ResponseWriter, r *http.Request, isAdd bool) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	data, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	if !data.Verify(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input := []uint64{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	message := ""
	for _, v := range input {
		plugin, _ := pluginInterface.Find(server.DB, v)
		if isAdd {
			_, err = data.AddPlugin(server.DB, plugin)
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, err)
				return
			}
			message = "Plugin added to the Organization"
		} else {
			_, err = data.RemovePlugin(server.DB, plugin)
			if err != nil {
				responses.ERROR(w, http.StatusInternalServerError, err)
				return
			}
			message = "Plugins removed to the Organization"
		}

	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": message,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: message,
	})
}

// RemovePluginFromOrganization godoc
// @Summary Remove plugin to  Organization
// @Description Remove Plugin from the Organization
// @Tags Organization
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 204 {string} No Content
// @Router /organization/plugin/{id} [delete]
func (server *Server) RemovePluginFromOrganization(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	data, err := orgInterface.Find(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}
	if !data.Verify(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
		return
	}

	vars := mux.Vars(r)
	pluginId, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	plugin, err := pluginInterface.Find(server.DB, pluginId)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	_, err = data.RemovePlugin(server.DB, plugin)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", pluginId))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func CreateOrganizationResponse(org *models.Organization) (*doc.Organization, error) {
	organizationResponse := doc.Organization{}
	orgBytes, _ := json.Marshal(org)
	err := json.Unmarshal(orgBytes, &organizationResponse)
	if err != nil {
		return nil, err
	}
	return &organizationResponse, nil
}
