package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"01cloud-api/api/models/doc"
	"01cloud-api/api/utils/helper"
	"01cloud-api/api/utils/prometheus"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	permission_helper "01cloud-api/api/utils/imrole"
	storage_helper "01cloud-api/api/utils/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var clusterInterface = models.NewCluster()

// CreateCluster godoc
// @Summary Create Cluster
// @Description Create Cluster
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param file formData file true  "This is create cluster file"
// @Param body body doc.Cluster true "Create Cluster"
// @Success 201 {object} doc.Cluster
// @Router /cluster [post]
func (server *Server) CreateCluster(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("file is required"))
		return
	}
	configPath, err := uploadConfigFile(file)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.Cluster{}
	data.Name = r.FormValue("name")
	data.ConfigPath = configPath
	data.Context = r.FormValue("context")
	data.Token = r.FormValue("token")
	data.Region = r.FormValue("region")
	data.Provider = r.FormValue("provider")
	data.PrometheusServerUrl = r.FormValue("prometheus_server_url")
	imageRegistryID, err := strconv.ParseUint(r.FormValue("image_registry_id"), 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid image registry id"))
		return
	}
	data.ImageRegistryID = imageRegistryID
	// data.StorageAccessKey = r.FormValue("storage_access_key")
	// data.StorageSecretKey = r.FormValue("storage_secret_key")
	pvCapacity, err := strconv.ParseInt(r.FormValue("pv_capacity"), 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid pv capacity"))
		return
	}
	data.PvCapacity = uint64(pvCapacity)
	data.Zone = r.FormValue("zone")
	//data.DnsZone = r.FormValue("dns_zone")
	data.Labels = r.FormValue("labels")
	node, err := strconv.ParseInt(r.FormValue("nodes"), 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid node"))
		return
	}
	data.Nodes = uint64(node)

	err = data.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	dataCreated, err := clusterInterface.Save(server.DB, data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	clusterResponse, err := CreateClusterDetailResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    clusterResponse,
		Success: 1,
		Message: "Success",
	})
}

// AddStorageDetail godoc
// @Summary Add Storage Detail
// @Description Add Storage Detail
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Add storage detail"
// @Success 200 {object} doc.Cluster
// @Router /cluster/{id}/change-storage [post]
func (server *Server) AddStorageDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, errors.New("invalid token"))
		return
	}
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	cluster, err := clusterInterface.FindWithDns(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cluster not found"))
		return
	}
	usr, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if cluster.OrganizationID == 0 && !usr.IsAdmin {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you cannot edit zerone cluster"))
		return
	}
	if cluster.OrganizationID != uint64(oid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you dont have access to edit this cluster"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	storage := map[string]string{}
	err = json.Unmarshal(body, &storage)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	js, err := json.Marshal(storage)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to read body"))
		return
	}
	if storage["provider"] == constants.GCP {
		storageclient, err := storage_helper.CreateStorageClientWithCredential(storage["credentials"])
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid credential"))
			return
		}
		_ = storage_helper.CreateBucket(storageclient, storage["project_id"], fmt.Sprintf("zerone-bucket-%d", cluster.ID))
	}
	if storage["provider"] == constants.EKS {
		_ = storage_helper.CreateAwsBucket(storage, cluster.ID)
	}
	_ = cluster.CloudStorage.Scan(js)

	cls, err := clusterInterface.Update(server.DB, *cluster)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, cls)
		return
	}
	clusterResponse, err := CreateClusterResponse(cls)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterResponse,
		Success: 1,
		Message: "Success",
	})

}

func uploadConfigFile(file multipart.File) (string, error) {
	filepath := fmt.Sprintf("/data/config/01cloud/%s", uuid.New().String())
	_, err := os.Stat("/data/config/01cloud")
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath, 0777)
		if err != nil {
			return "", err
		}
	}
	f, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", err
	}
	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		return "", err
	}
	return filepath, nil
}

// GetClusters godoc
// @Summary Get Clusters
// @Description Get Clusters
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} []doc.Cluster
// @Router /clusters [get]
func (server *Server) GetClusters(w http.ResponseWriter, r *http.Request) {
	datas, err := clusterInterface.FindAll(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, datas)
}

