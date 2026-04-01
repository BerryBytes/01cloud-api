package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"
	permission_helper "01cloud-api/api/utils/imrole"
	storage_helper "01cloud-api/api/utils/storage"
	"01cloud-api/api/utils/tfgenerator"

	"github.com/gorilla/mux"
	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
)

var clusterRequestInterface = models.NewClusterRequest()

// CreateClusterRequest godoc
// @Summary Create a new cluster
// @Description Create a new cluster with the input payload
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.ClusterRequest true "Create cluster"
// @Success 201 {object} doc.ClusterRequest
// @Router /create-cluster [post]
func (server *Server) CreateClusterRequest(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	data := models.ClusterRequest{}
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
	data.OrganizationID = uint64(oid)
	if data.OrganizationID != 0 {
		orgn, err := orgInterface.Find(server.DB, uint(data.OrganizationID))
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid organization"))
			return
		}
		data.Organization = orgn
	}

	cluster := models.Cluster{}
	cluster.Name = data.Name
	cluster.ConfigPath = ""
	cluster.Region = data.Region
	cluster.Provider = data.Provider
	cluster.ProjectName = data.ProjectId
	cluster.Zone = data.Region
	cluster.PvCapacity = 20
	cluster.OrganizationID = uint64(oid)
	cluster.Active = false
	if cluster.IsNameExists(server.DB) {
		responses.ERROR(w, http.StatusBadRequest, errors.New("duplicate name"))
		return
	}
	cs, _ := clusterInterface.Save(server.DB, cluster)
	data.ClusterID = uint64(cs.ID)
	dataCreated, err := clusterRequestInterface.Save(server.DB, &data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	cluster.ClusterRequestID = uint64(dataCreated.ID)
	cluster.ID = cs.ID
	_, _ = clusterInterface.Update(server.DB, cluster)
	cr, _ := clusterRequestInterface.Find(server.DB, uint64(dataCreated.ID))
	err = server.InitClusterCreation(cr)
	if err != nil {
		_, _ = clusterRequestInterface.Delete(server.DB, uint64(cr.ID))
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	// Checks if plan is created within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "plan", "cluster", data.Name, "")
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	clusterrequestResponse, err := CreateClusterRequestDetailResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    clusterrequestResponse,
		Success: 1,
		Message: "Success",
	})
}

// CreateClusterRequest godoc
// @Summary Create a new vcluster
// @Description Create a new cluster with the input payload
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.ClusterRequest true "Create VCluster"
// @Success 201 {object} doc.ClusterRequest
// @Router /create-cluster/vcluster [post]
func (server *Server) CreateVClusterRequest(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	req := models.VClusterRequest{}
	err = json.Unmarshal(body, &req)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = req.VClusterValidate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := req.NewClusterRequest()
	data.Status = "Provisioning"
	data.OrganizationID = uint64(oid)
	if data.OrganizationID != 0 {
		orgn, err := orgInterface.Find(server.DB, uint(data.OrganizationID))
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("organization not found"))
			return
		}
		data.Organization = orgn
	}

	cluster := models.Cluster{}
	cluster.Name = data.Name
	cluster.Region = data.Region
	cluster.Provider = data.Provider
	cluster.ProjectName = data.ProjectId
	cluster.Zone = data.Region
	cluster.PvCapacity = 20
	cluster.OrganizationID = uint64(oid)
	cluster.Active = false
	if cluster.IsNameExists(server.DB) {
		err := fmt.Errorf("already taken %s name", data.Name)
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	cs, _ := clusterInterface.Save(server.DB, cluster)
	data.ClusterID = uint64(cs.ID)
	dataCreated, err := clusterRequestInterface.Save(server.DB, data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	cluster.ClusterRequestID = uint64(dataCreated.ID)
	cluster.ID = cs.ID
	_, _ = clusterInterface.Update(server.DB, cluster)
	cr, _ := clusterRequestInterface.Find(server.DB, uint64(dataCreated.ID))
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "provision", "cluster", data.Name, "")
	}
	sessionId, _ := auth.ExtractSessionID(r)
	zeroneToken, err := auth.CreateToken(uid, oid, sessionId)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	zeroneCluster := getZeroneCluster()
	message := models.CreateClusterRequest{
		ID:              int64(dataCreated.ID),
		Name:            data.Name,
		BucketSecretKey: zeroneCluster.StorageSecretKey,
		BucketAccessKey: zeroneCluster.StorageAccessKey,
		Namespace:       helper.GetCreateClusterNamespace(dataCreated),
		ConfigPath:      zeroneCluster.ConfigPath,
		UserAccessKey:   fmt.Sprint(cs.ID),
		Type:            constants.VClusterProvision,
		Region:          data.Region,
		SubscriptonID:   data.SubscriptionID,
		VClusterAPIKEY:  req.VClusterAPIKEY,
		ZeroneAPIKEY:    zeroneToken,
		Kube_version:    req.Kube_version,
	}
	err = queue.Publish(constants.VClusterProvision, message)
	if err != nil {
		_, _ = clusterRequestInterface.Delete(server.DB, uint64(cr.ID))
		log.Error("vcluster create message published error :: ", err)
	}
	response, err := CreateClusterRequestDetailResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    response,
		Success: 1,
		Message: "Success",
	})
}

