package controllers

import (
	"01cloud-api/api/auth"
	"01cloud-api/api/models/doc"

	log "github.com/sirupsen/logrus"

	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"01cloud-api/api/models"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/formaterror"
	"01cloud-api/api/utils/helper"

	"github.com/docker/docker/pkg/archive"
	"github.com/gorilla/mux"
)

var pluginInterface = models.NewPlugin()
var pluginVersionInterface = models.NewPluginVersion()

// CreatePlugin godoc
// @Summary Create a new plugin
// @Description Create a new plugin with the input payload
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.Plugin true "Create Plugin"
// @Success 201 {object} doc.Plugin
// @Router /plugin [post]
func (server *Server) CreatePlugin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := models.Plugin{}
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
	add_ons := &struct {
		AddOns string `json:"add_ons"`
	}{}
	err = json.Unmarshal(body, &add_ons)
	if err == nil {
		var adonArray []string
		if len(add_ons.AddOns) > 0 {
			adonArray = strings.Split(add_ons.AddOns, ",")
		}
		plugins, err := pluginInterface.FindPluginByIds(server.DB, adonArray)
		if err == nil {
			data.AddOns = plugins
		}
	}
	dataCreated, err := pluginInterface.Save(server.DB, &data)
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
	pluginResponse, err := CreatePluginResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    pluginResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetPlugins godoc
// @Summary Get Plugins
// @Description Get list plugins from token
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param page query int true "Page"
// @Param size query int true "Size"
// @Param search query string true "Search"
// @Param sort-column query string true "Sort Column"
// @Param sort-direction query string true "Sort Direction"
// @Param support_ci query bool true "Support CI"
// @Success 200 {array} doc.Plugin
// @Router /plugins [get]
func (server *Server) GetPlugins(w http.ResponseWriter, r *http.Request) {
	_, oid, err := auth.ExtractTokenID(r)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	page, _ := strconv.ParseUint(r.FormValue("page"), 10, 32)
	size, _ := strconv.ParseUint(r.FormValue("size"), 10, 32)
	search := r.FormValue("search")
	sortColumn := r.FormValue("sort-column")
	sortDirection := r.FormValue("sort-direction")
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if sortColumn == "" {
		sortColumn = "id"
	}
	if sortDirection == "" {
		sortDirection = "desc"
	}
	supportCi, err := strconv.ParseBool(r.FormValue("support_ci"))
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	isManagedService := false
	if r.FormValue("is_managed_service") != "" {
		isManagedService, err = strconv.ParseBool(r.FormValue("is_managed_service"))
		if err != nil {
			responses.ERROR(w, http.StatusBadRequest, err)
			return
		}
	}
	datas, count, err := pluginInterface.FindBySupportCi(server.DB, supportCi, isManagedService, oid, page, size, search, sortColumn, sortDirection)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	dataresponse := []models.Plugin{}
	for _, d := range *datas {
		if d.Name != "helm chart" && d.Name != "docker" {
			dataresponse = append(dataresponse, d)
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataresponse)
		return
	}
	pluginResponseList := []doc.Plugin{}
	for _, item := range dataresponse {
		pluginResponse, err := CreatePluginResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		pluginResponseList = append(pluginResponseList, *pluginResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginResponseList,
		Success: 1,
		Message: "Success",
		Count:   count,
		Page:    int(page),
		Size:    int(size),
	})
}

