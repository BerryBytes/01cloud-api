package controllers

import (
	"01cloud-api/api/models"
	"encoding/json"

	log "github.com/sirupsen/logrus"
)

func (server *Server) OnCreatedStorage(storageData interface{}) error {
	data := models.StorageResponse{}
	dat := storageData.([]byte)
	err := json.Unmarshal(dat, &data)
	if err != nil {
		log.Error(err)
		return err
	}
	environment, err := (&models.Environment{}).Find(server.DB, data.EnvironmentID)
	if err != nil {
		log.Error(err)
		return err
	}
	for _, st := range data.StorageList {
		// Set other fields
		st.EnvironmentID = environment.ID
		st.UserID = uint(environment.Application.OwnerId)
		log.Debug("Saving storage :: ", st)
		_, err := storageInterface.SaveStorage(server.DB, st)
		if err != nil {
			log.Error("Error in saving storage", err)
		}
	}
	return nil
}
