package controllers

/* security_scanner_controller
This file is responsible for integrating 01-security project to 01-cloud.

APIs:
	route -> Method Responsible
	/scanner/plugins -> PluginList
	/scanner/{id}/scan -> ScanRequest
	/scanner/{id}/reports -> GetAllScanReports
	/scanner/{id}/reports/{reportId} -> GetScanReport

	PluginList(GET): this function will simply proxy the user's getPlugins
		request to 01-security microservice, and the response
		retrieved will be sent back to the user as it is.

	ScanRequest(POST): PARAM{id} which is environment ID and post body contains {name:""}
		on this function user will be sending the scanner
		plugin name which was retrieved from above PluginList api, and now
		this function will:
		1. check if the user has adequate permission on the environment i.e. READ
		2. retrieve the list of plugins {for checking the variables that needs to be supplied as JSON on request}
		3. retrieve the user plugin name
		4. check if the plugin exist and if true get the variables name that needs to be sent for the plugin to 01-security
		5. create a JSON for the variables, as  {"variable_name":"value"}, the value here will be dynamically determined
			as like if the key is namespace then the environment namespace will be the value, and if kubeconfig, then the
			config will be retrieved and sent as value.
		6. again prepare for a JSON as {"name": "scan_name", "plugin_name": "plugin", "variables": "upper_JSON"}
		7. request to the 01-security microservice and return the response from 01-security microservice

	GetAllScanReports(GET): PARAM{id} which is environment ID
		this function will check if the user has adequate permission for this environment i.e. READ
		and if true retrieve the namespace from the environment id and then sends request to
		01-security microservice for retrieving all the reports which are available for this namespace

	GetScanReport(GET): PARAM{id,reportId} which is environment ID and report ID
		this function will check if the user has adequate permission for this environment i.e. READ
		and if true sends the request to 01-security microservice for retrieving the report of this ID
*/

import (
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/helper"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var token string

// PluginList ScannerPlugins godoc
// @Summary Get the scanner Plugins
// @Description Use this url for fetching the scanner plugins list
// @Tags SecurityScanner
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /scanner/{id}/plugins [get]
func (server *Server) PluginList(w http.ResponseWriter, r *http.Request) {
	env, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", os.Getenv("SECURITY_SCANNER_API")+"plugins", nil)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return
	}
	_, _ = server.requestor(w, req, false, env.ServiceType, 2)
}

// ClusterPluginList ScannerPlugins godoc
// @Summary Get the cluster scanner Plugins
// @Description Use this url for fetching the cluster scanner plugins list
// @Tags SecurityScanner
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /scanner/cluster-plugins [get]
func (server *Server) ClusterPluginList(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequestWithContext(context.Background(), "GET", os.Getenv("SECURITY_SCANNER_API")+"plugins", nil)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return
	}
	_, _ = server.requestor(w, req, false, -1, 2)
}

// ScanRequest godoc
// @Summary Trigger security scanner Scan request
// @Description Use this url for starting the scan on security scanner
// @Tags SecurityScanner
// @Accept  json
// @Produce  json
// @Param body body doc.SecurityScanRequest true "Scan Request"
// @Param id path string true "envId"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /scanner/{id}/scan [post]
func (server *Server) ScanRequest(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil || code != 0 || environment == nil || user == nil {
		responses.ERROR(w, code, err)
		return
	}
	plugins, err := server.getListOfPlugins(w, r)
	if err != nil {
		return
	}
	name, err := getTheRequestedImageName(w, r)
	if err != nil {
		return
	}
	variables := checkWhichImageAndReturnParams(plugins, name)
	bodyVars := server.parseBodyVars(environment, variables)
	reqBody := map[string]interface{}{}
	reqBody["name"] = fmt.Sprintf("%s-%s", name, time.Now().Format("2006-01-02 15:04:05"))
	reqBody["plugin_name"] = name
	reqBody["variables"] = bodyVars
	body, _ := json.Marshal(reqBody)
	b := bytes.NewBufferString(string(body))

	ns := helper.GetNamespace(environment)
	req, err := http.NewRequestWithContext(context.Background(), "POST", os.Getenv("SECURITY_SCANNER_API")+"scan/"+ns, b)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return
	}
	_, _ = server.requestor(w, req, false, environment.ServiceType, 2)
}

