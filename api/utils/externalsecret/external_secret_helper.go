package externalsecret

import (
	"01cloud-api/api/models"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/gosimple/slug"
)

type ExternalSecretRequest struct {
	EnvironmentId     uint                   `json:"environment_id"`
	SecretName        string                 `json:"secret_name"`
	Keys              string                 `json:"keys"`
	Secrets           map[string]string      `json:"secrets"`
	Namespace         string                 `json:"namespace"`
	Enabled           bool                   `json:"enabled"`
	PluginUrl         string                 `json:"plugin_url"`
	KubeconfigPath    string                 `json:"kubeconfig_path"`
	Provider          string                 `json:"provider"`
	AwsCredential     AwsCredential          `json:"aws_credential"`
	GcpCredential     GcpCredential          `json:"gcp_credential"`
	VaultCredential   VaultCredential        `json:"vault_credential"`
	EnvironmentType   int                    `json:"environment_type"`
	AdditionalSetting map[string]interface{} `json:"additional_setting"`
	Annotations       map[string]string      `json:"annotations"`
	Labels            map[string]string      `json:"labels"`
}

type AwsCredential struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

type GcpCredential struct {
	ProjectId      string `json:"project_id"`
	ServiceAccount string `json:"service_account"`
}

type VaultCredential struct {
	Server  string `json:"server"`
	Path    string `json:"path"`
	Version string `json:"version"`
	Token   string `json:"token"`
}

type RetryExternalSecretRequest struct {
	Secrets           map[string]string      `json:"secrets"`
	Enabled           bool                   `json:"enabled"`
	Provider          string                 `json:"provider"`
	AwsCredential     AwsCredential          `json:"aws_credential"`
	GcpCredential     GcpCredential          `json:"gcp_credential"`
	VaultCredential   VaultCredential        `json:"vault_credential"`
	AdditionalSetting map[string]interface{} `json:"additional_setting"`
}

type ExternalSecretResponse struct {
	EnvironmentId uint   `json:"environment_id"`
	SecretName    string `json:"secret_name"`
	Namespace     string `json:"namespace"`
	StatusMessage string `json:"message"`
	Success       bool   `json:"success"`
	ErrorCode     int    `json:"error_code"`
}

func IsExternalSecretEnable(input *models.Environment) bool {
	return input.ExternalSecret.RawMessage != nil
}

func PrepareExternalSecretRequest(input *models.Environment) (*ExternalSecretRequest, error) {
	if input.Application.Cluster == nil || input.Application.Cluster.DNS == nil {
		return nil, errors.New("cluster and dns object nil :: ")
	}
	esr := &ExternalSecretRequest{}
	err := json.Unmarshal(input.ExternalSecret.RawMessage, esr)
	if err != nil {
		return nil, err
	}
	esr.EnvironmentId = input.ID
	esr.EnvironmentType = input.ServiceType
	esr.Annotations, esr.Labels = getAnnotationsAndLabels(input)
	esr.Namespace = helper.GetNamespace(input)
	esr.PluginUrl = input.PluginVersion.Url
	esr.KubeconfigPath = input.Application.Cluster.ConfigPath
	return esr, nil
}

func (e *ExternalSecretRequest) ValidateExternalSecretRequest() error {
	if e == nil {
		return nil
	}
	if e.EnvironmentId == 0 {
		return errors.New("required environment id")
	}
	if len(e.Secrets) == 0 {
		return errors.New("required secrets")
	}
	if e.Namespace == "" {
		return errors.New("required namespace")
	}
	if e.Provider == constants.EKS {
		if e.AwsCredential.AccessKey == "" {
			return errors.New("required access key")
		}
		if e.AwsCredential.SecretKey == "" {
			return errors.New("required secret key")
		}
		if e.AwsCredential.Region == "" {
			return errors.New("required region")
		}
	}
	if e.Provider == constants.GCP {
		if e.GcpCredential.ServiceAccount == "" {
			return errors.New("required  service credential file")
		}
		if e.GcpCredential.ProjectId == "" {
			return errors.New("required project id")
		}
	}
	if e.Provider == constants.Vault {
		if e.VaultCredential.Path == "" {
			return errors.New("required path")
		}
		if e.VaultCredential.Server == "" {
			return errors.New("required server url")
		}
		if e.VaultCredential.Token == "" {
			return errors.New("required token")
		}
	}
	return nil
}

