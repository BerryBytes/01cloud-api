package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

func TestClusteRequestToJSON(t *testing.T) {
	testCases := []struct {
		name       string
		data       *ClusterRequest
		errMessage error
	}{
		{
			name: "ClusteRequest_JSON_CASE",
			data: &ClusterRequest{
				Name:   "test-pro",
				Region: "hello ClusteRequest",
			},
			errMessage: nil,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			_, err = tc.data.ToJson()
			assert.Equal(t, tc.errMessage, err)
		})
	}
}
func TestClusteRequestPrepare(t *testing.T) {
	testCases := []struct {
		name       string
		data       *ClusterRequest
		errMessage error
	}{
		{
			name: "ClusteRequest_PREPARE_CASE",
			data: &ClusterRequest{
				Name:   "test-pro",
				Region: "  hello ClusteRequest  ",
			},
			errMessage: nil,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			tc.data.Prepare()
		})
	}
}

func TestClusteRequestValidate(t *testing.T) {
	testCases := []struct {
		name       string
		data       *ClusterRequest
		pid        uint
		errMessage error
	}{

		{
			name: "REQUIRE_NAME",
			data: &ClusterRequest{
				Name: "",
			},
			pid:        0,
			errMessage: errors.New("required name"),
		},
		{
			name: "REQUIRE_PROVIDER",
			data: &ClusterRequest{
				Name:     "Test-ClusteRequest",
				Provider: "",
			},
			pid:        0,
			errMessage: errors.New("required provider name"),
		},
		{
			name: "REQUIRE_REGION",
			data: &ClusterRequest{
				Name:     "Test-ClusteRequest",
				Provider: "cluster",
				Region:   "",
			},
			pid:        0,
			errMessage: errors.New("required region"),
		},
		{
			name: "REQUIRE_NODE_GROUP",
			data: &ClusterRequest{
				Name:            "Test-ClusteRequest",
				Provider:        "sss",
				Region:          "us-east",
				NodeGroupDetail: postgres.Jsonb{},
				//	VpcName:         "",
			},
			pid:        0,
			errMessage: errors.New("required node group details"),
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

func TestClusteRequestCreate(t *testing.T) {
	tClusteRequest := &ClusterRequest{
		Name:            "ClusteRequest_test",
		Provider:        "ClusteRequest_test is purposed to test",
		Region:          "us-east1",
		VpcName:         "san",
		NodeGroupDetail: postgres.Jsonb{},
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewClusterRequest()
	saved, err := data.Save(server.DB, tClusteRequest)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, saved.Name, tClusteRequest.Name)
}

func TestFindClusterRequest(t *testing.T) {
	tClusteRequest := &ClusterRequest{
		Name:            "ClusteRequest_test",
		Provider:        "ClusteRequest_test is purposed to test",
		Region:          "us-east1",
		VpcName:         "san",
		NodeGroupDetail: postgres.Jsonb{},
	}
	tClusteRequest.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "ClusteRequest_test"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewClusterRequest()
	ClusteRequest, err := data.Find(server.DB, uint64(tClusteRequest.ID))
	if err != nil {
		t.Errorf("this is the error getting one ClusteRequest: %v\n", err)
		return
	}
	assert.NotNil(t, ClusteRequest)
	assert.NoError(t, err)
}

func TestFindWithDNSClusterRequest(t *testing.T) {
	tClusteRequest := &ClusterRequest{
		Name:            "ClusteRequest_test",
		Provider:        "ClusteRequest_test is purposed to test",
		Region:          "us-east1",
		VpcName:         "san",
		NodeGroupDetail: postgres.Jsonb{},
	}
	tClusteRequest.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "ClusteRequest_test"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewClusterRequest()
	ClusteRequest, err := data.FindWithDns(server.DB, uint64(tClusteRequest.ID))
	if err != nil {
		t.Errorf("this is the error getting one ClusteRequest: %v\n", err)
		return
	}
	assert.NotNil(t, ClusteRequest)
	assert.NoError(t, err)
}

func TestFindAllClusteRequest(t *testing.T) {
	tClusteRequest := &ClusterRequest{
		Name:            "ClusteRequest_test",
		Provider:        "ClusteRequest_test is purposed to test",
		Region:          "us-east1",
		VpcName:         "san",
		NodeGroupDetail: postgres.Jsonb{},
	}
	tClusteRequest.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "organization_id"}).AddRow(1, time.Now(), time.Now(), "ClusteRequest_test", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewClusterRequest()
	ClusteRequest, err := data.FindAllWithInactive(server.DB)
	if err != nil {
		t.Errorf("this is the error getting one ClusteRequest: %v\n", err)
		return
	}
	assert.Equal(t, len(*ClusteRequest), 1)
}

func TestFindAllClusterByOrganization(t *testing.T) {
	tClusteRequest := &ClusterRequest{
		Name:            "ClusteRequest_test",
		Provider:        "ClusteRequest_test is purposed to test",
		Region:          "us-east1",
		VpcName:         "san",
		NodeGroupDetail: postgres.Jsonb{},
		OrganizationID:  1,
	}
	tClusteRequest.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "organization_id"}).AddRow(1, time.Now(), time.Now(), "ClusteRequest_test", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewClusterRequest()
	ClusteRequest, err := data.FindAllByOrganization(server.DB, int64(tClusteRequest.OrganizationID))
	if err != nil {
		t.Errorf("this is the error getting one ClusteRequest: %v\n", err)
		return
	}
	assert.Equal(t, len(*ClusteRequest), 1)
}

func TestDeleteClusteRequest(t *testing.T) {
	tClusteRequest := &ClusterRequest{
		Name:            "ClusteRequest_test",
		Provider:        "ClusteRequest_test is purposed to test",
		Region:          "us-east1",
		VpcName:         "san",
		NodeGroupDetail: postgres.Jsonb{},
		OrganizationID:  1,
	}
	tClusteRequest.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tClusteRequest.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewClusterRequest()
	row, err := data.Delete(server.DB, uint64(tClusteRequest.ID))
	if err != nil {
		t.Errorf("this is the error deleting ClusteRequest: %v\n", err)
		return
	}
	assert.NotNil(t, row)
}

func TestUpdateClusterRequest(t *testing.T) {
	tClusteRequest := &ClusterRequest{
		Name:            "ClusteRequest_test",
		Provider:        "ClusteRequest_test is purposed to test",
		Region:          "us-east1",
		VpcName:         "san",
		NodeGroupDetail: postgres.Jsonb{},
		OrganizationID:  1,
	}
	tClusteRequest.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tClusteRequest.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "aa@gmail.com"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewClusterRequest()
	updatedcluster, err := data.Update(server.DB, tClusteRequest)
	if err != nil {
		t.Errorf("this is the error updating the cluster: %v\n", err)
		return
	}
	assert.Equal(t, updatedcluster.Name, updatedcluster.Name)
}

// func TestUpdateCluster(t *testing.T) {
// 	tClusteRequest := &ClusterRequest{
// 		Name:            "ClusteRequest_test",
// 		Provider:        "ClusteRequest_test is purposed to test",
// 		Region:          "us-east1",
// 		VpcName:         "san",
// 		NodeGroupDetail: postgres.Jsonb{},
// 		OrganizationID:  1,
// 		Status:          "active",
// 	}
// 		tClusteRequest.ID = 1
// 	server.Mock.ExpectBegin()
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tClusteRequest.Name))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)