// GetCreateClusters godoc
// @Summary Get clusters
// @Description Get list of organization's clusters
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.ClusterRequest
// @Router /create-cluster/vcluster/validate [get]
func (server *Server) ValidateVClusterToken(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("vtoken")
	url := os.Getenv("VCLUSTER_API_URL")
	if url == "" {
		url = "https://zerone-4409-9534.01cloud.com/v1"
	}
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/users/verify-token", url), nil)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	if token == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("required token"))
		return
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	res, err := client.Do(req)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusAccepted {
		if res.StatusCode == http.StatusForbidden {
			responses.ERROR(w, http.StatusForbidden, errors.New("user is not authorized to create cluster"))
			return
		}
		responses.ERROR(w, http.StatusBadRequest, errors.New("bad request"))
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    "ok",
		Success: 1,
		Message: "Success",
	})
}

// GetCreateClusters godoc
// @Summary Get clusters
// @Description Get list of organization's clusters
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.ClusterRequest
// @Router /create-cluster [get]
func (server *Server) GetCreateClusterRequests(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
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

func (server *Server) InitClusterCreation(request *models.ClusterRequest) error {
	path := tfgenerator.GetTrafformFolder(request)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0777)
		if err != nil {
			return err
		}
	}
	err = helper.CopyDirectory(fmt.Sprintf("%s/%s", "/data/terraform", request.Provider), path)
	if err != nil {
		return err
	}

	if request.Credential != "" {
		err = helper.Copy(request.Credential, path+"/credentials.json")
		if err != nil {
			return err
		}
	}
	err = tfgenerator.DuplicateNodegroupname(request)
	if err != nil {
		return err
	}
	tfVars, variables := tfgenerator.GenerateTFvars(request)
	if tfVars == nil || variables == nil {
		return errors.New("unable to generate variables please check your input")
	}
	var mainTf []byte
	if request.Provider == constants.GCP {
		mainTf = tfgenerator.GenerateMainTfGcp(request)
	} else if request.Provider == constants.EKS {
		mainTf = tfgenerator.GenerateMainTfAws(request)
	}
	backendTf := tfgenerator.GenerateBackendTF(request)
	err = os.WriteFile(fmt.Sprintf("%s/%s", path, "terraform.tfvars"), tfVars, 0777)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s", path, "variables.tf"), variables, 0777)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s", path, "main.tf"), mainTf, 0777)
	if err != nil {
		return err
	}
	err = os.WriteFile(fmt.Sprintf("%s/%s", path, "backend.tf"), backendTf, 0777)
	if err != nil {
		return err
	}
	err = helper.CreateTarFile(path, getTarFileName(request))
	if err != nil {
		return err
	}
	err = os.RemoveAll(path)
	if err != nil {
		return err
	}
	bucketName := tfgenerator.GetBucketName(request.Organization)
	objectName := tfgenerator.GetObjectName(request)
	_ = storage_helper.CreateBucket(server.StorageClient, os.Getenv("GCLOUD_PROJECT"), bucketName)
	//if err != nil {
	//	return err
	//}
	err = storage_helper.UploadFile(server.StorageClient, tfgenerator.GetBucketName(request.Organization), objectName, getTarFileName(request))
	if err != nil {
		log.Error(err)
		return err
	}
	request.Status = constants.ClusterPlaning
	_, _ = clusterRequestInterface.Update(server.DB, request)
	//zeroneCluster := models.Cluster{}
	//_, err = zeroneCluster.FindZeroneCluster(server.DB)
	//if err != nil {
	//	return err
	//}
	//caFilePath := certificate.GetCertificateFolder(&zeroneCluster)
	//caFile := fmt.Sprintf("%s/ca.pem", caFilePath)
	zeroneCluster := getZeroneCluster()
	req := models.CreateClusterRequest{
		ID:              int64(request.ID),
		Name:            request.Name,
		BucketSecretKey: zeroneCluster.StorageSecretKey,
		BucketAccessKey: zeroneCluster.StorageAccessKey,
		UserAccessKey:   request.AccessKey,
		UserSecretKey:   request.SecretKey,
		BucketName:      bucketName,
		BucketPath:      objectName,
		Namespace:       helper.GetCreateClusterNamespace(request),
		ConfigPath:      zeroneCluster.ConfigPath,
		Type:            constants.ClusterPlan,
		Region:          request.Region,
	}
	err = queue.Publish(constants.CreateCluster, req)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// GetClusterConfig godoc
