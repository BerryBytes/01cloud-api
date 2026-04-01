package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestStorage_Find(t *testing.T) {
	storage := Storage{
		Name:     "aa",
		Capacity: 512,
	}
	storage.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "capacity"}).AddRow(1, time.Now(), time.Now(), "aa", 512))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewStorage()
	fpc, err := data.Find(server.DB, uint64(storage.ID))
	if err != nil {
		t.Errorf("this is the error getting one storage: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestStorageCreate(t *testing.T) {
// 	storage := Storage{
// 		Name:     "testname",
// 		Capacity: 512,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewStorage()
// 	saved, err := data.SaveStorage(server.DB, storage)
// 	if err != nil {
// 		t.Errorf("this is the error getting the Storage: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Capacity, storage.Capacity)
// }

func TestUpdateStorage(t *testing.T) {
	storage := Storage{
		Name:     "aa",
		Capacity: 512,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "capacity"}).AddRow(1, storage.Name, storage.Capacity))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "capacity"}).AddRow(1, "aa", 512))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	storage.ID = 1
	data := NewStorage()
	updatedstorage, err := data.Update(server.DB, storage)
	if err != nil {
		t.Errorf("this is the error updating the storage: %v\n", err)
		return
	}
	assert.Equal(t, updatedstorage.Capacity, storage.Capacity)
}

func TestDeleteStorage(t *testing.T) {
	storage := Storage{
		Name:     "aa",
		Capacity: 512,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "capacity"}).AddRow(1, storage.Name, storage.Capacity))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewStorage()
	id, err := data.Delete(server.DB, uint64(storage.ID))
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}
