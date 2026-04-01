package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var cronJobTest = &CronJob{
	EnvironmentID: 1,
	Name:          "cronjob_test",
	Image:         "image_test",
	UserID:        1,
}

func TestPrepareCronJob(t *testing.T) {
	cronJobTest.Prepare()
}

func TestValidateCronJob(t *testing.T) {
	testcases := []struct {
		Name    string
		CronJob *CronJob
		Err     error
	}{
		{
			Name: "EMPTY_NAME",
			CronJob: &CronJob{
				Name: "",
			},
			Err: errors.New("required name"),
		},
		{
			Name: "INVALID_NAME",
			CronJob: &CronJob{
				Name: "test@123",
			},
			Err: errors.New("allowed alphanumeric, underscore, hyphen and space only"),
		},
		{
			Name: "EMPTY_COMMAND",
			CronJob: &CronJob{
				Name:    "test",
				Command: "",
			},
			Err: errors.New("required command"),
		},
		{
			Name: "DATA_SHEDHULE",
			CronJob: &CronJob{
				Name:     "test",
				Command:  "echo test",
				Schedule: "* * * * *",
			},
		},
		{
			Name: "DATA_SHEDHULE_ERROR",
			CronJob: &CronJob{
				Name:     "test",
				Command:  "echo test",
				Schedule: "schedule_error",
			},
			Err: errors.New("expected exactly 5 fields, found 1: [schedule_error]"),
		},
	}
	for _, tc := range testcases {
		err := tc.CronJob.Validate()
		assert.Equal(t, err, tc.Err)
	}
}

func TestVerfyResourceCronJob(t *testing.T) {
	testcases := []struct {
		Name        string
		CronJob     *CronJob
		Environment *Environment
		Err         error
	}{
		{
			Name:    "NO_ERROR",
			CronJob: &CronJob{},
			Environment: &Environment{
				CronJob: []*CronJob{
					{
						Name: "test",
					},
				},
				Application: &Application{
					Project: &Project{
						Subscription: &Subscription{
							CronJob: 2,
						},
					},
				},
			},
		},
		{
			Name:    "LIMIT_ERROR",
			CronJob: &CronJob{},
			Environment: &Environment{
				CronJob: []*CronJob{
					{
						Name: "test_limit",
					},
				},
				Application: &Application{
					Project: &Project{
						Subscription: &Subscription{
							CronJob: 1,
						},
					},
				},
			},
			Err: errors.New("cronJob quota limit exceed"),
		},
		{
			Name: "NAME_ERROR",
			CronJob: &CronJob{
				Name: "same_name",
			},
			Environment: &Environment{
				CronJob: []*CronJob{
					{
						Name: "same_name",
					},
				},
				Application: &Application{
					Project: &Project{
						Subscription: &Subscription{
							CronJob: 2,
						},
					},
				},
			},
			Err: errors.New("name already exists"),
		},
	}
	for _, tc := range testcases {
		err := tc.CronJob.VerifyResource(tc.Environment)
		assert.Equal(t, err, tc.Err)
	}

}

// func TestSaveCronJob(t *testing.T) {
// 	cronData := &CronJob{
// 		Name: "test",
// 	}
// 	server.Mock.ExpectBegin()
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectExec(regexp.QuoteMeta(
// 		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewCronJob()
// 	_, err := data.SaveCronJob(server.DB, cronData, 1)
// 	if err != nil {
// 		t.Errorf("this is the error saving CronJob: %v\n", err)
// 		return
// 	}
// 	assert.NoError(t, err)
// }

func TestFindCronJob(t *testing.T) {
	cronJobTest.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "test"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCronJob()
	fpc, err := data.Find(server.DB, uint64(cronJobTest.ID))
	if err != nil {
		t.Errorf("this is the error getting one CronJob: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

func TestFindAllWithFiltersCronJob(t *testing.T) {
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, cronJobTest.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCronJob()
	testCronJob, err := data.FindAllWithFilters(server.DB, cronJobTest.EnvironmentID, 1, 1)
	if err != nil {
		t.Errorf("this is the error getting the cron job with filters: %v\n", err)
		return
	}
	assert.Equal(t, len(*testCronJob), 1)
}

func TestUpdateCronJob(t *testing.T) {
	test := &CronJob{
		Name:  "new_name",
		Image: "new_image",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "image"}).AddRow(1, test.Name, test.Image))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "image"}).AddRow(1, test.Name, test.Image))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	cronJobTest.ID = 1
	data := NewCronJob()
	newdata, err := data.Update(server.DB, cronJobTest)
	if err != nil {
		t.Errorf("this is the error updating the CronJob: %v\n", err)
		return
	}
	assert.Equal(t, cronJobTest.Name, newdata.Name)
	assert.Equal(t, cronJobTest.Image, newdata.Image)
}