// @Summary Get create cluster Config
// @Description Get create cluster config by provider
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param provider query string true "Provider"
// @Success 200 {object} doc.ObjectResponse
// @Router /create-cluster-config [get]
func (server *Server) GetClusterCreationConfig(w http.ResponseWriter, r *http.Request) {
	provider := r.FormValue("provider")
	if provider == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("provider is required"))
		return
	}
	res := make(map[string]interface{})
	file, err := os.ReadFile(fmt.Sprintf("/data/public/%s.config.schema.json", provider))
	if err == nil {
		err := json.Unmarshal(file, &res)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
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

func getTarFileName(request *models.ClusterRequest) string {
	organizationName := "zerone"
	if request.Organization != nil {
		organizationName = slug.Make(request.Organization.Name)
	}
	return fmt.Sprintf("/data/terraform/%s-%s-%s.tgz", organizationName, slug.Make(request.Name), request.Provider)
}

// ApplyTerraform godoc
// @Summary Initiate apply process
// @Description Initiate cluster apply process by id
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/apply [get]
func (server *Server) ApplyTerraform(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
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
	request, err := clusterRequestInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uint64(oid) != request.OrganizationID {
		responses.ERROR(w, http.StatusNotFound, errors.New("you dont have access to change this"))
		return
	}
	if request.OrganizationID != 0 {
		orgn, err := orgInterface.Find(server.DB, uint(request.OrganizationID))
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid organization"))
			return
		}
		request.Organization = orgn
	}

	//zeroneCluster := models.Cluster{}
	//_, err = zeroneCluster.FindZeroneCluster(server.DB)
	//if err != nil {
	//	responses.JSON(w, http.StatusInternalServerError, errors.New("Internal server error, please contact support team"))
	//	return
	//}
	err = isInProgress(request.Status)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, err)
		return
	}
	request.Status = constants.ClusterApplying
	_, _ = clusterRequestInterface.Update(server.DB, request)
	zeroneCluster := getZeroneCluster()
	req := models.CreateClusterRequest{
		ID:              int64(pid),
		Name:            request.Name,
		BucketSecretKey: zeroneCluster.StorageSecretKey,
		BucketAccessKey: zeroneCluster.StorageAccessKey,
		UserAccessKey:   request.AccessKey,
		UserSecretKey:   request.SecretKey,
		BucketName:      tfgenerator.GetBucketName(request.Organization),
		BucketPath:      tfgenerator.GetObjectName(request),
		Namespace:       helper.GetCreateClusterNamespace(request),
		ConfigPath:      zeroneCluster.ConfigPath,
		Type:            constants.ClusterApply,
		Region:          request.Region,
	}
	err = queue.Publish(constants.CreateCluster, req)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	// Checks if cluster is created within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "create", "cluster", request.Name, "")
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": "Cluster creation process initiated",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func isInProgress(status string) error {
	if status == constants.ClusterApplying || status == constants.ClusterPlaning || status == constants.PackageInstalling {
		return errors.New("your are not allowed to apply , another process is running")
	}
	if status == constants.ClusterStatusApplied {
		return errors.New("you cannot update applied cluster request")
	}
	return nil
}

