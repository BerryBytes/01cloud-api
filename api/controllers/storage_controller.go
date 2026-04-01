package controllers

import (
	"01cloud-api/api/models"
	"01cloud-api/api/models/doc"
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
	"01cloud-api/api/utils/prometheus"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var storageInterface = models.NewStorage()

// StorageFetch godoc
// @Summary trigger command for backend to fetch available storage
// @Description trigger command for backend to fetch available storage
// @Tags Storage
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} string
// @Router /environment/{id}/storage-fetch [get]
func (server *Server) StorageFetch(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, READ)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	err = queue.Publish(constants.FetchStorage, &map[string]interface{}{
		"environment": environment,
		"storage":     nil,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]interface{}{
			"message": "Fetching storage command",
			"storage": environment.Storage,
		})
		return
	}
	storageResponseList := []doc.Storage{}
	for _, item := range environment.Storage {
		storageResponse, err := CreateStorageResponse(item)
		if err != nil {
			log.Error(err)
			continue
		}
		storageResponseList = append(storageResponseList, *storageResponse)
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data:    storageResponseList,
		Success: 1,
		Message: "Fetching storage command",
	})

}

// CreateStorageEnvironment godoc
// @Summary trigger command for backend to create new storage
// @Description trigger command for backend to create new storage
// @Tags Storage
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} string
// @Router /environment/{id}/storage [post]
func (server *Server) CreateStorageEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, user, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	if environment.ServiceType < 1 {
		responses.ERROR(w, http.StatusForbidden, errors.New("not supported for this environment "))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	storage := &models.Storage{}
	err = json.Unmarshal(body, storage)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	storage.Prepare(nil)
	err = storage.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	if server.checkProjectStorageQuota(environment.Application.ProjectID, storage.Capacity, environment.Application.Project.Subscription.DiskSpace) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("storage capacity exceeds than subscription plan, Please upgrade subscription plan "))
		return
	}

	if checkPathDuplicate(environment.Storage, storage) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("path already mounted "))
		return
	}
	storage.UserID = user.ID
	storage.EnvironmentID = environment.ID

	// if the stroage type is default then set the access mode to ReadWriteOnce
	if storage.StorageType == nil || strings.EqualFold(*storage.StorageType, "default") {
		rwo := "ReadWriteOnce"
		storage.AccessModes = &rwo
	} else {
		storage.AccessModes = storage.StorageType
	}
	storage, err = storageInterface.SaveStorage(server.DB, *storage)
	if err != nil {
		log.Error(err)
		return
	}
	pName := "docker"
	if environment.ServiceType == 1 {
		pName = environment.Application.Plugin.Name
	}
	storageName := fmt.Sprintf("storage-%s-%d-binding", pName, storage.ID)
	storage.VolumeName = &storageName
	_, err = storageInterface.Update(server.DB, *storage)
	if err != nil {
		log.Error(err)
	}
	err = queue.Publish(constants.CreateStorage, &map[string]interface{}{
		"environment": environment,
		"storage":     storage,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]interface{}{
			"message": "Creating storage command",
			"storage": storage,
		})
		return
	}
	storageResponse, err := CreateStorageResponse(storage)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    storageResponse,
		Success: 1,
		Message: "Creating storage command",
	})
}

func checkPathDuplicate(listStorage []*models.Storage, storage *models.Storage) bool {
	if storage.MountPath == nil {
		return false
	}
	for _, s := range listStorage {
		if *s.MountPath == *storage.MountPath && s.ID != storage.ID {
			return true
		}
	}
	return false
}

func (server *Server) checkProjectStorageQuota(pid, capacity uint64, subscriptionCapacity uint32) bool {
	env := models.Environment{}
	usage, err := env.GetUsedResource(server.DB, pid, false)
	if err != nil {
		return false
	}
	if (capacity + usage.Disk) > uint64(subscriptionCapacity) {
		return true
	}
	return false
}

// UpdateStorageEnvironment godoc
// @Summary trigger command for backend to update storage
// @Description trigger command for backend to update storage
// @Tags Storage
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param sid path int true "Storage id"
// @Success 200 {object} string
// @Router /environment/{id}/storage/{sid} [put]
func (server *Server) UpdateStorageEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	vars := mux.Vars(r)
	sid, err := strconv.ParseUint(vars["sid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	storage, err := storageInterface.Find(server.DB, sid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	tempStorage := &models.Storage{}
	err = json.Unmarshal(body, tempStorage)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	tempStorage.ID = uint(sid)
	if server.checkProjectStorageQuota(environment.Application.ProjectID, tempStorage.Capacity-storage.Capacity, environment.Application.Project.Subscription.DiskSpace) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("storage capacity exceeds than subscription plan, Please upgrade subscription plan "))
		return
	}
	if environment.ServiceType > 0 && checkPathDuplicate(environment.Storage, tempStorage) {
		responses.ERROR(w, http.StatusUnprocessableEntity, errors.New("path already mounted "))
		return
	}
	storage.Prepare(tempStorage)
	err = storage.Validate()
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	storage, err = storageInterface.Update(server.DB, *storage)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	err = queue.Publish(constants.UpdateStorage, &map[string]interface{}{
		"environment": environment,
		"storage":     storage,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]interface{}{
			"message": "Updating storage command",
			"storage": storage,
		})
		return
	}
	storageResponse, err := CreateStorageResponse(storage)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    storageResponse,
		Success: 1,
		Message: "Updating storage command",
	})
}

