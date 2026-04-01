package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestUser_Find(t *testing.T) {
	user := User{
		Email: "bish@gmail.com",
	}
	user.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email"}).AddRow(1, time.Now(), time.Now(), "bish@gmail.com"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewUser()
	fpc, err := data.FindUserByID(server.DB, uint(user.ID))
	if err != nil {
		t.Errorf("this is the error getting one deduction: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

func TestUserByEmail_Find(t *testing.T) {
	user := User{
		Email: "bish@gmail.com",
	}
	user.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "email"}).AddRow(1, time.Now(), time.Now(), "bish@gmail.com"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewUser()
	fpc, err := data.FindUserByEmail(server.DB, user.Email)
	if err != nil {
		t.Errorf("this is the error getting one deduction: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestUserCreate(t *testing.T) {
// 	user := User{
// 		Email: "bish@gmail.com",
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	//saved, err := server.store.CreateDeduction(deduction)
// 	data := NewUser()
// 	saved, err := data.SaveUser(server.DB, &user)
// 	if err != nil {
// 		t.Errorf("this is the error getting the User: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Email, user.Email)
// }

func TestFindAllUser(t *testing.T) {
	var testuser *[]User
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "dd@gmail.com"))

	data := NewUser()
	testuser, err = data.FindAllUsers(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testuser), 1)
}

func TestFindAllUserByProject(t *testing.T) {
	user := User{
		Email: "bish@gmail.com",
	}
	user.ID = 1
	var testuser *[]User
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, "aa", "dd", "Nepal", 1, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "dd", "ff"))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name"}).AddRow(1, "dd", "ff"))
	data := NewUser()
	testuser, _, err = data.FindAllUsersWithFilters(server.DB, 1, 1, "", "", "")
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testuser), 1)
}
func TestUpdateUser(t *testing.T) {
	user := User{
		Email: "bish@gmail.com",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, user.Email))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, "aa@gmail.com"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	user.ID = 1
	data := NewUser()
	updateduser, err := data.UpdateAUser(server.DB, user.ID, &user)
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, updateduser.Email, user.Email)
}

func TestVerifyUser(t *testing.T) {
	user := User{
		Email:         "bish@gmail.com",
		Active:        false,
		EmailVerified: false,
	}
	user.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "active", "email_verified"}).AddRow(1, user.Email, user.Active, user.EmailVerified))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "active", "email_verified"}).AddRow(1, "bish@gmail.com", true, true))
	// server.Mock.ExpectExec(regexp.QuoteMeta(
	// 	`UPDATE`)).WillReturnResult(sqlmock.NewResult(1, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewUser()
	updateduser, err := data.VerifyEmail(server.DB, user.ID)
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, updateduser.EmailVerified, true)
}
func TestUpdatePassword(t *testing.T) {
	user := User{
		Email:    "bish@gmail.com",
		Password: "bish",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(1, user.Email, user.Password))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(1, "bish@gmail.com", "ccc"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	user.ID = 1
	data := NewUser()
	updateduser := data.UpdatePassword(server.DB, &user)
	assert.Nil(t, updateduser)

}
func TestDeleteUser(t *testing.T) {
	user := User{
		Email: "bish@gmail.com",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email"}).AddRow(1, user.Email))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewUser()
	id, err := data.DeleteAUser(server.DB, uint(user.ID))
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}

// func TestInvoice_Find(t *testing.T) {
// 	user := User{
// 		Action:    "aa",
// 		Module:    "dd",
// 		Remarks:   "Nepal",
// 		ProjectID: 1,
// 		UserID:    1,
// 	}
// 	user.ID = 1
// 	project := Project{
// 		Name: "bbb",
// 	}
// 	project.ID = 1
// 	user := User{
// 		FirstName: "bish",
// 	}
// 	user.ID = 1
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, user.Action, user.Module, user.Remarks, user.ProjectID, user.UserID))
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, project.Name))
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name"}).AddRow(1, user.FirstName))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewUser()
// 	fpc, err := data.FindAllUserByProject(server.DB, user.ProjectID, 20, 2, user.Action, "", "", int(user.UserID))
// 	if err != nil {
// 		t.Errorf("this is the error getting the users: %v\n", err)
// 		return
// 	}
// 	assert.NotNil(t, fpc)
// 	assert.NoError(t, err)
// }
