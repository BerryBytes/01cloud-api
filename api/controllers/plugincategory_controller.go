package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models/doc"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/gorilla/mux"
)

var pluginCategoryInterface = models.NewPluginCategory()

// CreatePluginCategory godoc
// @Summary Get Create Plugin Category
// @Description Create Plugin Category
// @Tags PluginCategory
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.PluginCategory true "Create Plugin Category"
// @Success 201 {object} doc.PluginCategory
// @Router /plugin-category [post]
func (server *Server) CreatePluginCategory(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {

		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.PluginCategory{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.Prepare()
	err = data.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataCreated, err := pluginCategoryInterface.Save(server.DB, &data)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	w.Header().Set("Location", fmt.Sprintf("%s%s/%d", r.Host, r.URL.Path, dataCreated.ID))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, dataCreated)
		return
	}
	pluginCategoryResponse, err := CreatePluginCategoryResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    pluginCategoryResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetPluginCategorys godoc
// @Summary Get Plugin Categorys
// @Description Get Plugin Categorys
// @Tags PluginCategory
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param query query string true "Query"
// @Param is_add_on query string true "Is Add On"
// @Success 200 {array} doc.PluginCategory
// @Router /plugin-category [get]
func (server *Server) GetPluginCategorys(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {

		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	query := r.FormValue("query")
	isaddon := r.FormValue("is_add_on")
	dataList, err := pluginCategoryInterface.FindAll(server.DB, isaddon, query)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataList)
		return
	}
	pluginCategoryList := []doc.PluginCategory{}
	for _, item := range *dataList {
		pluginCategoryResponse, err := CreatePluginCategoryResponse(&item)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			continue

		}
		pluginCategoryList = append(pluginCategoryList, *pluginCategoryResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginCategoryList,
		Success: 1,
		Message: "Success",
	})
}

// GetPluginCategory godoc
// @Summary Get Plugin Category
// @Description Get Plugin Category
// @Tags PluginCategory
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Get Plugin Category"
// @Success 200 {object} doc.PluginCategory
// @Router /plugin-category/{pid} [get]
func (server *Server) GetPluginCategory(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {

		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := pluginCategoryInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, data)
		return
	}
	pluginCategoryResponse, err := CreatePluginCategoryResponse(data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginCategoryResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdatePluginCategory godoc
// @Summary Update Plugin Category
// @Description Update Plugin Category
// @Tags PluginCategory
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Get Plugin Category"
// @Param body body doc.PluginCategory true "Update Plugin Category"
// @Success 200 {object} doc.PluginCategory
// @Router /plugin-category/{pid} [put]
func (server *Server) UpdatePluginCategory(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {

		responses.ERROR(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := pluginCategoryInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("PluginCategory not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.PluginCategory{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate.ID = data.ID
	dataUpdated, err := pluginCategoryInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	pluginCategoryResponse, err := CreatePluginCategoryResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginCategoryResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeletePluginCategory godoc
// @Summary Delete Plugin Category
// @Description Delete Plugin Category
// @Tags PluginCategory
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Delete Plugin Category"
// @Success 200
// @Router /plugin-category/{pid} [delete]
func (server *Server) DeletePluginCategory(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data, err := pluginCategoryInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("unauthorized"))
		return
	}
	if len(data.Plugins) > 0 {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("plugins should be removed first to delete plugin category"))
		return
	}

	_, err = pluginCategoryInterface.Delete(server.DB, int64(pid))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", pid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

// AddPluginsToPluginCategory godoc
// @Summary Add Plugins To PluginCategory
// @Description Add Plugins To PluginCategory
// @Tags PluginCategory
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Plugin Category Id"
// @Param body body doc.Plugin true "Add Plugins to PluginCategory"
// @Success 200 {string} string
// @Router /plugin-category/{pid}/category [post]
func (server *Server) AddPluginsToPluginCategory(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	_, err = pluginCategoryInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("PluginCategory not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var plugin models.Plugin
	err = json.Unmarshal(body, &plugin)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dbPlugin, err := pluginInterface.Find(server.DB, uint64(plugin.ID))
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}

	err = pluginCategoryInterface.AddPlugins(server.DB, dbPlugin)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]string{
			"message": fmt.Sprintf("%s Added", dbPlugin.Name),
		})
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: fmt.Sprintf("%s Added", dbPlugin.Name),
	})
}

// DeletePluginFromPluginCategory godoc
// @Summary Delete Plugin From PluginCategory
// @Description Delete Plugin From PluginCategory
// @Tags PluginCategory
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param pid path int true "Plugin Id"
// @Param body body doc.Plugin true "Delete plugin"
// @Success 200
// @Router /plugin-category/{pid}/category [delete]
func (server *Server) DeletePluginFromPluginCategory(w http.ResponseWriter, r *http.Request) {
	_, _, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	_, err = pluginCategoryInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("PluginCategory not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	var plugin *models.Plugin
	err = json.Unmarshal(body, &plugin)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	plugin, err = pluginInterface.Find(server.DB, uint64(plugin.ID))
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = pluginCategoryInterface.DeletePlugins(server.DB, plugin)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	w.Header().Set("Entity", fmt.Sprintf("%d", pid))
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusNoContent, "")
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}

func CreatePluginCategoryResponse(pluginCategory *models.PluginCategory) (*doc.PluginCategory, error) {
	pluginCategoryRes := doc.PluginCategory{}
	pluginCategoryResBytes, _ := json.Marshal(pluginCategory)
	err := json.Unmarshal(pluginCategoryResBytes, &pluginCategoryRes)
	if err != nil {
		return nil, err
	}
	return &pluginCategoryRes, nil

}
