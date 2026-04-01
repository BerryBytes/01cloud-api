package models

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestChartVersionFind(t *testing.T) {
	chartVersion := ChartVersion{
		Version: "1.25",
	}
	chartVersion.ID = "1"
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "version"}).AddRow("1", "1.25"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewChartVersion()
	cv, err := data.Find(server.DB, chartVersion.ID)
	if err != nil {
		t.Errorf("this is the error getting Chart Version: %v\n", err)
		return
	}
	assert.NotNil(t, cv)
	assert.NoError(t, err)
}

func TestChartVersionBeforeCreate(t *testing.T) {
	testCase := struct {
		name string
		data *ChartVersion
	}{
		name: "ChartVersion_Test",
		data: &ChartVersion{
			ID: "1",
		},
	}
	tc := testCase
	t.Run(tc.name, func(t *testing.T) {
		// assert.Error(tc.data.BeforeCreate(server.DB), err)
	})
}
