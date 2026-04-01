package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestResource_Find(t *testing.T) {
	resource := &Resource{
		Name:  "good",
		Cores: 5,
	}
	resource.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "cores"}).AddRow(1, time.Now(), time.Now(), "good", 5))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewResource()
	fpc, err := data.Find(server.DB, uint64(resource.ID))
	if err != nil {
		t.Errorf("this is the error getting one deduction: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestResourceCreate(t *testing.T) {
// 	resource := &Resource{
// 		Name:  "good",
// 		Cores: 5,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	//saved, err := server.store.CreateDeduction(deduction)
// 	data := NewResource()
// 	saved, err := data.Save(server.DB, resource)
// 	if err != nil {
// 		t.Errorf("this is the error getting the Resource: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Name, resource.Name)
// 	assert.Equal(t, saved.Cores, resource.Cores)
// }

func TestFindAllResource(t *testing.T) {
	var testresource *[]Resource
	var err error
	resource := &Resource{
		OrganizationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "cores"}).AddRow(1, "good", 5))

	data := NewResource()
	testresource, err = data.FindAll(server.DB, resource)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testresource), 1)
}

func TestFindAllWithInactive(t *testing.T) {
	var testresource *[]Resource
	var err error
	resource := &Resource{
		OrganizationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "cores"}).AddRow(1, "good", 5))

	data := NewResource()
	testresource, err = data.FindAllWithInactive(server.DB, resource)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testresource), 1)
}

func TestUpdateResource(t *testing.T) {
	resource := &Resource{
		Name:  "good",
		Cores: 5,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "cores"}).AddRow(1, resource.Name, resource.Cores))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "cores"}).AddRow(1, "aa", "db"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	resource.ID = 1
	data := NewResource()
	updatedresource, err := data.Update(server.DB, resource)
	if err != nil {
		t.Errorf("this is the error updating the resource: %v\n", err)
		return
	}
	assert.Equal(t, updatedresource.Name, resource.Name)
	assert.Equal(t, updatedresource.Cores, resource.Cores)
}

func TestDeleteResource(t *testing.T) {
	resource := &Resource{
		Name:  "good",
		Cores: 5,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "cores"}).AddRow(1, resource.Name, resource.Cores))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewResource()
	id, err := data.Delete(server.DB, uint64(resource.ID))
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}

func TestCheckResourceLimit(t *testing.T) {
	resource := &Resource{
		Name:   "good",
		Cores:  5,
		Memory: 5,
	}
	org := &Organization{
		OrganizationPlan: &OrganizationPlan{
			Cores:  8,
			Memory: 8,
		},
	}

	// server.Mock.ExpectExec(regexp.QuoteMeta(
	// 	`SELECT`)).WillReturnResult(sqlmock.NewResult(0, 1))
	// server.Mock.ExpectCommit()
	// server.Mock.MatchExpectationsInOrder(false)
	data := NewResource()
	dataresource := data.CheckResourceLimit(server.DB, org, resource)
	assert.Equal(t, dataresource, true)
}
