package git

import (
	"01cloud-api/api/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/google/go-github/github"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	github_endpoint "golang.org/x/oauth2/github"
)

type GitHub struct {
	gitUser *github.Client
}

var gitUserInterface = models.NewGitUser()

func NewGithub(gitUser models.GitUser) GitInterface {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(

		&oauth2.Token{AccessToken: gitUser.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &GitHub{gitUser: client}
}
func (g GitHub) GetRepos(orgId string) ([]KVResponse, error) {
	ctx := context.Background()
	resData := []KVResponse{}
	var repos []*github.Repository
	var err error
	if orgId == "" {
		repos, _, err = g.gitUser.Repositories.List(ctx, "", &github.RepositoryListOptions{
			Affiliation: "owner,collaborator",
			ListOptions: github.ListOptions{
				PerPage: 1000,
			},
		})
		if err != nil {
			return []KVResponse{}, err
		}

	} else {
		repos, _, err = g.gitUser.Repositories.ListByOrg(ctx, orgId, &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 1000}})
		if err != nil {
			return []KVResponse{}, err
		}
	}
	for _, d := range repos {
		resData = append(resData, KVResponse{
			Key:   *d.Name,
			Value: fmt.Sprint(*d.ID),
		})
	}
	return resData, nil
}

func (g GitHub) GetBranches(repoid string) ([]KVResponse, error) {
	ctx := context.Background()
	id, _ := strconv.Atoi(repoid)
	repos, _, err := g.gitUser.Repositories.GetByID(ctx, int64(id))
	if err != nil {
		return []KVResponse{}, err
	}
	branches, _, err := g.gitUser.Repositories.ListBranches(ctx,
		*repos.Owner.Login, *repos.Name, &github.ListOptions{PerPage: 1000})
	if err != nil {

		return []KVResponse{}, err
	}
	resData := []KVResponse{}
	for _, b := range branches {
		resData = append(resData, KVResponse{
			Key:   *b.Name,
			Value: *b.Name,
		})
	}
	return resData, nil

}
func (g GitHub) GetOrganizations() ([]KVResponse, error) {
	resData := []KVResponse{}
	ctx := context.Background()
	repos, _, err := g.gitUser.Organizations.List(ctx, "", &github.ListOptions{})
	if err != nil {
		return []KVResponse{}, errors.New("unable to get organization")
	}
	for _, d := range repos {
		resData = append(resData, KVResponse{
			Key:   *d.Login,
			Value: *d.Login,
		})
	}
	return resData, nil
}

func (g GitHub) GetRepoDetails(repoId string) (models.GitRepo, error) {
	id, _ := strconv.Atoi(repoId)
	ctx := context.Background()
	repo, _, err := g.gitUser.Repositories.GetByID(ctx, int64(id))
	if err != nil {
		return models.GitRepo{}, err
	}
	resRepo := models.GitRepo{
		ID:       strconv.Itoa(int(*repo.ID)),
		Owner:    *repo.Owner.Login,
		Name:     *repo.Name,
		HtmlUrl:  *repo.HTMLURL,
		CloneURL: *repo.CloneURL,
		GitURL:   *repo.GitURL,
	}
	return resRepo, nil
}
func (g GitHub) GetLatestCommit(environment models.Environment) (models.Commit, error) {
	ctx := context.Background()
	branch, _, err := g.gitUser.Repositories.GetBranch(ctx, environment.GitRepository.Owner, environment.GitRepository.Name, environment.GitBranch)
	if err != nil {
		return models.Commit{}, err
	}
	commit := models.Commit{
		SHA:     *branch.Commit.SHA,
		Message: *branch.Commit.Commit.Message,
	}
	author := branch.Commit.Author
	if author != nil {
		commit.Author = *author.Login
	}
	return commit, nil
}
func GithubLogin(externalLogin models.ExternalLogin) (models.GitUser, error) {
	ctx := context.Background()
	gitUser := models.GitUser{}
	client, accessToken, err := GetGithubClient(externalLogin)
	if err != nil {
		return models.GitUser{}, err
	}
	usr, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return models.GitUser{}, err
	}
	gitUser.GitUserID = uint64(*usr.ID)
	gitUser.AccessToken = accessToken
	gitUser.ServiceUserName = *usr.Login
	return gitUser, nil
}

func GetGithubClient(externalLogin models.ExternalLogin) (*github.Client, string, error) {
	ctx := context.Background()
	conf := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint:     github_endpoint.Endpoint,
	}
	tok, err := conf.Exchange(ctx, externalLogin.Code)
	if err != nil {
		return nil, "", err
	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: tok.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return client, tok.AccessToken, nil

}

func GithubUserLogin(externalLogin models.ExternalLogin) (*github.User, string, error) {
	ctx := context.Background()
	client, token, err := GetGithubClient(externalLogin)
	if err != nil {
		return nil, "", err
	}
	usr, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, "", err
	}
	emails, _, err := client.Users.ListEmails(ctx, &github.ListOptions{})
	if err != nil {
		return nil, "", err
	}
	for _, email := range emails {
		if *email.Primary {
			usr.Email = email.Email
			break
		}
	}
	return usr, token, nil
}