// GetAllClusters godoc
// @Summary Get All Clusters
// @Description Get All Clusters
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param oid query string true "Organization Id"
// @Param region query string true "Region"
// @Success 200 {array} doc.Cluster
// @Router /get-clusters [get]
func (server *Server) GetAllClusters(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	region := r.URL.Query().Get("region")
	if oid == 0 && region == "" {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("either region or oid must be specified"))
		return
	}
	datas, err := clusterInterface.FindAllClustersByRegion(server.DB, oid, region)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	clusterResponseList := []doc.Cluster{}
	for _, cluster := range *datas {
		clusterResponse, err := CreateClusterResponse(&cluster)
		if err != nil {
			log.Error(err)
		}
		clusterResponseList = append(clusterResponseList, *clusterResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetClusterEnvironments godoc
// @Summary Get Cluster Environments
// @Description Get Cluster Environments
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Cluster Environment"
// @Success 200 {object} doc.Environment
// @Router /cluster/{id}/environments [get]
func (server *Server) GetClusterEnvironments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Environment{}
	datas, err := data.FindEnvironmentByCluster(server.DB, cid)
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

// GetClustersForAdmin godoc
// @Summary Get Clusters For Admin
// @Description Get Clusters For Admin
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} []doc.Cluster
// @Router /admin/clusters [get]
func (server *Server) GetClustersForAdmin(w http.ResponseWriter, r *http.Request) {
	datas, err := clusterInterface.FindAllWithInActive(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	clusterResponseList := []doc.ClusterDetails{}
	for _, cluster := range *datas {
		clusterResponse, err := CreateClusterDetailResponse(&cluster)
		if err != nil {
			log.Error(err)
		}
		clusterResponseList = append(clusterResponseList, *clusterResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterResponseList,
		Success: 1,
		Message: "Success",
	})
}

// GetClustersByOrgForAdmin godoc
// @Summary Get Clusters By Org For Admin
// @Description Get Clusters By Orgnization For Admin
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param oid path int true "Organization Id"
// @Success 200 {object} []doc.ClusterRequest
// @Router /admin/org/cluster/{oid} [get]
func (server *Server) GetClustersByOrgForAdmin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	oid, err := strconv.ParseUint(vars["oid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	datas, err := clusterRequestInterface.FindAllByOrganization(server.DB, int64(oid))
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	clusterrequestResponse := []doc.ClusterRequestDetails{}
	for _, item := range *datas {
		response, err := CreateClusterRequestDetailResponse(&item)
		if err != nil {
			log.Error(err)
		}
		clusterrequestResponse = append(clusterrequestResponse, *response)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterrequestResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetRegions godoc
// @Summary Get Regions
// @Description Get Regions
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} string
// @Router /regions [get]
func (server *Server) GetRegions(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	datas, err := clusterInterface.FindRegions(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    datas,
		Success: 1,
		Message: "Success",
	})
}

// GetAllRegions godoc
// @Summary Get All Regions
// @Description Get All Regions
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} string
// @Router /get-regions [get]
func (server *Server) GetAllRegions(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	datas, err := clusterInterface.FindAllRegions(server.DB, oid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    datas,
		Success: 1,
		Message: "Success",
	})
}

// GetCluster godoc
// @Summary Get Cluster
// @Description Get Cluster
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Cluster"
// @Success 200 {object} doc.Cluster
// @Router /cluster/{id} [get]
func (server *Server) GetCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := clusterInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	clusterResponse, err := CreateClusterDetailResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetClusterWithDetail godoc
// @Summary Get Cluster With Detail
// @Description Get Cluster With Detail
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Cluster Detail"
// @Success 200 {object} doc.Cluster
// @Router /admin/cluster/{id} [get]
func (server *Server) GetClusterWithDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := clusterInterface.FindWithDetails(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	clusterResponse, err := CreateClusterDetailResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdateCluster godoc
// @Summary Update Cluster
// @Description Update Cluster
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param file formData file true  "This is Update cluster file"
// @Param body body doc.Cluster true "Update Cluster"
// @Param id path int true "Update Cluster"
// @Success 200 {object} doc.Cluster
// @Router /cluster/{id} [put]
func (server *Server) UpdateCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Cluster{}
	err = server.DB.Model(models.Cluster{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cluster not found"))
		return
	}

	if err := r.ParseMultipartForm(2 << 20); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	usr, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if data.OrganizationID == 0 && !usr.IsAdmin {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you cannot edit zerone cluster"))
		return
	}
	if data.OrganizationID != uint64(oid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you dont have access to edit this cluster"))
		return
	}
	file, _, err := r.FormFile("file")
	configPath := ""
	if err == nil {
		configPath, _ = uploadConfigFile(file)
	}
	dataUpdate := models.Cluster{}
	dataUpdate.Name = r.FormValue("name")
	dataUpdate.ConfigPath = configPath
	dataUpdate.Context = r.FormValue("context")
	dataUpdate.Token = r.FormValue("token")
	dataUpdate.Region = r.FormValue("region")
	dataUpdate.Provider = r.FormValue("provider")
	dataUpdate.PrometheusServerUrl = r.FormValue("prometheus_server_url")
	imageRegistryID, err := strconv.ParseUint(r.FormValue("image_registry_id"), 10, 64)
	if err == nil {
		dataUpdate.ImageRegistryID = imageRegistryID
	}
	dnsId, err := strconv.ParseUint(r.FormValue("dns_id"), 10, 64)
	if err == nil {
		dataUpdate.DNSId = dnsId
	}
	// dataUpdate.StorageAccessKey = r.FormValue("storage_access_key")
	// dataUpdate.StorageSecretKey = r.FormValue("storage_secret_key")
	pvCapacity, _ := strconv.ParseInt(r.FormValue("pv_capacity"), 10, 64)
	data.PvCapacity = uint64(pvCapacity)
	dataUpdate.Zone = r.FormValue("zone")
	dataUpdate.Labels = r.FormValue("labels")
	active, err := strconv.ParseBool(r.FormValue("active"))
	if err != nil {
		active = data.Active
	}
	dataUpdate.Active = active
	node, _ := strconv.ParseInt(r.FormValue("nodes"), 10, 64)
	dataUpdate.Nodes = uint64(node)
	dataUpdate.ID = data.ID
	dataUpdated, err := clusterInterface.Update(server.DB, dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if dnsId != 0 {
		cluster, err := clusterInterface.FindWithDns(server.DB, uint64(data.ID))
		if err != nil {
			log.Errorln("error on finding Dns with cluster Id", err)
			responses.ERROR(w, http.StatusInternalServerError, err)
		}

		clusterRequest, err := clusterRequestInterface.FindWithDns(server.DB, cluster.ClusterRequestID)
		if err != nil {
			log.Errorln("error on getting Cluster request with Id", err)
			responses.ERROR(w, http.StatusNotFound, err)
		}
		if clusterRequest.Status == "applied" || clusterRequest.Status == "imported" {
			CheckV2Response(dataUpdated, r, w)
			return
		}
		file, _ := os.ReadFile("/data/helmfile/values.schema-update.json")
		var packages []models.PackageRequest
		err = json.Unmarshal(file, &packages)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
		data := PackageManagerModel{
			userId:         uid,
			organizationId: oid,
			packages:       packages,
			typ:            constants.UpdatePackage,
			clusterId:      data.ClusterRequestID,
		}
		log.Debug("Package manager data ::", data)
		err = server.PackageManager(data)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}

	}
	CheckV2Response(dataUpdated, r, w)
}

func CheckV2Response(data *models.Cluster, r *http.Request, w http.ResponseWriter) {
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data)
		return
	}
	clusterResponse, err := CreateClusterDetailResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterResponse,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Cluster{}
	err = server.DB.Model(models.Cluster{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}
	_, err = clusterInterface.Delete(server.DB, pid)
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

// ValidateDNSPermission godoc
// @Summary Validate DNS Permission
// @Description Validate DNS Permission
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.DNS true "Validate DNS"
// @Success 200 {object} map[string]interface{}
// @Router /check-dns-permissions [post]
func (server *Server) ValidateDNSPermission(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

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
	err = data.ValidateCredentials()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var hasPermission []string
	var noPermission []string
	if data.Provider == constants.GCP {
		hasPermission, noPermission, err = permission_helper.CheckGCPDNSPermission(data)
	} else if data.Provider == constants.EKS {
		hasPermission, noPermission, err = permission_helper.CheckAwsDNSPermissions(data)
	} else if data.Provider == constants.CloudFlare {
		hasPermission, noPermission, err = permission_helper.CheckCloudFlareDNSPermission(data)
	}
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	} else {
		responses.JSON(w, http.StatusOK, map[string]interface {
		}{
			"has_permission": hasPermission,
			"no_permission":  noPermission,
		})
	}
}

// GetClusterOverview godoc
// @Summary Get Cluster Overview
// @Description Get Cluster Overview
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Cluster Overview"
// @Success 200 {object} map[string]interface{}
// @Router /create-cluster/{id}/insights [post]
func (server *Server) GetClusterOverview(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid id"))
		return
	}
	cluster, err := clusterRequestInterface.Find(server.DB, pid)
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
	res, err := prometheus.GetInsightOverviewCluster(request, cluster.Cluster)
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

// GetClusterPipelineLogs godoc
// @Summary Get Cluster Pipeline Logs
// @Description Get Cluster Pipeline Logs
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Get Cluster Pipeline Logs"
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Bad Request"
// @Router /cluster/{id}/{pipeline}/{task}/{step} [get]
func (server *Server) GetClusterPipelineLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipeline := vars["pipeline"]
	task := vars["task"]
	step := vars["step"]
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	err = models.ValidatePipeline(pipeline, task, step)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	_, err = clusterInterface.Find(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	data, err := server.StoreClient.ClusterWOrkflow().GetCreateClusterWorkflow(pipeline)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("pipeline not found"))
		return
	}
	jsonbody, err := json.Marshal(data)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("invalid pipeline data"))
		return
	}
	meta := models.ClusterWorkflowMetadata{}
	if err := json.Unmarshal(jsonbody, &meta); err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("invalid pipeline data"))
		return
	}
	pipelineLogFilePATH := os.Getenv("PIPELINE_LOG_PATH")
	if pipelineLogFilePATH == "" {
		pipelineLogFilePATH = "/data/logs/cluster"
	}
	logFilePath := fmt.Sprintf("%s/%s/%s/%s/%s.log", pipelineLogFilePATH, meta.Namespace, pipeline, task, step)
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

