package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}

type Response struct {
	Success int         `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Count   int         `json:"count,omitempty"`
	Page    int         `json:"page,omitempty"`
	Size    int         `json:"size,omitempty"`
}

func ERROR(w http.ResponseWriter, statusCode int, err error) {
	if err != nil {
		JSON(w, statusCode, Response{
			Message: err.Error(),
			Success: 0,
		})
		return
	}
	JSON(w, http.StatusBadRequest, nil)
}
