package controllers

import (
	"01cloud-api/api/mailer"
	"01cloud-api/api/responses"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func (server *Server) ListEmailTemplate(w http.ResponseWriter, r *http.Request) {
	keys := make([]string, len(mailer.EMAIL_DATA))
	i := 0
	for k := range mailer.EMAIL_DATA {
		keys[i] = k
		i++
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data:    keys,
	})
}

func (server *Server) GetEmailTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailTemplate := vars["template"]
	body := mailer.GetMessageBody(emailTemplate)
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
		Data:    body,
	})
}

func (server *Server) SetEmailTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	emailTemplate := vars["template"]
	body, err := io.ReadAll(r.Body)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	payload := struct {
		MessageBody string `json:"data"`
	}{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		responses.ERROR(w, http.StatusUnprocessableEntity, err)
		return
	}
	err = mailer.SetMessageBody(payload.MessageBody, emailTemplate)
	if err != nil {
		responses.ERROR(w, http.StatusBadRequest, err)
		return
	}
	responses.JSON(w, http.StatusOK, responses.Response{
		Success: 1,
		Message: "Success",
	})
}
