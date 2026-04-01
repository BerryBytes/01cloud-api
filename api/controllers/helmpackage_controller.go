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
	"time"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"
	"01cloud-api/api/utils/tfgenerator"

	store "github.com/berrybytes/01cloud-store"
	"github.com/ghodss/yaml"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type PackageManagerModel struct {
	clusterId      uint64
	userId         uint
	organizationId uint
	packages       []models.PackageRequest
	typ            string
}

func (server *Server) InstallUninstallPackage(w http.ResponseWriter, r *http.Request, typ string) {
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
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to read body"))
		return
	}
	var packages []models.PackageRequest
	err = json.Unmarshal(body, &packages)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("unable to parse body"))
		return
	}
	data := PackageManagerModel{
		userId:         uid,
		organizationId: oid,
		packages:       packages,
		typ:            typ,
		clusterId:      pid,
	}
	err = server.PackageManager(data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": "Package installation triggered",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) PackageManager(managePkgModel PackageManagerModel) error {
	request, err := clusterRequestInterface.FindWithDns(server.DB, managePkgModel.clusterId)
	if err != nil {
		logrus.Errorln("Error on getting Cluster request with Id", err)
		return err
	}
	if uint64(managePkgModel.organizationId) != request.OrganizationID {
		return errors.New("you dont have access to change this")
	}
	if request.Cluster.DNSId == 0 {
		return errors.New("please add dns detail first to install packages")
	}
	if request.OrganizationID != 0 {
		orgn, err := orgInterface.Find(server.DB, uint(request.OrganizationID))
		if err != nil {
			return errors.New("invalid organization")
		}
		request.Organization = orgn
	}
	zeroneCluster := getZeroneCluster()
	kubeconfig := []byte("")
	if request.Cluster.ConfigPath != "" {
		kubeconfig, err = os.ReadFile(request.Cluster.ConfigPath)
		if err != nil {
			logrus.Errorln("unable to open config file ", err)
			return errors.New("unable to open config file")
		}
	}
	packagesYamlWithEnv := []models.PackageYaml{}
	for _, packageRequest := range managePkgModel.packages {
		packageYaml := models.PackageYaml{}
		packageYaml.Name = packageRequest.Name
		packageYaml.Namespace = GetPackageNamespace(packageRequest.Name)
		packageYaml.Chart = packageRequest.Chart
		packageYaml, err = injectDnsEnvs(packageRequest, packageYaml, request)
		if err != nil {
			return err
		}
		packageYaml, err = injectSecretPatcherEnvs(packageRequest, packageYaml, request)
		if err != nil {
			return err
		}
		packageYaml, err = injectVeleroEnvs(packageRequest, packageYaml, request)
		if err != nil {
			return err
		}
		packageYaml, err = injectJobEnvs(packageRequest, packageYaml, request)
		if err != nil {
			return err
		}
		packageYaml, _ = injectPrometheusAlertManagerEnvs(packageRequest, packageYaml, request)
		if packageYaml.Name == "contour" {
			packageYaml.Set = append(packageYaml.Set, getPackageEnvironment("envoy.useHostPort", "false"))
		}
		packagesYamlWithEnv = append(packagesYamlWithEnv, packageYaml)
	}
	repoUrl := os.Getenv("HELM_CHART_URL")
	if repoUrl == "" {
		repoUrl = "https://berrybytes.github.io/helm-chart-org"
	}
	helmFile := map[string]interface{}{
		"repositories": []map[string]string{
			{"name": "zerone", "url": repoUrl},
		},
		"releases": packagesYamlWithEnv,
	}
	helmbyte, err := yaml.Marshal(helmFile)
	if err != nil {
		return errors.New("invalid input")
	}
	request.Status = constants.PackageInstalling
	_, _ = clusterRequestInterface.Update(server.DB, request)

	req := models.CreateClusterRequest{
		ID:               int64(managePkgModel.clusterId),
		Name:             request.Name,
		BucketSecretKey:  zeroneCluster.StorageSecretKey,
		BucketAccessKey:  zeroneCluster.StorageAccessKey,
		UserAccessKey:    request.AccessKey,
		UserSecretKey:    request.SecretKey,
		BucketName:       tfgenerator.GetBucketName(request.Organization),
		BucketPath:       tfgenerator.GetObjectName(request),
		Namespace:        helper.GetCreateClusterNamespace(request),
		ConfigPath:       zeroneCluster.ConfigPath,
		Type:             managePkgModel.typ,
		Region:           request.Region,
		KubeConfigString: string(kubeconfig),
		HelmFileString:   string(helmbyte),
	}

	err = queue.Publish(constants.CreateCluster, req)
	if err != nil {
		return err
	}

	// Checks if package is install or uninstall within Orgination level then we have to add audit activity
	if managePkgModel.organizationId > 0 {
		_, _ = server.SaveOrganizationActivityWithJson(managePkgModel.userId, managePkgModel.organizationId, "update", "cluster", request.Name, "by "+managePkgModel.typ)
	}
	return nil
}

