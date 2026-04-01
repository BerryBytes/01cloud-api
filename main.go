package main

import (
	"01cloud-api/api"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// @title 01cloud-api
// @version 1.0.1
// @description 01cloud-api
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email info@01cloud.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8081
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, not coming through %v", err)
	} else {
		log.Info("We are getting the env values")
	}
	logLevel := log.InfoLevel
	if os.Getenv("DEBUG") == "true" {
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)

	api.Run()
}
