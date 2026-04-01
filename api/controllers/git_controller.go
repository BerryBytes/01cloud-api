package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/notifications"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/git"
	"01cloud-api/api/utils/helper"

	"01cloud-api/api/mailer"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var gitUserInterface = models.NewGitUser()

// ConnectToGit godoc
// @Summary connect to git
// @Description Connect to git
// @Tags GitRoutes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.ExternalLogin true "External Login"
// @Success 201 {object} doc.GitUser
// @Router /external/connect/git [post]
// @Router /git/connect [post]
func (server *Server) ConnectToGit(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	_, err = userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not founnd"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	externalLogin := models.ExternalLogin{}
	err = json.Unmarshal(body, &externalLogin)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = externalLogin.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	gitUser, err := git.GitLogin(externalLogin)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	gitUser.ServiceName = externalLogin.Service
	gitUser.UserID = uint64(uid)
	//	gtUsr := models.GitUser{}
	oldUser, _ := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), externalLogin.Service)
	if oldUser != nil {
		gitUser.ID = oldUser.ID
		oldUser, err = gitUserInterface.Update(server.DB, &gitUser)
	} else {
		oldUser, err = gitUserInterface.Save(server.DB, &gitUser)
	}
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uint(oldUser.UserID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	linkUnlinkEmail := mailer.SendLinkUnlinkEmail(user.Email, oldUser.ServiceName, true)
	err = notifications.NotifyEmail(server.NotifyClient, &linkUnlinkEmail)
	if err != nil {
		log.Error("notify email error :: ", err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"user": oldUser,
		})
		return
	}
	gitUserResponse, err := CreateGitUserResponse(oldUser)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"user": gitUserResponse,
		},
		Success: 1,
		Message: "Success",
	})
}

// GetGitServices godoc
// @Summary connect to git
// @Description Connect to git
// @Tags GitRoutes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User Id"
// @Success 200 {object} doc.GitUser
// @Router /external/connections [post]
// @Router /git/connections [post]
func (server *Server) GetGitServices(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	datas, err := gitUserInterface.FindByUserId(server.DB, uint64(uid))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	gitUserList := []doc.GitUser{}
	for _, item := range datas {
		gitUserResponse, err := CreateGitUserResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		gitUserList = append(gitUserList, *gitUserResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    gitUserList,
		Success: 1,
		Message: "Success",
	})
}

// RevokeGitToken godoc
// @Summary Revoke git token
// @Description Revoke git Token
// @Tags GitRoutes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "User Id"
// @Success 200 {object} doc.GitUser
// @Router /external/revoke/{id} [delete]
// @Router /git/revoke/{id} [delete]
func (server *Server) RevokeGitToken(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	oldUser, err := gitUserInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if oldUser.UserID != uint64(uid) {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("you are not authorized to revoke this token"))
		return
	}
	_, err = gitUserInterface.Delete(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	user, err := userInterface.FindUserByID(server.DB, uint(oldUser.UserID))
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("user not found"))
		return
	}
	linkUnlinkEmail := mailer.SendLinkUnlinkEmail(user.Email, oldUser.ServiceName, false)
	err = notifications.NotifyEmail(server.NotifyClient, &linkUnlinkEmail)
	if err != nil {
		log.Error("notify email error :: ", err)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, &map[string]interface{}{
			"message": "Token deleted successfully",
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})

}
func (server *Server) GetGithubClient(userId uint64, environment models.Environment) (*github.Client, error) {
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(userId), environment.Application.GitService)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gtUsr.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	cl := github.NewClient(tc)
	return cl, nil
}

func (server *Server) GetTag(uid uint, environment *models.Environment) (*models.Commit, *models.GitUser, error) {
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), environment.Application.GitService)
	if err != nil {
		return nil, nil, err
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			return nil, nil, err
		}
	}
	branch, err := gitI.GetLatestCommit(*environment)
	if err != nil {
		return nil, nil, err
	}
	return &branch, gtUsr, nil
}

func (server *Server) CreateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig) (*git.Hook, *http.Response, error) {
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(userId), environment.Application.GitService)
	if err != nil {
		return nil, nil, err
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			return nil, nil, err
		}
	}
	webhook, response, err := gitI.CreateWebhook(userId, environment, ciConfig)
	if err != nil {
		return nil, nil, err
	}
	return webhook, response, nil

}

func (server *Server) UpdateWebhook(userId uint64, environment models.Environment, ciConfig models.CiConfig, hookId string) (*git.Hook, *http.Response, error) {
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(userId), environment.Application.GitService)
	if err != nil {
		return nil, nil, err
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			return nil, nil, err
		}
	}
	webhook, response, err := gitI.UpdateWebhook(userId, environment, ciConfig, hookId)
	if err != nil {
		return nil, nil, err
	}
	return webhook, response, nil

}

func (server *Server) DeleteWebhook(environment *models.Environment, hookId string, r *http.Request) error {
	userId, _, err := auth.ExtractTokenID(r)
	if err != nil {
		return err
	}
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(userId), environment.Application.GitService)
	if err != nil {
		return err
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			return err
		}
	}
	err = gitI.DeleteWebhook(environment, hookId)
	return err

}

