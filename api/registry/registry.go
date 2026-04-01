package registry

import (
	"01cloud-api/api/models"
	"context"
	"errors"
	"time"
)

type WebhookResponse struct {
	Tag string `json:"tag"`
	URL string `json:"url"`
}

type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Repository struct {
	RepositoryType string `json:"repository_type"`
	Status         int    `json:"status"`
	Description    string `json:"description"`
	IsPrivate      bool   `json:"is_private"`
	IsAutomated    bool   `json:"is_automated"`
	CanEdit        bool   `json:"can_edit"`
	LastUpdated    string `json:"last_updated"`
	Name           string `json:"name"`
	Uri            string `json:"uri"`
	RepositoryArn  string `json:"repository_arn"`
}

type RepositoryList struct {
	Description    string `json:"description"`
	IsPrivate      bool   `json:"is_private"`
	LastUpdated    string `json:"last_updated"`
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	PullCount      int    `json:"pull_count"`
	RepositoryType string `json:"repository_type"`
	Uri            string `json:"uri"`
}

type Tags struct {
	Creator       int       `json:"creator"`
	ID            string    `json:"id"`
	LastUpdated   time.Time `json:"last_updated"`
	LastUpdatedBy string    `json:"last_updated_by"`
	Name          string    `json:"name"`
	Repository    int       `json:"repository"`
	FullSize      int       `json:"full_size"`
	TagStatus     string    `json:"tag_status"`
	TagLastPulled time.Time `json:"tag_last_pulled"`
	TagLastPushed time.Time `json:"tag_last_pushed"`
}
type BaseRegistry struct{}
type RegistryInterface interface {
	CreateWebhook(ctx context.Context, namespace, repo, name, url string) error
	ParseWebhook(body []byte) (*WebhookResponse, error)
	GetOrganizations(ctx context.Context, limit int) ([]Organization, error)
	GetRepositories(ctx context.Context, namespace string) ([]RepositoryList, error)
	GetCurrentUser(ctx context.Context) (*User, error)
	GetRegistry(ctx context.Context, namespace, repo string) (*Repository, error)
	GetTags(ctx context.Context, namespace, repo string, limit int) ([]Tags, error)
}

func NewRegistry(git models.GitUser) (RegistryInterface, error) {
	switch {
	case git.ServiceName == "dockerhub":
		return NewDockerhub(git)
	case git.ServiceName == "aws_ecr":
		return NewEcr(git)
	default:
		return nil, nil
	}
}

func RegistryLogin(externalLogin models.ExternalLogin) (*models.GitUser, error) {
	switch {
	case externalLogin.Service == "dockerhub":
		return DockerhubLogin(externalLogin)
	case externalLogin.Service == "aws_ecr":
		return EcrLogin(externalLogin)
	default:
		return nil, errors.New("unable to registry login")
	}
}