type ZeroneCluster struct {
	Name             string
	StorageAccessKey string
	StorageSecretKey string
	ConfigPath       string
}

func getZeroneCluster() ZeroneCluster {
	return ZeroneCluster{
		Name:             "GKE",
		StorageAccessKey: os.Getenv("STORAGE_ACCESS_KEY"),
		StorageSecretKey: os.Getenv("STORAGE_SECRET_KEY"),
		ConfigPath:       os.Getenv("ZERONE_CONFIG_PATH"),
	}
}

func injectDnsEnvs(packageRequest models.PackageRequest, packageYaml models.PackageYaml, request *models.ClusterRequest) (models.PackageYaml, error) {
	if packageRequest.RequiredDNS {
		credentials := ""
		if request.Cluster.DNS.Provider == constants.GCP {
			if request.Cluster.DNS.Credential == "" {
				return packageYaml, errors.New("DNS is not setup properly")
			}
			credentials = helper.ReadFileInB64String(request.Cluster.DNS.Credential)
			if credentials == "" {
				return packageYaml, errors.New("invalid gcp credentials")
			}
		}
		if request.Cluster.DNS.Provider == constants.CloudFlare {
			if request.Cluster.DNS.Credential == "" {
				return packageYaml, errors.New("DNS is not setup properly")
			}
			request.Cluster.DNS.ZoneID = request.Cluster.DNS.BaseDomain
			credentials = request.Cluster.DNS.Credential
		}
		envs := packageRequest.Set
		envs = append(envs, getPackageEnvironment("dns.provider", request.Cluster.DNS.Provider))
		envs = append(envs, getPackageEnvironment("dns.baseDomain", request.Cluster.DNS.BaseDomain))
		envs = append(envs, getPackageEnvironment("cluster.provider", request.Cluster.Provider))
		envs = append(envs, getPackageEnvironment("dns.zone", request.Cluster.DNS.ZoneID))
		envs = append(envs, getPackageEnvironment("gcp.gcloudProject", request.Cluster.DNS.ProjectId))
		envs = append(envs, getPackageEnvironment("gcp.googleAppCred", credentials))
		envs = append(envs, getPackageEnvironment("cloudflare.email", request.Cluster.DNS.ProjectId))
		envs = append(envs, getPackageEnvironment("cloudflare.apiKey", request.Cluster.DNS.Credential))
		envs = append(envs, getPackageEnvironment("aws.awsRegion", request.Cluster.DNS.Region))
		envs = append(envs, getPackageEnvironment("aws.awsSecretKey", request.Cluster.DNS.SecretKey))
		envs = append(envs, getPackageEnvironment("aws.awsAccessKey", request.Cluster.DNS.AccessKey))
		packageYaml.Set = envs
	}
	return packageYaml, nil
}

