package git

import (
	"01cloud-api/api/models"
	"01cloud-api/api/utils/datatype"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	"github.com/xanzy/go-gitlab"

	"golang.org/x/oauth2"
	gitlab_endpoint "golang.org/x/oauth2/gitlab"
)

type GitLab struct {
	gitUser *gitlab.Client
}

func NewGitLab(gitUser models.GitUser) (GitInterface, error) {
	var gitClient *gitlab.Client
	var err error
	if gitUser.IsOauth {
		gitClient, err = gitlab.NewOAuthClient(gitUser.AccessToken)
		if err != nil {
			return nil, err
		}
	} else {
		if gitUser.ServiceUrl == "" {
			gitUser.ServiceUrl = "https://gitlab.com"
		}
		gitUrl := gitlab.WithBaseURL(gitUser.ServiceUrl)
		gitClient, err = gitlab.NewClient(gitUser.AccessToken, gitUrl)
		if err != nil {
			log.Warnf("Failed to create client: %v", err)
		}
	}
	_, _, err = gitClient.Users.CurrentUser()
	if err != nil {
		return nil, errors.New("invalid credintials")
	}
	return &GitLab{gitUser: gitClient}, nil
}
func (g GitLab) GetRepos(orgId string) ([]KVResponse, error) {
	resData := []KVResponse{}
	var repos []*gitlab.Project
	var err error
	if orgId == "" {
		opt := &gitlab.ListProjectsOptions{Owned: gitlab.Bool(true)}
		repos, _, err = g.gitUser.Projects.ListProjects(opt)
		if err != nil {
			return nil, err
		}
	} else {
		repos, _, err = g.gitUser.Groups.ListGroupProjects(orgId, &gitlab.ListGroupProjectsOptions{ListOptions: gitlab.ListOptions{PerPage: 1000}})
		if err != nil {
			return nil, err
		}
	}
	for _, d := range repos {

		resData = append(resData, KVResponse{
			Key:   d.Name,
			Value: fmt.Sprint(d.ID),
		})
	}
	return resData, nil
}

func (g GitLab) GetBranches(repoid string) ([]KVResponse, error) {

	bl := &gitlab.ListBranchesOptions{}
	branches, _, err := g.gitUser.Branches.ListBranches(repoid, bl)
	if err != nil {
		return nil, err
	}

	resData := []KVResponse{}
	for _, d := range branches {

		resData = append(resData, KVResponse{
			Key:   d.Name,
			Value: d.Name,
		})
	}
	return resData, nil
}

func (g GitLab) GetOrganizations() ([]KVResponse, error) {
	groups, _, err := g.gitUser.Groups.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return []KVResponse{}, errors.New("unable to get organizations")
	}
	resData := []KVResponse{}
	for _, d := range groups {

		resData = append(resData, KVResponse{
			Key:   d.Name,
			Value: fmt.Sprint(d.ID),
		})
	}
	return resData, nil
}

func (g GitLab) GetRepoDetails(repoId string) (models.GitRepo, error) {
	var userName string
	repo, _, err := g.gitUser.Projects.GetProject(repoId, &gitlab.GetProjectOptions{})
	if err != nil {
		return models.GitRepo{}, err
	}
	if repo.Owner == nil {
		userName = repo.Namespace.Name
	} else {
		userName = repo.Owner.Username
	}
	resRepo := models.GitRepo{
		ID:       strconv.Itoa(int(repo.ID)),
		Name:     repo.Name,
		Owner:    userName,
		HtmlUrl:  repo.WebURL,
		CloneURL: repo.HTTPURLToRepo,
	}
	return resRepo, nil
}
func (g GitLab) GetLatestCommit(environment models.Environment) (models.Commit, error) {
	branch, _, err := g.gitUser.Branches.GetBranch(environment.GitRepository.ID, environment.GitBranch)
	if err != nil {
		return models.Commit{}, err
	}
	commit := models.Commit{
		SHA:     branch.Commit.ShortID,
		Message: branch.Commit.Message,
		Author:  branch.Commit.AuthorName,
	}
	return commit, nil
}

