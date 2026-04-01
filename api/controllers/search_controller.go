package controllers

import (
	"net/http"

	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/helper"

	log "github.com/sirupsen/logrus"
)

func (server *Server) Search(w http.ResponseWriter, r *http.Request) {
	uid, _, _ := auth.ExtractTokenID(r)
	query := r.FormValue("query")
	projects, _ := iproject.SearchProject(server.DB, uid, query)
	apps, _ := applicationInterface.SearchApplication(server.DB, uid, query)
	env := models.Environment{}
	envs, _ := env.SearchEnvironment(server.DB, uid, query)
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"projects":     projects,
			"applications": apps,
			"environments": envs,
		})
		return
	}
	projectsResponseList := []doc.Project{}
	for _, item := range projects {
		projectsResponse, err := CreateProjectResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		projectsResponseList = append(projectsResponseList, *projectsResponse)
	}
	applicationResponseList := []doc.Application{}
	for _, application := range apps {
		applicationResponse, err := CreateApplicationResponse(application)
		if err != nil {
			log.Error(err)
			continue
		}
		applicationResponseList = append(applicationResponseList, *applicationResponse)
	}
	environmentResponseList := []doc.Environment{}
	for _, environment := range envs {
		environmentResponse, err := CreateEnvironmentResponse(environment)
		if err != nil {
			log.Error(err)
			continue
		}
		environmentResponseList = append(environmentResponseList, *environmentResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: &map[string]interface{}{
			"projects":     projectsResponseList,
			"applications": applicationResponseList,
			"environments": environmentResponseList,
		},
		Success: 1,
		Message: "Success",
	})
}

// SearchUser godoc
// @Summary Autocompletion on user
// @Description Create autocompletion features on user search
// @Tags Search
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param query query string true "Search string"
// @Success 200 {object} doc.User
// @Router /search/user [get]
func (server *Server) SearchUser(w http.ResponseWriter, r *http.Request) {
	uid, oid, _ := auth.ExtractTokenID(r)
	query := r.FormValue("query")

	user := models.User{}
	users, _ := user.SearchUserName(server.DB, uid, oid, query)

	responses.JSON(w, http.StatusOK, &map[string]interface{}{
		"users": users,
	})
}
