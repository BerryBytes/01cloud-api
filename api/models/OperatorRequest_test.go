package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var operatorRequest = &OperatorRequest{
	PackageName:         "TestName",
	CsvValue:            "{test:test}",
	InstallationMode:    "test-mode",
	InstallPlanApproval: "test-app",
	Channel:             "test-channel",
	Cluster: &Cluster{
		Name:       "test1",
		ConfigPath: "test/path",
		Context:    "test-context",
		Token:      "testadkfjakfmdkkjk",
		Region:     "us-east",
	},
}

// func TestSaveOpearatorRequest(t *testing.T) {

// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewOperatorRequest()
// 	saved, err := data.Save(server.DB, *operatorRequest)
// 	if err != nil {
// 		t.Errorf("this is the error getting the operatorRequest: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.PackageName, operatorRequest.PackageName)
// }

func TestFindAllOperatorRequest(t *testing.T) {
	operatorRequest.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "package_name"}).AddRow(1, time.Now(), time.Now(), operatorRequest.PackageName))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewOperatorRequest()
	apps, err := data.FindAll(server.DB, *operatorRequest)
	if err != nil {
		t.Errorf("this is the error getting all operatorRequest: %v\n", err)
		return
	}
	assert.Equal(t, len(*apps), 1)
}

func TestFindOperatorRequest(t *testing.T) {
	operatorRequest.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "package_name"}).AddRow(1, time.Now(), time.Now(), operatorRequest.PackageName))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "version"}).AddRow(1, "dd", "ff"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewOperatorRequest()
	app, err := data.Find(server.DB, uint64(operatorRequest.ID))
	if err != nil {
		t.Errorf("this is the error getting one operatorRequest: %v\n", err)
		return
	}
	assert.NotNil(t, app)
	assert.NoError(t, err)
}

func TestFindByNameAndClusterID(t *testing.T) {
	operatorRequest.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "package_name"}).AddRow(1, time.Now(), time.Now(), operatorRequest.PackageName))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "version"}).AddRow(1, "dd", "ff"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewOperatorRequest()
	app, err := data.FindByNameAndClusterID(server.DB, uint64(operatorRequest.ID), operatorRequest.PackageName)
	if err != nil {
		t.Errorf("this is the error getting one operatorRequest: %v\n", err)
		return
	}
	assert.NotNil(t, app)
}
func TestUpdate(t *testing.T) {
	operatorRequest.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "package_name"}).AddRow(1, time.Now(), time.Now(), operatorRequest.PackageName))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "package_name"}).AddRow(1, time.Now(), time.Now(), operatorRequest.PackageName))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewOperatorRequest()
	updatedOperatorReq, err := data.Update(server.DB, *operatorRequest)
	if err != nil {
		t.Errorf("this is the error updating the operator: %v\n", err)
		return
	}
	assert.Equal(t, updatedOperatorReq.PackageName, operatorRequest.PackageName)
}

func TestDelete(t *testing.T) {
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "package_name"}).AddRow(1, time.Now(), time.Now(), operatorRequest.PackageName))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewOperatorRequest()
	_, err := data.Delete(server.DB, uint64(operatorRequest.ID))
	if err != nil {
		t.Errorf("this is the error deleting OperatorRequest: %v\n", err)
		return
	}
	assert.Nil(t, err)
}

func TestValidate(t *testing.T) {

	testcases := []struct {
		Name            string
		OperatorRequest *OperatorRequest
		cvalue          string
		Err             error
	}{
		{
			Name:            "CLUSTER_OK",
			OperatorRequest: operatorRequest,
			Err:             nil,
		},
		{
			Name: "EMPTY_NAME",
			OperatorRequest: &OperatorRequest{
				PackageName: "",
			},
			Err: errors.New("required package name"),
		},
		{
			Name: "EMPTY_CONFIG",
			OperatorRequest: &OperatorRequest{
				PackageName:      "test",
				InstallationMode: "",
			},
			Err: errors.New("required installationMode"),
		},
	}
	for _, tc := range testcases {
		err := tc.OperatorRequest.Validate()
		assert.Equal(t, err, tc.Err)
	}

}
