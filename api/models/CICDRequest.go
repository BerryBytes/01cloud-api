package models

import (
	"errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CIRequest struct {
	EnvironmentId     int64                   `json:"environment_id"`
	Namespace         string                  `json:"namespace"`
	ConfigPath        string                  `json:"config_path"`
	Name              string                  `json:"name"`
	ImageRepoUsername string                  `json:"image_repo_username"`
	ImageRepoPassword string                  `json:"image_repo_password"`
	ImageRepoService  string                  `json:"image_repo_service"`
	ImageRepoProject  string                  `json:"image_repo_project"`
	ImageRepoProvider string                  `json:"image_repo_provider"`
	GitUrl            string                  `json:"git_url"`
	GitBranch         string                  `json:"git_branch"`
	GitUserName       string                  `json:"git_username"`
	GitAccessToken    string                  `json:"git_access_token"`
	SubDirectory      string                  `json:"sub_directory"`
	RepositoryImage   RepositoryImage         `json:"repository_image"`
	PluginUrl         string                  `json:"plugin_url"`
	Author            string                  `json:"author"`
	CommitMessage     string                  `json:"commit_message"`
	BaseImage         string                  `json:"base_image"`
	BaseTag           string                  `json:"base_tag"`
	BuildScript       string                  `json:"base_script"`
	RunScript         string                  `json:"run_script"`
	CloneEnv          *CloneEnvironment       `json:"clone_environment"`
	CISteps           interface{}             `json:"ci_steps"`
	CiFiles           *map[string]interface{} `json:"ci_files"`
}

type CICDOptions struct {
	WorkflowName      string
	TagName           string
	ManualBuildAuthor string
	Raw               map[string]interface{}
}

type WorkflowMetadata struct {
	Workflow  map[string]interface{} `json:"workflow"`
	CIRequest CIRequest              `json:"ci_request"`
	Namespace string                 `json:"namespace"`
	Pipeline  *CIPipeline            `json:"pipeline"`
	LogSteps  []LogSteps             `json:"log_steps"`
	CloneEnv  *CloneEnvironment      `json:"clone_environment"`
}

type CIPipeline struct {
	Name           string      `json:"name"`
	Namespace      string      `json:"namespace"`
	Runner         string      `json:"runner"`
	Tasks          []Task      `json:"tasks"`
	CreationTime   metav1.Time `json:"creation_time"`
	CompletionTime metav1.Time `json:"completion_time"`
	Status         string      `json:"status"`
}

type Task struct {
	Name           string      `json:"name"`
	RunAfter       []string    `json:"run_after"`
	Status         string      `json:"status"`
	CreationTime   metav1.Time `json:"creation_time"`
	CompletionTime metav1.Time `json:"completion_time"`
	Steps          []Step      `json:"steps"`
}

type Step struct {
	Name           string      `json:"name"`
	Status         string      `json:"status"`
	CreationTime   metav1.Time `json:"creation_time"`
	CompletionTime metav1.Time `json:"completion_time"`
}

type ClusterWorkflowMetadata struct {
	Workflow  map[string]interface{}
	CIRequest CreateClusterRequest
	Namespace string
	Type      string
}

type LogSteps struct {
	Type string
	Step int
}

type CIResponse struct {
	Name           string            `json:"Name"`
	Namespace      string            `json:"Namespace"`
	Type           string            `json:"Type"`
	Status         string            `json:"Status"`
	IsCustomCI     bool              `json:"IsCustomCI"`
	HelmFileString string            `json:"HelmFileString"`
	CloneEnv       *CloneEnvironment `json:"CloneEnv"`
}

type CreateClusterRequest struct {
	ID                int64       `json:"id"`
	Name              string      `json:"name"`
	Namespace         string      `json:"namespace"`
	ConfigPath        string      `json:"config_path"`
	ImageRepoUsername string      `json:"image_repo_username"`
	ImageRepoPassword string      `json:"image_repo_password"`
	ImageRepoService  string      `json:"image_repo_service"`
	ImageRepoProject  string      `json:"image_repo_project"`
	BucketAccessKey   string      `json:"bucket_access_key"`
	BucketSecretKey   string      `json:"bucket_secret_key"`
	UserAccessKey     string      `json:"user_access_key"`
	UserSecretKey     string      `json:"user_secret_key"`
	BucketName        string      `json:"bucket_name"`
	BucketPath        string      `json:"bucket_path"`
	Type              string      `json:"type"`
	Region            string      `json:"region"`
	HelmFileString    string      `json:"helm_file_string"`
	KubeConfigString  string      `json:"kube_config_string"`
	SubscriptonID     string      `json:"subscription_id"`
	VClusterAPIKEY    string      `json:"vcluster_api_key"`
	ZeroneAPIKEY      string      `json:"zerone_api_key"`
	Custom            interface{} `json:"custom"`
	Kube_version      string      `json:"kube_version"`
}

func ValidatePipeline(pipeline, task, step string) error {
	if pipeline == "" {
		return errors.New("required pipeline name")
	}
	if task == "" {
		return errors.New("required task name")
	}
	if step == "" {
		return errors.New("required step name")
	}
	return nil
}

func (pipeline *CIPipeline) IsTaskAvailable(taskName string) bool {
	for _, task := range pipeline.Tasks {
		if task.Name == taskName {
			return true
		}
	}
	return false
}

func (pipeline *CIPipeline) IsStepAvailable(taskName, stepName string) bool {
	for _, task := range pipeline.Tasks {
		if task.Name == taskName {
			for _, step := range task.Steps {
				if step.Name == stepName {
					return true
				}
			}
			return false
		}
	}
	return false
}