// CancelPlan godoc
// @Summary Cancel cluster creation process
// @Description Cancel cluster creation process by id
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param workflow_name query string true "workflow_name"
// @Param id path int true "Cluster request id"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/cancel [get]
func (server *Server) CancelPlan(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
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
	//data := models.ClusterRequest{}
	request, err := clusterRequestInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}

	workFlowName := r.FormValue("workflow_name")
	if workFlowName == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("workflow name is required"))
		return
	}
	if uint64(oid) != request.OrganizationID {
		responses.ERROR(w, http.StatusNotFound, errors.New("you dont have access to change this"))
		return
	}
	if request.OrganizationID != 0 {
		orgn, err := orgInterface.Find(server.DB, uint(request.OrganizationID))
		if err != nil {
			responses.ERROR(w, http.StatusNotFound, errors.New("invalid organization"))
			return
		}
		request.Organization = orgn
	}

	if request.Status == constants.ClusterApply || request.Status == constants.ClusterStatusPackageInstalled {
		responses.ERROR(w, http.StatusForbidden, errors.New("you cannot update applied cluster request"))
		return
	}
	meta, err := server.StoreClient.ClusterWOrkflow().GetCreateClusterWorkflow(workFlowName)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if meta == nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("workflow not found"))
		return
	}
	req := models.CreateClusterRequest{
		ID:              int64(pid),
		Name:            request.Name,
		BucketSecretKey: meta.CIRequest.BucketSecretKey,
		BucketAccessKey: meta.CIRequest.BucketAccessKey,
		UserAccessKey:   request.AccessKey,
		UserSecretKey:   request.SecretKey,
		BucketName:      tfgenerator.GetBucketName(request.Organization),
		BucketPath:      tfgenerator.GetObjectName(request),
		Namespace:       helper.GetCreateClusterNamespace(request),
		ConfigPath:      meta.CIRequest.ConfigPath,
		Type:            constants.ClusterApply,
		Region:          request.Region,
	}
	err = queue.Publish(constants.CreateCluster, req)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": "Plan cancel process initiated",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// EnableDisableCluster godoc
