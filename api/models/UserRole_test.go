package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUserRole_Find(t *testing.T) {
	UserRole := UserRole{
		Name: "aa",
		Code: 112,
	}
	UserRole.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "code"}).AddRow(1, time.Now(), time.Now(), "aa", 112))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewUserRole()
	fpc, err := data.Find(server.DB, uint64(UserRole.ID))
	if err != nil {
		t.Errorf("this is the error getting one UserRole: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestUserRoleCreate(t *testing.T) {
// 	UserRole := &UserRole{
// 		Name: "aa",
// 		Code: 112,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewUserRole()
// 	saved, err := data.Save(server.DB, UserRole)
// 	if err != nil {
// 		t.Errorf("this is the error getting the UserRole: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Name, UserRole.Name)
// 	assert.Equal(t, saved.Code, UserRole.Code)
// }

func TestFindAllUserRole(t *testing.T) {
	var testUserRole *[]UserRole
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code"}).AddRow(1, "aa", 112))

	data := NewUserRole()
	testUserRole, err = data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testUserRole), 1)
}
func TestUpdateUserRole(t *testing.T) {
	UserRole := &UserRole{
		Name: "aa",
		Code: 112,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code"}).AddRow(1, UserRole.Name, UserRole.Code))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code"}).AddRow(1, "aa", 1212))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	UserRole.ID = 1
	data := NewUserRole()
	updatedUserRole, err := data.Update(server.DB, UserRole)
	if err != nil {
		t.Errorf("this is the error updating the UserRole: %v\n", err)
		return
	}
	assert.Equal(t, updatedUserRole.Name, UserRole.Name)
	assert.Equal(t, updatedUserRole.Code, UserRole.Code)
}

func TestDeleteUserRole(t *testing.T) {
	UserRole := UserRole{
		Name: "aa",
		Code: 112,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "code"}).AddRow(1, UserRole.Name, UserRole.Code))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewUserRole()
	id, err := data.Delete(server.DB, uint64(UserRole.ID))
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}