func GitlabLogin(externalLogin models.ExternalLogin) (models.GitUser, error) {
	gitUser := models.GitUser{}
	if !externalLogin.IsOauth {
		gitUser.AccessToken = externalLogin.Code
		if externalLogin.Url == "" {
			externalLogin.Url = "https://gitlab.com"
		}
		gitClient, _, err := GetGitlabClient(externalLogin)
		if err != nil {
			return models.GitUser{}, errors.New("failed to create client")
		}
		usr, _, err := gitClient.Users.CurrentUser()
		if err != nil {
			return models.GitUser{}, errors.New("invalid credentials")
		}
		gitUser.ServiceUrl = externalLogin.Url
		gitUser.ServiceUserName = usr.Name
		gitUser.GitUserID = uint64(usr.ID)
	} else {
		client, accessToken, err := GetGitlabClient(externalLogin)
		if err != nil {
			return models.GitUser{}, errors.New("failed to create client")
		}
		user, _, err := client.Users.CurrentUser()
		if err != nil {
			return models.GitUser{}, err
		}
		gitUser.GitUserID = uint64(user.ID)
		gitUser.AccessToken = accessToken
		gitUser.ServiceUserName = user.Username
		gitUser.ServiceUrl = externalLogin.Url
		gitUser.IsOauth = externalLogin.IsOauth
	}
	return gitUser, nil
}

func GetGitlabClient(externalLogin models.ExternalLogin) (*gitlab.Client, string, error) {
	var client *gitlab.Client
	var err error
	var accessToken string
	if !externalLogin.IsOauth {
		gitUrl := gitlab.WithBaseURL(externalLogin.Url)
		client, err = gitlab.NewClient(externalLogin.Code, gitUrl)
		if err != nil {
			return nil, "", errors.New("failed to create client")
		}
	} else {
		ctx := context.Background()
		endpoint := gitlab_endpoint.Endpoint
		conf := &oauth2.Config{
			ClientID:     os.Getenv("GITLAB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITLAB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITLAB_REDIRECT_URL"),
			Scopes: []string{
				"api", "sudo", "profile", "email", "openid", "read_user", "write_repository", "write_registry",
			},
			Endpoint: endpoint,
		}
		tok, err := conf.Exchange(ctx, externalLogin.Code)
		if err != nil {
			return nil, "", errors.New("invalid service code")
		}
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: tok.AccessToken},
		)
		token, err := ts.Token()
		if err != nil {
			return nil, "", err
		}
		accessToken = token.AccessToken
		client, err = gitlab.NewOAuthClient(token.AccessToken)
		if err != nil {
			return nil, "", err
		}
	}
	return client, accessToken, nil
}

func GitlabUserLogin(externalLogin models.ExternalLogin) (*gitlab.User, error) {
	var client *gitlab.Client
	var err error
	if !externalLogin.IsOauth {
		if externalLogin.Url == "" {
			externalLogin.Url = "https://gitlab.com"
		}
		client, _, err = GetGitlabClient(externalLogin)
		if err != nil {
			return nil, err
		}
	} else {
		client, _, err = GetGitlabClient(externalLogin)
		if err != nil {
			return nil, err
		}
	}
	usr, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, err
	}
	return usr, nil

}

func (g GitLab) CreateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig) (*Hook, *http.Response, error) {
	if strings.TrimSpace(ciConfig.Events) == "" {
		return nil, nil, nil
	}
	pushEvents := strings.Contains(ciConfig.Events, "push")
	tagEvents := strings.Contains(ciConfig.Events, "release")
	url := fmt.Sprintf("%s/webhook/%d/%d", os.Getenv("BASE_URL"), environment.ID, userId)

	hook, response, err := g.gitUser.Projects.AddProjectHook(environment.GitRepository.ID, &gitlab.AddProjectHookOptions{
		URL:           gitlab.String(url),
		PushEvents:    &pushEvents,
		TagPushEvents: &tagEvents,
	})
	if err != nil {
		return nil, nil, err
	}
	webhook := Hook{
		ID: strconv.Itoa(int(hook.ID)),
	}

	return &webhook, response.Response, nil
}

