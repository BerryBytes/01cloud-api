package controllers

import (
	"01cloud-api/api/responses"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/logging"
	"encoding/json"
	"io"
	"net/http"
)

func (server *Server) UpdateExternalLogging(w http.ResponseWriter, r *http.Request) {
	environment, _, code, err := server.GetEnvironmentWithUserPermission(r, WRITE)
	if err != nil {
		responses.ERROR(w, code, err)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	updateRequest := &logging.LoggingRequest{}
	err = json.Unmarshal(body, updateRequest)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	//This is the case for updating logging so we are calling PrepareLoggingUpdate
	if environment.ExternalLogging.RawMessage != nil {
		body, err = updateRequest.PrepareLoggingUpdate(environment)
		if err != nil {
			responses.ERROR(w, http.StatusUnprocessableEntity, err)
			return
		}
	}
	environment.ExternalLogging.RawMessage = body
	err = logging.ValidateLoggingRequest(environment.ExternalLogging)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	updatedEnv, err := environment.UpdateLogging(server.DB)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	updatedLoggingRequest, err := logging.PrepareLoggingRequest(updatedEnv)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = queue.Publish(constants.ExternalLogging, updatedLoggingRequest)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
