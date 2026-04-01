package logging

import (
	"01cloud-api/api/models"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
	"01cloud-api/api/utils/imrole"

	"encoding/json"
	"errors"

	"github.com/jinzhu/gorm/dialects/postgres"
)

type LoggingRequest struct {
	EnvironmentId     uint                   `json:"environment_id"`
	Namespace         string                 `json:"namespace"`
	Enabled           bool                   `json:"enabled"`
	KubeconfigPath    string                 `json:"kubeconfig_path"`
	Provider          string                 `json:"provider"`
	GCPCredential     GCPCredential          `json:"gcp_credential"`
	AwsCredential     AwsCredential          `json:"aws_credential"`
	ElasticCredential ElasticCredential      `json:"elastic_credential"`
	LokiCredential    LokiCredential         `json:"loki_credential"`
	KafkaCredential   KafkaCredential        `json:"kafka_credential"`
	AdditionalSetting map[string]interface{} `json:"additional_setting"`
}

type AwsCredential struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
	Bucket    string `json:"bucket"`
	Path      string `json:"path"`
	Format    string `json:"format"`
	LogGroup  string `json:"log_group"`
	LogStream string `json:"log_stream"`
}

type GCPCredential struct {
	Project string `json:"project"`
	Bucket  string `json:"bucket"`
	KeyFile string `json:"key_file"`
	Path    string `json:"path"`
	StoreAs string `json:"store_as"`
}

type ElasticCredential struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	Scheme    string `json:"scheme"`
	Path      string `json:"path"`
	SSLVerify bool   `json:"ssl_verify"`
}

type LokiCredential struct {
	Host             string `json:"host"`
	Tenant           string `json:"tenant"`
	Labels           Label  `json:"labels"`
	ExtraLabels      Label  `json:"extra_labels"`
	Username         string `json:"username"`
	Password         string `json:"password"`
	LineFormat       string `json:"line_format"`
	KubernetesLabels bool   `json:"kubernetes_labels"`
}

