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

var imageRegistry = ImageRegistry{
	Name:           "test",
	Provider:       "aws",
	Service:        "test-service",
	UserName:       "test-username",
	OrganizationID: 1,
	Password:       "pass",
	Active:         true,
	Credentials:    postgres.Jsonb{},
	ProjectName:    "pro-name",
}

func TestSaveImageRegistry(t *testing.T) {
	for _, data := range testCondition {
		if data == "fail" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnError(err)
		} else {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(false)
		datas := NewImageRegistry()
		saved, err := datas.Save(server.DB, &imageRegistry)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NotNil(t, saved)
		assert.NoError(t, err)
	}

}

func TestIsNameExistImageRegistry(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewImageRegistry()
	result := datas.IsNameExists(server.DB, uint(imageRegistry.OrganizationID), imageRegistry.Name)
	assert.NotNil(t, result)
	assert.Equal(t, reflect.TypeOf(result).Kind(), reflect.Bool)
}

func TestImageRegistryFindAll(t *testing.T) {
	imageRegistry.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), imageRegistry.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewImageRegistry()
	imageRegistryData, err := data.FindAll(server.DB, &imageRegistry)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*imageRegistryData), 1)
}

func TestImageRegistryWithInactive(t *testing.T) {
	imageRegistry.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), imageRegistry.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewImageRegistry()
	imageRegistryData, err := data.FindAllWithInactive(server.DB, &imageRegistry)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*imageRegistryData), 1)
}

func TestFindImageRegistry(t *testing.T) {
	imageRegistry.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewImageRegistry()
	cls, err := datas.Find(server.DB, uint64(cluster.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)
}

func TestImageRegistryUpdate(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`UPDATE`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewImageRegistry()
	cls, err := datas.Update(server.DB, &imageRegistry)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)
}

func TestDeleteImageRegistry(t *testing.T) {
	imageRegistry.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(true)
	datas := NewImageRegistry()
	_, err := datas.Delete(server.DB, uint64(imageRegistry.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NoError(t, err)

}

func TestPrepareImageRegistry(t *testing.T) {
	imageRegistry.Prepare()
}

func TestValidateImageRegistry(t *testing.T) {
	testcases := []struct {
		Name          string
		ImageRegistry *ImageRegistry
		Err           error
	}{
		{
			Name:          "CLUSTER_OK",
			ImageRegistry: &imageRegistry,
			Err:           nil,
		},
		{
			Name: "EMPTY_NAME",
			ImageRegistry: &ImageRegistry{
				Name: "",
			},
			Err: errors.New("required name"),
		},
		{
			Name: "EMPTY_PROVIDER",
			ImageRegistry: &ImageRegistry{
				Name:     "Test",
				Provider: "",
			},
			Err: errors.New("required provider"),
		},
	}
	for _, tc := range testcases {
		err := tc.ImageRegistry.Validate()
		assert.Equal(t, err, tc.Err)
	}
}