func injectJobEnvs(packageRequest models.PackageRequest, packageYaml models.PackageYaml, request *models.ClusterRequest) (models.PackageYaml, error) {
	if packageRequest.Name == "zerone-jobs" {
		envs := packageYaml.Set
		envs = append(envs, getPackageEnvironment("api.endpoint", os.Getenv("BASE_URL")+"/joblog"))
		envs = append(envs, getPackageEnvironment("secrets.value", os.Getenv("API_SECRET")))
		packageYaml.Set = envs
	}
	return packageYaml, nil
}

func injectPrometheusAlertManagerEnvs(packageRequest models.PackageRequest, packageYaml models.PackageYaml, request *models.ClusterRequest) (models.PackageYaml, error) {
	if packageRequest.Name == "prometheus-operator" {
		envs := packageYaml.Set
		envs = append(envs, getPackageEnvironment("prometheus-node-exporter.hostNetwork", "false"))
		cortexUrl := os.Getenv("CORTEX_URL")
		if cortexUrl == "" {
			cortexUrl = "https://cortex.test.01cloud.dev"
		}
		resp, err := alertManagerRequest(request.ClusterID)
		if err != nil {
			logrus.Info("alertmanager request err :: ", err)
			return packageYaml, err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				logrus.Error("Error reading response body:", err)
				return packageYaml, err
			}
			type respStruct struct {
				Message string `json:"message"`
				Success int    `json:"success"`
				Data    struct {
					TenantName string `json:"tenant_name"`
					Password   string `json:"password"`
				} `json:"data"`
			}
			userResp := respStruct{}
			err = json.Unmarshal(respBody, &userResp)
			if err != nil {
				logrus.Error("error unmarshalling json respnose:", err)
				return packageYaml, err
			}
			if userResp.Data.TenantName != "" && userResp.Data.Password != "" {
				envs = append(envs, getPackageEnvironment("server.remoteWrite[0].url", cortexUrl+"/api/prom/push"))
				envs = append(envs, getPackageEnvironment("server.remoteWrite[0].basic_auth.username", userResp.Data.TenantName))
				envs = append(envs, getPackageEnvironment("server.remoteWrite[0].basic_auth.password", userResp.Data.Password))
				packageYaml.Set = envs
			}
		}
	}
	return packageYaml, nil
}

func alertManagerRequest(clusterId uint64) (*http.Response, error) {
	requestURL := fmt.Sprintf("%s/monitoring/tenants/%d?token=%s", os.Getenv("BASE_URL"), clusterId, os.Getenv("API_SECRET"))
	resp, err := http.Get(requestURL)
	if err != nil {
		logrus.Error("Error sending get request to alertmanager ::", err)
		return nil, err
	}
	if resp.StatusCode == http.StatusBadRequest {
		resp, err = http.Post(requestURL, "application/json", nil)
		if err != nil {
			logrus.Error("Error sending post request to alertmanager::", err)
			return nil, err
		}
	}
	return resp, nil
}

