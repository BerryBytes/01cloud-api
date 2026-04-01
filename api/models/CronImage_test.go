package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var CronImageTest = &CronImage{
	Name:      "test_name",
	ImageName: "test_image",
	Version:   "latest",
	Active:    true,
}

func TestPrepareCronImage(t *testing.T) {
	CronImageTest.Prepare()
}

func TestValidateCronImage(t *testing.T) {
	testcases := []struct {
		Name      string
		CronImage *CronImage
		Err       error
	}{
		{
			Name: "EMPTY_NAME",
			CronImage: &CronImage{
				Name: "",
			},
			Err: errors.New("required name"),
		},
		{
			Name: "EMPTY_IMAGE_NAME",
			CronImage: &CronImage{
				Name:      "test",
				ImageName: "",
			},
			Err: errors.New("required image name"),
		},
		{
			Name: "NO_ERROR",
			CronImage: &CronImage{
				Name:      "test",
				ImageName: "test_image",
			},
		},
	}
	for _, tc := range testcases {
		err := tc.CronImage.Validate()
		assert.Equal(t, err, tc.Err)
	}
}

// func TestSaveCronImage(t *testing.T) {
// 	CronImageTest.ID = 1
// 	server.Mock.ExpectBegin()
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewCronImage()
// 	_, err := data.Save(server.DB, CronImageTest)
// 	if err != nil {
// 		t.Errorf("this is the error saving CronImage: %v\n", err)
// 		return
// 	}
// 	assert.NoError(t, err)
// }

func TestFindAllCronImage(t *testing.T) {
	var testCronImage *[]CronImage
	var err error
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, CronImageTest.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCronImage()
	testCronImage, err = data.FindAll(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the CronImage: %v\n", err)
		return
	}
	assert.Equal(t, len(*testCronImage), 1)
}

func TestFindAllwithInactiveCronImage(t *testing.T) {
	var testCronImage *[]CronImage
	var err error
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, CronImageTest.Name))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCronImage()
	testCronImage, err = data.FindAllWithInactive(server.DB)
	if err != nil {
		t.Errorf("this is the error getting the inactive CronImage: %v\n", err)
		return
	}
	assert.Equal(t, len(*testCronImage), 1)
}

func TestFindCronImage(t *testing.T) {
	CronImageTest.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "test"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCronImage()
	fpc, err := data.Find(server.DB, uint64(CronImageTest.ID))
	if err != nil {
		t.Errorf("this is the error getting one CronImage: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

func TestUpdateCronImage(t *testing.T) {
	test := &CronImage{
		Name:      "new_name",
		ImageName: "new_image",
		Version:   "1.25",
		Active:    false,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "image_name"}).AddRow(1, test.Name, test.ImageName))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "image_name"}).AddRow(1, test.Name, test.ImageName))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	CronImageTest.ID = 1
	data := NewCronImage()
	newdata, err := data.Update(server.DB, CronImageTest)
	if err != nil {
		t.Errorf("this is the error updating the CronImage: %v\n", err)
		return
	}
	assert.Equal(t, CronImageTest.ImageName, newdata.ImageName)
	assert.Equal(t, CronImageTest.Name, newdata.Name)
}

func TestDeleteCronImage(t *testing.T) {
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, CronImageTest.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewCronImage()
	CronImageTest.ID = 1
	id, err := data.Delete(server.DB, uint64(CronImageTest.ID))
	if err != nil {
		t.Errorf("this is the error deleting the CronImage: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}