// ClusterScanRequest godoc
// @Summary Trigger security scanner cluster Scan request
// @Description Use this url for starting the cluster scan on security scanner
// @Tags SecurityScanner
// @Accept  json
// @Produce  json
// @Param body body doc.SecurityScanRequest true "Scan Request"
// @Param id path string true "clusterId"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /scanner/{id}/cluster-scan [post]
func (server *Server) ClusterScanRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["id"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	cluster, err := clusterInterface.Find(server.DB, uint64(clusterId))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", os.Getenv("SECURITY_SCANNER_API")+"plugins", nil)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return
	}
	_, plugins := server.requestor(w, req, true, -1, 2)
	name, err := getTheRequestedImageName(w, r)
	if err != nil {
		return
	}
	variables := checkWhichImageAndReturnParams(plugins, name)
	bodyVars := map[string]string{}
	for _, vars := range variables {
		switch vars {
		case "k8s_kubeconfig":
			kubefile, err := os.ReadFile(cluster.ConfigPath)
			if err != nil {
				log.Errorf("cannot read the kubeconfig file. ERR: %v", err)
			}
			base64Encoded := base64.StdEncoding.EncodeToString(kubefile)
			bodyVars["k8s_kubeconfig"] = string(base64Encoded)
		}
	}
	reqBody := map[string]interface{}{}
	reqBody["name"] = fmt.Sprintf("%s-%s", name, time.Now().Format("2006-01-02 15:04:05"))
	reqBody["plugin_name"] = name
	reqBody["variables"] = bodyVars
	body, _ := json.Marshal(reqBody)
	b := bytes.NewBufferString(string(body))
	clusterNs := helper.GetClusterNamespace(cluster)
	req, err = http.NewRequestWithContext(context.Background(), "POST", os.Getenv("SECURITY_SCANNER_API")+"scan/"+clusterNs, b)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return
	}
	_, _ = server.requestor(w, req, false, -1, 2)
}

// parseBodyVars
//
// this function is responsible for generating the variables key field
// value for sending the POST request to the security scanner microservice
func (server *Server) parseBodyVars(environment *models.Environment, variables []string) map[string]string {
	body := map[string]string{}
	for _, vars := range variables {
		switch vars {
		case "k8s_namespace":
			body["k8s_namespace"] = helper.GetNamespace((environment))
		case "k8s_kubeconfig":
			kubefile, err := os.ReadFile(environment.Application.Cluster.ConfigPath)
			if err != nil {
				log.Errorf("cannot read the kubeconfig file. ERR: %v", err)
			}
			base64Encoded := base64.StdEncoding.EncodeToString(kubefile)
			body["k8s_kubeconfig"] = string(base64Encoded)
		case "git_repo":
			body["git_repo"] = strings.Split(environment.GitUrl, "https://")[1]
		case "registry_image":
			body["registry_image"] = environment.ImageUrl + ":" + environment.ImageTag
		case "registry_username":
			gituser, _ := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uint(environment.Application.OwnerId)), environment.Application.ImageService)
			body["registry_username"] = gituser.ServiceUserName
		case "registry_password":
			gituser, _ := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uint(environment.Application.OwnerId)), environment.Application.ImageService)
			body["registry_password"] = gituser.AccessToken
		case "registry_host":
			body["registry_host"] = environment.Application.Cluster.ImageRegistry.Service
		case "git_token":
			gituser, _ := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uint(environment.Application.OwnerId)), environment.Application.GitService)
			body["git_token"] = gituser.AccessToken
		case "git_branch":
			body["git_branch"] = environment.GitBranch
		case "sonar_projectkey":
			body["sonar_projectkey"] = fmt.Sprintf("%s-%d", environment.Name, environment.ID)
		case "sonar_projectname":
			body["sonar_projectname"] = fmt.Sprintf("%s-%d-%s", environment.Name, environment.ID, environment.Application.Name)
		case "sonar_host":
			body["sonar_host"] = os.Getenv("SECURITY_SCANNER_SONAR_HOST")
		case "sonar_username":
			body["sonar_username"] = os.Getenv("SECURITY_SCANNER_SONAR_USERNAME")
		case "sonar_password":
			body["sonar_password"] = os.Getenv("SECURITY_SCANNER_SONAR_PASSWORD")
		}
	}
	return body
}

