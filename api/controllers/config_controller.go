package controllers

import (
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/helper"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func (server *Server) PutAdminConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	vars := mux.Vars(r)
	fileName := vars["filename"]
	if fileName == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("required file name"))
		return
	}
	response, err := saveConfig("/data/admin", fileName, body)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data:    response,
	})
}

func (server *Server) GetAdminConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["filename"]
	if fileName == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("required file name"))
		return
	}
	resultMap, err := getConfig("/data/admin/" + fileName)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data:    resultMap,
	})
}

func (server *Server) PutPublicConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["filename"]
	if fileName == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("required file name"))
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	response, err := saveConfig("/data/public", fileName, body)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data:    response,
	})
}

func (server *Server) GetPublicConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileName := vars["filename"]
	if fileName == "" {
		responses.ERROR(w, http.StatusBadRequest, errors.New("required file name"))
		return
	}
	resultMap, err := getConfig("/data/public/" + fileName)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data:    resultMap,
	})
}

func saveConfig(path, fileName string, body []byte) (map[string]interface{}, error) {
	if _, err := os.Stat(path); err != nil {
		err := os.MkdirAll(path, 0600)
		if err != nil {
			return nil, err
		}
	}
	err := os.WriteFile(path+"/"+fileName, body, 0777)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(body, &resultMap)
	if err != nil {
		return nil, err
	}
	return resultMap, nil
}

func getConfig(path string) (map[string]interface{}, error) {
	dataBytes, err := helper.ReadConfigFile(path)
	if err != nil {
		return nil, err
	}
	var resultMap map[string]interface{}
	err = json.Unmarshal(dataBytes, &resultMap)
	if err != nil {
		return nil, err

	}
	return resultMap, nil
}