// @Summary Change cluster status
// @Description Change cluster status by id
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Cluster request id"
// @Param body body doc.ClusterRequest true "Create Resource"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/change-active-status [post]
func (server *Server) EnableDisableCluster(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
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
	request, err := clusterRequestInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if uint64(oid) != request.OrganizationID {
		responses.ERROR(w, http.StatusNotFound, errors.New("you dont have access to change this"))
		return
	}
	orgn, err := request.ValidateOrganization(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	request.Organization = orgn

	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	input := models.ClusterRequest{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if input.Active && request.Status != constants.ClusterStatusPackageInstalled {
		responses.ERROR(w, http.StatusNotFound, errors.New("complete cluster setup first to change status of this cluster"))
		return
	}
	if request.Cluster == nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("cluster not available"))
		return
	}
	input.ID = uint(pid)
	request.Cluster.Active = input.Active
	_, err = clusterInterface.Update(server.DB, *request.Cluster)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to change status"))
		return
	}
	// Checks if cluster is enabled or Disabled within Orgination level then we have to add audit activity
	var currentStatus string = "inactive"
	if request.Active {
		currentStatus = "active"
	}
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "update", "cluster", request.Name, "to status "+currentStatus)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": "Status changed successfully",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// UpdateClusterRequest godoc
// @Summary Update a ClusterRequest
// @Description Update a ClusterRequest with the input payload
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Param body body doc.ClusterRequest true "Update ClusterRequest"
// @Success 200 {object} doc.ClusterRequest
// @Router /create-cluster/{id} [put]
func (server *Server) UpdateClusterRequest(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	data := models.ClusterRequest{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	//dm := models.ClusterRequest{}
	dataMain, err := clusterRequestInterface.Find(server.DB, cid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cluster request not found"))
		return
	}
	if dataMain.Status == constants.ClusterPlaning || dataMain.Status == constants.ClusterApplying {
		responses.ERROR(w, http.StatusNotFound, errors.New("cannot edit cluster in planing state"))
		return
	}
	if uint64(oid) != dataMain.OrganizationID {
		responses.ERROR(w, http.StatusForbidden, errors.New("you dont have access to change this"))
		return
	}
	if data.Version != "" {
		file, _ := os.ReadFile(fmt.Sprintf("/data/public/%s.config.schema.json", dataMain.Provider))
		res := make(map[string]interface{})
		err = json.Unmarshal(file, &res)
		if err != nil {
			responses.ERROR(w, http.StatusForbidden, errors.New("schema not available"))
			return
		}
		versionList := res["properties"].(map[string]interface{})["cluster_version"].(map[string]interface{})["items"].([]interface{})

		if indexOf(dataMain.Version, versionList) < indexOf(data.Version, versionList) {
			responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("you cannot downgrade version"))
			return
		}
	}
	data.ID = uint(cid)
	dataUpdated, err := clusterRequestInterface.Update(server.DB, &data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	cr, _ := clusterRequestInterface.Find(server.DB, uint64(dataUpdated.ID))
	if dataMain.Type != constants.TypeImported {
		cr.Status = constants.ClusterPlaning
		_, _ = clusterRequestInterface.Update(server.DB, cr)
		err = server.InitClusterCreation(cr)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, errors.New("input validation failed"))
			return
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	clusterrequestResponse, err := CreateClusterRequestDetailResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterrequestResponse,
		Success: 1,
		Message: "Success",
	})
}

func indexOf(word string, data []interface{}) int {
	for k, v := range data {
		if word == v.(string) {
			return k
		}
	}
	return -1
}

// CheckPermission godoc
// @Summary Check credential permission
// @Description Check credential permission by provider
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.ClusterRequest true "Update ClusterRequest"
// @Success 200 {object} doc.ObjectResponse
// @Router /check-permissions [post]
func (server *Server) CheckPermission(w http.ResponseWriter, r *http.Request) {
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

	data := models.ClusterRequest{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = data.ValidatePermission()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var hasPermission []string
	var noPermission []string
	if data.Provider == constants.GCP {
		hasPermission, noPermission, err = permission_helper.CheckGcpPermissions(data)
	} else {
		hasPermission, noPermission, err = permission_helper.CheckAwsPermissions(data)
	}
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	} else {
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, map[string]interface {
			}{
				"has_permission": hasPermission,
				"no_permission":  noPermission,
			})
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data: map[string]interface {
			}{
				"has_permission": hasPermission,
				"no_permission":  noPermission,
			},
			Success: 1,
			Message: "Success",
		})
	}
}

// GetClusterRequest godoc
// @Summary Get cluster request by id
// @Description Get ClusterRequest request by id from token
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Success 200 {object} doc.ClusterRequest
// @Router /create-cluster/{id} [get]
func (server *Server) GetClusterRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := clusterRequestInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	clusterrequestResponse, err := CreateClusterRequestDetailResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    clusterrequestResponse,
		Success: 1,
		Message: "Success",
	})
}

