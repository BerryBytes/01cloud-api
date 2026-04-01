package controllers

import (
	"01cloud-api/api/responses"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (server *Server) GetByNameController(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	model := vars["model"]
	name := vars["name"]
	v := r.URL.Query()
	app := v.Get("app")
	app_id := v.Get("appid")
	org_id := v.Get("orgid")
	proj := v.Get("proj")
	proj_id := v.Get("projid")
	pass := map[string]string{
		"org_id":  org_id,
		"proj":    proj,
		"proj_id": proj_id,
		"app_id":  app_id,
		"app":     app,
	}

	data, err := server.matchModel(model, name, pass)
	if err != nil {
		fmt.Println(err)
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}

	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    data,
		Success: 1,
		Message: "Success",
	})
}

func (server *Server) matchModel(model, name string, queries map[string]string) (interface{}, error) {
	switch model {
	case "project":
		var org_id uint64 = 0
		if queries["org_id"] != "" {
			id, err := strconv.Atoi(queries["org_id"])
			if err == nil {
				org_id = uint64(id)
			}
		}
		return iproject.GetProjectByName(server.DB, name, org_id)
	case "environment":
		app_id := server.ParseQueriesID(queries, "app_id")
		if app_id == 0 {
			proj_id := server.ParseQueriesID(queries, "proj_id")
			if proj_id == 0 {
				org_id := server.ParseQueriesID(queries, "org_id")
				proj_name := server.ParseQueries(queries, "proj")
				proj, err := iproject.GetProjectByName(server.DB, proj_name, org_id)
				if err == nil {
					proj_id = uint64(proj.ID)
				}
			}
			app_name := server.ParseQueries(queries, "app")
			app, err := applicationInterface.GetApplicationByName(server.DB, app_name, proj_id)
			if err == nil {
				app_id = uint64(app.ID)
			}
		}
		return environmentInterface.GetEnvironmentByName(server.DB, name, app_id)
	case "application":
		proj_id := server.ParseQueriesID(queries, "proj_id")
		if proj_id == 0 {
			org_id := server.ParseQueriesID(queries, "org_id")
			proj_name := server.ParseQueries(queries, "proj")
			proj, err := iproject.GetProjectByName(server.DB, proj_name, org_id)
			if err == nil {
				proj_id = uint64(proj.ID)
			}
		}

		return applicationInterface.GetApplicationByName(server.DB, name, proj_id)
	default:
		return nil, errors.New("no such model found")
	}
}

func (server *Server) ParseQueriesID(queries map[string]string, pickVal string) uint64 {
	var id uint64 = 0
	if queries[pickVal] != "" {
		val, err := strconv.Atoi(queries[pickVal])
		if err == nil {
			id = uint64(val)
		}
	}
	return id
}

func (server *Server) ParseQueries(queries map[string]string, pickVal string) string {
	return queries[pickVal]
}
