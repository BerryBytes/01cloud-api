package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var orgPlanInterface = models.NewOrganizationPlan()

// CreateOrganizationPlan godoc
// @Summary Create a new OrganizationPlan
// @Description Create a new OrganizationPlan with the input paylod
// @Tags OrganizationPlan
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.OrganizationPlan true "Create Organization Plan"
// @Success 201 {object} doc.OrganizationPlan
// @Router /organizationPlan [post]
func (server *Server) CreateOrganizationPlan(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	organizationPlan := models.OrganizationPlan{}
	err = json.Unmarshal(body, &organizationPlan)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	organizationPlan.Prepare()
	err = organizationPlan.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	organizationPlanCreated, err := orgPlanInterface.Save(server.DB, &organizationPlan)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, organizationPlanCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, organizationPlanCreated)
		return
	}
	orgPlanResponse, err := CreateOrganizationPlanResponse(organizationPlanCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    orgPlanResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetOrganizationPlans godoc
// @Summary Get OrganizationPlans
// @Description Get all OrganizationPlan
// @Tags OrganizationPlan
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.OrganizationPlan
// @Router /organizationPlans [get]
func (server *Server) GetOrganizationPlans(w http.ResponseWriter, r *http.Request) {
	organizationPlans, err := orgPlanInterface.FindAll(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, organizationPlans)
		return
	}
	orgPlanResponseList := []doc.OrganizationPlan{}
	for _, item := range *organizationPlans {
		orgResponse, err := CreateOrganizationPlanResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		orgPlanResponseList = append(orgPlanResponseList, *orgResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    orgPlanResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetOrganizationPlansForAdmin godoc
// @Summary Get OrganizationPlans
// @Description Get all OrganizationPlan
// @Tags OrganizationPlan
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.OrganizationPlan
// @Router /admin/organizationPlans [get]
func (server *Server) GetOrganizationPlansForAdmin(w http.ResponseWriter, r *http.Request) {
	organizationPlans, err := orgPlanInterface.FindAllWithInActive(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, organizationPlans)
		return
	}
	orgPlanResponseList := []doc.OrganizationPlan{}
	for _, item := range *organizationPlans {
		orgResponse, err := CreateOrganizationPlanResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		orgPlanResponseList = append(orgPlanResponseList, *orgResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    orgPlanResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetOrganizationPlan godoc
// @Summary Get OrganizationPlans
// @Description Get all OrganizationPlan
// @Tags OrganizationPlan
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "OrganizationPlan id"
// @Success 200 {object} doc.OrganizationPlan
// @Router /organizationPlan/{id} [get]
func (server *Server) GetOrganizationPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	organizationPlanReceived, err := orgPlanInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, organizationPlanReceived)
		return
	}
	orgPlanResponse, err := CreateOrganizationPlanResponse(organizationPlanReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    orgPlanResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateOrganizationPlan godoc
// @Summary Update a OrganizationPlan
// @Description Update a OrganizationPlan with the input paylod
// @Tags OrganizationPlan
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.OrganizationPlan true "Update Organization Plan"
// @Success 200 {object} doc.OrganizationPlan
// @Router /organizationPlan [post]
func (server *Server) UpdateOrganizationPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	organizationPlanReceived, err := orgPlanInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	organizationPlanUpdate := models.OrganizationPlan{}
	err = json.Unmarshal(body, &organizationPlanUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	organizationPlanUpdate.ID = organizationPlanReceived.ID
	organizationPlanUpdated, err := orgPlanInterface.Update(server.DB, &organizationPlanUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, organizationPlanUpdated)
		return
	}
	orgPlanResponse, err := CreateOrganizationPlanResponse(organizationPlanUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    orgPlanResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeleteOrganizationPlan godoc
// @Summary Delete a OrganizationPlan
// @Description Delete a  OrganizationPlan with id
// @Tags OrganizationPlan
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "OrganizationPlan id"
// @Success 204 {string} No content
// @Router /organizationPlan/{id} [delete]
func (server *Server) DeleteOrganizationPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Is a valid organizationPlan id given to us?
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	_, err = orgPlanInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	_, err = orgPlanInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
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

func CreateOrganizationPlanResponse(orgPlan *models.OrganizationPlan) (*doc.OrganizationPlan, error) {
	organizationPlanResponse := doc.OrganizationPlan{}
	orgPlanBytes, _ := json.Marshal(orgPlan)
	err := json.Unmarshal(orgPlanBytes, &organizationPlanResponse)
	if err != nil {
		return nil, err
	}
	return &organizationPlanResponse, nil
}
