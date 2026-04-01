package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserInvite_Find(t *testing.T) {
	UserInvite := UserInvite{
		FirstName: "aa",
		LastName:  "mmm",
		Email:     "vv@gmail.com",
	}
	UserInvite.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "memory", "disk_space,apps"}).AddRow(1, time.Now(), time.Now(), "aa", 512, 200))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewUserInvite()
	fpc, err := data.Find(server.DB, uint64(UserInvite.ID))
	if err != nil {
		t.Errorf("this is the error getting one UserInvite: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}
func TestByEmail_Find(t *testing.T) {
	UserInvite := UserInvite{
		FirstName: "aa",
		LastName:  "mmm",
		Email:     "vv@gmail.com",
	}
	UserInvite.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "memory", "disk_space,apps"}).AddRow(1, time.Now(), time.Now(), "aa", 512, 200))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewUserInvite()
	fpc, err := data.FindByEmail(server.DB, UserInvite.Email)
	if err != nil {
		t.Errorf("this is the error getting one UserInvite: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}
func TestByToken_Find(t *testing.T) {
	UserInvite := UserInvite{
		FirstName: "aa",
		LastName:  "mmm",
		Email:     "vv@gmail.com",
		Token:     "asdss",
	}
	UserInvite.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "memory", "disk_space,apps"}).AddRow(1, time.Now(), time.Now(), "aa", 512, 200))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewUserInvite()
	fpc, err := data.FindByToken(server.DB, UserInvite.Token)
	if err != nil {
		t.Errorf("this is the error getting one UserInvite: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestUserInviteCreate(t *testing.T) {
// 	UserInvite := UserInvite{
// 		FirstName: "aa",
// 		LastName:  "mmm",
// 		Email:     "vv@gmail.com",
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewUserInvite()
// 	saved, err := data.Save(server.DB, UserInvite)
// 	if err != nil {
// 		t.Errorf("this is the error getting the UserInvite: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.FirstName, UserInvite.FirstName)
// 	assert.Equal(t, saved.LastName, UserInvite.LastName)
// }

func TestFindAllUserInvite(t *testing.T) {
	var testUserInvite *[]UserInvite
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "memory", "apps"}).AddRow(1, "dd", 512, 200))

	data := NewUserInvite()
	testUserInvite, err = data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testUserInvite), 1)
}

func TestUpdateUserInvite(t *testing.T) {
	UserInvite := UserInvite{
		FirstName: "aa",
		LastName:  "mmm",
		Email:     "vv@gmail.com",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).AddRow(1, UserInvite.FirstName, UserInvite.LastName, UserInvite.Email))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).AddRow(1, "aa", "mmm", "zz"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	UserInvite.ID = 1
	data := NewUserInvite()
	updatedUserInvite, err := data.Update(server.DB, UserInvite)
	if err != nil {
		t.Errorf("this is the error updating the UserInvite: %v\n", err)
		return
	}
	assert.Equal(t, updatedUserInvite.FirstName, UserInvite.FirstName)
	assert.Equal(t, updatedUserInvite.LastName, UserInvite.LastName)
}

func TestDeleteUserInvite(t *testing.T) {
	UserInvite := UserInvite{
		FirstName: "aa",
		LastName:  "mmm",
		Email:     "vv@gmail.com",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name", "email"}).AddRow(1, UserInvite.FirstName, UserInvite.LastName, UserInvite.Email))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewUserInvite()
	id, err := data.Delete(server.DB, uint64(UserInvite.ID))
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}