type Label struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type KafkaCredential struct {
	Brokers      string `json:"host"`
	DefaultTopic string `json:"default_topic"`
	FormatType   string `json:"format_type"`
	SASLoverSSl  bool   `json:"sasl_over_ssl"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func (updateLoggingRequest *LoggingRequest) PrepareLoggingUpdate(input *models.Environment) ([]byte, error) {
	if input.ExternalLogging.RawMessage == nil {
		return nil, errors.New("logging is not enabled")
	}
	if updateLoggingRequest == nil {
		return nil, errors.New("required logging update field")
	}
	lg := &LoggingRequest{}
	err := json.Unmarshal(input.ExternalLogging.RawMessage, lg)
	if err != nil {
		return nil, err
	}
	if updateLoggingRequest.Provider != "" {
		lg.Provider = updateLoggingRequest.Provider
	}
	lg.Enabled = updateLoggingRequest.Enabled
	//This is for S3 & cloudwatch
	if updateLoggingRequest.AwsCredential.AccessKey != "" {
		lg.AwsCredential.AccessKey = updateLoggingRequest.AwsCredential.AccessKey
	}
	if updateLoggingRequest.AwsCredential.SecretKey != "" {
		lg.AwsCredential.SecretKey = updateLoggingRequest.AwsCredential.SecretKey
	}
	if updateLoggingRequest.AwsCredential.Bucket != "" {
		lg.AwsCredential.Bucket = updateLoggingRequest.AwsCredential.Bucket
	}
	if updateLoggingRequest.AwsCredential.Region != "" {
		lg.AwsCredential.Region = updateLoggingRequest.AwsCredential.Region
	}
	// This is for Elastic
	if updateLoggingRequest.ElasticCredential.Host != "" {
		lg.ElasticCredential.Host = updateLoggingRequest.ElasticCredential.Host
	}
	if updateLoggingRequest.ElasticCredential.Port != "" {
		lg.ElasticCredential.Port = updateLoggingRequest.ElasticCredential.Port
	}
	if updateLoggingRequest.ElasticCredential.User != "" {
		lg.ElasticCredential.User = updateLoggingRequest.ElasticCredential.User
	}
	if updateLoggingRequest.ElasticCredential.Password != "" {
		lg.ElasticCredential.Password = updateLoggingRequest.ElasticCredential.Password
	}
	if updateLoggingRequest.ElasticCredential.Scheme != "" {
		lg.ElasticCredential.Scheme = updateLoggingRequest.ElasticCredential.Scheme
	}
	//This is for loki
	if updateLoggingRequest.LokiCredential.Host != "" {
		lg.LokiCredential.Host = updateLoggingRequest.LokiCredential.Host
	}
	//This is for kafka
	if updateLoggingRequest.KafkaCredential.Brokers != "" {
		lg.KafkaCredential.Brokers = updateLoggingRequest.KafkaCredential.Brokers
	}
	if updateLoggingRequest.KafkaCredential.DefaultTopic != "" {
		lg.KafkaCredential.DefaultTopic = updateLoggingRequest.KafkaCredential.DefaultTopic
	}
	if updateLoggingRequest.KafkaCredential.FormatType != "" {
		lg.KafkaCredential.FormatType = updateLoggingRequest.KafkaCredential.FormatType
	}
	//This is for GCP bucket
	if updateLoggingRequest.GCPCredential.Project != "" {
		lg.GCPCredential.Project = updateLoggingRequest.GCPCredential.Project
	}
	if updateLoggingRequest.GCPCredential.Bucket != "" {
		lg.GCPCredential.Bucket = updateLoggingRequest.GCPCredential.Bucket
	}
	if updateLoggingRequest.GCPCredential.KeyFile != "" {
		lg.GCPCredential.KeyFile = updateLoggingRequest.GCPCredential.KeyFile
	}

	//This is for additional setting
	if updateLoggingRequest.AdditionalSetting != nil {
		lg.AdditionalSetting = updateLoggingRequest.AdditionalSetting
	}
	lgBytes, err := json.Marshal(lg)
	if err != nil {
		return nil, err
	}
	return lgBytes, nil

}

func PrepareLoggingRequest(input *models.Environment) (*LoggingRequest, error) {
	if input.Application.Cluster == nil || input.Application.Cluster.DNS == nil {
		return nil, errors.New("cluster and dns object nil :: ")
	}
	lr := &LoggingRequest{}
	err := json.Unmarshal(input.ExternalLogging.RawMessage, lr)
	if err != nil {
		return nil, err
	}
	lr.EnvironmentId = input.ID
	lr.Namespace = helper.GetNamespace(input)
	lr.KubeconfigPath = input.Application.Cluster.ConfigPath
	return lr, nil
}

func ValidateLoggingRequest(data postgres.Jsonb) error {
	if data.RawMessage == nil || len(data.RawMessage) == 0 {
		return nil
	}
	e := &LoggingRequest{}
	err := json.Unmarshal(data.RawMessage, e)
	if err != nil {
		return err
	}
	if e.Provider == constants.S3 {
		if e.AwsCredential.AccessKey == "" {
			return errors.New("required access key")
		}
		if e.AwsCredential.SecretKey == "" {
			return errors.New("required secret key")
		}
		if e.AwsCredential.Region == "" {
			return errors.New("required region")
		}
		if e.AwsCredential.Bucket == "" {
			return errors.New("required bucket")
		}
		_, _, err := CheckAwsLoggingPermissions(e.AwsCredential)
		if err != nil {
			return err
		}
	}
	if e.Provider == constants.Cloudwatch {
		if e.AwsCredential.AccessKey == "" {
			return errors.New("required access key")
		}
		if e.AwsCredential.SecretKey == "" {
			return errors.New("required secret key")
		}
		if e.AwsCredential.Region == "" {
			return errors.New("required region")
		}
		_, _, err := CheckCloudwatchLoggingPermissions(e.AwsCredential)
		if err != nil {
			return err
		}
	}
	if e.Provider == constants.GCP {
		if e.GCPCredential.Bucket == "" {
			return errors.New("required access bucket")
		}
		if e.GCPCredential.KeyFile == "" {
			return errors.New("required secret keyfile")
		}
		if e.GCPCredential.Bucket == "" {
			return errors.New("required bucket")
		}
		_, _, err := CheckGcpLoggingPermissions(e.GCPCredential)
		if err != nil {
			return err
		}
	}
	if e.Provider == constants.Elastic {
		if e.ElasticCredential.Host == "" {
			return errors.New("required host url")
		}
		if e.ElasticCredential.Port == "" {
			return errors.New("required port")
		}
		if e.ElasticCredential.User == "" {
			return errors.New("required user")
		}
		if e.ElasticCredential.Password == "" {
			return errors.New("required password")
		}
		if e.ElasticCredential.Scheme == "" {
			return errors.New("required scheme")
		}

	}
	if e.Provider == constants.Loki {
		if e.LokiCredential.Host == "" {
			return errors.New("required host url")
		}
	}
	if e.Provider == constants.Kafka {
		if e.KafkaCredential.Brokers == "" {
			return errors.New("required broker")
		}
		if e.KafkaCredential.DefaultTopic == "" {
			return errors.New("required default topic")
		}
		if e.KafkaCredential.FormatType == "" {
			return errors.New("required format type")
		}

	}
	return nil
}

func CheckAwsLoggingPermissions(request AwsCredential) ([]string, []string, error) {
	requiredAction := constants.AwsLoggerPermission
	return imrole.CheckAWSCredentials(requiredAction, request.AccessKey, request.SecretKey, request.Region)
}

func CheckCloudwatchLoggingPermissions(request AwsCredential) ([]string, []string, error) {
	requiredAction := constants.CloudwatchLoggerPermission
	return imrole.CheckAWSCredentials(requiredAction, request.AccessKey, request.SecretKey, request.Region)
}
func CheckGcpLoggingPermissions(request GCPCredential) ([]string, []string, error) {
	requiredRoles := constants.GCPLoggerPermission
	return imrole.CheckGCPCredentials(requiredRoles, request.KeyFile)
}