// GetAllActivePlugins godoc
// @Summary Get All Active Plugins
// @Description Get list of All Active plugins from token
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param page query int true "Page"
// @Param size query int true "Size"
// @Param search query string true "Search"
// @Param sort-column query string true "Sort Column"
// @Param sort-direction query string true "Sort Direction"
// @Success 200 {array} doc.Plugin
// @Router /plugins/active [get]
func (server *Server) GetAllActivePlugins(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseUint(r.FormValue("page"), 10, 32)
	size, _ := strconv.ParseUint(r.FormValue("size"), 10, 32)
	search := r.FormValue("search")
	sortColumn := r.FormValue("sort-column")
	sortDirection := r.FormValue("sort-direction")
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	if sortColumn == "" {
		sortColumn = "id"
	}
	if sortDirection == "" {
		sortDirection = "desc"
	}
	key := "plugins-active"
	var value interface{}
	if ok := server.Cache.Get(key, &value); ok {
		responses.JSON(w, http.StatusOK, value)
		return
	}
	datas, count, err := pluginInterface.FindAllActiveWithFilters(server.DB, page, size, search, sortColumn, sortDirection)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	dataresponse := []models.Plugin{}
	for _, d := range *datas {
		if d.Name != "helm chart" && d.Name != "docker" {
			dataresponse = append(dataresponse, d)
		}
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataresponse)
		return
	}
	pluginResponseList := []doc.Plugin{}
	for _, item := range dataresponse {
		pluginResponse, err := CreatePluginResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		pluginResponseList = append(pluginResponseList, *pluginResponse)
	}
	response := responses.Response{
		Data:    pluginResponseList,
		Success: 1,
		Message: "Success",
		Count:   count,
		Page:    int(page),
		Size:    int(size),
	}
	server.Cache.Set(key, response)
	responses.JSON(w, http.StatusOK, response)
}

// GetAllPlugins godoc
// @Summary Get All Plugins
// @Description Get list of All plugins from token
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param page query int true "Page"
// @Param size query int true "Size"
// @Param search query string true "Search"
// @Param sort-column query string true "Sort Column"
// @Param sort-direction query string true "Sort Direction"
// @Success 200 {array} doc.Plugin
// @Router /admin/plugins [get]
func (server *Server) GetAllPlugins(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.ParseUint(r.FormValue("page"), 10, 32)
	size, _ := strconv.ParseUint(r.FormValue("size"), 10, 32)
	search := r.FormValue("search")
	sortColumn := r.FormValue("sort-column")
	sortDirection := r.FormValue("sort-direction")
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 100
	}
	if sortColumn == "" {
		sortColumn = "id"
	}
	if sortDirection == "" {
		sortDirection = "desc"
	}
	datas, count, err := pluginInterface.FindAllWithFilters(server.DB, page, size, search, sortColumn, sortDirection)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, &map[string]interface{}{
			"count": count,
			"data":  datas,
		})
		return
	}
	pluginResponseList := []doc.Plugin{}
	for _, item := range *datas {
		pluginResponse, err := CreatePluginResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		pluginResponseList = append(pluginResponseList, *pluginResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginResponseList,
		Success: 1,
		Message: "Success",
		Count:   count,
		Page:    int(page),
		Size:    int(size),
	})
}

// GetPlugin godoc
// @Summary Get Plugin by id
// @Description Get Plugin by id from token
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 200 {object} doc.Plugin
// @Router /plugin/{id} [get]
func (server *Server) GetPlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := pluginInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	pluginResponse, err := CreatePluginResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginResponse,
		Success: 1,
		Message: "Success",
	})
}

// GetAddOns godoc
// @Summary Get addons by plugin id
// @Description Get Plugin by id from token
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 200 {array} doc.Plugin
// @Router /plugin/{id}/addons [get]
func (server *Server) GetAddOns(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return

	}
	catid := r.FormValue("cat_id")
	query := r.FormValue("query")
	var catids []string
	if catid != "" {
		catids = strings.Split(catid, ",")
	}
	dataReceived, err := pluginInterface.FindAddOns(server.DB, pid, query, catids)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	pluginResponseList := []doc.Plugin{}
	for _, item := range dataReceived {
		pluginResponse, err := CreatePluginResponse(item)
		if err != nil {
			log.Error(err)
			continue
		}
		pluginResponseList = append(pluginResponseList, *pluginResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginResponseList,
		Success: 1,
		Message: "Success",
	})
}