// DestroyClusterRequest godoc
// @Summary Destroy cluster request by id
// @Description Destroy ClusterRequest request by id from token
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Success 200 {object} doc.ClusterRequest
// @Router /create-cluster/{id}/destroy [delete]
func (server *Server) DestroyClusterRequest(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
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
	data, err := clusterRequestInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("unauthorized"))
		return
	}
	if data.Status == constants.ClusterStatusApplied || data.Status == constants.ClusterStatusPackageInstalled {
		zeroneCluster := getZeroneCluster()
		req := models.CreateClusterRequest{
			ID:               int64(pid),
			Name:             data.Name,
			BucketSecretKey:  zeroneCluster.StorageSecretKey,
			BucketAccessKey:  zeroneCluster.StorageAccessKey,
			UserAccessKey:    data.AccessKey,
			UserSecretKey:    data.SecretKey,
			BucketName:       tfgenerator.GetBucketName(data.Organization),
			BucketPath:       tfgenerator.GetObjectName(data),
			Namespace:        helper.GetCreateClusterNamespace(data),
			ConfigPath:       zeroneCluster.ConfigPath,
			Type:             constants.ClusterDestroy,
			Region:           data.Region,
			KubeConfigString: data.Cluster.ConfigPath,
		}
		err = queue.Publish(constants.CreateCluster, req)
		if err == nil {
			data.Cluster.Active = false
			_, _ = clusterInterface.Update(server.DB, *data.Cluster)
		}
	}
	// Checks if cluster is destroyed within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "destroy", "cluster", data.Name, "")
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": "Cluster destroy process initiated",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
	//_, err = data.Delete(server.DB, pid)
	//if err != nil {
	//	responses.ERROR(w, http.StatusBadRequest, err)
	//	return
	//}
	//responses.JSON(w, http.StatusNoContent, "")
}

// DeleteClusterRequest godoc
// @Summary Delete cluster request by id
// @Description Delete ClusterRequest request by id from token
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Success 200 {object} doc.ClusterRequest
// @Router /create-cluster/{id} [delete]
func (server *Server) DeleteClusterRequest(w http.ResponseWriter, r *http.Request) {
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
	data := models.ClusterRequest{}
	err = server.DB.Model(models.ClusterRequest{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("unauthorized"))
		return
	}
	if data.Status == constants.ClusterStatusApplied {
		responses.ERROR(w, http.StatusOK, errors.New("you can not delete running cluster"))
	}
	_, err = clusterRequestInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	_, err = clusterInterface.DeleteByClusterRequest(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	// Checks if cluster is destroyed within Orgination level then we have to add audit activity
	if oid > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, oid, "delete", "cluster", data.Name, "")
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

// GetClusterRequestWorkflows godoc
// @Summary Get cluster request workflow by id
// @Description Get cluster request workflow list by id
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Param page query int true "page number"
// @Param limit query int true "limit"
// @Success 200 {object} doc.ObjectResponse
// @Router /create-cluster/{id}/workflows [get]
func (server *Server) GetClusterRequestWorkflows(w http.ResponseWriter, r *http.Request) {
	uid, orgId, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	page, err := strconv.ParseUint(r.FormValue("page"), 10, 64)
	if err != nil || page < 1 {
		page = 0
	} else {
		page = page - 1
	}
	limit, err := strconv.ParseUint(r.FormValue("limit"), 10, 64)
	if err != nil {
		limit = 100
	}
	clusterReceived, err := clusterRequestInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if clusterReceived.OrganizationID != uint64(orgId) && !user.IsAdmin {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view workflow"))
		return
	}

	status, err := server.StoreClient.ClusterWOrkflow().GetClusterWorkflowDetail(helper.GetCreateClusterNamespace(clusterReceived), int(page), int(limit))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"page":  page + 1,
			"limit": limit,
			"data":  status,
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]interface{}{
			"page":  page + 1,
			"limit": limit,
			"data":  status,
		},
		Success: 1,
		Message: "Success",
	})
}

