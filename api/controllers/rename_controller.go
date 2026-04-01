package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Rename godoc
// @Summary Rename org, project, app and env
// @Description rename
// @Tags Rename
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param model path string true "Model"
// @Param body body doc.RenameRequest true "Rename Model"
// @Success 200  {object} doc.SuccessResponse
// @Router /rename/{model}/{id} [put]
func (server *Server) RenameModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	model := vars["model"]
	id, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	uid, _, _ := auth.ExtractTokenID(r)
	_, err = userInterface.FindUserByID(server.DB, uid)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("user not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	input := models.RenameRequest{}
	err = json.Unmarshal(body, &input)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if model == "" || input.Name == "" {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("required model and name field"))
		return
	}
	err = server.Rename(id, uid, model, input.Name)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) Rename(id uint64, userID uint, model, name string) error {
	switch model {
	case "application":
		return server.RenameApp(id, name)
	case "environment":
		return server.RenameEnvironment(id, name)
	case "project":
		return server.RenameProject(id, userID, name)
	default:
		return errors.New("invalid rename model")
	}
}

func (server *Server) RenameProject(id uint64, userID uint, name string) error {
	var iproject = models.NewIProject()
	proj, err := iproject.Find(server.DB, id)
	if err != nil {
		return err
	}

	exists := iproject.IsNameExists(server.DB, userID, name, proj)
	if exists {
		return errors.New("name already exists")
	}

	return iproject.Rename(server.DB, name, proj)
}

func (server *Server) RenameApp(id uint64, name string) error {
	app, err := applicationInterface.Find(server.DB, id)
	if err != nil {
		return err
	}
	if name != app.Name && applicationInterface.IsNameExists(server.DB, app.ProjectID, name) {
		return errors.New("name already exists")
	}
	err = applicationInterface.RenameApp(server.DB, app.ID, name)
	if err != nil {
		return err
	}
	return nil
}

func (server *Server) RenameEnvironment(id uint64, name string) error {
	data := models.Environment{}
	env, err := data.FindWithProApp(server.DB, id)
	if err != nil {
		return err
	}
	if name != "" {
		if name != env.Name && env.IsNameExists(server.DB, uint(env.ApplicationID), name) {
			return errors.New("name already exists")
		}
	}
	err = env.RenameEnvironment(server.DB, env.ID, name)
	if err != nil {
		return err
	}
	return nil
}