// UpdateLabelAndColor godoc
// @Summary Update Label And Color
// @Description Update Label And Color
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Update Label and Color"
// @Param body body doc.Cluster true "Update Label and Color"
// @Success 200 {object} doc.Cluster
// @Router /create-cluster/{id}/label-color [put]
func (server *Server) UpdateLabelAndColor(w http.ResponseWriter, r *http.Request) {
	_, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Cluster{}
	request, err := clusterInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if request.OrganizationID != uint64(oid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to update cluster request"))
		return
	}
	data.ID = request.ID
	updatedData, err := clusterInterface.UpdateLabelsAndColor(server.DB, data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, updatedData)
		return
	}
	clusterResponse, err := CreateClusterResponse(updatedData)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterResponse,
		Success: 1,
		Message: "Success",
	})
}

func CreateClusterResponse(cluster *models.Cluster) (*doc.Cluster, error) {
	clusterResponse := doc.Cluster{}
	clusterBytes, _ := json.Marshal(cluster)
	err := json.Unmarshal(clusterBytes, &clusterResponse)
	if err != nil {
		return nil, err
	}
	return &clusterResponse, nil

}

func CreateClusterDetailResponse(cluster *models.Cluster) (*doc.ClusterDetails, error) {
	clusterResponse := doc.ClusterDetails{}
	clusterBytes, _ := json.Marshal(cluster)
	err := json.Unmarshal(clusterBytes, &clusterResponse)
	if err != nil {
		return nil, err
	}
	return &clusterResponse, nil

}
