package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var pluginCategory = &PluginCategory{
	Name:        "test-name",
	Description: "test-desc",
	IsAddOn:     true,
}

var pluginModel = &Plugin{
	Name:        "Apache",
	Description: "Plugin_test",
	MinCpu:      500,
	MinMemory:   500,
	SourceUrl:   "dev.com",
}

func TestSavePluginCategory(t *testing.T) {
	for _, data := range testCondition {
		if data == "fail" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnError(err)
		} else {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(false)
		datas := NewPluginCategory()
		saved, err := datas.Save(server.DB, pluginCategory)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NotNil(t, saved)
		assert.NoError(t, err)
	}
}

func TestFindAllPluginCategory(t *testing.T) {
	pluginCategory.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), pluginCategory.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewPluginCategory()
	getDatas, err := datas.FindAll(server.DB, "true", "test-query")
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*getDatas), 1)
}

func TestFindPliginCategory(t *testing.T) {
	pluginCategory.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), pluginCategory.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewPluginCategory()
	getDatas, err := datas.Find(server.DB, uint64(pluginCategory.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, getDatas)
}

func TestFindCategoryByIds(t *testing.T) {
	pluginCategory.ID = 1
	ids := []string{"10", "20"}

	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), pluginCategory.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewPluginCategory()
	getDatas, err := datas.FindCategoryByIds(server.DB, ids)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(getDatas), 1)

}

func TestUpdatePluginCategory(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`UPDATE`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewPluginCategory()
	cls, err := datas.Update(server.DB, pluginCategory)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, cls)
	assert.NoError(t, err)
}

func TestDeletePluginCategory(t *testing.T) {
	pluginCategory.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, pluginCategory.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewPluginCategory()
	row, err := data.Delete(server.DB, int64(pluginCategory.ID))
	if err != nil {
		t.Errorf("this is the error deleting pluginCategory: %v\n", err)
		return
	}
	assert.NotNil(t, row)
}

func TestPluginsDelete(t *testing.T) {
	pluginCategory.ID = 1
	pluginModel.ID = 1

	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, pluginCategory.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPluginCategory()
	err := data.DeletePlugins(server.DB, tPlugin)
	if err != nil {
		assert.Error(t, err)
		return
	}
}

func TestPrearedPluginCategory(t *testing.T) {
	pluginCategory.Prepare()
}
func TestPluginCategoryValidate(t *testing.T) {
	testcases := []struct {
		Name           string
		PluginCategory *PluginCategory
		Err            error
	}{
		{
			Name:           "CLUSTER_OK",
			PluginCategory: pluginCategory,
			Err:            nil,
		},
		{
			Name: "EMPTY_NAME",
			PluginCategory: &PluginCategory{
				Name: "",
			},
			Err: errors.New("required name"),
		},
		{
			Name: "INVALID_NAME",
			PluginCategory: &PluginCategory{
				Name: "test@##$()",
			},
			Err: errors.New("allowed alphanumeric, underscore, hyphen and space only"),
		},
	}
	for _, tc := range testcases {
		err := tc.PluginCategory.Validate()
		assert.Equal(t, err, tc.Err)
	}
}
