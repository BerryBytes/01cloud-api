package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

var tPlugin = &Plugin{
	Model: gorm.Model{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	Name:          "Apache",
	Description:   "Plugin_test",
	MinCpu:        500,
	MinMemory:     500,
	SourceUrl:     "dev.com",
	Image:         "test-image",
	Attributes:    "test-attribute",
	ServiceDetail: addons.ServiceDetail,
	AddOns: []*Plugin{
		{
			Name:        "Plugin-2",
			Description: "plugin 2 test",
			Categories: []*PluginCategory{
				{
					Name:        "plugincat",
					Description: "plugin category test",
				},
			},
		},
	},
	Categories: []*PluginCategory{
		{
			Name:        "plugincat",
			Description: "plugin category test",
		},
	},
}
var addons = &Plugin{
	Model: gorm.Model{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	Name:        "Plugin-2",
	Description: "plugin 2 test",
}
var categories = &PluginCategory{
	Model: gorm.Model{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	Name:        "plugincat",
	Description: "plugin category test",
}

func TestPluginPrepare(t *testing.T) {
	// data := &Plugin{
	// 	Name:        "test-app",
	// 	ProjectID:   1,
	// 	PluginID:    1,
	// 	ClusterID:   1,
	// 	ServiceType: 2,
	// }
	tPlugin.Prepare()
}

func TestPluginValidate(t *testing.T) {
	testCases := []struct {
		name       string
		data       *Plugin
		pid        uint
		errMessage error
	}{
		{
			name: "VALID_CASE",
			data: &Plugin{
				Name:        "test-app",
				Description: "aasssd",
				SourceUrl:   "aa",
				MinCpu:      1,
				MinMemory:   2,
			},
			errMessage: nil,
		},
		{
			name: "REQUIRE_NAME",
			data: &Plugin{
				Name:        "",
				Description: "aasssd",
				SourceUrl:   "aa",
				MinCpu:      1,
				MinMemory:   2,
			},
			errMessage: errors.New("required name"),
		},
		{
			name: "REQUIRE_NAME",
			data: &Plugin{
				Name:        "asb!&&",
				Description: "aasssd",
				SourceUrl:   "aa",
				MinCpu:      1,
				MinMemory:   2,
			},
			errMessage: errors.New("allowed alphanumeric, underscore, hyphen and space only"),
		},
		{
			name: "VALID_NAME_REQUIRE",
			data: &Plugin{
				Name:        "asb",
				Description: "",
				SourceUrl:   "aa",
				MinCpu:      1,
				MinMemory:   2,
			},
			errMessage: errors.New("required description"),
		},
		{
			name: "SOURCE_REQUIRE",
			data: &Plugin{
				Name:        "asb",
				Description: "aasssd",
				SourceUrl:   "",
				MinCpu:      1,
				MinMemory:   2,
			},
			errMessage: errors.New("required source url"),
		},
		{
			name: "MINCPU_REQUIRE",
			data: &Plugin{
				Name:        "asb",
				Description: "aasssd",
				SourceUrl:   "aa",
				MinCpu:      0,
				MinMemory:   2,
			},
			errMessage: errors.New("required minimum cpu"),
		},
		{
			name: "MINMEMORY_REQUIRE",
			data: &Plugin{
				Name:        "asb",
				Description: "aasssd",
				SourceUrl:   "aa",
				MinCpu:      1,
				MinMemory:   0,
			},
			errMessage: errors.New("required minimum memory"),
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

func TestCreatePlugin(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPlugin()
	saved, err := data.Save(server.DB, tPlugin)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, saved.Name, tPlugin.Name)
}

func TestFindPlugin(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "Plugin_test"))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(addons.ID, time.Now(), time.Now(), addons.Name))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(categories.ID, time.Now(), time.Now(), categories.Name))

	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPlugin()
	Plugin, err := data.Find(server.DB, uint64(tPlugin.ID))
	if err != nil {
		t.Errorf("this is the error getting one Plugin: %v\n", err)
		return
	}
	assert.NotNil(t, Plugin)
	assert.NoError(t, err)
}

func TestFindAllPlugin(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(tPlugin.ID, time.Now(), time.Now(), tPlugin.Name))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(addons.ID, time.Now(), time.Now(), addons.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPlugin()
	Plugin, err := data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting one Plugin: %v\n", err)
		return
	}
	assert.Equal(t, len(*Plugin), 1)
}

func TestDeletePlugin(t *testing.T) {
	tPlugin.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tPlugin.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewPlugin()
	row, err := data.Delete(server.DB, uint64(tPlugin.ID))
	if err != nil {
		t.Errorf("this is the error deleting Plugin: %v\n", err)
		return
	}
	assert.NotNil(t, row)
}

func TestFindPluginByIds(t *testing.T) {
	// plugins:=&[]Plugin{
	// 	{
	// 		Name: "test",
	// 		Description: "test describe",
	// 	},
	// }
	ids := []string{"1"}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(tPlugin.ID, time.Now(), time.Now(), tPlugin.Name))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(addons.ID, time.Now(), time.Now(), addons.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPlugin()
	Plugin, err := data.FindPluginByIds(server.DB, ids)
	if err != nil {
		t.Errorf("this is the error getting one Plugin: %v\n", err)
		return
	}
	assert.Equal(t, len(Plugin), 1)
}