// checkWhichImageAndReturnParams
//
// problem which we solve here -> user will send us the name of the plugin to scan with,
// but we don't specifically know where the plugin exists like if in image, k8s or repo
//
// so this function will check for every plugin names on separate loops for image, k8s and repo
// if any of the loop finds the name of the plugin to use than,
// variables value will be set to that and will be sent as the response
func checkWhichImageAndReturnParams(plugins *models.PluginsResponse, name string) []string {
	for _, plugin := range plugins.Image {
		if plugin.Name == name {
			return plugin.Variables
		}
	}
	for _, plugin := range plugins.Repo {
		if plugin.Name == name {
			return plugin.Variables
		}
	}
	for _, plugin := range plugins.K8S {
		if plugin.Name == name {
			return plugin.Variables
		}
	}
	for _, plugin := range plugins.Cluster {
		if plugin.Name == name {
			return plugin.Variables
		}
	}
	return []string{}
}

// getListOfPlugins
//
// this function retrieves the list of plugins from the security scanner microservice
// This function is only used by the scan method, as Plugin List method handles this in its own proxy
func (server *Server) getListOfPlugins(w http.ResponseWriter, r *http.Request) (*models.PluginsResponse, error) {
	env, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return nil, err
	}
	req, err := http.NewRequestWithContext(context.Background(), "GET", os.Getenv("SECURITY_SCANNER_API")+"plugins", nil)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return nil, err
	}
	_, plugins := server.requestor(w, req, true, env.ServiceType, 2)
	return plugins, err
}

// getTheRequestedImageName
//
// this function will retrieve the name of the image that the user wants to scan with
func getTheRequestedImageName(w http.ResponseWriter, r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return "", err
	}
	input := models.ScannerScanRequest{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return "", err
	}
	return input.Name, nil
}

// GetScanReport godoc
// @Summary Fetch a specific report generated by security scanner
// @Description Use this url for fetching detailed report of the scan
// @Tags SecurityScanner
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path string true "envId"
// @Param reportId path string true "reportId"
// @Success 200 {object} map[string]interface{}
// @Router /scanner/{id}/reports/{reportId} [get]
func (server *Server) GetScanReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// the request can be for the cluster report so, ignoring for now
	// _, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	// if err != nil {
	// responses.ERROR(w, code, err)
	// return
	// }
	req, err := http.NewRequestWithContext(context.Background(), "GET", os.Getenv("SECURITY_SCANNER_API")+"report/"+vars["reportId"], nil)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return
	}
	// as we donot need the service type so ignoring the field totally
	_, _ = server.requestor(w, req, false, -1, 2)
}

func (server *Server) DeleteScanReport(env *models.Environment) {
	ns := helper.GetNamespace(env)
	req, err := http.NewRequestWithContext(context.Background(), "DELETE", os.Getenv("SECURITY_SCANNER_API")+"reports/"+ns, nil)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		return
	}
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		log.Error("error sending request ::", err)
		return
	}
}

// GetAllScanReports godoc
// @Summary Fetch all reports generated by security scanner as per namespace
// @Description Use this url for fetching all reports of the scan
// @Tags SecurityScanner
// @Accept  json
// @Param id path string true "envId"
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /scanner/{id}/reports [get]
func (server *Server) GetAllScanReports(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	ns := helper.GetNamespace(environment)
	req, err := http.NewRequestWithContext(context.Background(), "GET", os.Getenv("SECURITY_SCANNER_API")+"reports/"+ns, nil)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return
	}
	_, _ = server.requestor(w, req, false, environment.ServiceType, 2)
}