// DeleteStorageEnvironment godoc
// @Summary trigger command for backend to delete storage
// @Description trigger command for backend to delete storage
// @Tags Storage
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Param sid path int true "Storage id"
// @Success 200 {object} string
// @Router /environment/{id}/storage/{sid} [delete]
func (server *Server) DeleteStorageEnvironment(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, ADMIN)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	if environment.ServiceType == 0 {
		responses.ERROR(w, http.StatusForbidden, errors.New("not supported for this environment "))
		return
	}
	vars := mux.Vars(r)
	sid, err := strconv.ParseUint(vars["sid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	storage, err := storageInterface.Find(server.DB, sid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	err = queue.Publish(constants.DeleteStorage, &map[string]interface{}{
		"environment": environment,
		"storage":     storage,
	})
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	storage, err = storageInterface.SaveStorage(server.DB, *storage)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	_, err = storageInterface.Delete(server.DB, sid)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusCreated, &map[string]interface{}{
			"message": "Delete storage command",
			"storage": storage,
		})
		return
	}
	storageResponse, err := CreateStorageResponse(storage)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return

	}
	responses.JSON(w, http.StatusCreated, responses.Response{
		Data:    storageResponse,
		Success: 1,
		Message: "Delete storage command",
	})
}

// ListStorageInEnv godoc
// @Summary Get List of storage available in that environment
// @Description Get List of storage available in that environment
// @Tags Storage
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param id path int true "Environment id"
// @Success 200 {object} object
// @Router /environment/{id}/storage [get]
func (server *Server) ListStorageInEnv(w http.ResponseWriter, r *http.Request) {
	environment, err := GetEnvironmentFromRequest(server.DB, r)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	_, err = CheckPermission(server.DB, r, READ, environment)
	if err != nil {
		responses.ERROR(w, http.StatusUnauthorized, err)
		return
	}
	storage, err := server.StoreClient.Storage().GetStorageState(helper.GetNamespace(environment))
	if err != nil {
		responses.JSON(w, http.StatusInternalServerError, err)
		return

	}
	insight := prometheus.GetAvailableSizeInPVC(environment.Application.Cluster, helper.GetNamespace(environment))

	storageArray := []*models.Storage{}
	for _, str := range environment.Storage {
		if vol, ok := insight[*str.VolumeName]; ok {
			usedStorage := uint64(vol)
			str.UsedStorage = &usedStorage
		}
		storageArray = append(storageArray, str)
	}
	if !helper.IsV2(r) {
		responses.JSON(w, http.StatusOK, map[string]interface{}{
			"storage":  storageArray,
			"manifest": storage,
		})
		return
	}
	storages := []doc.Storage{}
	for _, storageArr := range storageArray {
		storageData, err := CreateStorageResponse(storageArr)
		if err == nil {
			storages = append(storages, *storageData)
		}
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]interface{}{
			"storage":  storages,
			"manifest": storage,
		},
		Success: 1,
		Message: "Success",
	})
}
func CreateStorageResponse(storage *models.Storage) (*doc.Storage, error) {
	storageResponse := doc.Storage{}
	storageBytes, _ := json.Marshal(storage)
	err := json.Unmarshal(storageBytes, &storageResponse)
	if err != nil {
		return nil, err
	}
	return &storageResponse, nil
}
func getAvailableStorageInProject(server *Server, pid uint64, subscriptionCapacity uint32) (uint64, error) {
	env := models.Environment{}
	usage, err := env.GetUsedResource(server.DB, pid, false)
	if err != nil {
		return 0, err
	}
	if uint64(subscriptionCapacity) <= usage.Disk {
		return 0, nil
	}
	return uint64(subscriptionCapacity) - usage.Disk, nil
}

func (server *Server) GetAvailableStorageInProject(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pid, err := strconv.ParseUint(vars["pid"], 10, 64)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	project, err := iproject.Find(server.DB, pid)
	if err != nil {
		responses.ERROR(w, http.StatusNotFound, err)
		return
	}
	availableStorage, err := getAvailableStorageInProject(server, pid, project.Subscription.DiskSpace)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Data: map[string]uint64{
			"available_storage": availableStorage,
		},
		Success: 1,
		Message: "Success",
	})
}

// func (server *Server) NewStorageInterface(r *http.Request) models.StorageInterface {
// 	test := r.URL.Query().Get("storage")
// 	if test != "" {
// 		testdata := server.MockInterface.(models.StorageInterface)
// 		return testdata
// 	}
// 	data := models.NewStorage(server.DB)
// 	return data
// }
