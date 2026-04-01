package git

import (
	"01cloud-api/api/models"
	"01cloud-api/api/utils/helper"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	bitbucket "github.com/ktrysmt/go-bitbucket"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	bitbucket_endpoint "golang.org/x/oauth2/bitbucket"
)

type BitBucket struct {
	gitUser *bitbucket.Client
}

func NewBitBucket(gitUser models.GitUser) (GitInterface, error) {
	gitClient := bitbucket.NewOAuthbearerToken(gitUser.AccessToken)
	return &BitBucket{gitUser: gitClient}, nil
}

func BitBucketLogin(externalLogin models.ExternalLogin) (models.GitUser, error) {
	ctx := context.Background()
	gitUser := models.GitUser{}
	conf := &oauth2.Config{
		ClientID:     os.Getenv("BITBUCKET_CLIENT_ID"),
		ClientSecret: os.Getenv("BITBUCKET_CLIENT_SECRET"),
		Endpoint:     bitbucket_endpoint.Endpoint,
	}
	tok, err := conf.Exchange(ctx, externalLogin.Code)
	if err != nil {
		log.Warnf("error occured :: %v", err)
		return models.GitUser{}, err
	}
	client := bitbucket.NewOAuthbearerToken(tok.AccessToken)
	usr, err := client.User.Profile()
	if err != nil {
		log.Warnf("error occured :: %v", err)
		return models.GitUser{}, nil
	}
	data, err := json.Marshal(tok)
	if err != nil {
		log.Warnf("error occured :: %v", err)
		return models.GitUser{}, err
	}
	var token *postgres.Jsonb
	err = json.Unmarshal(data, &token)
	if err != nil {
		log.Warnf("error occured :: %v", err)
		return models.GitUser{}, nil
	}
	gitUser.ServiceUserName = usr.Username
	gitUser.IsOauth = true
	gitUser.Token = token
	gitUser.AccessToken = tok.AccessToken
	return gitUser, nil
}

func (g BitBucket) GetRepos(orgId string) ([]KVResponse, error) {
	resData := []KVResponse{}
	if orgId == "" {
		_, err := g.gitUser.User.Profile()
		if err != nil {
			logrus.Info("error while getting profile of bitbucket user :: ", err)
			return []KVResponse{}, err
		}
		//set the owner to 'default' so that error won't occur as bitbucket only has workspace/org support
		repos, err := g.gitUser.Workspaces.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{Owner: "default"})
		if err != nil {
			logrus.Info("error while getting repos of bitbucket user :: ", err)
			return []KVResponse{}, err
		}
		for _, d := range repos.Items {
			resData = append(resData, KVResponse{
				Key:   d.Name,
				Value: d.Slug,
			})
		}
	} else {
		repos, err := g.gitUser.Workspaces.Repositories.ListForAccount(&bitbucket.RepositoriesOptions{Owner: orgId})
		if err != nil {
			logrus.Info("error while getting repos of bitbucket user :: ", err)
			return []KVResponse{}, err
		}
		for _, d := range repos.Items {
			resData = append(resData, KVResponse{
				Key:   d.Name,
				Value: d.Slug,
			})
		}
	}
	return resData, nil
}

func (g BitBucket) GetBranches(repoid string) ([]KVResponse, error) {
	owner, rslug, err := g.GetWorkSpaceAndRslug(repoid)
	if err != nil {
		return []KVResponse{}, err
	}
	branches, err := g.gitUser.Repositories.Repository.ListBranches(&bitbucket.RepositoryBranchOptions{Owner: owner, RepoSlug: rslug})
	if err != nil {
		return []KVResponse{}, err
	}
	resData := []KVResponse{}
	for _, b := range branches.Branches {
		resData = append(resData, KVResponse{
			Key:   b.Name,
			Value: b.Name,
		})
	}
	return resData, nil

}

func (g BitBucket) GetOrganizations() ([]KVResponse, error) {
	resData := []KVResponse{}
	wsList, err := g.gitUser.Workspaces.List()
	if err != nil {
		return []KVResponse{}, errors.New("unable to get organization " + err.Error())
	}
	for _, workspace := range wsList.Workspaces {
		resData = append(resData, KVResponse{
			Key:   workspace.Name,
			Value: workspace.Slug,
		})
	}
	return resData, nil
}

