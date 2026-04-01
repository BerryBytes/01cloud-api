package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/auth"
	"01cloud-api/api/models/doc"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var idns = models.NewDNS()

// CreateDns godoc
// @Summary Create a new dns
// @Description Create a new dns with the input payload
// @Tags Dns
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.DNS true "Create Dns"
// @Success 201 {object} doc.DNS
// @Router /dns [post]
func (server *Server) CreateDns(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.DNS{}
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

	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
	if idns.IsNameExists(server.DB, uint64(oid), data.Name) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New(data.Name+" dns already exists"))
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	if oid > 0 {
		_, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}

	data.OrganizationID = uint64(oid)
	dataCreated, err := idns.Save(server.DB, &data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if Dns is Created within Orgination level then we are adding audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "create", "dns", data.Name, "")
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	dnsResponse, err := CreateDnsDetailResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    dnsResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetDnsList godoc
// @Summary Get Dns list
// @Description Get list of dns
// @Tags Dns
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.DNS
// @Router /dns [get]
func (server *Server) GetDnsList(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
	_, err = userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	data := models.DNS{}
	data.OrganizationID = uint64(oid)
	datas, err := idns.FindAllByOrganization(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	dnsResponseList := []doc.DNSDetails{}
	for _, dns := range *datas {
		dnsResponse, err := CreateDnsDetailResponse(&dns)
		if err != nil {
			log.Error(err)
		}
		dnsResponseList = append(dnsResponseList, *dnsResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    dnsResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetDnssForAdmin godoc
// @Summary Get Dnss by organization
// @Description Get list of dns by organization
// @Tags Dns
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.DNS
// @Router /admin/dns [get]
func (server *Server) GetDnssForAdmin(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	data := models.DNS{}
	data.OrganizationID = uint64(oid)
	datas, err := idns.FindAllWithInactive(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	dnsResponseList := []doc.DNSDetails{}
	for _, dns := range *datas {
		dnsResponse, err := CreateDnsDetailResponse(&dns)
		if err != nil {
			log.Error(err)
		}
		dnsResponseList = append(dnsResponseList, *dnsResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    dnsResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetDns godoc
// @Summary Get Dns by id
// @Description Get Dns by id from token
// @Tags Dns
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Dns id"
// @Success 200 {object} doc.DNS
// @Router /dns/{id} [get]
func (server *Server) GetDns(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}

	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	//data := models.DNS{}
	dataReceived, err := idns.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	dnsResponse, err := CreateDnsDetailResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    dnsResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateDns godoc
// @Summary Update a Dns
// @Description Update a Dns with the input payload
// @Tags Dns
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Dns id"
// @Param body body doc.DNS true "Update Dns"
// @Success 200 {object} doc.DNS
// @Router /dns/{id} [put]
func (server *Server) UpdateDns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
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

	data := models.DNS{}
	err = server.DB.Model(models.DNS{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("dns not found "))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	count := 0
	err = server.DB.Model(models.Cluster{}).Where("dns_id = ? and active = ?", pid, true).Count(&count).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	if count > 0 {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("you cannot update this dns, it is used by some active clusters"))
		return
	}
	dataUpdate := models.DNS{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	if oid > 0 {
		_, err = orgInterface.Find(server.DB, oid)
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found "))
			return
		}
	}

	dataUpdate.ID = data.ID
	dataUpdated, err := idns.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	// Checks if Dns is Updated within Orgination level then we are adding audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "update", "dns", dataUpdated.Name, "")
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	dnsResponse, err := CreateDnsDetailResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    dnsResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteDns godoc
// @Summary Delete a Dns
// @Description Delete a Dns with the input payload
// @Tags Dns
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Dns id"
// @Success 204 {object} doc.DNS
// @Router /dns/{id} [delete]
func (server *Server) DeleteDns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized "))
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if !user.IsAdmin && oid == 0 {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("not authorized "))
		return
	}

	data := models.DNS{}
	err = server.DB.Model(models.DNS{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	count := 0
	err = server.DB.Model(models.Cluster{}).Where("dns_id = ? and active = ?", pid, true).Count(&count).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotAcceptable, err)
		return
	}
	if count > 0 {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("you cannot delete this dns, it is used by some active clusters"))
		return
	}
	_, err = idns.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// Checks if Dns is Deleted within Orgination level then we are adding audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "delete", "dns", data.Name, "")
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

func CreateDnsResponse(dns *models.DNS) (*doc.DNS, error) {
	dnsResponse := doc.DNS{}
	dnsBytes, _ := json.Marshal(dns)
	err := json.Unmarshal(dnsBytes, &dnsResponse)
	if err != nil {
		return nil, err
	}
	return &dnsResponse, nil
}

func CreateDnsDetailResponse(dns *models.DNS) (*doc.DNSDetails, error) {
	dnsResponse := doc.DNSDetails{}
	dnsBytes, _ := json.Marshal(dns)
	err := json.Unmarshal(dnsBytes, &dnsResponse)
	if err != nil {
		return nil, err
	}
	return &dnsResponse, nil
}