// func TestFindBySupportCi(t *testing.T) {
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(tPlugin.ID, time.Now(), time.Now(), tPlugin.Name))
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(categories.ID, time.Now(), time.Now(), categories.Name))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewPlugin()
// 	org := []uint{
// 		0,
// 	}
// 	for _, oid := range org {
// 		Plugin, _, err := data.FindBySupportCi(server.DB, true, oid, 0, 0, "", "", "")
// 		if err != nil {
// 			t.Errorf("this is the error getting one Plugin: %v\n", err)
// 			return
// 		}
// 		assert.Equal(t, len(*Plugin), 1)
// 	}

// // }
// func TestFindAllActiveWithFilters(t *testing.T) {
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(tPlugin.ID, time.Now(), time.Now(), tPlugin.Name))
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(categories.ID, time.Now(), time.Now(), categories.Name))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewPlugin()
// 	Plugin, _, err := data.FindAllActiveWithFilters(server.DB, 0, 0, "", "", "")
// 	if err != nil {
// 		t.Errorf("this is the error getting one Plugin: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, len(*Plugin), 1)
// }

func TestFindAllWithFilters(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(tPlugin.ID, time.Now(), time.Now(), tPlugin.Name))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(categories.ID, time.Now(), time.Now(), categories.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPlugin()
	Plugin, _, err := data.FindAllWithFilters(server.DB, 0, 0, "", "", "")
	if err != nil {
		t.Errorf("this is the error getting one Plugin: %v\n", err)
		return
	}
	assert.Equal(t, len(*Plugin), 1)
}

func TestFindAddons(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(tPlugin.ID, time.Now(), time.Now(), tPlugin.Name))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(addons.ID, time.Now(), time.Now(), addons.Name))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(categories.ID, time.Now(), time.Now(), categories.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPlugin()
	catid := []string{}
	Plugin, err := data.FindAddOns(server.DB, 0, "", catid)
	if err != nil {
		t.Errorf("this is the error getting one Plugin: %v\n", err)
		return
	}
	assert.Equal(t, len(Plugin), 0)
}

func TestUpdatePlugin(t *testing.T) {
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tPlugin.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "aa"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	//storage.ID = 1
	data := NewPlugin()
	updatedplugin, err := data.Update(server.DB, tPlugin)
	if err != nil {
		t.Errorf("this is the error updating the storage: %v\n", err)
		return
	}
	assert.Equal(t, updatedplugin.Name, tPlugin.Name)
}

var pluginVersion = &PluginVersion{
	Model: gorm.Model{
		ID:        1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	Version:    "test-version",
	ChangeLogs: "test-change-log",
	Url:        "test",
	Active:     true,
}

func TestPluginVersionPrepared(t *testing.T) {
	pluginVersion.Prepare()
}

func TestPluginVersionValidate(t *testing.T) {
	testCases := []struct {
		name       string
		data       *PluginVersion
		errMessage error
	}{
		{
			name: "VALID_CASE",
			data: &PluginVersion{
				Version:    "test-app",
				ChangeLogs: "aasssd",
				Url:        "test url",
				PluginID:   1,
			},
			errMessage: nil,
		},
		{
			name:       "REQUIRE_VERSION",
			data:       &PluginVersion{},
			errMessage: errors.New("required version"),
		},
		{
			name: "REQUIRE_CHANGELOGS",
			data: &PluginVersion{
				Version: "testV1",
			},
			errMessage: errors.New("required changelogs"),
		},
		{
			name: "VALID_NAME_REQUIRE",
			data: &PluginVersion{
				Version:    "testV1",
				ChangeLogs: "chageLog",
			},
			errMessage: errors.New("required Url"),
		},
		{
			name: "SOURCE_REQUIRE",
			data: &PluginVersion{
				Version:    "testV1",
				ChangeLogs: "chageLog",
				Url:        "ajhfjad",
			},
			errMessage: errors.New("required plugin"),
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err = tc.data.Validate()
			assert.Equal(t, tc.errMessage, err)
		})
	}
}

func TestSavePluginVersion(t *testing.T) {
	for _, data := range testCondition {
		if data == "fail" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnError(err)
		} else {
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(false)
		data := NewPluginVersion()
		saved, err := data.Save(server.DB, pluginVersion)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.Equal(t, saved.Attributes, pluginVersion.Attributes)
	}
}

func TestFindAllPluginVersion(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPluginVersion()
	saved, err := data.FindAll(server.DB)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*saved), 1)

}

func TestFindAllWithInactivePluginVersion(t *testing.T) {

	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPluginVersion()
	saved, err := data.FindAllWithInactive(server.DB)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*saved), 1)

}

func TestFindAllByPluginPluginVersion(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPluginVersion()
	saved, err := data.FindAllByPlugin(server.DB, uint64(tPlugin.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(*saved), 1)

}

func TestFindAllByPluginVersion(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPluginVersion()
	saved, err := data.Find(server.DB, uint64(pluginVersion.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, saved)

}
func TestFindLatestPluginVersion(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPluginVersion()
	saved, err := data.FindLatestPluginVersion(server.DB, uint64(pluginVersion.ID))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, saved)

}

func TestUpdatePluginVersion(t *testing.T) {
	server.Mock.ExpectBegin()
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewPluginVersion()
	updatedplugin, err := data.Update(server.DB, pluginVersion)
	if err != nil {
		t.Errorf("this is the error updating the storage: %v\n", err)
		return
	}
	assert.Equal(t, updatedplugin.Attributes, pluginVersion.Attributes)
}

func TestDeletePluginVersion(t *testing.T) {
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewPluginVersion()
	row, err := data.Delete(server.DB, uint64(pluginVersion.ID))
	if err != nil {
		t.Errorf("this is the error deleting PluginVersion: %v\n", err)
		return
	}
	assert.NotNil(t, row)
}
