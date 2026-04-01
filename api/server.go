package api

import (
	"os"

	log "github.com/sirupsen/logrus"

	"01cloud-api/api/controllers"
	"01cloud-api/api/seed"
	"01cloud-api/api/utils/constants"
	"01cloud-api/api/utils/helper"
)

var server = controllers.Server{}

// func init() {
// 	log.SetFormatter(&log.JSONFormatter{})
// }

func Run() {
	queue, err := helper.NewQueue().GetQueue()
	if err != nil {
		log.Fatalf("Error getting queue: %v", err)
		os.Exit(1)
	}
	functionEnvList := map[string]func(interface{}) error{
		constants.BuildWorkflowSuccess:     server.OnCompleteCi,
		constants.CreateClusterStatus:      server.OnReceiveClusterStatus,
		constants.CreatedTemplateStorage:   server.OnCreatedStorage,
		constants.PostCreateEnvironment:    server.CloneEnvironmentFunctionSet,
		constants.ExternalSecretStatus:     server.ExternalSecretStatus,
		constants.ExternalSecretSyncStatus: server.ExternalSecretSyncStatus,
		constants.ErrorQueue:               server.OnReceiveErrorQueue,
	}
	server.Initialize(os.Getenv("DB_DRIVER"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_PORT"), os.Getenv("DB_HOST"), os.Getenv("DB_NAME"), os.Getenv("BASE_URL"))
	seed.Load(server.DB)
	for name, function := range functionEnvList {
		go func(name string, function func(interface{}) error) {
			err := queue.Subscribe(name, function)
			if err != nil {
				log.Errorf("Error subscribing to queue %s: %v", name, err)
			}
		}(name, function)
	}
	server.Run(":8081")
	//<-ch
}