func (retryRequest *RetryExternalSecretRequest) PrepareExternalSecretUpdate(input *models.Environment) ([]byte, error) {
	if input.ExternalSecret.RawMessage == nil {
		return nil, errors.New("external secret is not enabled")
	}
	if retryRequest == nil {
		return nil, errors.New("required update field become nil")
	}
	esr := &ExternalSecretRequest{}
	err := json.Unmarshal(input.ExternalSecret.RawMessage, esr)
	if err != nil {
		return nil, err
	}
	if len(retryRequest.Secrets) != 0 {
		esr.Secrets = retryRequest.Secrets
	}
	if retryRequest.Provider != "" {
		esr.Provider = retryRequest.Provider
	}
	if retryRequest.AwsCredential.AccessKey != "" {
		esr.AwsCredential.AccessKey = retryRequest.AwsCredential.AccessKey
	}
	if retryRequest.AwsCredential.SecretKey != "" {
		esr.AwsCredential.SecretKey = retryRequest.AwsCredential.SecretKey
	}
	if retryRequest.AwsCredential.Region != "" {
		esr.AwsCredential.Region = retryRequest.AwsCredential.Region
	}
	if retryRequest.GcpCredential.ProjectId != "" {
		esr.GcpCredential.ProjectId = retryRequest.GcpCredential.ProjectId
	}
	if retryRequest.GcpCredential.ServiceAccount != "" {
		esr.GcpCredential.ServiceAccount = retryRequest.GcpCredential.ServiceAccount
	}
	if retryRequest.VaultCredential.Server != "" {
		esr.VaultCredential.Server = retryRequest.VaultCredential.Server
	}
	if retryRequest.VaultCredential.Path != "" {
		esr.VaultCredential.Path = retryRequest.VaultCredential.Path
	}
	if retryRequest.VaultCredential.Token != "" {
		esr.VaultCredential.Token = retryRequest.VaultCredential.Token
	}
	if len(retryRequest.AdditionalSetting) != 0 {
		esr.AdditionalSetting = retryRequest.AdditionalSetting
	}
	esrBytes, err := json.Marshal(esr)
	if err != nil {
		return nil, err
	}
	return esrBytes, nil
}

func getAnnotationsAndLabels(input *models.Environment) (map[string]string, map[string]string) {
	environmentType := "template"
	if input.ServiceType == 1 {
		environmentType = "non-template"
	} else if input.ServiceType == 2 {
		environmentType = "image"
	} else if input.ServiceType == 3 {
		environmentType = "helm-chart"
	} else if input.ServiceType == 4 {
		environmentType = "operator"
	}
	baseDomain := input.Application.Cluster.DNS.BaseDomain
	ingressClass := "zerone"
	dnsZone := fmt.Sprintf("%s.%s", helper.GetClusterNamespace(input.Application.Cluster), baseDomain)
	target := dnsZone
	if input.LoadBalancer != nil {
		ingressClass = helper.GetContourNamespace(input.LoadBalancer)
		target = ingressClass + "." + baseDomain
	}
	zone := input.Application.Cluster.DNS.ZoneID
	if input.Application.Cluster.DNS.Provider == "cloudflare" {
		zone = input.Application.Cluster.DNS.BaseDomain
	}
	return map[string]string{
			"app.01cloud.io/zone":         zone,
			"app.01cloud.io/dns-zone":     dnsZone,
			"app.01cloud.io/project_name": input.Application.Cluster.DNS.ProjectId,
			"app.01cloud.io/base_domain":  baseDomain,
			"app.01cloud.io/target":       target,
			"ingress_class":               ingressClass,
		}, map[string]string{
			"app.01cloud.io/environment":  slug.Make(input.Name),
			"app.01cloud.io/created_from": slug.Make("external-secret-service"),
			"app.01cloud.io/env_id":       slug.Make(fmt.Sprintf("%d", input.ID)),
			"app.01cloud.io/helm-chart":   slug.Make(input.Application.Plugin.Name),
			"app.01cloud.io/helm-version": input.PluginVersion.Version,
			"app.01cloud.io/app_name":     slug.Make(input.Application.Name),
			"app.01cloud.io/app_id":       slug.Make(fmt.Sprintf("%d", input.Application.ID)),
			"app.01cloud.io/project_name": slug.Make(input.Application.Project.Name),
			"app.01cloud.io/project_id":   slug.Make(fmt.Sprintf("%d", input.Application.ProjectID)),
			"app.01cloud.io/provider":     slug.Make(input.Application.Cluster.DNS.Provider),
			"app.01cloud.io/managed-by":   "helm",
			"env":                         os.Getenv("GCLOUD_NAMESPACE"),
			"environment_type":            environmentType,
			"ingress_class":               ingressClass,
			"app.01cloud.io/client_name":  slug.Make(input.Application.Project.User.FirstName + "_" + input.Application.Project.User.LastName),
			"app.01cloud.io/status":       "active",
		}
}