// UpdatePlugin godoc
// @Summary Update a Plugin
// @Description Update a Plugin with the input payload
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Param body body doc.Plugin true "Update Plugin"
// @Success 200 {object} doc.Plugin
// @Router /plugin/{id} [put]
func (server *Server) UpdatePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Plugin{}
	err = server.DB.Model(models.Plugin{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("plugin not found"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate := models.Plugin{}
	err = json.Unmarshal(body, &dataUpdate)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	//dataUpdate.Prepare()
	err = dataUpdate.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	dataUpdate.ID = data.ID

	add_ons := &struct {
		AddOns string `json:"add_ons"`
	}{}
	err = json.Unmarshal(body, &add_ons)
	if err == nil {
		var adonArray []string
		if len(add_ons.AddOns) > 0 {
			adonArray = strings.Split(add_ons.AddOns, ",")
		}
		plugins, err := pluginInterface.FindPluginByIds(server.DB, adonArray)
		if err == nil {
			err = server.DB.Model(&data).Association("AddOns").Clear().Error
			if err != nil {
				log.Error(err)
			}
			err = server.DB.Model(&data).Association("AddOns").Append(plugins).Error
			if err != nil {
				log.Error(err)
			}
			dataUpdate.AddOns = plugins
		}
	}
	categories := &struct {
		Categories string `json:"Categories"`
	}{}
	err = json.Unmarshal(body, categories)
	if err == nil {
		var categoriesArray []string
		if len(categories.Categories) > 0 {

			categoriesArray = strings.Split(categories.Categories, ",")
		}
		dbCats, err := pluginCategoryInterface.FindCategoryByIds(server.DB, categoriesArray)
		if err == nil {
			err = server.DB.Model(&data).Association("Categories").Clear().Error
			if err != nil {
				log.Error(err)
			}
			err = server.DB.Model(&data).Association("Categories").Append(dbCats).Error
			if err != nil {
				log.Error(err)
			}
		}
	}
	dataUpdated, err := pluginInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	pluginResponse, err := CreatePluginResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeletePlugin godoc
// @Summary Delete a Plugin
// @Description Delete a Plugin with the input payload
// @Tags Plugin
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 204 {object} doc.Plugin
// @Router /plugin/{id} [delete]
func (server *Server) DeletePlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.Plugin{}
	err = server.DB.Model(models.Plugin{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("Unauthorized"))
		return
	}
	_, err = pluginInterface.Delete(server.DB, pid)
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

// //////////////
// CreatePluginVersion godoc
// @Summary Create a new plugin-version
// @Description Create a new plugin-version with the input paylod
// @Tags PluginVersion
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Security ApiKeyAuth
// @Param body body doc.PluginVersion true "Create PluginVersion"
// @Success 201 {object} doc.PluginVersion
// @Router /plugin-version [post]
func (server *Server) CreatePluginVersion(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.PluginVersion{}
	pluginId, err := strconv.ParseInt(r.FormValue("plugin_id"), 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data.PluginID = uint64(pluginId)
	data.Version = r.FormValue("version")
	data.ChangeLogs = r.FormValue("change_logs")
	data.Attributes = r.FormValue("attributes")
	data.ReleaseDate = time.Now()
	data.Active = true

	upgradableVal := r.FormValue("upgradable")
	upgradable, err := strconv.ParseBool(upgradableVal)
	if err != nil {
		data.Upgradable = upgradable
	}

	////////// start foreignkey
	dataFor, err := pluginInterface.Find(server.DB, data.PluginID)
	println(err)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("invalid plugin"))
		return
	}

	data.Plugin = dataFor
	pluginPath, err := uploadPluginFile(r, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data.Url = pluginPath
	err = data.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	///////// end foreignkey
	dataCreated, err := pluginVersionInterface.Save(server.DB, &data)
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
	pluginVersionResponse, err := CreatePluginVersionResponse(dataCreated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    pluginVersionResponse,
		Success: 1,
		Message: "Success",
	})
}

func uploadPluginFile(r *http.Request, pluginVersion *models.PluginVersion) (string, error) {

	tarFile, _, err := r.FormFile("package")
	if err != nil {
		return "", errors.New("plugin package file is required")
	}

	path := fmt.Sprintf("/data/plugin/%s/%s", pluginVersion.Plugin.Name, pluginVersion.Version)
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0777)
		if err != nil {
			return "", err
		}
	}

	err = archive.Untar(tarFile, path, &archive.TarOptions{})
	if err != nil {
		log.Error(err)
		return "", errors.New("invalid tar file ")
	}
	return path, nil
}

