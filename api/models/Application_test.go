package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestApplicationPrepare(t *testing.T) {
	data := &Application{
		Name:        "test-app",
		ProjectID:   1,
		PluginID:    1,
		ClusterID:   1,
		ServiceType: 2,
	}
	data.Prepare()
}

func TestApplicationToJSON(t *testing.T) {
	data := &Application{
		Name:        "test-app",
		ProjectID:   1,
		PluginID:    1,
		ClusterID:   1,
		ServiceType: 2,
	}
	mg, err := data.ToJson()
	assert.NoError(t, err)
	assert.NotNil(t, mg)
}

func TestApplicationValidate(t *testing.T) {
	testCases := []struct {
		name       string
		data       *Application
		pid        uint
		errMessage error
	}{
		{
			name: "VALID_CASE",
			data: &Application{
				Name:        "test-app",
				ProjectID:   1,
				PluginID:    1,
				ClusterID:   1,
				ServiceType: 2,
			},
			errMessage: nil,
		},
		{
			name: "REQUIRE_NAME",
			data: &Application{
				Name: "",
			},
			errMessage: errors.New("required Name"),
		},
		{
			name: "VALID_NAME_REQUIRE",
			data: &Application{
				Name: "1234&&",
			},
			errMessage: errors.New("allowed alphanumeric, underscore, hyphen and space only"),
		},
		{
			name: "PROJECT_ID_REQUIRE",
			data: &Application{
				Name: "test-app",
			},
			errMessage: errors.New("required Project Id"),
		},
		{
			name: "PLUGIN_REQUIRE",
			data: &Application{
				Name:        "test-app",
				ProjectID:   1,
				PluginID:    0,
				ServiceType: 1,
			},
			errMessage: errors.New("required Plugin"),
		},
		{
			name: "CLUSTER_ID_REQUIRE",
			data: &Application{
				Name:        "test-app",
				ProjectID:   1,
				PluginID:    1,
				ServiceType: 2,
			},
			errMessage: errors.New("region Plugin"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			tc.data.ID = tc.pid
			err = tc.data.Validate()
			assert.Equal(t, tc.errMessage, err)
		})
	}
}

// func TestApplicationCreate(t *testing.T) {
// 	app := &Application{
// 		Name:      "test-app",
// 		ProjectID: 1,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewApplication()
// 	saved, err := data.Save(server.DB, app)
// 	if err != nil {
// 		t.Errorf("this is the error getting the Application: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Name, app.Name)
// }

func TestFindApplication(t *testing.T) {
	app := &Application{
		Name:      "test-app",
		ProjectID: 1,
	}
	app.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "test-app"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	app, err := data.Find(server.DB, uint64(app.ID))
	if err != nil {
		t.Errorf("this is the error getting one application: %v\n", err)
		return
	}
	assert.NotNil(t, app)
	assert.NoError(t, err)
}

func TestIsNameExistsApplication(t *testing.T) {
	app := &Application{
		Name:      "test-app",
		ProjectID: 1,
	}
	app.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "test-app"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	isAppExist := data.IsNameExists(server.DB, app.ProjectID, "app")
	if err != nil {
		t.Errorf("this is the error getting one application: %v\n", err)
		return
	}
	assert.NotNil(t, app)
	assert.Equal(t, isAppExist, false)
}

func TestFindAllApplication(t *testing.T) {
	app := &Application{
		Name:      "app1",
		ProjectID: 1,
		Active:    true,
		Project: &Project{
			Name:   "project",
			UserID: 1,
		},
	}
	app.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "app"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	apps, err := data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting all applcations: %v\n", err)
		return
	}
	assert.Equal(t, len(*apps), 1)
}

func TestSearchApplication(t *testing.T) {
	app := &Application{
		Name:      "app1",
		ProjectID: 1,
		Active:    true,
		Project: &Project{
			Name:   "project",
			UserID: 1,
		},
	}
	app.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "organization_id"}).AddRow(1, time.Now(), time.Now(), "project_test", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	apps, err := data.SearchApplication(server.DB, uint(app.Project.UserID), "")
	if err != nil {
		t.Errorf("this is the error getting search applications : %v\n", err)
		return
	}
	assert.Equal(t, len(apps), 1)
}

func TestFindAllApplicationByProject(t *testing.T) {
	app := &Application{
		Name:      "app1",
		ProjectID: 1,
		Active:    true,
		Project: &Project{
			Name:   "project",
			UserID: 1,
		},
	}
	app.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "app"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	apps, err := data.FindAllApplicationByProject(server.DB, app.ProjectID, app.Project.UserID, true)
	if err != nil {
		t.Errorf("this is the error getting all application by project: %v\n", err)
		return
	}
	assert.Equal(t, len(apps), 1)
}

func TestFindAllApplicationCountByProject(t *testing.T) {
	app := &Application{
		Name:      "app1",
		ProjectID: 1,
		Active:    true,
		Project: &Project{
			Name:   "project",
			UserID: 1,
		},
	}
	app.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "app"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	apps := data.FindAllApplicationCountByProject(server.DB, app.ProjectID, app.Project.UserID)
	if err != nil {
		t.Errorf("this is the error getting in counting application by project: %v\n", err)
		return
	}
	assert.Equal(t, apps, 0)
}

func TestFindAllApplicationByProjectForAdmin(t *testing.T) {
	app := &Application{
		Name:      "app1",
		ProjectID: 1,
		Active:    true,
		Project: &Project{
			Name:   "project",
			UserID: 1,
		},
	}
	app.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), app.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	apps, err := data.FindAllApplicationByProjectForAdmin(server.DB, app.ProjectID, app.Project.UserID)
	if err != nil {
		t.Errorf("this is the error getting all admin applications: %v\n", err)
		return
	}
	assert.Equal(t, len(*apps), 1)
}

func TestUpdateApplication(t *testing.T) {
	app := &Application{
		Name:      "test-app",
		ProjectID: 1,
	}
	app.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "project_id"}).AddRow(1, app.Name, app.ProjectID))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "project_id"}).AddRow(1, app.Name, app.ProjectID))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	updatedApp, err := data.Update(server.DB, app)
	if err != nil {
		t.Errorf("this is the error updating the application: %v\n", err)
		return
	}
	assert.NotNil(t, updatedApp)
}

func TestChangeIsActiveApplication(t *testing.T) {
	app := &Application{
		Name:      "test",
		ProjectID: 1,
		Active:    false,
	}
	app.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "project_id", "active"}).AddRow(1, app.Name, app.ProjectID, false))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewApplication()
	err := data.ChangeIsActive(server.DB, true, app.ID)
	if err != nil {
		t.Errorf("this is the error change status the application: %v\n", err)
		return
	}
	assert.Nil(t, err)
}

func TestDeleteApplication(t *testing.T) {
	app := &Application{
		Name:      "test-app",
		ProjectID: 1,
	}
	app.ID = 1
	authorization := &Authorization{
		Model: gorm.Model{
			ID: 1,
		},
		Email: "test@gmail.com",
		User: &User{
			Model: gorm.Model{
				ID: 1,
			},
			FirstName: "test",
		},
		UserID: 1,
		Project: &Project{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-project",
		},
		ProjectID: 1,
		Application: &Application{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-app",
		},
		ApplicationID: 1,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT * from "Authorization" where (ApplicationID=$1)`)).WithArgs(app.ID).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "application_id"}).AddRow(1, authorization.Email, app.ID))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))

	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, app.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewApplication()
	id, _ := data.Delete(server.DB, uint64(app.ID))

	assert.Equal(t, int64(id), int64(0))
}