func (g BitBucket) GetRepoDetails(repoId string) (models.GitRepo, error) {
	owner, rslug, err := g.GetWorkSpaceAndRslug(repoId)
	if err != nil {
		return models.GitRepo{}, err
	}
	repo, err := g.gitUser.Workspaces.Repositories.Repository.Get(&bitbucket.RepositoryOptions{Owner: owner, RepoSlug: rslug})
	if err != nil {
		return models.GitRepo{}, err
	}
	value := repo.Links
	htmlUrl, _ := helper.GetValues(value, "html", "href")
	webUrl, _ := helper.GetValues(value, "downloads", "href")
	cloneUrl := ""
	if cresp, ok := value["clone"].([]interface{}); ok {
		for _, j := range cresp {
			clonemap := j.(map[string]interface{})
			cloneUrl = fmt.Sprintf("%v", (clonemap["href"]))
			break
		}
	}
	if len(cloneUrl) == 0 {
		return models.GitRepo{}, errors.New("clone url not found")
	}
	url := strings.Split(cloneUrl, "@")
	if len(url) > 1 {
		cloneUrl = fmt.Sprintf("https://%s", url[1])
	}
	resRepo := models.GitRepo{
		ID:       repo.Slug,
		Owner:    owner,
		Name:     repo.Name,
		HtmlUrl:  fmt.Sprintf("%v", htmlUrl),
		CloneURL: cloneUrl,
		GitURL:   fmt.Sprintf("%v", webUrl),
	}
	return resRepo, nil
}
func (g BitBucket) GetLatestCommit(environment models.Environment) (models.Commit, error) {
	commit, err := g.gitUser.Repositories.Commits.GetCommits(&bitbucket.CommitsOptions{Owner: environment.GitRepository.Owner, RepoSlug: environment.GitRepository.ID, Branchortag: environment.GitBranch})
	if err != nil {
		return models.Commit{}, err
	}
	type MyCommit struct {
		Author  interface{} `json:"author"`
		Message string      `json:"message"`
		SHA     string      `json:"hash"`
		Date    string      `json:"date"`
	}
	mycommit := []MyCommit{}
	res, ismap := commit.(map[string]interface{})
	if ismap {
		value := res["values"]
		resp, err := json.Marshal(value)
		if err != nil {
			return models.Commit{}, err
		}
		err = json.Unmarshal(resp, &mycommit)
		if err != nil {
			return models.Commit{}, err
		}
		for i := 0; i < len(mycommit); i++ {
			storeData, ok := mycommit[i].Author.(map[string]interface{})
			if ok {
				mycommit[i].Author = storeData["raw"]
			}
		}
	}
	if len(mycommit) == 0 {
		return models.Commit{}, nil
	}
	latestcommit := models.Commit{
		SHA:     mycommit[0].SHA,
		Message: mycommit[0].Message,
		Author:  fmt.Sprintf("%v", mycommit[0].Author),
	}
	return latestcommit, nil
}

func (g BitBucket) CreateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig) (*Hook, *http.Response, error) {
	if strings.TrimSpace(ciConfig.Events) == "" {
		return nil, nil, nil
	}
	var events []string
	pushEvents := strings.Contains(ciConfig.Events, "push")
	tagEvents := strings.Contains(ciConfig.Events, "release")
	if pushEvents || tagEvents {
		events = append(events, "repo:push")
	}
	url := fmt.Sprintf("%s/webhook/%d/%d", os.Getenv("BASE_URL"), environment.ID, userId)
	args := &bitbucket.WebhooksOptions{
		Owner:    environment.GitRepository.Owner,
		RepoSlug: environment.GitRepository.ID,
		Events:   events,
		Url:      url,
		Active:   true,
	}
	Webhookscreated, err := g.gitUser.Repositories.Webhooks.Create(args)
	if err != nil {
		log.Warnf("error occured on webhook :%v", err)
	}
	hookid := Webhookscreated.Uuid
	webhook := Hook{
		ID: hookid,
	}
	return &webhook, &http.Response{}, nil
}

func (g BitBucket) UpdateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig, hookId string) (*Hook, *http.Response, error) {
	if strings.TrimSpace(ciConfig.Events) == "" {
		delargs := &bitbucket.WebhooksOptions{
			Owner:    environment.GitRepository.Owner,
			RepoSlug: environment.GitRepository.ID,
			Uuid:     hookId,
		}
		_, err := g.gitUser.Repositories.Webhooks.Delete(delargs)
		return nil, &http.Response{}, err
	}
	var events []string
	pushEvents := strings.Contains(ciConfig.Events, "push")
	tagEvents := strings.Contains(ciConfig.Events, "release")
	if pushEvents || tagEvents {
		events = append(events, "repo:push")
	}
	url := fmt.Sprintf("%s/webhook/%d/%d", os.Getenv("BASE_URL"), environment.ID, userId)
	updateargs := &bitbucket.WebhooksOptions{
		Owner:    environment.GitRepository.Owner,
		RepoSlug: environment.GitRepository.ID,
		Uuid:     hookId,
		Events:   events,
		Url:      url,
		Active:   true,
	}
	Webhookupdated, err := g.gitUser.Repositories.Webhooks.Update(updateargs)
	if err != nil {
		return nil, nil, err
	}
	hookid := Webhookupdated.Uuid
	webhook := Hook{
		ID: hookid,
	}
	return &webhook, &http.Response{}, nil
}