// DownloadTfScripts godoc
// @Summary Download terraform scripts by id
// @Description Download terraform scripts by cluster request id
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/download-files [get]
func (server *Server) DownloadTfScripts(w http.ResponseWriter, r *http.Request) {
	uid, orgId, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	clusterReceived, err := clusterRequestInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if clusterReceived.OrganizationID != uint64(orgId) && !user.IsAdmin {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	downloadUrl := ""
	if clusterReceived.Cluster != nil && clusterReceived.Cluster.ConfigPath != "" {
		sourcePath := clusterReceived.Cluster.ConfigPath
		destPath := strings.Replace(sourcePath, "config", "uploads", 1)
		log.Debugf("File Source Path :: %s Destination Path %s:: ", sourcePath, destPath)
		err := helper.Copy(sourcePath, destPath)
		if err != nil {
			responses.ERROR(w, http.StatusBadRequest, errors.New("unable to copy config file"))
			return
		}
		downloadUrl = fmt.Sprintf("%s%s", os.Getenv("BASE_URL"), destPath)
	}

	if os.Getenv("DOWNLOAD_SOURCE") == "s3" {
		objectName := tfgenerator.GetObjectName(clusterReceived)
		downloadUrl, err = storage_helper.DownloadFile(server.StorageClient, tfgenerator.GetBucketName(clusterReceived.Organization), objectName)
		if err != nil {
			responses.ERROR(w, http.StatusBadRequest, errors.New("unable to download files"))
			return
		}
	}
	if downloadUrl == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("unable to download files"))
		return
	}
	//bucketName := tfgenerator.GetBucketName(clusterReceived.Organization)
	//err = storage_helper.CreateBucket(server.StorageClient, os.Getenv("GCLOUD_PROJECT"), bucketName)

	// Checks if user opens export script within Orgination level then we have to add audit activity
	if orgId > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(uid, orgId, "open", "cluster", clusterReceived.Name, "export script")
	}
	log.Debugf("Download file URL :: %s", downloadUrl)
	w.Header().Set("Content-Type", "text/plain")
	file, err := os.Open(downloadUrl)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("unable to download files"))
		return
	}
	defer file.Close()
	_, err = io.Copy(w, file)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("unable to download files"))
		return
	}
}

func (server *Server) GetClusterRequestWorkflowLog(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	workflowName := r.FormValue("workflow_name")
	if workflowName == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("workflow name is required"))
		return
	}
	clusterReceived, err := clusterRequestInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	if clusterReceived.OrganizationID != uint64(oid) && !user.IsAdmin {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to view workflow"))
		return
	}
	status, _ := server.StoreClient.ClusterWOrkflow().GetCreateClusterWorkflowLog(workflowName)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, status)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    status,
		Success: 1,
		Message: "Success",
	})
}

// GetPermission godoc
// @Summary Get Permission
// @Description Get Permission
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param provider query string true "Provider"
// @Success 200 {string} string
// @Router /get-permissions [get]
func (server *Server) GetPermission(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	provider := r.URL.Query().Get("provider")
	if provider == "" {
		responses.ERROR(w, http.StatusNotAcceptable, errors.New("provider must be specified"))
		return
	}

	if provider != constants.GCP && provider != constants.EKS {
		responses.ERROR(w, http.StatusBadRequest, errors.New("provide valid provider"))
		return
	}
	if provider == constants.GCP {
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, constants.GcpRequiredPermission)
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    constants.GcpRequiredPermission,
			Success: 1,
			Message: "Success",
		})
	} else {
		if !helper.IsV2(r) {
			responses.JSON(w, http.StatusOK, constants.AwsRequiredPermission)
			return
		}
		responses.JSON(w, http.StatusOK, responses.Response{
			Data:    constants.AwsRequiredPermission,
			Success: 1,
			Message: "Success",
		})
	}
}

func CreateClusterRequestResponse(clusterRequest *models.ClusterRequest) (*doc.ClusterRequest, error) {
	response := doc.ClusterRequest{}
	requestBytes, _ := json.Marshal(clusterRequest)
	err := json.Unmarshal(requestBytes, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func CreateClusterRequestDetailResponse(clusterRequest *models.ClusterRequest) (*doc.ClusterRequestDetails, error) {
	response := doc.ClusterRequestDetails{}
	requestBytes, _ := json.Marshal(clusterRequest)
	err := json.Unmarshal(requestBytes, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil

}

//todo : check roles
//api should be enabled:
//Compute Engine API - compute.googleapis.com
//Kubernetes Engine API - container.googleapis.com
//Filestore instance API - file.googleapis.com
//roles/compute.networkAdmin    Compute Network Admin
//roles/compute.securityAdmin   Compute Security Admin
//roles/file.editor             Cloud Filestore Editor
//roles/compute.viewer          Compute Viewer
//roles/container.admin         Kubernetes Engine Admin
//roles/container.clusterAdmin  Kubernetes Engine Cluster Admin
//roles/container.developer     Kubernetes Engine Developer
//roles/iam.serviceAccountAdmin Service Account Admin
//roles/iam.serviceAccountUser  Service Account User