func (g GitLab) UpdateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig, hookId string) (*Hook, *http.Response, error) {
	hID, err := strconv.Atoi(hookId)
	if err != nil {
		return nil, nil, err
	}

	if strings.TrimSpace(ciConfig.Events) == "" {
		res, err := g.gitUser.Projects.DeleteProjectHook(environment.GitRepository.ID, hID)
		return nil, res.Response, err
	}
	pushEvents := strings.Contains(ciConfig.Events, "push")
	tagEvents := strings.Contains(ciConfig.Events, "release")
	url := fmt.Sprintf("%s/webhook/%d/%d", os.Getenv("BASE_URL"), environment.ID, userId)
	hook, res, err := g.gitUser.Projects.EditProjectHook(environment.GitRepository.ID, hID, &gitlab.EditProjectHookOptions{
		URL:           gitlab.String(url),
		PushEvents:    &pushEvents,
		TagPushEvents: &tagEvents,
	},
	)
	if err != nil {
		return nil, nil, err
	}

	webhook := Hook{
		ID: strconv.Itoa(int(hook.ID)),
	}
	return &webhook, res.Response, nil
}

func (g GitLab) DeleteWebhook(environment *models.Environment, hookId string) error {
	hID, err := strconv.Atoi(hookId)
	if err != nil {
		return err
	}
	_, err = g.gitUser.Projects.DeleteProjectHook(environment.GitRepository.ID, hID)
	return err
}

func (g GitLab) GetWorkflowContent(environment *models.Environment) map[string]interface{} {
	workflows := map[string]interface{}{}
	trees, _, err := g.gitUser.Repositories.ListTree(environment.GitRepository.ID, &gitlab.ListTreeOptions{
		Path:      datatype.StringPtr(pipelineDir),
		Ref:       datatype.StringPtr(environment.GitBranch),
		Recursive: datatype.BoolPtr(true),
	})
	if err == nil {
		for _, tree := range trees {
			if tree.Type == "blob" {
				content, _, err := g.gitUser.Repositories.RawBlobContent(environment.GitRepository.ID, tree.ID)
				if err == nil {
					workflows[tree.Name] = string(content)
				}
			}
		}
	}
	return workflows
}

func (g GitLab) GetDockerfile(environment *models.Environment) ([]byte, error) {
	var dockerfile []byte
	trees, _, err := g.gitUser.Repositories.ListTree(environment.GitRepository.ID, &gitlab.ListTreeOptions{
		Path:      datatype.StringPtr(getSubDir(environment)),
		Ref:       datatype.StringPtr(environment.GitBranch),
		Recursive: datatype.BoolPtr(true),
	})
	if err == nil {
		for _, tree := range trees {
			if tree.Type == "blob" {
				if tree.Name == dockerFile {
					content, _, err := g.gitUser.Repositories.RawBlobContent(environment.GitRepository.ID, tree.ID)
					if err == nil {
						dockerfile = content
					}
				}
			}
		}
	}
	return dockerfile, err
}

func GitlabHookPayload(payload map[string]interface{}, repo models.GitRepo) (string, string, error) {
	var err error
	pushEvent := gitlab.PushEvent{}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", "", err
	}
	err = json.Unmarshal(data, &pushEvent)
	if err != nil {
		return "", "", err
	}
	payloadrepoId := pushEvent.ProjectID
	repoId, err := strconv.Atoi(repo.ID)
	if err != nil {
		return "", "", err
	}
	if payloadrepoId != repoId {
		return "", "", errors.New("invalid payload")
	}
	branch := pushEvent.Ref
	branch = strings.ReplaceAll(branch, "refs/heads/", "")
	tag := pushEvent.CheckoutSHA
	return tag, branch, nil
}

func (g GitLab) GitExpiredToken(gitUser models.GitUser, DB *gorm.DB) (GitInterface, error) {
	_, err := gitUserInterface.Delete(DB, uint64(gitUser.ID))
	if err != nil {
		return nil, err
	}
	return nil, errors.New("token expired")
}

func (g GitLab) Profile() error {
	return nil
}
