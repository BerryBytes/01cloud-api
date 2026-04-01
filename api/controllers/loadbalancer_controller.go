package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var loadbalancerInterface = models.NewLoadBalancer()

// CreateLoadBalancer godoc
// @Summary Create a new LoadBalancer
// @Description Create a new LoadBalancer with the input paylod
// @Tags LoadBalancer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param body body doc.LoadBalancer true "Create LoadBalancer"
// @Success 201 {object} doc.LoadBalancer
// @Router /loadbalancer [post]
func (server *Server) CreateLoadBalancer(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.LoadBalancer{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.Prepare()
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	_, err = userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	project, err := iproject.Find(server.DB, data.ProjectID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid project"))
		return
	}
	projectLb, err := loadbalancerInterface.FindAllByProject(server.DB, data.ProjectID)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if uint64(len(*projectLb)) >= project.Subscription.LoadBalancer {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("sorry you cannot create new loadbalancer in this project, Quota limit exceeded"))
		return
	}
	region := &struct {
		Region string `json:"region"`
	}{}
	err = json.Unmarshal(body, &region)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	regionId := region.Region
	cls, err := clusterInterface.FindByRegionWithDns(server.DB, regionId, oid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("region not found"))
		return
	}
	data.ClusterID = uint64(cls.ID)

	err = data.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	dataCreated, err := loadbalancerInterface.Save(server.DB, &data)

	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	dataCreated.Cluster = cls
	err = queue.Publish(constants.CreateLoadBalancer, dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	loadbalancerResponse, err := CreateLoadbalancerResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    loadbalancerResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetLoadBalancers godoc
// @Summary Get all LoadBalancers
// @Description Get array of LoadBalancers with the project id
// @Tags LoadBalancer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Project id"
// @Success 200 {array} doc.LoadBalancer
// @Router /project/{id}/loadbalancers [get]
func (server *Server) GetLoadBalancerByProject(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	projectId, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	project, err := iproject.Find(server.DB, projectId)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("project not found"))
		return
	}
	if uid != uint(project.UserID) {
		if !authInterface.IsAuthorizedProject(server.DB, uint64(uid), projectId) {
			responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized, you are not authorized"))
			return
		}
	}
	datas, err := loadbalancerInterface.FindAllByProject(server.DB, projectId)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	loadbalancerResponse := []doc.LoadBalancer{}
	for _, item := range *datas {
		response, err := CreateLoadbalancerResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		loadbalancerResponse = append(loadbalancerResponse, *response)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    loadbalancerResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetLoadBalancers godoc
// @Summary Get all LoadBalancers
// @Description Get array of LoadBalancers
// @Tags LoadBalancer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.LoadBalancer
// @Router /loadbalancers [get]
func (server *Server) GetLoadBalancersForAdmin(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
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

	data := models.LoadBalancer{}
	datas, err := data.FindAllWithInactive(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	loadbalancerResponse := []doc.LoadBalancer{}
	for _, item := range *datas {
		response, err := CreateLoadbalancerResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		loadbalancerResponse = append(loadbalancerResponse, *response)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    loadbalancerResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetEnvironmentByLoadBalancer(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	vars := mux.Vars(r)
	lid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
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

	data := models.Environment{}
	datas, err := data.FindEnvironmentByLoadBalancer(server.DB, lid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	environmentResponseList := []doc.Environment{}
	for _, environment := range *datas {
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

// FetchLoadBalancer status godoc
// @Summary Trigger get LoadBalancer service
// @Description Trigger Get LoadBalancer status service by id
// @Tags LoadBalancer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "LoadBalancer id"
// @Success 200 {object} doc.SuccessResponse
// @Router /loadbalancer/{id}/fetch-status [get]
func (server *Server) FetchLoadBalancerStatus(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := loadbalancerInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = queue.Publish(constants.FetchLoadBalancerStatus, dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{"message": "Fetching loadbalancer status"})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetLoadBalancer godoc
// @Summary Get LoadBalancer
// @Description Get LoadBalancer by id
// @Tags LoadBalancer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "LoadBalancer id"
// @Success 200 {object} doc.LoadBalancer
// @Router /loadbalancer/{id} [get]
func (server *Server) GetLoadBalancer(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := loadbalancerInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	loadbalancerResponse, err := CreateLoadbalancerResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    loadbalancerResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetLoadBalancer status godoc
// @Summary Get LoadBalancer status
// @Description Get LoadBalancer status by id
// @Tags LoadBalancer
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "LoadBalancer id"
// @Success 200 {object} doc.LoadBalancerStatusResponse
// @Router /loadbalancer/{id}/status [get]
func (server *Server) GetLoadBalancerStatus(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
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
	_, err = loadbalancerInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	serviceDetail, err := server.StoreClient.Loadbalancer().GetLoadbalancerStatus(fmt.Sprintf("%s-lb-%d", os.Getenv("GCLOUD_NAMESPACE"), pid))
	if err != nil {
		responses.ERROR(w, http.StatusNoContent, errors.New("no data available"))
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, serviceDetail)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    serviceDetail,
		Success: 1,
		Message: "Success",
	})
}

// DeleteLoadBalancer godoc
// @Summary Delete LoadBalancer by id
// @Description Delete a LoadBalancer with LoadBalancer id
// @Tags LoadBalancer
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "LoadBalancer id"
// @Success 200 {object} doc.LoadBalancer
// @Router /loadBalancer/{id} [delete]
func (server *Server) DeleteLoadBalancer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	_, err = userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	organization := &models.Organization{}
	if oid > 0 && organization.CheckRole(server.DB, uint64(uid), uint64(oid)) {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	data, err := loadbalancerInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("unauthorized"))
		return
	}
	if data.IsLoadbalancerUsed(server.DB) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("you cannot delete used loadbalancer"))
		return
	}
	err = queue.Publish(constants.DestroyLoadBalancer, data)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	go func() {
		time.Sleep(5 * time.Second)
		_, _ = loadbalancerInterface.Delete(server.DB, pid)
	}()
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

func CreateLoadbalancerResponse(loadbalancer *models.LoadBalancer) (*doc.LoadBalancer, error) {
	LoadbalancerResponse := doc.LoadBalancer{}
	loadbalancerBytes, _ := json.Marshal(loadbalancer)
	err := json.Unmarshal(loadbalancerBytes, &LoadbalancerResponse)
	if err != nil {
		return nil, err
	}
	return &LoadbalancerResponse, nil
}
