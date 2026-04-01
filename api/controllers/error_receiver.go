package controllers

import (
	"01cloud-api/api/models"
	"01cloud-api/api/websocket"
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
)

func (server *Server) OnReceiveErrorQueue(errorData interface{}) error {
	errorMsg := &models.ErrorMessage{}
	inputBytes := errorData.([]byte)
	err := json.Unmarshal(inputBytes, errorMsg)
	if err != nil {
		return err
	}
	if errorMsg.ClusterId != 0 {
		crInterface := models.NewClusterRequest()
		clusterRequest, err := crInterface.Find(server.DB, uint64(errorMsg.ClusterId))
		if err != nil {
			return err
		}
		if errorMsg.Code != 0 {
			clusterRequest.ErrorMessage.RawMessage = inputBytes
			err := crInterface.UpdateErrorMessage(server.DB, clusterRequest)
			if err != nil {
				return err
			}
			wsConn, mutex := websocket.WebsocketConn(fmt.Sprintf("cluster-%d", clusterRequest.ID))
			go websocket.EmitErrorMessage(wsConn, mutex, errorMsg)
			return nil
		}
		if clusterRequest.ErrorMessage.RawMessage != nil {
			clusterRequest.ErrorMessage.RawMessage = nil
			err := crInterface.UpdateErrorMessage(server.DB, clusterRequest)
			if err != nil {
				return err
			}
		}
		return nil
	}
	environment := models.Environment{}
	env, err := environment.Find(server.DB, uint64(errorMsg.EnvironmentId))
	if err != nil {
		log.Error("error fetch environment :: ", err)
		return err
	}
	if errorMsg.Code != 0 {
		env.ErrorMessage.RawMessage = inputBytes
		err := env.UpdateErrorMessage(server.DB)
		if err != nil {
			return err
		}
		wsConn, mutex := websocket.WebsocketConn(fmt.Sprintf("env-%d", env.ID))
		go websocket.EmitErrorMessage(wsConn, mutex, errorMsg)
		return nil
	}
	if env.ErrorMessage.RawMessage != nil {
		env.ErrorMessage.RawMessage = nil
		err := env.UpdateErrorMessage(server.DB)
		if err != nil {
			return err
		}
	}
	return nil
}
