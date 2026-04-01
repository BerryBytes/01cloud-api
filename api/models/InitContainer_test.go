package models

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

var intiContainerModel = InitContainer{
	EnvironmentID: 1,
	Image:         "test-image",
	Command:       "test-command",
	Name:          "test-name",
}

func TestSaveInitContainer(t *testing.T) {
	pid := 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	datas := NewInitContainer()
	saved, err := datas.SaveInitContainer(server.DB, uint64(pid), &intiContainerModel)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, saved)
	assert.NoError(t, err)

}

func TestInitContainerPrepared(t *testing.T) {
	intiContainerModel.Prepare()
}

func TestInitContainerValidate(t *testing.T) {
	testcases := []struct {
		Name          string
		InitContainer *InitContainer
		Err           error
	}{
		{
			Name:          "CLUSTER_OK",
			InitContainer: &intiContainerModel,
			Err:           nil,
		},
		{
			Name: "EMPTY_NAME",
			InitContainer: &InitContainer{
				Name: "",
			},
			Err: errors.New("required name"),
		},
		{
			Name: "INVALID_NAME",
			InitContainer: &InitContainer{
				Name: "test@##$()",
			},
			Err: errors.New("allowed alphanumeric, underscore, hyphen and space only"),
		},
		{
			Name: "EMPTY_NAME",
			InitContainer: &InitContainer{
				Name:    "test-name",
				Command: "",
			},
			Err: errors.New("required command"),
		},
	}
	for _, tc := range testcases {
		err := tc.InitContainer.Validate()
		assert.Equal(t, err, tc.Err)
	}
}

func TestInitContainerUpdate(t *testing.T) {
	intiContainerModel.ID = 1
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "command", "image"}).AddRow(1, intiContainerModel.Name, intiContainerModel.Command, intiContainerModel.Image))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "command", "image"}).AddRow(1, intiContainerModel.Name, intiContainerModel.Command, intiContainerModel.Image))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewInitContainer()
	updatedactivity, err := data.Update(server.DB, &intiContainerModel)
	if err != nil {
		t.Errorf("this is the error updating the initContainer: %v\n", err)
		return
	}
	assert.Equal(t, updatedactivity.Name, intiContainerModel.Name)
}

func TestFindAllWithFilter(t *testing.T) {
	intiContainerModel.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "command", "image"}).AddRow(1, intiContainerModel.Name, intiContainerModel.Command, intiContainerModel.Image))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewInitContainer()
	datas, err := data.FindAllWithFilters(server.DB, intiContainerModel.EnvironmentID, 1, uint64(1))
	if err != nil {
		t.Errorf("this is the error getting the inticontainer: %v\n", err)
		return
	}
	assert.Equal(t, len(*datas), 1)
}

func TestInitContainerFind(t *testing.T) {
	intiContainerModel.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "command", "image"}).AddRow(1, intiContainerModel.Name, intiContainerModel.Command, intiContainerModel.Image))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewInitContainer()
	datas, err := data.Find(server.DB, uint64(intiContainerModel.ID))
	if err != nil {
		t.Errorf("this is the error getting the inticontainer: %v\n", err)
		return
	}
	assert.NotEmpty(t, datas)
}
