package models

// ScannerScanRequest
//
// actual request by the user for starting the scan
type ScannerScanRequest struct {
	Name string `json:"name"`
}

// PluginsResponse
//
// this is the struct which will be returned from the security scanner microservice
type PluginsResponse struct {
	Image   []PluginBase `json:"image"`
	K8S     []PluginBase `json:"k8s"`
	Repo    []PluginBase `json:"repo"`
	Cluster []PluginBase `json:"cluster"`
}

type PluginBase struct {
	Description string   `json:"description"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	Logo        string   `json:"logo"`
	Variables   []string `json:"variables"`
}

// LoginResponse
//
// this is the login response from the security scanner API
type LoginResponse struct {
	Data struct {
		Token string `json:"token"`
		User  string `json:"user"`
	} `json:"data"`
	Message string `json:"message"`
}
