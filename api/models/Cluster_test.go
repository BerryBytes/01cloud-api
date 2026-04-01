package models

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

var cluster = &Cluster{
	Name:                "test1",
	ConfigPath:          "test/path",
	Context:             "test-context",
	Token:               "testadkfjakfmdkkjk",
	Region:              "us-east",
	Provider:            "test Provider",
	PrometheusServerUrl: "pormuthes url",
	ImageRegistryID:     1,
	PvCapacity:          2,
	Zone:                "east",
	Nodes:               11,
	Labels:              "ajdkajd",
	Active:              true,
	OrganizationID:      1,
	ProjectName:         "test project name",
	Attributes:          "att",
	DNSId:               12,
	CloudStorage:        postgres.Jsonb{},
	Weight:              1,
	ClusterRequestID:    2,
	Color:               "red",
}

func TestClusterSave(t *testing.T) {
	for _, data := range testCondition {
		if data == "fail" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnError(err)
		} else {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(false)
		datas := NewCluster()
		saved, err := datas.Save(server.DB, *cluster)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NotNil(t, saved)
		assert.NoError(t, err)
	}
}

func TestClusterFindAll(t *testing.T) {
	cluster.ID = 1

	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), cluster.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCluster()
	cluster, err := data.FindAll(server.DB)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*cluster), 1)
}

func TestFindAllClustersByRegion(t *testing.T) {
	cluster.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), cluster.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCluster()
	cluster, err := data.FindAllClustersByRegion(server.DB, uint(cluster.OrganizationID), cluster.Region)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*cluster), 1)

}

func TestFindAllWithInActive(t *testing.T) {
	cluster.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), cluster.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCluster()
	cluster, err := data.FindAllWithInActive(server.DB)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*cluster), 1)

}

// func TestFindRegions(t *testing.T) {
// 	cluster.ID = 1
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "active"}).AddRow(1, time.Now(), time.Now(), cluster.Name, true))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewCluster()
// 	cluster, err := data.FindRegions(server.DB, uint(cluster.OrganizationID))
// 	if err != nil {
// 		assert.Error(t, err)
// 		return
// 	}
// 	assert.Equal(t, len(*cluster), 1)
// }

func TestFindAllRegions(t *testing.T) {
	cluster.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "active"}).AddRow(1, time.Now(), time.Now(), cluster.Name, true))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCluster()
	cluster, err := data.FindAllRegions(server.DB, uint(cluster.OrganizationID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*cluster), 0)

}

func TestFindCluster(t *testing.T) {
	cluster.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewCluster()
	cls, err := datas.Find(server.DB, uint64(cluster.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)

}

func TestFindWithDns(t *testing.T) {
	cluster.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewCluster()
	cls, err := datas.FindWithDns(server.DB, uint64(cluster.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)

}

func TestFindZeroneCluster(t *testing.T) {

	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewCluster()
	cls, err := datas.FindZeroneCluster(server.DB)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)

}

func TestFindAllClusterWithOrganizationId(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "active"}).AddRow(1, time.Now(), time.Now(), cluster.Name, true))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewCluster()
	cls, err := datas.FindAllClusterWithOrganizationId(server.DB, uint64(cluster.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*cls), 1)

}

func TestFindWithDetails(t *testing.T) {

	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "active"}).AddRow(1, time.Now(), time.Now(), cluster.Name, true))

	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewCluster()
	cls, err := datas.FindWithDetails(server.DB, uint64(cluster.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)

}

func TestFindByRegion(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "active"}).AddRow(1, time.Now(), time.Now(), cluster.Name, true))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewCluster()
	cls, err := datas.FindByRegion(server.DB, cluster.Region, uint(cluster.OrganizationID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)

}

// func TestFindByRegionWithDns(t *testing.T) {
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	datas := NewCluster()
// 	cls, err := datas.FindByRegionWithDns(server.DB, cluster.Region, uint(cluster.OrganizationID))
// 	if err != nil {
// 		assert.Error(t, err)
// 		return
// 	}
// 	assert.NotNil(t, cls)
// 	assert.NoError(t, err)

// }

func TestClusterUpdate(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`UPDATE`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewCluster()
	cls, err := datas.Update(server.DB, *cluster)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)

}

func TestClusterDelete(t *testing.T) {
	for _, data := range testCondition {
		if data == "fail" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`DELETE`)).WillReturnError(err)
		} else {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "active"}).AddRow(1, time.Now(), time.Now(), cluster.Name, true))
			server.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(true)
		datas := NewCluster()
		_, err := datas.Delete(server.DB, uint64(cluster.ID))
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NoError(t, err)
	}

}

func TestDeleteByClusterRequest(t *testing.T) {
	for _, data := range testCondition {
		if data == "fail" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`DELETE`)).WillReturnError(err)
		} else {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "active"}).AddRow(1, time.Now(), time.Now(), cluster.Name, true))
			server.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(false)
		datas := NewCluster()
		_, err := datas.DeleteByClusterRequest(server.DB, uint64(cluster.ID))
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NoError(t, err)
	}

}
func TestUpdateLabelsAndColor(t *testing.T) {
	for _, data := range testCondition {
		if data == "fail" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`UPDATE`)).WillReturnError(err)
		} else {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`UPDATE`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(true)
		datas := NewCluster()
		saved, err := datas.UpdateLabelsAndColor(server.DB, *cluster)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NotNil(t, saved)
		assert.NoError(t, err)
	}

}

func TestPrepared(t *testing.T) {
	testcases := []struct {
		Name    string
		Cluster *Cluster
		cvalue  string
		Err     error
	}{
		{
			Name:    "CLUSTER_OK",
			Cluster: cluster,
			Err:     nil,
		},
		{
			Name: "EMPTY_NAME",
			Cluster: &Cluster{
				Name: "",
			},
			Err: errors.New("required name"),
		},
		{
			Name: "EMPTY_CONFIG",
			Cluster: &Cluster{
				ConfigPath: "",
				Name:       cluster.Name,
			},
			Err: errors.New("config file is required"),
		},
		{
			Name: "EMPTY_PROMETHES",
			Cluster: &Cluster{
				PrometheusServerUrl: "",
				Name:                cluster.Name,
				ConfigPath:          cluster.ConfigPath,
			},
			Err: errors.New("prometheus server url  is required"),
		},
		{
			Name: "EMPTY_PV",
			Cluster: &Cluster{
				PvCapacity:          0,
				Name:                cluster.Name,
				ConfigPath:          cluster.ConfigPath,
				PrometheusServerUrl: cluster.PrometheusServerUrl,
			},
			Err: errors.New("number of pv is required"),
		},
	}
	for _, tc := range testcases {
		err := tc.Cluster.Validate()
		assert.Equal(t, err, tc.Err)
	}

}

func TestIsClusterNameExists(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "active"}).AddRow(1, time.Now(), time.Now(), cluster.Name, true))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := cluster
	result := datas.IsNameExists(server.DB)
	assert.NotNil(t, result)
	assert.Equal(t, reflect.TypeOf(result).Kind(), reflect.Bool)
}