// GetRepos godoc
// @Summary Get Git Repos
// @Description Get Git Repository
// @Tags GitRoutes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param service path string true "Service"
// @Param org query string true "Organization Id"
// @Success 200 {array} doc.KVResponse
// @Router /git/repos [get]
// @Router /git/repo/{service} [get]
func (server *Server) GetRepos(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	serviceName := vars["service"]
	orgId := r.URL.Query().Get("org")
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("not authorized to github, please connect to github"))
		return
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
	}
	resData, err := gitI.GetRepos(orgId)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, resData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resData,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetGitUser(w http.ResponseWriter, r *http.Request) (*github.User, *github.Client, error) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return nil, nil, err
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return nil, nil, err
	}
	serviceMap := struct {
		Service string `json:"service_name"`
	}{}
	err = json.Unmarshal(body, &serviceMap)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return nil, nil, err
	}
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceMap.Service)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("not authorized to github, please connect to github"))
		return nil, nil, err
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gtUsr.AccessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	usr, _, err := client.Users.Get(ctx, "")
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return nil, nil, err
	}
	return usr, client, nil
}

// GetOrganizations godoc
// @Summary Get organizations
// @Description Get organizations
// @Tags GitRoutes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param service path string true "Service Name"
// @Success 200 {array} doc.KVResponse
// @Router /git/organizations [post]
// @Router /git/org/{service} [post]
func (server *Server) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	serviceName := vars["service"]
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("not authorized to github, please connect to github"))
		return
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
	}
	resData, err := gitI.GetOrganizations()
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, resData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resData,
		Success: 1,
		Message: "Success",
	})
}

// GetReposByOrganization godoc
// @Summary Get repos by organization
// @Description Get Repos by Organization
// @Tags GitRoutes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param name path string true "name"
// @Success 200 {array} map[string]string
// @Router /git/organization/{name} [post]
// @Router /git/repos/{name} [post]
func (server *Server) GetReposByOrganization(w http.ResponseWriter, r *http.Request) {
	_, client, err := server.GetGitUser(w, r)
	ctx := context.Background()
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to get user repos"))
		return
	}
	vars := mux.Vars(r)
	orgId := vars["name"]
	if orgId == "" {
		responses.ERROR(w, http.StatusNotFound, errors.New("organization is required"))
		return
	}
	repos, _, err := client.Repositories.ListByOrg(ctx, orgId, &github.RepositoryListByOrgOptions{ListOptions: github.ListOptions{PerPage: 1000}})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, errors.New("unable to get user repos"))
		return
	}
	resData := []map[string]string{}
	for _, d := range repos {
		resData = append(resData, map[string]string{
			"name":  *d.Name,
			"value": fmt.Sprint(*d.ID),
		})
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, resData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resData,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetRepository(uid uint, application *models.Application) (*models.GitRepo, error) {
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), application.GitService)
	if err != nil {
		log.Error("not connected to git provider :: ", err)
		return nil, errors.New("application not linked to any provider, please check your account link")
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			log.Error("invalid git token :: ", err)
			return nil, errors.New("invalid git token error")
		}
	}
	repo, err := gitI.GetRepoDetails(application.GitUrl)
	if err != nil {
		log.Error("failed to get repo details :: ", err)
		return nil, errors.New("failed to get repo details")
	}
	return &repo, nil
}

// GetBranches godoc
// @Summary Get Branches
// @Description Get Branches
// @Tags GitRoutes
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param service path string true "service"
// @Param repo query int true "Repo Id"
// @Param uid query int true "user Id"
// @Success 200 {object} doc.KVResponse
// @Router /git/branches [post]
// @Router /git/branch/{service} [post]
func (server *Server) GetBranches(w http.ResponseWriter, r *http.Request) {
	uid, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	serviceName := vars["service"]
	repoId := r.URL.Query().Get("repo")
	user := r.URL.Query().Get("uid")
	if user != "" {
		userid, err := strconv.ParseUint(user, 10, 64)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		uid = uint(userid)
	}
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(uid), serviceName)
	if err != nil {
		responses.ERROR(w, http.StatusForbidden, errors.New("not authorized to github, please connect to github"))
		return
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			responses.ERROR(w, http.StatusUnauthorized, err)
			return
		}
	}
	resData, err := gitI.GetBranches(repoId)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, resData)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    resData,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) GetDockerfile(env *models.Environment) ([]byte, error) {
	var dockerfile []byte
	gtUsr, err := gitUserInterface.FindByUserIdAndService(server.DB, uint64(env.Application.OwnerId), env.Application.GitService)
	if err != nil {
		return dockerfile, err
	}
	gitI, err := git.NewGit(*gtUsr)
	if err != nil {
		gitI, err = git.ExpiredToken(*gtUsr, server.DB)
		if err != nil {
			return dockerfile, err
		}
	}
	dockerfile, err = gitI.GetDockerfile(env)
	if err != nil {
		return dockerfile, err
	}
	if len(dockerfile) == 0 {
		dockerfilePath := fmt.Sprintf("%s/%s", env.PluginVersion.Url, constants.Dockerfile)
		dockerfile, err = os.ReadFile(dockerfilePath)
		if err != nil {
			return dockerfile, err
		}
	}
	return dockerfile, err
}

func CreateGitUserResponse(gitUser *models.GitUser) (*doc.GitUser, error) {
	gitUserResponse := doc.GitUser{}
	gitUserBytes, _ := json.Marshal(gitUser)
	err := json.Unmarshal(gitUserBytes, &gitUserResponse)
	if err != nil {
		return nil, err
	}
	return &gitUserResponse, nil
}
