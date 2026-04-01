package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestSubscription_Find(t *testing.T) {
	subscription := Subscription{
		Name:      "aa",
		Memory:    512,
		DiskSpace: 200,
		Apps:      1,
	}
	subscription.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "memory", "disk_space,apps"}).AddRow(1, time.Now(), time.Now(), "aa", 512, 200))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewSubscription()
	fpc, err := data.Find(server.DB, uint64(subscription.ID))
	if err != nil {
		t.Errorf("this is the error getting one subscription: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestSubscriptionCreate(t *testing.T) {
// 	subscription := Subscription{
// 		Name:      "testname",
// 		Memory:    512,
// 		DiskSpace: 200,
// 		Apps:      1,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewSubscription()
// 	saved, err := data.Save(server.DB, subscription)
// 	if err != nil {
// 		t.Errorf("this is the error getting the Subscription: %v\n", err)
// 		return
// 	}
// 	assert.NotNil(t, saved)
// }

func TestFindAllSubscription(t *testing.T) {
	var testsubscription *[]Subscription
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "memory", "apps"}).AddRow(1, "dd", 512, 200))

	data := NewSubscription()
	testsubscription, err = data.FindAll(server.DB, 1)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testsubscription), 1)
}

// // func TestFindAllSubscriptionByProject(t *testing.T) {
// // 	subscription := Subscription{
// // 		Action:    "aa",
// // 		Module:    "dd",
// // 		Remarks:   "Nepal",
// // 		ProjectID: 1,
// // 		UserID:    1,
// // 	}
// // 	subscription.ID = 1
// // 	var testsubscription *[]Subscription
// // 	var err error
// // 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// // 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, "dd", "ff", "gg", 1, 1))
// // 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// // 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, "dd", "ff", "gg", 1, 1))
// // 	data := NewSubscription(server.DB)
// // 	testsubscription, err = data.FindAllSubscriptionByProject(subscription.ProjectID, 20, 2, subscription.Action, "", "", int(subscription.UserID))
// // 	if err != nil {
// // 		t.Errorf("this is the error getting the users: %v\n", err)
// // 		return
// // 	}
// // 	assert.Equal(t, len(*testsubscription), 1)
// // }
func TestUpdateSubscription(t *testing.T) {
	subscription := Subscription{
		Name:      "aa",
		Memory:    512,
		DiskSpace: 200,
		Apps:      1,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "memory", "apps"}).AddRow(1, subscription.Name, subscription.Memory, subscription.Apps))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "memory", "apps"}).AddRow(1, "aa", 512, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	subscription.ID = 1
	data := NewSubscription()
	updatedsubscription, err := data.Update(server.DB, subscription)
	if err != nil {
		t.Errorf("this is the error updating the subscription: %v\n", err)
		return
	}
	assert.Equal(t, updatedsubscription.Name, subscription.Name)
	assert.Equal(t, updatedsubscription.Apps, subscription.Apps)
}

func TestDeleteSubscription(t *testing.T) {
	subscription := Subscription{
		Name:      "aa",
		Memory:    512,
		DiskSpace: 200,
		Apps:      1,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "memory", "apps"}).AddRow(1, subscription.Name, subscription.Memory, subscription.Apps))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewSubscription()
	id, err := data.Delete(server.DB, uint64(subscription.ID))
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}

// func TestInvoice_Find(t *testing.T) {
// 	subscription := Subscription{
// 		Name:      "aa",
// 		Memory:    512,
// 		DiskSpace: 200,
// 		Apps:      1,
// 	}
// 	subscription.ID = 1
// 	user.ID = 1
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, subscription.Action, subscription.Module, subscription.Remarks, subscription.ProjectID, subscription.UserID))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewSubscription(server.DB)
// 	fpc, err := data.FindAllSubscriptionByProject(subscription.ProjectID, 20, 2, subscription.Action, "", "", int(subscription.UserID))
// 	if err != nil {
// 		t.Errorf("this is the error getting the users: %v\n", err)
// 		return
// 	}
// 	assert.NotNil(t, fpc)
// 	assert.NoError(t, err)
// }