// 	data := NewClusterRequest()
// 	updatedcluster, err := data.ChangeStatus(server.DB, tClusteRequest)
// 	if err != nil {
// 		t.Errorf("this is the error updating the cluster: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, updatedcluster.Name, updatedcluster.Name)
// }

// func TestEnabledisableClusterRequest(t *testing.T) {
// 	tClusteRequest := &ClusterRequest{
// 		Name:            "ClusteRequest_test",
// 		Provider:        "ClusteRequest_test is purposed to test",
// 		Region:          "us-east1",
// 		VpcName:         "san",
// 		NodeGroupDetail: postgres.Jsonb{},
// 		Active:          true,
// 	}
// 	tClusteRequest.ID = 1
// 	server.Mock.ExpectBegin()
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tClusteRequest.Name))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)

// 	data := NewClusterRequest()
// 	updatedcluster, err := data.EnableDisableCluster(server.DB, tClusteRequest)
// 	if err != nil {
// 		t.Errorf("this is the error updating the cluster: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, updatedcluster.Name, updatedcluster.Name)
// }

// func TestUpdateEnableClusterRequest(t *testing.T) {
// 	tClusteRequest := &ClusterRequest{
// 		Name:            "ClusteRequest_testdd",
// 		Provider:        "ClusteRequest_test is purposed to testdd",
// 		Region:          "us-east1",
// 		VpcName:         "sanff",
// 		NodeGroupDetail: postgres.Jsonb{},
// 		OrganizationID:  1,
// 	}
// 	tClusteRequest.ID = 1
// 	server.Mock.ExpectBegin()
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tClusteRequest.Name))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)

// 	data := NewClusterRequest()
// 	updatedcluster, err := data.EnableDisableCluster(server.DB, tClusteRequest)
// 	if err != nil {
// 		t.Errorf("this is the error updating the cluster: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, updatedcluster.Name, updatedcluster.Name)
// }
