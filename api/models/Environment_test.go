package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFileManagerStatus(t *testing.T) {
	app := &Application{
		Name: "test app",
	}
	resource := &Resource{
		Name: "test app",
	}
	time := time.Now()
	env := &Environment{
		Name:               "test",
		FileManagerEnabled: &time,
		Application:        app,
		ApplicationID:      1,
		Resource:           resource,
		ResourceID:         1,
	}
	env.ID = 1
	for _, data := range testCondition {
		server.Mock.MatchExpectationsInOrder(false)
		if data == "ok" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			server.Mock.ExpectExec(regexp.QuoteMeta(
				`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		} else {
			server.Mock.MatchExpectationsInOrder(true)
			server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnError(err)
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		datas := NewEnvironment()
		env, err := datas.UpdateFileManagerStatus(server.DB, env)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NotNil(t, env)
	}

}
