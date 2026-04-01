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

var GUser = &GitUser{
	Model: gorm.Model{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	ServiceUserName: "bish",
	AccessToken:     "asdf",
	ServiceName:     "service1",
	GitUserID:       1,
	UserID:          1,
}

func TestGitUserValidate(t *testing.T) {
	testCases := []struct {
		name       string
		data       *GitUser
		pid        uint
		errMessage error
	}{
		{
			name: "VALID_CASE",
			data: &GitUser{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ServiceUserName: "bish",
				AccessToken:     "asdf",
				ServiceName:     "service1",
				GitUserID:       1,
				UserID:          1,
			},
			errMessage: nil,
		},
		{
			name: "REQUIRE_USERID",
			data: &GitUser{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ServiceUserName: "bish",
				AccessToken:     "asdf",
				ServiceName:     "service1",
				GitUserID:       1,
				UserID:          0,
			},
			errMessage: errors.New("user id required"),
		},
		{
			name: "REQUIRE_USERID",
			data: &GitUser{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ServiceUserName: "bish",
				AccessToken:     "asdf",
				ServiceName:     "service1",
				GitUserID:       0,
				UserID:          1,
			},
			errMessage: errors.New("git user id required"),
		},
		{
			name: "VALID_ACCESS_TOKEN",
			data: &GitUser{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ServiceUserName: "bish",
				AccessToken:     "",
				ServiceName:     "service1",
				GitUserID:       1,
				UserID:          1,
			},
			errMessage: errors.New("access token is required"),
		},
		{
			name: "SERVICENAME_REQUIRE",
			data: &GitUser{
				Model: gorm.Model{
					ID:        1,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				ServiceUserName: "bish",
				AccessToken:     "asdf",
				ServiceName:     "",
				GitUserID:       1,
				UserID:          1,
			},
			errMessage: errors.New("service name  is required"),
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

func TestGitUser_Find(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "remarks"}).AddRow(1, time.Now(), time.Now(), "good"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewGitUser()
	fpc, err := data.Find(server.DB, uint64(GUser.ID))
	if err != nil {
		t.Errorf("this is the error getting one deduction: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

func TestFindByGitUserID_Find(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "remarks"}).AddRow(1, time.Now(), time.Now(), "good"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewGitUser()
	fpc, err := data.FindByGitUserID(server.DB, uint64(GUser.GitUserID))
	if err != nil {
		t.Errorf("this is the error getting one deduction: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

func TestFindByUserID_Find(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "remarks"}).AddRow(1, time.Now(), time.Now(), "good"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewGitUser()
	fpc, err := data.FindByUserId(server.DB, uint64(GUser.UserID))
	if err != nil {
		t.Errorf("this is the error getting one deduction: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

func TestUserIDAndservice_Find(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "remarks"}).AddRow(1, time.Now(), time.Now(), "good"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewGitUser()
	fpc, err := data.FindByUserIdAndService(server.DB, uint64(GUser.ID), GUser.ServiceName)
	if err != nil {
		t.Errorf("this is the error getting one deduction: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestGitUserCreate(t *testing.T) {
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewGitUser()
// 	saved, err := data.Save(server.DB, GUser)
// 	if err != nil {
// 		t.Errorf("this is the error getting the GitUser: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.ServiceName, GUser.ServiceName)
// 	assert.Equal(t, saved.AccessToken, GUser.AccessToken)
// }

// func TestFindAll(t *testing.T) {
// 	var testGitUser *[]GitUser
// 	var err error
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "service_name"}).AddRow(1, time.Now(), time.Now(), "test-git"))
// 	data := NewGitUser()
// 	testGitUser, err = data.FindAll(server.DB)
// 	if err != nil {
// 		t.Errorf("this is the error getting the users: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, len(*testGitUser), 1)
// }

func TestUpdateGitUser(t *testing.T) {
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "service_name", "access_token"}).AddRow(1, GUser.ServiceName, GUser.AccessToken))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "service_name", "access_token"}).AddRow(1, "good", "db"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewGitUser()
	updatedGitUser, err := data.Update(server.DB, GUser)
	if err != nil {
		t.Errorf("this is the error updating the GitUser: %v\n", err)
		return
	}
	assert.Equal(t, updatedGitUser.ServiceName, GUser.ServiceName)
	assert.Equal(t, updatedGitUser.AccessToken, GUser.AccessToken)
}

// func TestDeleteGitUser(t *testing.T) {
// 	server.Mock.ExpectBegin()
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "service_name", "access_token"}).AddRow(1, GUser.ServiceName, GUser.AccessToken))
// 	server.Mock.ExpectExec(regexp.QuoteMeta(
// 		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)

// 	data := NewGitUser()
// 	id, err := data.Delete(server.DB, uint64(GUser.ID))
// 	if err != nil {
// 		t.Errorf("this is the error updating the user: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, int64(id), int64(1))
// }
