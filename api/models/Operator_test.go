package models

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var testCondition = []string{
	"ok",
	"fail",
}

var operator = &Operator{
	Name:        "test name ",
	PackageName: "test packageName",
	ThumbUrl:    "test.com",
	Provider:    "aws",
}

func TestSaveOperator(t *testing.T) {
	for _, data := range testCondition {
		if data == "fail" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnError(err)
		} else {
			server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		}
		server.Mock.ExpectCommit()
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(false)
		datas := NewOperatorRepo()
		saved, err := datas.Save(server.DB, operator)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NotNil(t, saved)
		assert.NoError(t, err)
	}
}

func TestEnableDisable(t *testing.T) {
	operator.ID = 1
	for _, data := range testCondition {
		server.Mock.ExpectBegin()
		if data == "ok" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "packageName", "provider", "thumbUrl"}).AddRow(operator.ID, operator.Name, operator.PackageName, operator.Provider, operator.ThumbUrl))
			server.Mock.ExpectExec(regexp.QuoteMeta(
				`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "packageName", "provider", "thumbUrl"}).AddRow(operator.ID, operator.Name, operator.PackageName, operator.Provider, operator.ThumbUrl))
		}
		server.Mock.ExpectCommit()
		server.Mock.MatchExpectationsInOrder(false)
		datas := NewOperatorRepo()
		err := datas.EnableDisable(server.DB, operator.PackageName, false)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.Nil(t, err)
	}

}

func TestUpdateOperator(t *testing.T) {
	server.Mock.ExpectQuery(regexp.QuoteMeta(`UPDATE`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewOperatorRepo()
	err := datas.UpdateOperator(server.DB, operator)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NoError(t, err)

}

func TestFindAllOperator(t *testing.T) {
	operator.ID = 1

	for _, data := range testCondition {
		if data == "ok" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "packageName", "provider", "thumbUrl"}).AddRow(operator.ID, operator.Name, operator.PackageName, operator.Provider, operator.ThumbUrl))
		}
	}
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewOperatorRepo()
	operator, err := data.FindAllOperator(server.DB, true, uint64(1), uint64(10))
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, len(operator), 2)

}

func TestFindOperator(t *testing.T) {
	operator.ID = 1
	for _, data := range testCondition {
		if data == "ok" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "packageName", "provider", "thumbUrl"}).AddRow(operator.ID, operator.Name, operator.PackageName, operator.Provider, operator.ThumbUrl))
		}
	}
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewOperatorRepo()
	_, err := data.FindOperator(server.DB, operator.Name)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Nil(t, err)

}

func TestSyncOperator(t *testing.T) {
	operator.ID = 1
	for _, data := range testCondition {
		server.Mock.ExpectBegin()
		server.Mock.MatchExpectationsInOrder(false)
		if data == "ok" {
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "packageName", "provider", "thumbUrl"}).AddRow(operator.ID, operator.Name, operator.PackageName, operator.Provider, operator.ThumbUrl))
			server.Mock.ExpectExec(regexp.QuoteMeta(
				`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
			server.Mock.ExpectQuery(regexp.QuoteMeta(
				`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "packageName", "provider", "thumbUrl"}).AddRow(operator.ID, operator.Name, operator.PackageName, operator.Provider, operator.ThumbUrl))
		} else {
			server.Mock.MatchExpectationsInOrder(true)
			server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnError(err)

		}
		server.Mock.ExpectCommit()
		datas := NewOperatorRepo()
		err := datas.SyncOperator(server.DB, operator)
		if err != nil {
			assert.Error(t, err)
			return
		}
		assert.NoError(t, err)
	}
}
