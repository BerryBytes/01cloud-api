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

func TestAuthorizationValidate(t *testing.T) {
	testCases := []struct {
		name       string
		data       *Authorization
		pid        uint
		errMessage error
	}{
		{
			name: "VALID_CASE",
			data: &Authorization{
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
				ProjectID:  1,
				UserRoleID: 1,
				Group: &Group{
					Model: gorm.Model{
						ID: 1,
					},
					Name: "test-group",
					Organization: &Organization{
						Model: gorm.Model{
							ID: 1,
						},
						Name: "test-org",
					},
					OrganizationID: 1,
				},
			},
			pid:        1,
			errMessage: nil,
		},
		{
			name:       "EMAIL_REQUIRE_ERROR",
			data:       &Authorization{},
			pid:        1,
			errMessage: errors.New("required Email or Group Id"),
		},
		{
			name: "USER_ROLE_REQUIRE_ERROR",
			data: &Authorization{
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
			},
			pid:        1,
			errMessage: errors.New("required UserRole"),
		},
		{
			name: "PROJECT_REQUIRED_ERROR",
			data: &Authorization{
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
				UserRoleID: 1,
				UserID:     1,
			},
			pid:        1,
			errMessage: errors.New("required Project"),
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

func TestAuthorizationToJson(t *testing.T) {
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
		Environment: &Environment{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-env",
			Application: &Application{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-app",
				Project: &Project{
					Model: gorm.Model{
						ID: 1,
					},
					Name: "test-project",
				},
				ProjectID: 1,
			},
			ApplicationID: 1,
		},
		EnvironmentID: 1,
		Group: &Group{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-group",
			Organization: &Organization{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-org",
			},
			OrganizationID: 1,
		},
	}
	mapResponse, err := authorization.ToJson()
	assert.NoError(t, err)
	assert.NotNil(t, mapResponse)
}

func TestAuthorizationPrepare(t *testing.T) {
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
		Environment: &Environment{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-env",
			Application: &Application{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-app",
				Project: &Project{
					Model: gorm.Model{
						ID: 1,
					},
					Name: "test-project",
				},
				ProjectID: 1,
			},
			ApplicationID: 1,
		},
		EnvironmentID: 1,
	}
	authorization.Prepare()
}

func TestCreateAuthorization(t *testing.T) {
	authorization := &Authorization{
		Email: "test@gmail.com",
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewAuthorizationRepo()
	saved, err := data.Save(server.DB, authorization)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, saved)
	assert.NoError(t, err)
}

func TestCreateAuthorizationErr(t *testing.T) {
	authorization := &Authorization{
		Email: "test@gmail.com",
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnError(errors.New("error occurs while saving..."))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewAuthorizationRepo()
	auth, err := data.Save(server.DB, authorization)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Empty(t, auth)
	assert.Error(t, err)
}

func TestAuthorizationFind(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
	authorization := &Authorization{
		Model: gorm.Model{
			ID: 1,
		},
		Email: "test@gmail.com",
		Application: &Application{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-app",
		},
		ApplicationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth, err := d.Find(server.DB, uint64(authorization.ID))
	assert.NotNil(t, auth)
	assert.NoError(t, err)
}

func TestGetRoleProject(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth, err := d.GetRoleProject(server.DB, authorization.UserID, authorization.ProjectID)
	assert.NotNil(t, auth)
	assert.NoError(t, err)
}

func TestFindAllAuthorization(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "application_id"}).AddRow(1, time.Now(), time.Now(), "test@gmail.com", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewAuthorizationRepo()
	auths, err := data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting authorizations: %v\n", err)
		return
	}
	assert.Equal(t, len(*auths), 1)
}

func TestFindAllInProject(t *testing.T) {
	authorization := &Authorization{
		Model: gorm.Model{
			ID: 1,
		},
		Project: &Project{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-project",
		},
		ProjectID: 1,
		Email:     "test@gmail.com",
		Application: &Application{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-app",
		},
		ApplicationID: 1,
		Environment: &Environment{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-env",
		},
		EnvironmentID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "application_id"}).AddRow(1, time.Now(), time.Now(), "test@gmail.com", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewAuthorizationRepo()
	auths, err := data.FindAllInProject(server.DB, uint64(authorization.ProjectID))
	if err != nil {
		t.Errorf("this is the error getting projects: %v\n", err)
		return
	}
	assert.Equal(t, len(*auths), 1)
}

func TestFindAllInEnv(t *testing.T) {
	authorization := &Authorization{
		Model: gorm.Model{
			ID: 1,
		},
		Project: &Project{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-project",
		},
		ProjectID: 1,
		Email:     "test@gmail.com",
		Application: &Application{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-app",
		},
		ApplicationID: 1,
		Environment: &Environment{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-env",
		},
		EnvironmentID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email", "application_id"}).AddRow(1, time.Now(), time.Now(), "test@gmail.com", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewAuthorizationRepo()
	auths, err := data.FindAllInEnv(server.DB, uint64(authorization.EnvironmentID))
	if err != nil {
		t.Errorf("this is the error getting environment: %v\n", err)
		return
	}
	assert.Equal(t, len(*auths), 1)
}

func TestGetRoleApp(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth, err := d.GetRoleApp(server.DB, authorization.UserID, authorization.ProjectID, uint64(authorization.ID))
	assert.NotNil(t, auth)
	assert.NoError(t, err)
}

func TestGetRoleEnv(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth, err := d.GetRoleEnv(server.DB, authorization.UserID, authorization.ProjectID, authorization.ApplicationID, uint64(authorization.ID))
	assert.NotNil(t, auth)
	assert.NoError(t, err)
}

func TestIsAuthorizedProject(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAuthorizedProject(server.DB, authorization.UserID, authorization.ProjectID)
	assert.NotNil(t, auth)
}

func TestIsAuthorizedOrganization(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
		Group: &Group{
			Model: gorm.Model{
				ID: 1,
			},
			Name:           "test-group",
			OrganizationID: 1,
			Organization: &Organization{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-org",
			},
		},
		GroupID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAuthorizedOrganization(server.DB, uint(authorization.UserID), uint(authorization.Group.OrganizationID))
	assert.NotNil(t, auth)
}

func TestIsWriteAuthorizedProject(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsWriteAuthorizedProject(server.DB, authorization.UserID, authorization.ProjectID)
	assert.NotNil(t, auth)
}

func TestIsAdminOfProject(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAdminOfProject(server.DB, authorization.UserID, authorization.ProjectID)
	assert.NotNil(t, auth)
}

func TestIsAuthorizedApplication(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAuthorizedApplication(server.DB, authorization.UserID, authorization.ApplicationID, authorization.ProjectID)
	assert.NotNil(t, auth)
}

func TestIsAdminOfApplication(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAdminOfApplication(server.DB, authorization.UserID, authorization.ApplicationID, authorization.ProjectID)
	assert.NotNil(t, auth)
}

func TestIsWriteAuthorizedApplication(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsWriteAuthorizedApplication(server.DB, authorization.UserID, authorization.ApplicationID, authorization.ProjectID)
	assert.NotNil(t, auth)
}

func TestIsAdminOfEnvironment(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
		Environment: &Environment{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-env",
			Application: &Application{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-app",
				Project: &Project{
					Model: gorm.Model{
						ID: 1,
					},
					Name: "test-project",
				},
				ProjectID: 1,
			},
			ApplicationID: 1,
		},
		EnvironmentID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAdminOfEnvironment(server.DB, authorization.UserID, authorization.Environment)
	assert.NotNil(t, auth)
}

func TestIsWriteAuthorizedEnvironment(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
		Environment: &Environment{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-env",
			Application: &Application{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-app",
				Project: &Project{
					Model: gorm.Model{
						ID: 1,
					},
					Name: "test-project",
				},
				ProjectID: 1,
			},
			ApplicationID: 1,
		},
		EnvironmentID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsWriteAuthorizedEnvironment(server.DB, authorization.UserID, authorization.Environment)
	assert.NotNil(t, auth)
}

func TestIsAuthorizedEnvironment(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
		Environment: &Environment{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-env",
			Application: &Application{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-app",
				Project: &Project{
					Model: gorm.Model{
						ID: 1,
					},
					Name: "test-project",
				},
				ProjectID: 1,
			},
			ApplicationID: 1,
		},
		EnvironmentID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAuthorizedEnvironment(server.DB, authorization.UserID, authorization.Environment)
	assert.NotNil(t, auth)
}

func TestIsAdminOfHelmEnvironment(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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

	env := &HelmEnvironment{
		Model: gorm.Model{
			ID: 1,
		},
		Name: "test-env",
		Application: &Application{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-app",
			Project: &Project{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-project",
			},
			ProjectID: 1,
		},
		ApplicationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAdminOfHelmEnvironment(server.DB, authorization.UserID, env)
	assert.NotNil(t, auth)
}

func TestIsWriteAuthorizedHelmEnvironment(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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

	env := &HelmEnvironment{
		Model: gorm.Model{
			ID: 1,
		},
		Name: "test-env",
		Application: &Application{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-app",
			Project: &Project{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-project",
			},
			ProjectID: 1,
		},
		ApplicationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsWriteAuthorizedHelmEnvironment(server.DB, authorization.UserID, env)
	assert.NotNil(t, auth)
}

func TestIsAuthorizedHelmEnvironment(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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

	env := &HelmEnvironment{
		Model: gorm.Model{
			ID: 1,
		},
		Name: "test-env",
		Application: &Application{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-app",
			Project: &Project{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-project",
			},
			ProjectID: 1,
		},
		ApplicationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	auth := d.IsAuthorizedHelmEnvironment(server.DB, authorization.UserID, env)
	assert.NotNil(t, auth)
}

func TestDeleteAuthorization(t *testing.T) {
	authorization := &Authorization{
		Model: gorm.Model{
			ID: 1,
		},
		Email: "test@gmail.com",
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
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, authorization.Email))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewAuthorizationRepo()
	row, err := data.Delete(server.DB, uint64(authorization.ID))
	if err != nil {
		t.Errorf("this is the error deleting authorization: %v\n", err)
		return
	}
	assert.NotNil(t, row)
}

func TestUpdateAuthorization(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "test@gmail.com"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	updateAuth, err := d.Update(server.DB, authorization)
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.NotNil(t, updateAuth)
}

func TestIsUserExists(t *testing.T) {
	authColumns := []string{"id", "created_at", "updated_at", "email", "application_id"}
	appColumns := []string{"id", "created_at", "updated_at", "name"}
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
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(authColumns).AddRow(authorization.ID, time.Now(), time.Now(), "test@gmail.com", authorization.ApplicationID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(appColumns).AddRow(authorization.ID, time.Now(), time.Now(), authorization.Application.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "test@gmail.com"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewAuthorizationRepo()
	isExist := d.IsUserExists(server.DB, authorization)
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.NotNil(t, isExist)
}

func TestDeleteOrganizationMember(t *testing.T) {
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
		Environment: &Environment{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-env",
			Application: &Application{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-app",
				Project: &Project{
					Model: gorm.Model{
						ID: 1,
					},
					Name: "test-project",
				},
				ProjectID: 1,
			},
			ApplicationID: 1,
		},
		EnvironmentID: 1,
		Group: &Group{
			Model: gorm.Model{
				ID: 1,
			},
			Name: "test-group",
			Organization: &Organization{
				Model: gorm.Model{
					ID: 1,
				},
				Name: "test-org",
			},
			OrganizationID: 1,
		},
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, authorization.Email))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewAuthorizationRepo()
	row, err := data.DeleteOrganizationMember(server.DB, authorization.UserID, authorization.Group.OrganizationID)
	if err != nil {
		t.Errorf("this is the error deleting authorization: %v\n", err)
		return
	}
	assert.NotNil(t, row)
}