func injectVeleroEnvs(packageRequest models.PackageRequest, packageYaml models.PackageYaml, request *models.ClusterRequest) (models.PackageYaml, error) {
	if packageRequest.Name == "velero" {
		packageYaml.Namespace = "velero"
		if request.Cluster.CloudStorage.RawMessage != nil {
			envs := packageYaml.Set
			var storageMap map[string]string
			err := json.Unmarshal(request.Cluster.CloudStorage.RawMessage, &storageMap)
			if err != nil {
				return packageYaml, errors.New("invalid storage data")
			}
			if storageMap["provider"] == constants.GCP {
				cred, ok := storageMap["credentials"]
				if !ok {
					return packageYaml, errors.New("unable tp open config file")
				}
				credentials := helper.ReadFileInString(cred)
				if credentials == "" {
					return packageYaml, errors.New("invalid gcp credentials")
				}
				credentials = strings.ReplaceAll(credentials, `"`, `\"`)
				credentials = strings.ReplaceAll(credentials, `\n`, `\\\n`)
				envs = append(envs, getPackageEnvironment("credentials.secretContents.cloud", credentials))
				envs = append(envs, getPackageEnvironment("initContainers[0].name", "velero-plugin-for-gcp"))
				envs = append(envs, getPackageEnvironment("initContainers[0].image", "velero/velero-plugin-for-gcp:v1.1.0"))
			}
			if storageMap["provider"] == constants.EKS {
				creds := "[default]\naws_access_key_id=" + storageMap["access_key"] + "\naws_secret_access_key=" + storageMap["secret_key"]
				envs = append(envs, getPackageEnvironment("credentials.secretContents.cloud", creds))
				envs = append(envs, getPackageEnvironment("initContainers[0].name", "velero-plugin-for-aws"))
				envs = append(envs, getPackageEnvironment("initContainers[0].image", "velero/velero-plugin-for-aws:v1.1.0"))
			}
			bucket := fmt.Sprintf("zerone-bucket-%d", request.Cluster.ID)
			envs = append(envs, getPackageEnvironment("configuration.backupStorageLocation[0].provider", storageMap["provider"]))
			envs = append(envs, getPackageEnvironment("configuration.backupStorageLocation[0].bucket", bucket))
			envs = append(envs, getPackageEnvironment("configuration.volumeSnapshotLocation[0].provider", storageMap["provider"]))
			envs = append(envs, getPackageEnvironment("initContainers[0].volumeMounts[0].mountPath", "/target"))
			envs = append(envs, getPackageEnvironment("initContainers[0].volumeMounts[0].name", "plugins"))
			packageYaml.Set = envs
		} else {
			return packageYaml, errors.New("cloud storage is not setup properly")
		}
	}
	return packageYaml, nil
}

func injectSecretPatcherEnvs(packageRequest models.PackageRequest, packageYaml models.PackageYaml, request *models.ClusterRequest) (models.PackageYaml, error) {
	if packageRequest.Name == "secret-patcher" && request.Cluster.ImageRegistry != nil {
		if request.Cluster.ImageRegistry.Credentials.RawMessage != nil {
			var secretMap map[string]interface{}
			err := json.Unmarshal(request.Cluster.ImageRegistry.Credentials.RawMessage, &secretMap)
			if err == nil {
				envs := packageYaml.Set
				if envs == nil {
					envs = []map[string]interface{}{}
				}
				for k, v := range secretMap {
					if request.Cluster.ImageRegistry.Provider == constants.GCP && k == "gcp" {
						cred := v.(map[string]interface{})["google_app_cred"].(string)
						if cred == "" {
							return packageYaml, errors.New("invalid gcp credentials")
						}
						credentials := helper.ReadFileInB64String(cred)
						envs = append(envs, getPackageEnvironment(fmt.Sprintf("%s.%s", "gcp", "google_app_cred"), credentials))
						continue
					} else if request.Cluster.ImageRegistry.Provider == constants.EKS && k == "aws" {
						for k, vi := range v.(map[string]interface{}) {
							envs = append(envs, getPackageEnvironment(fmt.Sprintf("%s.%s", "aws", k), vi.(string)))
						}
						continue
					}
				}
				envs = append(envs, getPackageEnvironment("provider", request.Cluster.ImageRegistry.Provider))
				packageYaml.Set = envs
			}
		}
		if request.Cluster.ImageRegistry.Provider == "custom" {
			envs := packageYaml.Set
			envs = append(envs, getPackageEnvironment("credentials.docker_username", request.Cluster.ImageRegistry.UserName))
			envs = append(envs, getPackageEnvironment("credentials.docker_password", request.Cluster.ImageRegistry.Password))
			envs = append(envs, getPackageEnvironment("credentials.docker_registry_server", request.Cluster.ImageRegistry.Service))
			packageYaml.Set = envs
		}
	}
	return packageYaml, nil
}

func getPackageEnvironment(key string, value string) map[string]interface{} {
	domain := map[string]interface{}{
		"name":  key,
		"value": value,
	}
	return domain
}