// GetPluginVersions godoc
// @Summary Get PluginVersions
// @Description Get list plugin versions from token
// @Tags PluginVersion
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.PluginVersion
// @Router /plugin-versions [get]
func (server *Server) GetPluginVersions(w http.ResponseWriter, r *http.Request) {
	datas, err := pluginVersionInterface.FindAll(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	pluginVersionList := []doc.PluginVersion{}
	for _, item := range *datas {
		pluginResponse, err := CreatePluginVersionResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		pluginVersionList = append(pluginVersionList, *pluginResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginVersionList,
		Success: 1,
		Message: "Success",
	})
}

// GetPluginVersions godoc
// @Summary Get PluginVersions
// @Description Get all plugin versions
// @Tags PluginVersion
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Success 200 {array} doc.PluginVersion
// @Router /admin/plugin-versions [get]
func (server *Server) GetPluginVersionsForAdmin(w http.ResponseWriter, r *http.Request) {
	datas, err := pluginVersionInterface.FindAllWithInactive(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	pluginVersionList := []doc.PluginVersion{}
	for _, item := range *datas {
		pluginResponse, err := CreatePluginVersionResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		pluginVersionList = append(pluginVersionList, *pluginResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginVersionList,
		Success: 1,
		Message: "Success",
	})
}

// GetProject godoc
// @Summary GetPluginVersionByPluginId by id
// @Description Get PluginVersion by plugin id from token
// @Tags PluginVersion
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 200 {array} doc.PluginVersion
// @Router /plugin/{id}/versions [get]
func (server *Server) GetPluginVersionByPluginId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	datas, err := pluginVersionInterface.FindAllByPlugin(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, datas)
		return
	}
	pluginVersionList := []doc.PluginVersion{}
	for _, item := range *datas {
		pluginResponse, err := CreatePluginVersionResponse(&item)
		if err != nil {
			log.Error(err)
			continue
		}
		pluginVersionList = append(pluginVersionList, *pluginResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginVersionList,
		Success: 1,
		Message: "Success",
	})

}

// GetProject godoc
// @Summary GetLatestPluginVersionByPluginId by id
// @Description Get latest plugin by plugin id from token
// @Tags PluginVersion
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 200 {object} doc.PluginVersion
// @Router /plugin/{id}/latest-version [get]
func (server *Server) GetLatestPluginVersionByPluginId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	pluginVersion, err := pluginVersionInterface.FindLatestPluginVersion(server.DB, pid)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	file, _ := os.ReadFile(pluginVersion.Url + "/versions.json")
	var res []map[string]interface{}
	err = json.Unmarshal(file, &res)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	pluginVersion.Versions = res
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, pluginVersion)
		return
	}
	pluginVersionResponse, err := CreatePluginVersionResponse(pluginVersion)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginVersionResponse,
		Success: 1,
		Message: "Success",
	})

}

// GetPlugin godoc
// @Summary Get Latest Plugin Version Config by id
// @Description Get latest plugin version by plugin id from token
// @Tags PluginVersion
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 200 {object} doc.ObjectResponse
// @Router /plugin/{id}/config [get]
func (server *Server) GetPluginVersionConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := pluginVersionInterface.Find(server.DB, pid)

	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	file, _ := os.ReadFile(dataReceived.Url + "/values.schema.json")
	res := make(map[string]interface{})
	err = json.Unmarshal([]byte(file), &res)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, res)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "success",
		Data:    res,
	})
}

// GetPluginVersion godoc
// @Summary Get PluginVersion by id
// @Description Get PluginVersion by id from token
// @Tags PluginVersion
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "PluginVersion id"
// @Success 200 {object} doc.PluginVersion
// @Router /plugin-version/{id} [get]
func (server *Server) GetPluginVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := pluginVersionInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataReceived)
		return
	}
	pluginVersionResponse, err := CreatePluginVersionResponse(dataReceived)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginVersionResponse,
		Success: 1,
		Message: "Success",
	})
}