// GetClusterAllScanReports godoc
// @Summary Fetch all reports generated by security scanner as per cluster namespace
// @Description Use this url for fetching all reports of the cluster scan
// @Tags SecurityScanner
// @Accept  json
// @Param id path string true "envId"
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{}
// @Router /scanner/{id}/cluster-reports [get]
func (server *Server) GetClusterAllScanReports(w http.ResponseWriter, r *http.Request) {
	clusterId, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	cluster, err := clusterInterface.Find(server.DB, uint64(clusterId))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	ns := helper.GetClusterNamespace(cluster)
	req, err := http.NewRequestWithContext(context.Background(), "GET", os.Getenv("SECURITY_SCANNER_API")+"reports/"+ns, nil)
	if err != nil {
		log.Errorf("new request cannot be initiated. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return
	}
	_, _ = server.requestor(w, req, false, -1, 2)
}

// requestor
//
// this function is responsible for requesting the security scanner
// and is used by all the methods available here for proxying
func (server *Server) requestor(w http.ResponseWriter, req *http.Request, scanRequest bool, serviceType, retryCount int) (int, *models.PluginsResponse) {
	if retryCount == 0 {
		responses.ERROR(w, 500, fmt.Errorf("retry count exceeded"))
		return 500, nil
	}
	if token == "" {
		err := server.securityScannerLoginMiddleware()
		if err != nil {
			log.Errorf("cannot login to security scanner. ERR: %v", err)
			responses.ERROR(w, 400, err)
			return 400, nil
		}
	}
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+token)
	req.Header = headers
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("received err while sending the request. ERR %v", err)
		responses.ERROR(w, 500, err)
		return resp.StatusCode, nil
	}
	if resp.StatusCode == 401 {
		err := server.securityScannerLoginMiddleware()
		if err != nil {
			log.Errorf("cannot login to security scanner. ERR: %v", err)
			responses.ERROR(w, 400, err)
			return 400, nil
		}
		server.requestor(w, req, scanRequest, serviceType, retryCount-1)
		// return resp.StatusCode, nil
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("Error while closing body. ERR: %v", err)
		}
	}(resp.Body)

	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("Error while reading the response body. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return resp.StatusCode, nil
	}
	if scanRequest {
		var plugins models.PluginsResponse
		err := json.Unmarshal(resBody, &plugins)
		if err != nil {
			return resp.StatusCode, nil
		}
		return resp.StatusCode, &plugins
	}
	if req.URL.String() == os.Getenv("SECURITY_SCANNER_API")+"plugins" {
		var plugins models.PluginsResponse
		err = json.Unmarshal(resBody, &plugins)
		if err != nil {
			responses.ERROR(w, 500, err)
			return resp.StatusCode, nil
		}
		plugins = filterPluginsByEnvID(plugins, serviceType)
		pluginResp, _ := json.Marshal(plugins)
		_, err = w.Write(pluginResp)
		if err != nil {
			log.Errorf("error while writing on the request. ERR: %v", err)
			responses.ERROR(w, 500, err)
			return resp.StatusCode, nil
		}
		return resp.StatusCode, &plugins
	}
	log.Infof("response body: %s", resBody)
	_, err = w.Write(resBody)
	if err != nil {
		log.Errorf("error while writing on the request. ERR: %v", err)
		responses.ERROR(w, 500, err)
		return resp.StatusCode, nil
	}
	return 200, nil
}

// filterPluginsByEnvID
//
// this function is responsible for filtering the plugins
// as per the environment ID, as the security scanner microservice
func filterPluginsByEnvID(plugins models.PluginsResponse, serviceId int) models.PluginsResponse {
	var filteredPlugins models.PluginsResponse
	filteredPlugins.K8S = plugins.K8S
	switch serviceId {
	case -1:
		filteredPlugins.Cluster = plugins.Cluster
		filteredPlugins.K8S = nil
	case 1:
		filteredPlugins.Repo = plugins.Repo
	case 2:
		filteredPlugins.Image = plugins.Image
	}
	return filteredPlugins
}

// securityScannerLoginMiddleware
//
// this acts as a middleware which gets called by every other
// function,if the scanner server returns 401 i.e. UNAUTHORIZED
//
// this function then hits the login API and then updates the
// token global variable and returns,
// and then the caller function re-initiates the request
func (server *Server) securityScannerLoginMiddleware() error {
	buf := bytes.NewBufferString(`{
		"email": "` + os.Getenv("SECURITY_SCANNER_USERNAME") + `",
		"password": "` + os.Getenv("SECURITY_SCANNER_PASSWORD") + `"
	}`)
	req, err := http.NewRequestWithContext(context.Background(), "POST", os.Getenv("SECURITY_SCANNER_API")+"users/login", buf)
	if err != nil {
		log.Errorf("cannot create a new request for login to security scanner. ERR: %v", err)
		return err
	}
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("cannot login to security scanner. ERR: %v", err)
		return err
	}
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("error while reading the io response from security scanner on login. ERR: %v", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error while closing the io. ERR: %v", err)
		}
	}(resp.Body)
	var loginResponse models.LoginResponse
	err = json.Unmarshal(resBody, &loginResponse)
	if err != nil {
		log.Errorf("cannot unmarshal the login response. ERR: %v", err)
		return err
	}
	token = loginResponse.Data.Token
	return nil
}