func GetPackageNamespace(pkg string) string {
	packages := GetHelmPackages()
	if packages == nil {
		packages = constants.PackagesNamespaces
	}
	if v, ok := packages[pkg]; ok {
		return v
	}
	return pkg
}

func GetPackageClusterStatus(pkgStatus map[string]interface{}) map[string]interface{} {
	if pkgStatus == nil {
		return nil
	}
	file, _ := os.ReadFile("/data/public/helmfile.schema.json")
	packageData := make(map[string]interface{})
	err := json.Unmarshal(file, &packageData)
	if err != nil {
		logrus.Error("error while getting data :: ", err)
		return nil
	}
	updatedPkgStatus := make(map[string]interface{})
	packages, exists := packageData["packages"].([]interface{})
	if exists {
		for _, pkg := range packages {
			pkgMap, isMap := pkg.(map[string]interface{})
			if !isMap {
				continue
			}
			namespace, namespaceExists := pkgMap["namespace"].(string)
			title, titleExists := pkgMap["title"].(string)
			if namespaceExists && titleExists {
				if val, exists := pkgStatus[namespace]; exists {
					updatedPkgStatus[title] = val
				}
			}
		}
	}
	return updatedPkgStatus
}

func GetHelmPackages() map[string]string {
	repoUrl := os.Getenv("HELM_CHART_URL")
	if repoUrl == "" {
		repoUrl = "https://berrybytes.github.io/helm-chart-org"
	}
	response, err := http.Get(repoUrl + "/packages/namespaces/namespace.json")
	if err != nil {
		logrus.Error("getting namespace error :: ", err)
		return nil
	}
	defer response.Body.Close()
	var packages map[string]string
	err = json.NewDecoder(response.Body).Decode(&packages)
	if err != nil {
		logrus.Error("Error while unmarshalling the body :: ", err)
		return nil
	}
	return packages
}

// UnInstallPackage godoc
// @Summary Uninstall package
// @Description Uninstall package in cluster
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Param body body  doc.PackageRequestBody true "package install body"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/uninstall-package [post]
func (server *Server) UnInstallPackage(w http.ResponseWriter, r *http.Request) {
	server.InstallUninstallPackage(w, r, constants.UninstallPackage)
}

// InstallPackage godoc
// @Summary Install package
// @Description Install package in cluster
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Param body body doc.PackageRequestBody true "package install body"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/install-package [post]
func (server *Server) InstallPackage(w http.ResponseWriter, r *http.Request) {
	server.InstallUninstallPackage(w, r, constants.InstallPackage)
}

// InstallPackage godoc
// @Summary Install package
// @Description Install package in cluster
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Param body body doc.PackageRequestBody true "package install body"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/install-package [post]
func (server *Server) UpdatePackage(w http.ResponseWriter, r *http.Request) {
	uid, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	vars := mux.Vars(r)
	cid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	file, _ := os.ReadFile("/data/public/helmfile.schema-update.json")
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
		clusterId:      cid,
	}
	err = server.PackageManager(data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": "Package update triggered",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// GetPackageConfig godoc
// @Summary Get package config
// @Description Get package install config
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} doc.SuccessResponse
// @Router /package-config [get]
func (server *Server) GetPackageConfig(w http.ResponseWriter, r *http.Request) {
	file, _ := os.ReadFile("/data/public/helmfile.schema.json")
	res := make(map[string]interface{})
	err := json.Unmarshal(file, &res)
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

func (server *Server) UpdatePackageConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if _, err = os.Stat("/data/helmfile"); err != nil {
		err = os.Mkdir("/data/helmfile", 0600)
		if err != nil {
			responses.ERROR(w, http.StatusInternalServerError, err)
			return
		}
	}
	err = os.WriteFile("/data/public/helmfile.schema.json", body, 0777)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, "Successfully Updated")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// CheckPackageStatus godoc
