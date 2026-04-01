package models

type PackageRequest struct {
	Name        string                   `json:"name"`
	Namespace   string                   `json:"namespace"`
	Chart       string                   `json:"chart"`
	RequiredDNS bool                     `json:"required_dns"`
	Set         []map[string]interface{} `json:"set,omitempty"`
	Needs       []string                 `json:"needs,omitempty"`
}

type PackageYaml struct {
	Name        string                   `json:"name"`
	Namespace   string                   `json:"namespace,omitempty"`
	Chart       string                   `json:"chart"`
	RequiredDNS bool                     `json:"-"`
	Set         []map[string]interface{} `json:"set,omitempty"`
	Needs       []string                 `json:"needs,omitempty"`
}
