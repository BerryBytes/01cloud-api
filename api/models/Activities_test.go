package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestActivity_Find(t *testing.T) {
	activity := Activity{
		Remarks: "good",
	}
	activity.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "remarks"}).AddRow(1, time.Now(), time.Now(), "good"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewActivity()
	fpc, err := data.Find(server.DB, uint64(activity.ID))
	if err != nil {
		t.Errorf("this is the error getting one activity: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestActivityCreate(t *testing.T) {
// 	activity := Activity{
// 		Action:    "aa",
// 		Module:    "dd",
// 		Remarks:   "Nepal",
// 		ProjectID: 1,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewActivity()
// 	saved, err := data.Save(server.DB, activity)
// 	if err != nil {
// 		t.Errorf("this is the error getting the Activity: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Action, activity.Action)
// 	assert.Equal(t, saved.Module, activity.Module)
// }

func TestFindAllActivity(t *testing.T) {
	var testactivity *[]Activity
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks"}).AddRow(1, "dd", "ff", "gg"))

	data := NewActivity()
	testactivity, err = data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testactivity), 1)
}

func TestFindAllActivityByProject(t *testing.T) {
	activity := Activity{
		Action:    "aa",
		Module:    "dd",
		Remarks:   "Nepal",
		ProjectID: 1,
		UserID:    1,
	}
	activity.ID = 1
	var testactivity *[]Activity
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, "aa", "dd", "Nepal", 1, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "dd", "ff"))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name"}).AddRow(1, "dd", "ff"))
	data := NewActivity()
	testactivity, _, err = data.FindAllActivityByProject(server.DB, activity.ProjectID, 20, 2, activity.Action, "", "", "", int(activity.UserID))
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testactivity), 1)
}
func TestUpdateActivity(t *testing.T) {
	activity := Activity{
		Action:  "aa",
		Module:  "dd",
		Remarks: "Nepal",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks"}).AddRow(1, activity.Action, activity.Module, activity.Remarks))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks"}).AddRow(1, "aa", "db", "USA"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	activity.ID = 1
	data := NewActivity()
	updatedactivity, err := data.Update(server.DB, activity)
	if err != nil {
		t.Errorf("this is the error updating the activity: %v\n", err)
		return
	}
	assert.Equal(t, updatedactivity.Action, activity.Action)
	assert.Equal(t, updatedactivity.Module, activity.Module)
}

func TestDeleteActivity(t *testing.T) {
	activity := Activity{
		Action:  "aa",
		Module:  "dd",
		Remarks: "Nepal",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks"}).AddRow(1, activity.Action, activity.Module, activity.Remarks))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewActivity()
	id, err := data.Delete(server.DB, uint64(activity.ID))
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}

func TestActivtyFind(t *testing.T) {
	activity := Activity{
		Action:    "aa",
		Module:    "dd",
		Remarks:   "Nepal",
		ProjectID: 1,
		UserID:    1,
	}
	activity.ID = 1
	project := Project{
		Name: "bbb",
	}
	project.ID = 1
	user := User{
		FirstName: "bish",
	}
	user.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, activity.Action, activity.Module, activity.Remarks, activity.ProjectID, activity.UserID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, project.Name))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name"}).AddRow(1, user.FirstName))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewActivity()
	fpc, _, err := data.FindAllActivityByProject(server.DB, activity.ProjectID, 20, 2, activity.Action, "", "", "", int(activity.UserID))
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}