// @Summary Check package status
// @Description Initiate get package status process
// @Tags Cluster
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/package-status [get]
func (server *Server) CheckPackageStatus(w http.ResponseWriter, r *http.Request) {
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
	if dataReceived.Cluster == nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cluster is not created"))
		return
	}
	dataReceived.Cluster.ClusterRequestID = uint64(dataReceived.ID)
	err = queue.Publish(constants.FetchPackageStatus, dataReceived.Cluster)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	packageCollection := fmt.Sprintf("%d-cluster-package-status", dataReceived.ClusterID)
	packageStatus, err := store.NewStore().Loadbalancer().GetLoadbalancerStatus(packageCollection)
	if err != nil {
		time.Sleep(2 * time.Second)
		packageStatus, err = store.NewStore().Loadbalancer().GetLoadbalancerStatus(packageCollection)
		if err != nil {
			logrus.Error("error gettings package status from store :: ", err)
		}
	}
	resp := GetPackageClusterStatus(packageStatus)
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data:    resp,
	})
}

// ImportCluster godoc
// @Summary Import a cluster
// @Description Import new cluster with the input payload
// @Tags Cluster
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Security ApiKeyAuth
// @Param file formData file true  "This is Import cluster file"
// @Param body body doc.ClusterRequest true "Create cluster"
// @Success 201 {object} doc.ClusterRequest
// @Router /import-cluster [post]
func (server *Server) ImportCluster(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	configPath := ""
	file, _, err := r.FormFile("file")
	if err == nil {
		configPath, err = uploadConfigFile(file)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	}
	data := models.Cluster{}
	data.Name = r.FormValue("name")
	data.ConfigPath = configPath
	data.Region = r.FormValue("region")
	data.Provider = r.FormValue("provider")
	imageRegistryID, err := strconv.ParseUint(r.FormValue("image_registry_id"), 10, 64)
	if err == nil {
		data.ImageRegistryID = imageRegistryID
	}
	data.Zone = r.FormValue("zone")
	data.Active = false
	data.PvCapacity = 20
	data.OrganizationID = uint64(oid)
	data.Labels = r.FormValue("labels")

	err = data.ValidateImportCluster()
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
	cr, _ := clusterInterface.FindWithDns(server.DB, uint64(dataCreated.ID))
	clusterRequest := models.ClusterRequest{}
	clusterRequest.Name = data.Name
	clusterRequest.Region = data.Region
	clusterRequest.Provider = data.Provider
	clusterRequest.ProviderName = r.FormValue("provider_name")
	//clusterRequest.Cluster = dataCreated
	clusterRequest.OrganizationID = uint64(oid)
	clusterRequest.ClusterID = uint64(dataCreated.ID)
	clusterRequest.Active = true
	clusterRequest.Cluster = cr
	clusterRequest.Type = constants.TypeImported
	clusterRequest.Status = constants.TypeImported
	clRequest, err := clusterRequestInterface.Save(server.DB, &clusterRequest)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	data.ClusterRequestID = uint64(clRequest.ID)
	data.ID = dataCreated.ID
	_, _ = clusterInterface.Update(server.DB, data)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, clRequest)
		return
	}
	clusterResponse, err := CreateClusterRequestResponse(clRequest)
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

// ValidateKubeconfig godoc
// @Summary Validate kubeconfig
// @Description Validate kubeconfig file
// @Tags Cluster
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "ClusterRequest id"
// @Param body body doc.ClusterRequest true "Validate kubeconfig"
// @Success 200 {object} doc.SuccessResponse
// @Router /create-cluster/{id}/validate [post]
func (server *Server) ValidateKubeconfig(w http.ResponseWriter, r *http.Request) {
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
	if dataReceived.Cluster == nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("cluster is not created"))
		return
	}
	dataReceived.Cluster.ClusterRequestID = uint64(dataReceived.ID)

	req := models.CreateClusterRequest{
		ID:         int64(pid),
		Name:       dataReceived.Name,
		Namespace:  helper.GetCreateClusterNamespace(dataReceived),
		ConfigPath: dataReceived.Cluster.ConfigPath,
	}
	err = queue.Publish(constants.ValidateConfig, req)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{"message": "Validating cluster config"})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
