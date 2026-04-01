package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCiConfigToJson(t *testing.T) {
	testCases := []struct {
		Name       string
		CiConfig   *CiConfig
		errMessage error
	}{
		{
			Name: "OK",
			CiConfig: &CiConfig{
				WebhookUrl:    "https://test.com/webhook",
				Emails:        "test@gmail.com",
				EnvironmentID: 1,
				EventType:     "push",
			},
			errMessage: nil,
		},
		{
			Name:       "ERROR",
			CiConfig:   &CiConfig{},
			errMessage: nil,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			data, err := tc.CiConfig.ToJson()
			if err != nil {
				assert.Equal(t, tc.errMessage, err)
				return
			}
			assert.NotNil(t, data)
		})
	}
}

func TestCiConfigPrepare(t *testing.T) {
	testCases := []struct {
		Name       string
		CiConfig   *CiConfig
		errMessage error
	}{
		{
			Name: "OK",
			CiConfig: &CiConfig{
				WebhookUrl:    "https://test.com/webhook",
				Emails:        "test@gmail.com",
				EnvironmentID: 1,
				EventType:     "push",
			},
			errMessage: nil,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			tc.CiConfig.Prepare()
		})
	}
}

func TestCiConfigValidate(t *testing.T) {
	testCases := []struct {
		Name       string
		CiConfig   *CiConfig
		errMessage error
	}{
		{
			Name: "NORMAL",
			CiConfig: &CiConfig{
				WebhookUrl:    "https://test.com/webhook",
				Emails:        "test@gmail.com",
				EnvironmentID: 1,
				EventType:     "normal",
			},
			errMessage: nil,
		},
		{
			Name: "ERROR",
			CiConfig: &CiConfig{
				WebhookUrl:    "https://test.com/webhook",
				Emails:        "test@gmail.com",
				EnvironmentID: 1,
				EventType:     "error",
			},
			errMessage: nil,
		},
		{
			Name: "ALL",
			CiConfig: &CiConfig{
				WebhookUrl:    "https://test.com/webhook",
				Emails:        "test@gmail.com",
				EnvironmentID: 1,
			},
			errMessage: nil,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.Name, func(t *testing.T) {
			tc.CiConfig.Validate()
		})
	}
}

func TestCiConfigCreate(t *testing.T) {
	ciConfig := &CiConfig{
		WebhookUrl:    "https://test.com/webhook",
		Emails:        "test@gmail.com",
		EnvironmentID: 1,
		EventType:     "push",
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCiConfigRepo()
	saved, err := data.Save(server.DB, ciConfig)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, saved)
}

func TestCiConfigCreateErr(t *testing.T) {
	ciConfig := &CiConfig{
		WebhookUrl:    "https://test.com/webhook",
		Emails:        "test@gmail.com",
		EnvironmentID: 1,
		EventType:     "push",
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT`)).WillReturnError(errors.New("error while inserting.."))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCiConfigRepo()
	saved, err := data.Save(server.DB, ciConfig)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, saved)
	assert.NoError(t, err)
}
func TestCiConfigFind(t *testing.T) {
	ciConfig := &CiConfig{
		WebhookUrl:    "https://test.com/webhook",
		Emails:        "test@gmail.com",
		EnvironmentID: 1,
		EventType:     "push",
	}
	ciConfig.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "webhook_url"}).AddRow(1, time.Now(), time.Now(), "https://test.com/webhook"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCiConfigRepo()
	ciConfig, err := data.Find(server.DB, uint64(ciConfig.ID))
	if err != nil {
		t.Errorf("this is the error getting one ciConfig: %v\n", err)
		return
	}
	assert.NotNil(t, ciConfig)
	assert.NoError(t, err)
}

func TestFindByEnvironment(t *testing.T) {
	ciConfig := &CiConfig{
		WebhookUrl:    "https://test.com/webhook",
		Emails:        "test@gmail.com",
		EnvironmentID: 1,
		EventType:     "push",
	}
	ciConfig.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "webhook_url"}).AddRow(1, time.Now(), time.Now(), "https://test.com/webhook"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCiConfigRepo()
	ciConfig, err := data.FindByEnvironment(server.DB, ciConfig.EnvironmentID)
	if err != nil {
		t.Errorf("this is the error getting one ciConfig: %v\n", err)
		return
	}
	assert.NotNil(t, ciConfig)
	assert.NoError(t, err)
}

func TestFindAllCiConfig(t *testing.T) {
	ciConfig := &[]CiConfig{
		{
			WebhookUrl:    "https://test.com/webhook",
			Emails:        "test@gmail.com",
			EnvironmentID: 1,
			EventType:     "push",
		},
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "webhook_url"}).AddRow(1, time.Now(), time.Now(), "https://test.com/webhook"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCiConfigRepo()
	ciConfigs, err := data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting ciConfigs: %v\n", err)
		return
	}
	assert.Equal(t, len(*ciConfigs), len(*ciConfig))
}

func TestUpdateCiConfig(t *testing.T) {
	ciConfig := &CiConfig{
		WebhookUrl:    "https://test.com/webhook",
		Emails:        "test@gmail.com",
		EnvironmentID: 1,
		EventType:     "push",
	}
	ciConfig.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "webhook_url", "environment_id", "emails"}).AddRow(1, ciConfig.WebhookUrl, ciConfig.EnvironmentID, ciConfig.Emails))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "webhook_url", "environment_id", "emails"}).AddRow(1, ciConfig.WebhookUrl, ciConfig.EnvironmentID, "testabc@gmail.com"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCiConfigRepo()
	updatedCiConfig, err := data.Update(server.DB, ciConfig)
	if err != nil {
		t.Errorf("this is the error updating the ciConfig: %v\n", err)
		return
	}
	assert.NotNil(t, updatedCiConfig)
}

func TestDeleteCiConfig(t *testing.T) {
	ciConfig := &CiConfig{
		WebhookUrl:    "https://test.com/webhook",
		Emails:        "test@gmail.com",
		EnvironmentID: 1,
		EventType:     "push",
	}
	ciConfig.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "webhook_url"}).AddRow(1, ciConfig.WebhookUrl))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCiConfigRepo()
	row, err := data.Delete(server.DB, uint64(ciConfig.ID))
	if err != nil {
		t.Errorf("this is the error deleting ciConfig: %v\n", err)
		return
	}
	assert.NotNil(t, row)
}
