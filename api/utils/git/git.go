package git

import (
	"01cloud-api/api/models"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jinzhu/gorm"
)

type KVResponse struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

const (
	pipelineDir string = ".01cloud"
	dockerFile  string = "Dockerfile"
)

type Hook struct {
	ID     string                 `json:"id,omitempty"`
	Name   string                 `json:"name,omitempty"`
	URL    string                 `json:"url,omitempty"`
	Events []string               `json:"events,omitempty"`
	Active bool                   `json:"active,omitempty"`
	Config map[string]interface{} `json:"config,omitempty"`
}

type GitInterface interface {
	GetRepos(org string) ([]KVResponse, error)
	GetBranches(repoId string) ([]KVResponse, error)
	GetOrganizations() ([]KVResponse, error)
	GetRepoDetails(repoId string) (models.GitRepo, error)
	GetLatestCommit(environment models.Environment) (models.Commit, error)
	CreateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig) (*Hook, *http.Response, error)
	UpdateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig, hookId string) (*Hook, *http.Response, error)
	DeleteWebhook(environment *models.Environment, hookId string) error
	GetWorkflowContent(environment *models.Environment) map[string]interface{}
	GetDockerfile(environment *models.Environment) ([]byte, error)
	GitExpiredToken(git models.GitUser, db *gorm.DB) (GitInterface, error)
	Profile() error
}

func NewGit(git models.GitUser) (GitInterface, error) {
	var err error
	switch {
	case git.ServiceName == "github":
		client := NewGithub(git)
		err = client.Profile()
		if err != nil {
			return nil, err
		}
		return client, nil
		//return NewGithub(git), nil
	case git.ServiceName == "gitlab":
		client, err := NewGitLab(git)
		if err != nil {
			return nil, err
		}
		return client, nil
		//return NewGitLab(git)
	case git.ServiceName == "bitbucket":
		client, _ := NewBitBucket(git)
		err = client.Profile()
		if err != nil {
			return nil, err
			// client, err = RefreshToken(git, db)
			// if err != nil {
			// 	return nil, err
			// }
		}
		return client, nil
	default:
		return nil, nil
	}
}

func GitLogin(externalLogin models.ExternalLogin) (models.GitUser, error) {
	switch {
	case externalLogin.Service == "github":
		return GithubLogin(externalLogin)
	case externalLogin.Service == "gitlab":
		return GitlabLogin(externalLogin)
	case externalLogin.Service == "bitbucket":
		return BitBucketLogin(externalLogin)
	default:
		return models.GitUser{}, errors.New("unable to git login")
	}
}

func ParseWebHookPayload(payload map[string]interface{}, repo models.GitRepo, environmentReceived *models.Environment) (string, string, error) {
	switch {
	case environmentReceived.Application.GitService == "github":
		return GithubHookPayload(payload, repo)
	case environmentReceived.Application.GitService == "bitbucket":
		return BitBucketHookPayload(payload, repo)
	case environmentReceived.Application.GitService == "gitlab":
		return GitlabHookPayload(payload, repo)
	default:
		return "", "", errors.New("unable to find git service")
	}
}

func ExpiredToken(git models.GitUser, db *gorm.DB) (GitInterface, error) {
	switch {
	case git.ServiceName == "bitbucket":
		var bitBucket BitBucket
		client, err := bitBucket.GitExpiredToken(git, db)
		if err != nil {
			return nil, err
		}
		return client, nil
	case git.ServiceName == "github":
		var github GitHub
		client, err := github.GitExpiredToken(git, db)
		if err != nil {
			return nil, err
		}
		return client, nil
	case git.ServiceName == "gitlab":
		var gitLab GitLab
		client, err := gitLab.GitExpiredToken(git, db)
		if err != nil {
			return nil, err
		}
		return client, nil
	default:
		return nil, errors.New("token expired")
	}
}

// Get content from provide url
func getFileContent(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func getSubDir(req *models.Environment) string {
	subDir := ""
	envScripts := map[string]interface{}{}
	err := json.Unmarshal(req.Scripts.RawMessage, &envScripts)
	if err == nil {
		if s, ok := envScripts["sub_dir"].(string); ok {
			subDir = s
		}
	}
	return subDir
}
