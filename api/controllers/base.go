package controllers

import (
	"fmt"
	"net/http"
	"os"

	"01cloud-api/api/middlewares"
	"01cloud-api/api/models"
	"01cloud-api/api/utils/cache"
	grpc_helper "01cloud-api/api/utils/grpc"
	"01cloud-api/api/utils/helper"
	storage_helper "01cloud-api/api/utils/storage"

	"01cloud-api/api/websocket"

	"cloud.google.com/go/storage"
	store "github.com/berrybytes/01cloud-store"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"    //mysql database driver
	_ "github.com/jinzhu/gorm/dialects/postgres" //postgres database driver
)

type Server struct {
	MockInterface interface{}
	DB            *gorm.DB
	Router        *mux.Router
	Mock          sqlmock.Sqlmock
	StoreClient   store.StoreInterface
	StorageClient *storage.Client
	NotifyClient  *grpc.ClientConn
	BackupClient  *grpc.ClientConn
	Cache         cache.ICache
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
}

func (server *Server) Initialize(Dbdriver, DbUser, DbPassword, DbPort, DbHost, DbName, baseurl string) {
	var err error
	queue, err = helper.NewQueue().GetQueue()
	if err != nil {
		log.Error("Cannot connect to queue:", err)
	}
	if Dbdriver == "mysql" {
		DBURL := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local", DbUser, DbPassword, DbHost, DbPort, DbName)
		server.DB, err = gorm.Open(Dbdriver, DBURL)
		if err != nil {
			log.Warnf("Cannot connect to %s database: %v", Dbdriver, err)
		}
	}
	if Dbdriver == "postgres" {
		DBURL := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", DbHost, DbPort, DbUser, DbName, DbPassword)
		server.DB, err = gorm.Open(Dbdriver, DBURL)
		if err != nil {
			log.Warnf("Cannot connect to %s database: %v", Dbdriver, err)
		}
	}
	if os.Getenv("DBDEBUG") == "true" {
		server.DB = server.DB.Debug()
	}
	middlewares.DB = server.DB
	server.Migrate()
	sc, err := storage_helper.CreateStorageClient()
	if err != nil {
		log.Error("Cannot connect to cloud storage:", err)
	}

	backupClient, err := grpc_helper.GRPCConnection(os.Getenv("GRPC_BACKUP_SERVER"))
	if err != nil {
		log.Warnf("Cannot connect to backup client: %v", err)
	}
	server.StoreClient = store.NewStore()
	notifyClient, err := grpc_helper.GRPCConnection(os.Getenv("GRPC_NOTIFICATION_SERVER"))
	if err != nil {
		log.Warnf("Cannot connect to notification client: %v", err)
	}
	server.StorageClient = sc
	server.BackupClient = backupClient
	server.NotifyClient = notifyClient
	server.Cache = cache.NewCache(models.CacheDuration())
	server.Router = mux.NewRouter()
	server.initializeRoutes(baseurl)
	go websocket.H.Run()
}

func (server *Server) Migrate() {
	server.DB.AutoMigrate(models.User{},
		models.UserRole{},
		models.GitUser{},
		models.Authorization{},
		models.ResetPassword{},
		models.Subscription{},
		models.Project{},
		models.Plugin{},
		models.PluginVersion{},
		models.Application{},
		models.Resource{},
		models.Cluster{},
		models.Environment{},
		models.CiConfig{},
		models.Activity{},
		models.CronJob{},
		models.CronImage{},
		models.Storage{},
		models.Group{},
		models.InitContainer{},
		models.OrganizationPlan{},
		models.Organization{},
		models.LoadBalancer{},
		models.OrganizationMembers{},
		models.ClusterRequest{},
		models.UserInvite{},
		models.DNS{},
		models.ImageRegistry{},
		models.PluginCategory{},
		models.HelmEnvironment{},
		models.Chart{},
		models.ChartVersion{},
		models.Repo{},
		models.OperatorRequest{},
		models.Operator{},
		models.Token{},
		models.UsageHistory{},
		models.Session{},
	)
	err := server.DB.Model(&models.CiConfig{}).ModifyColumn("hook_id", "text").Error
	if err != nil {
		log.Errorln(err)
	}
	//_, e := server.DB.DB().Exec(constants.AlterTableProject)
	//log.Errorln(e)
	//_, e = server.DB.DB().Exec(constants.AlterTableEnvironment)
	//log.Errorln(e)
	//e = server.DB.Model(&models.CronJob{}).ModifyColumn("command", "text").Error
	//e = server.DB.Model(&models.InitContainer{}).ModifyColumn("command", "text").Error
	//log.Errorln(e)
}

func (server *Server) Run(addr string) {
	log.Info("Listening to port http://localhost:8081")
	log.Fatal(http.ListenAndServe(addr, server.Router))
}