func (g BitBucket) DeleteWebhook(environment *models.Environment, hookId string) error {
	delargs := &bitbucket.WebhooksOptions{
		Owner:    environment.GitRepository.Owner,
		RepoSlug: environment.GitRepository.ID,
		Uuid:     hookId,
	}
	_, err := g.gitUser.Repositories.Webhooks.Delete(delargs)
	return err
}

func (g BitBucket) GetWorkflowContent(environment *models.Environment) map[string]interface{} {
	workflows := map[string]interface{}{}
	files, err := g.gitUser.Repositories.Repository.ListFiles(&bitbucket.RepositoryFilesOptions{
		Owner:    environment.GitRepository.Owner,
		RepoSlug: environment.GitRepository.Name,
		Ref:      environment.GitBranch,
		Path:     pipelineDir,
	})
	if err != nil {
		return workflows
	}
	for _, file := range files {
		content, err := g.gitUser.Repositories.Repository.GetFileContent(&bitbucket.RepositoryFilesOptions{
			Owner:    environment.GitRepository.Owner,
			RepoSlug: environment.GitRepository.Name,
			Ref:      environment.GitBranch,
			Path:     file.String(),
		})
		if err == nil {
			workflows[file.String()] = string(content)
		}
	}
	return workflows
}

func (g BitBucket) GetDockerfile(environment *models.Environment) ([]byte, error) {
	var dockerfile []byte
	files, err := g.gitUser.Repositories.Repository.ListFiles(&bitbucket.RepositoryFilesOptions{
		Owner:    environment.GitRepository.Owner,
		RepoSlug: environment.GitRepository.Name,
		Ref:      environment.GitBranch,
		Path:     getSubDir(environment),
	})
	for _, file := range files {
		if file.String() == dockerFile {
			content, err := g.gitUser.Repositories.Repository.GetFileContent(&bitbucket.RepositoryFilesOptions{
				Owner:    environment.GitRepository.Owner,
				RepoSlug: environment.GitRepository.Name,
				Ref:      environment.GitBranch,
				Path:     file.String(),
			})
			if err == nil {
				dockerfile = content
			}
		}
	}
	return dockerfile, err
}

func BitBucketHookPayload(payload map[string]interface{}, repo models.GitRepo) (string, string, error) {
	preponame := payload["repository"].(map[string]interface{})["full_name"].(string)
	reponame := fmt.Sprintf("%s/%s", repo.Owner, repo.ID)
	if preponame != reponame {
		return "", "", errors.New("invalid payload")
	}
	changes := payload["push"].(map[string]interface{})["changes"]
	result := changes.([]interface{})
	var tag, branch string
	for _, j := range result {
		commit := j.(map[string]interface{})["commits"]
		branchcode := j.(map[string]interface{})["new"]
		tag = commit.([]interface{})[0].(map[string]interface{})["hash"].(string)
		branch = branchcode.(map[string]interface{})["name"].(string)
	}
	return tag, branch, nil
}

func (g BitBucket) GetWorkSpaceAndRslug(repoId string) (string, string, error) {
	var owner, rslug string
	// usr, err := g.gitUser.User.Profile()
	// if err != nil {
	// 	return "", "", err
	// }
	if strings.Contains(repoId, "/") {
		workspace := strings.Split(repoId, "/")
		owner = workspace[0]
		rslug = workspace[1]
	} else {
		owner = repoId
		rslug = repoId
	}
	return owner, rslug, nil
}

func (g BitBucket) GitExpiredToken(gitUser models.GitUser, DB *gorm.DB) (GitInterface, error) {
	conf := &oauth2.Config{
		ClientID:     os.Getenv("BITBUCKET_CLIENT_ID"),
		ClientSecret: os.Getenv("BITBUCKET_CLIENT_SECRET"),
		Endpoint:     bitbucket_endpoint.Endpoint,
	}
	oauthToken := oauth2.Token{}
	databyte, err := json.Marshal(gitUser.Token)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(databyte, &oauthToken)
	if err != nil {
		return nil, err
	}
	tokenSource := conf.TokenSource(context.Background(), &oauthToken)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, err
	}
	accessToken := oauthToken.AccessToken
	if newToken.AccessToken != oauthToken.AccessToken {
		data, err := json.Marshal(newToken)
		if err != nil {
			return nil, err
		}
		var token *postgres.Jsonb
		err = json.Unmarshal(data, &token)
		if err != nil {
			return nil, err
		}
		gitUser.Token = token
		gitUser.AccessToken = newToken.AccessToken
		_, err = gitUserInterface.Update(DB, &gitUser)
		if err != nil {
			return nil, err
		}
		accessToken = newToken.AccessToken
	}
	gitClient := bitbucket.NewOAuthbearerToken(accessToken)
	return &BitBucket{gitUser: gitClient}, nil
}
func (g BitBucket) Profile() error {
	_, err := g.gitUser.User.Profile()
	if err != nil {
		return err
	}
	return nil
}
