package controllers

import (
	"01cloud-api/api/responses"
	"encoding/json"
	"io"
	"net/http"
)

type PublishMessage struct {
	Title string                 `json:"title"`
	Data  map[string]interface{} `json:"data"`
}

func (server *Server) publishMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	publishMessage := PublishMessage{}
	err = json.Unmarshal(body, &publishMessage)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = queue.Publish(publishMessage.Title, publishMessage.Data)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
}
