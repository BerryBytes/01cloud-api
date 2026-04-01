package models

import (
	"database/sql"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"

	log "github.com/sirupsen/logrus"
)

type TestServer struct {
	DB   *gorm.DB
	Mock sqlmock.Sqlmock
}

var server TestServer

func TestMain(m *testing.M) {
	Database()
	os.Exit(m.Run())
}

func Database() {
	var err error
	var db *sql.DB
	TestDbDriver := "test"
	db, server.Mock, err = sqlmock.New()
	if err != nil {
		panic(err)
	}
	server.DB, err = gorm.Open(TestDbDriver, db)
	if err != nil {
		log.Warnf("This is the error: %v", err)
	}
}
