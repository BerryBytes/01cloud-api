package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

// func TestSavedetails(t *testing.T) {
// 	ResetPassword := &ResetPassword{
// 		Model: gorm.Model{
// 			ID:        1,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		Email: "bish@gmail.com",
// 		Token: "dsfrgea",
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewResetPassword()
// 	saved, err := data.SaveDatails(server.DB, ResetPassword)
// 	assert.Nil(t, err)
// 	assert.NotNil(t, saved)

// }

func TestDeleteResetPassword(t *testing.T) {
	ResetPassword := &ResetPassword{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Email: "bish@gmail.com",
		Token: "dsfrgea",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "email", "token"}).AddRow(1, ResetPassword.Email, ResetPassword.Token))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewResetPassword()
	id, err := data.DeleteDetails(server.DB, ResetPassword)
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}
