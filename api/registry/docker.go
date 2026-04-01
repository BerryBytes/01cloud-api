package registry

import (
	"01cloud-api/api/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	dockerhub "github.com/berrybytes/dockerhub-go"
)

type DockerRegistry struct {
	BaseRegistry
	hub *dockerhub.Client
}

type PushData struct {
	Tag string `json:"tag"`
}

type Repo struct {
	RepoURL string `json:"repo_url"`
}
type DockerhubWebhook struct {
	PushData   PushData `json:"push_data"`
	Repository Repo     `json:"repository"`
}

func NewDockerhub(user models.GitUser) (*DockerRegistry, error) {
	client, err := NewDockerHubRegistrty(context.Background(), user.ServiceUserName, user.AccessToken)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewDockerHubRegistrty(ctx context.Context, username, token string) (*DockerRegistry, error) {
	client := dockerhub.NewClient(nil)
	err := client.Auth.Login(ctx, username, token)

	if err != nil {
		return nil, err
	}
	registry := &DockerRegistry{
		hub: client,
	}
	return registry, nil
}

func DockerhubLogin(externalLogin models.ExternalLogin) (*models.GitUser, error) {
	client, err := NewDockerHubRegistrty(context.Background(), externalLogin.User, externalLogin.Code)
	if err != nil {
		return nil, errors.New("invalid credentials")

	}
	err = client.hub.Auth.Login(context.Background(), externalLogin.User, externalLogin.Code)
	if err != nil {
		return nil, errors.New("unauthorized user")
	}
	gitUser := models.GitUser{}
	gitUser.AccessToken = externalLogin.Code
	gitUser.ServiceName = externalLogin.Service
	gitUser.ServiceUserName = externalLogin.User
	return &gitUser, nil
}

func (r *DockerRegistry) CreateWebhook(ctx context.Context, namespace, repo, name, url string) error {
	_, err := r.hub.Webhook.CreateWebhook(context.Background(), namespace, repo, name, url)

	if err != nil {
		return err
	}
	return nil
}

func (r *DockerRegistry) ParseWebhook(body []byte) (*WebhookResponse, error) {
	hook := &DockerhubWebhook{}
	err := json.Unmarshal(body, &hook)
	if err != nil {
		return nil, err
	}
	response := &WebhookResponse{
		Tag: hook.PushData.Tag,
		URL: hook.Repository.RepoURL,
	}
	return response, err
}

func (r *DockerRegistry) GetOrganizations(ctx context.Context, limit int) ([]Organization, error) {
	org, err := r.hub.Organization.GetOrganizations(ctx, limit)

	var orgs []Organization

	if err != nil {
		return nil, err
	}

	for _, organization := range org.Results {
		orgs = append(orgs, Organization{
			ID:   organization.ID,
			Name: organization.Orgname,
		})
	}
	return orgs, nil
}

func (r *DockerRegistry) GetRepositories(ctx context.Context, namespace string) ([]RepositoryList, error) {
	repos, err := r.hub.Repositories.GetRepositories(ctx, namespace, &dockerhub.ListOptions{
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		return nil, err
	}
	var repoNames []RepositoryList
	for _, repo := range repos.Results {
		repoNames = append(repoNames, RepositoryList{
			Description:    repo.Description,
			IsPrivate:      repo.IsPrivate,
			LastUpdated:    repo.LastUpdated,
			Name:           repo.Name,
			Namespace:      repo.Namespace,
			PullCount:      repo.PullCount,
			RepositoryType: repo.RepositoryType,
		})
	}
	return repoNames, nil
}

func (r *DockerRegistry) GetCurrentUser(ctx context.Context) (*User, error) {
	user, err := r.hub.User.GetLoggedInUser(ctx)

	if err != nil {
		return nil, err
	}

	return &User{
		ID:       user.ID,
		Username: user.Username,
	}, nil
}

func (r *DockerRegistry) GetRegistry(ctx context.Context, namespace, repo string) (*Repository, error) {
	registryData, err := r.hub.Repositories.GetRepository(ctx, namespace, repo)

	if err != nil {
		return nil, err
	}
	return &Repository{
		RepositoryType: registryData.RepositoryType,
		Status:         registryData.Status,
		Description:    registryData.Description,
		IsPrivate:      registryData.IsPrivate,
		IsAutomated:    registryData.IsAutomated,
		CanEdit:        registryData.CanEdit,
		LastUpdated:    registryData.LastUpdated,
	}, nil
}

func (r *DockerRegistry) GetTags(ctx context.Context, namespace, repo string, limit int) ([]Tags, error) {
	tags, err := r.hub.Tag.GetTags(ctx, namespace, repo, limit)
	var tagResponse []Tags

	if err != nil {
		return nil, err
	}
	for _, tag := range tags.Results {
		tagResponse = append(tagResponse, Tags{
			Creator:       tag.Creator,
			ID:            fmt.Sprint(tag.ID),
			LastUpdated:   tag.LastUpdated,
			LastUpdatedBy: tag.LastUpdaterUsername,
			Name:          tag.Name,
			Repository:    tag.Repository,
			FullSize:      tag.FullSize,
			TagStatus:     tag.TagStatus,
			TagLastPulled: tag.TagLastPulled,
			TagLastPushed: tag.TagLastPushed,
		})
	}
	return tagResponse, nil
}
