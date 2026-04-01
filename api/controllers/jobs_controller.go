package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"01cloud-api/api/responses"

	"github.com/berrybytes/01cloud-store/model"
)

func (server *Server) StoreJobLogs(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	data := model.JobLog{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = server.StoreClient.CronJob().StoreCronJobLogs(data)
	if err != nil {
		responses.ERROR(w, http.StatusInternalServerError, err)
		return
	}
	responses.JSON(w, http.StatusOK, map[string]string{
		"message": "log stored",
	})

}