// UpdatePluginVersion godoc
// @Summary Update a PluginVersion
// @Description Update a PluginVersion with the input payload
// @Tags PluginVersion
// @Accept  x-www-form-urlencoded
// @Produce  json
// @Security ApiKeyAuth
// @Param package formData file true  "This is Update Plugin file"
// @Param id path int true "PluginVersion id"
// @Param body body doc.PluginVersion true "Update PluginVersion"
// @Success 200 {object} doc.PluginVersion
// @Router /plugin-version/{id} [put]
func (server *Server) UpdatePluginVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.PluginVersion{}
	err = server.DB.Model(models.PluginVersion{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("pluginVersion not found"))
		return
	}
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataUpdate := models.PluginVersion{}
	pluginId, err := strconv.ParseInt(r.FormValue("plugin_id"), 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataUpdate.PluginID = uint64(pluginId)
	dataUpdate.Version = r.FormValue("version")
	dataUpdate.ChangeLogs = r.FormValue("change_logs")
	dataUpdate.Attributes = r.FormValue("attributes")
	dataUpdate.ReleaseDate = time.Now()
	active, err := strconv.ParseBool(r.FormValue("active"))
	if err != nil {
		active = data.Active
	}
	dataUpdate.Active = active

	upgradableVal := r.FormValue("upgradable")
	upgradable, err := strconv.ParseBool(upgradableVal)
	if err != nil {
		data.Upgradable = upgradable
	}

	////////// start foreignkey
	dataFor, err := pluginInterface.Find(server.DB, data.PluginID)
	println(err)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("invalid plugin"))
		return
	}

	dataUpdate.Plugin = dataFor
	_, _, e := r.FormFile("package")
	if e == nil {
		pluginPath, err := uploadPluginFile(r, &dataUpdate)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
		dataUpdate.Url = pluginPath
	}
	//err = dataUpdate.Validate()
	//if err != nil {
	//	responses.ERROR(w, http.StatusUnprocessableEntity, err)
	//	return
	//}
	///////// end foreignkey
	dataUpdate.ID = data.ID
	dataUpdated, err := pluginVersionInterface.Update(server.DB, &dataUpdate)
	if err != nil {
		formattedError := formaterror.FormatError(err.Error())
		responses.ERROR(w, http.StatusInternalServerError, formattedError)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, dataUpdated)
		return
	}
	pluginVersionResponse, err := CreatePluginVersionResponse(dataUpdated)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    pluginVersionResponse,
		Success: 1,
		Message: "Success",
	})
}

// DeletePluginVersion godoc
// @Summary Delete a PluginVersion
// @Description Delete a PluginVersion with the input payload
// @Tags PluginVersion
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "PluginVersion id"
// @Success 204 {object} doc.PluginVersion
// @Router /plugin-version/{id} [delete]
func (server *Server) DeletePluginVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	data := models.PluginVersion{}
	err = server.DB.Model(models.PluginVersion{}).Where("id = ?", pid).Take(&data).Error

	if err != nil {
		responses.ERROR(w, http.StatusNotFound, errors.New("not found"))
		return
	}

	_, err = pluginVersionInterface.Delete(server.DB, pid)
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

// GetPlugin godoc
// @Summary Get Latest Plugin Version Setting by id
// @Description Get latest plugin version setting by plugin id from token
// @Tags PluginVersion
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Plugin id"
// @Success 200 {object} doc.ObjectResponse
// @Router /plugin/{id}/settings [get]
func (server *Server) GetPluginVersionSetting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["id"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	dataReceived, err := pluginVersionInterface.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	file, _ := os.ReadFile(dataReceived.Url + "/config.json")
	res := make(map[string]interface{})
	err = json.Unmarshal([]byte(file), &res)
	if err != nil {
		log.Error("unmarshall err :: ", err)
	}
	data := map[string]interface{}{
		"setting": map[string]interface{}{
			"properties": map[string]interface{}{},
		},
	}
	if setting, ok := res["setting"]; ok {
		data["setting"] = setting.(map[string]interface{})
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, res)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "success",
		Data:    data,
	})
}

func CreatePluginResponse(plugin *models.Plugin) (*doc.Plugin, error) {
	pluginResponse := doc.Plugin{}
	pluginBytes, _ := json.Marshal(plugin)
	err := json.Unmarshal(pluginBytes, &pluginResponse)
	if err != nil {
		return nil, err
	}
	return &pluginResponse, nil

}

func CreatePluginVersionResponse(pluginVersion *models.PluginVersion) (*doc.PluginVersion, error) {
	pluginVersionRes := doc.PluginVersion{}
	pluginVersionBytes, _ := json.Marshal(pluginVersion)
	err := json.Unmarshal(pluginVersionBytes, &pluginVersionRes)
	if err != nil {
		return nil, err
	}
	return &pluginVersionRes, nil

}