func (g GitHub) CreateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig) (*Hook, *http.Response, error) {

	if strings.TrimSpace(ciConfig.Events) == "" {
		return nil, nil, nil
	}
	events := strings.Split(ciConfig.Events, ",")
	//err = json.Unmarshal([]byte(ciConfig.Events), &events)
	//if err != nil {
	//	return nil, nil, nil
	//}

	active := true
	url := fmt.Sprintf("%s/webhook/%d/%d", os.Getenv("BASE_URL"), environment.ID, userId)

	config := map[string]interface{}{
		"url":          url,
		"content_type": "json",
	}

	hook, response, err := g.gitUser.Repositories.CreateHook(context.Background(), environment.GitRepository.Owner, environment.GitRepository.Name, &github.Hook{
		Events: events,
		Active: &active,
		Config: config,
	})
	if err != nil {
		if response.StatusCode == 404 {
			return nil, nil, errors.New("the application owner does not have the admin permission")
		}
		return nil, nil, err
	}
	webhook := Hook{
		ID: strconv.Itoa(int(*hook.ID)),
	}

	return &webhook, response.Response, nil
}

func (g GitHub) UpdateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig, hookId string) (*Hook, *http.Response, error) {
	hID, err := strconv.Atoi(hookId)
	if err != nil {
		return nil, nil, err
	}
	if strings.TrimSpace(ciConfig.Events) == "" {
		res, err := g.gitUser.Repositories.DeleteHook(context.Background(), environment.GitRepository.Owner, environment.GitRepository.Name, int64(hID))
		return nil, res.Response, err
	}
	events := strings.Split(ciConfig.Events, ",")
	active := true
	url := fmt.Sprintf("%s/webhook/%d/%d", os.Getenv("BASE_URL"), environment.ID, userId)

	config := map[string]interface{}{
		"url":          url,
		"content_type": "json",
	}
	hook, res, err := g.gitUser.Repositories.EditHook(context.Background(), environment.GitRepository.Owner, environment.GitRepository.Name, int64(hID), &github.Hook{
		Events: events,
		Active: &active,
		Config: config,
	})
	if err != nil {
		return nil, nil, err
	}

	webhook := Hook{
		ID: strconv.Itoa(int(*hook.ID)),
	}
	return &webhook, res.Response, nil
}

func (g GitHub) DeleteWebhook(environment *models.Environment, hookId string) error {
	hID, err := strconv.Atoi(hookId)
	if err != nil {
		return err
	}
	_, err = g.gitUser.Repositories.DeleteHook(context.Background(), environment.GitRepository.Owner, environment.GitRepository.Name, int64(hID))
	return err
}

func (g GitHub) GetWorkflowContent(environment *models.Environment) map[string]interface{} {
	workflows := map[string]interface{}{}
	re, s, _, err := g.gitUser.Repositories.GetContents(context.Background(), environment.GitRepository.Owner, environment.GitRepository.Name, fmt.Sprintf("%s/%s", getSubDir(environment), pipelineDir), &github.RepositoryContentGetOptions{
		Ref: environment.GitBranch,
	})
	if err == nil {
		logrus.Debugf("github content response :: %v", re)
		for _, pipeline := range s {
			if *pipeline.Type == "file" {
				logrus.Debugf("pipeline file name :: %s", *pipeline.Name)
				content, err := getFileContent(pipeline.GetDownloadURL())
				if err == nil {
					workflows[*pipeline.Name] = string(content)
				}
			}
		}
	}
	return workflows
}

func (g GitHub) GetDockerfile(environment *models.Environment) ([]byte, error) {
	var dockerfile []byte
	re, s, _, err := g.gitUser.Repositories.GetContents(context.Background(), environment.GitRepository.Owner, environment.GitRepository.Name, getSubDir(environment), &github.RepositoryContentGetOptions{
		Ref: environment.GitBranch,
	})
	if err == nil {
		logrus.Debugf("github content response :: %v", re)
		for _, st := range s {
			if *st.Type == "file" {
				logrus.Debugf("File Name :: %s Downlaod URL :: %s ", *st.Name, st.GetDownloadURL())
				if *st.Name == dockerFile {
					dockerfile, _ = getFileContent(st.GetDownloadURL())
				}
			}
		}
	}
	return dockerfile, err
}

func GithubHookPayload(payload map[string]interface{}, repo models.GitRepo) (string, string, error) {
	var branch, tag string
	r, ok := payload["repository"].(map[string]interface{})["id"].(float64)
	repoId := int64(r)
	if !ok {
		repoId = payload["repository"].(map[string]interface{})["id"].(int64)
	}
	id, _ := strconv.Atoi(repo.ID)
	if repoId != int64(id) {
		return "", "", errors.New("invalid payload")
	}
	if payload["base_ref"] != nil {
		return "", "", errors.New("invalid action")
	}
	if v, ok := payload["ref"]; ok && v != nil {
		branch = v.(string)
		branch = strings.ReplaceAll(branch, "refs/heads/", "")
	} else {
		if v, ok := payload["action"]; ok && v != nil && strings.ToLower(v.(string)) == "published" {
			branch = payload["release"].(map[string]interface{})["target_commitish"].(string)
			tag = payload["release"].(map[string]interface{})["tag_name"].(string)
		} else {
			return "", "", errors.New("invalid action")
		}
	}
	return tag, branch, nil
}

func (g GitHub) GitExpiredToken(gitUser models.GitUser, DB *gorm.DB) (GitInterface, error) {
	_, err := gitUserInterface.Delete(DB, uint64(gitUser.ID))
	if err != nil {
		return nil, err
	}
	return nil, errors.New("token expired")
}

func (g GitHub) Profile() error {
	_, _, err := g.gitUser.Users.Get(context.Background(), "")
	if err != nil {
		return err
	}
	return nil
}
