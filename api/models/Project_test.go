package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestProjectToJSON(t *testing.T) {
	testCases := []struct {
		name       string
		data       *Project
		errMessage error
	}{
		{
			name: "PROJECT_JSON_CASE",
			data: &Project{
				Name:        "test-pro",
				Description: "hello project",
			},
			errMessage: nil,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			_, err = tc.data.ToJson()
			assert.Equal(t, tc.errMessage, err)
		})
	}
}
func TestProjectPrepare(t *testing.T) {
	testCases := []struct {
		name       string
		data       *Project
		errMessage error
	}{
		{
			name: "PROJECT_PREPARE_CASE",
			data: &Project{
				Name:        "test-pro",
				Description: "  hello project  ",
			},
			errMessage: nil,
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			tc.data.Prepare()
		})
	}
}

func TestProjectValidate(t *testing.T) {
	testCases := []struct {
		name       string
		data       *Project
		pid        uint
		errMessage error
	}{
		{
			name: "VALID_CASE",
			data: &Project{
				Name:           "test-pro",
				SubscriptionID: 1,
			},
			pid:        1,
			errMessage: nil,
		},
		{
			name: "REQUIRE_NAME",
			data: &Project{
				Name: "",
			},
			pid:        0,
			errMessage: errors.New("required name "),
		},
		{
			name: "REQUIRE_SUBCRIPTION",
			data: &Project{
				Name:           "Test-Project",
				SubscriptionID: 0,
			},
			pid:        0,
			errMessage: errors.New("required subscription "),
		},
		{
			name: "NAME_LENGTH_LESS_THAN_64",
			data: &Project{
				Name:           "ttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt",
				SubscriptionID: 1,
			},
			pid:        1,
			errMessage: errors.New("name must not exceed 64 characters"),
		},
		{
			name: "INVALID_CHARACTER",
			data: &Project{
				Name:           "tttttt,&",
				SubscriptionID: 1,
			},
			pid:        1,
			errMessage: errors.New("invalid characters in name"),
		},
		{
			name: "DESCRIPTION_LENGTH_ERROR",
			data: &Project{
				Name:           "tttttt",
				SubscriptionID: 1,
				Description:    "Sample Testing means the analyses to be performed by either Party using the applicable Samples, as described in the Sample Testing Schedule.means those results arising from the Sample Testing which are to be shared between Lilly and Sponsor, as set forth in the Sample Testing Schedule. means the schedule attached hereto as Appendix B. means, with respect to a given Compound, the set of requirements for such Compound as set forth in the Quality Agreement. has the meaning set forth in the preamble. means any one or more bi- or multi- specific molecule that agonizes CD137 (4-1BB). means cinrebafusp alfa (PRS-343), a bivalent, bispecific fusion protein targeting CD137 (4-1BB) and HER2 excluding, however, any biosimilar of PRS-343 other than a biosimilar owned or controlled by Sponsor or its Affiliates. CONFIDENTIAL Pieris Study Certain confidential information contained in this document, marked by brackets, has been omitted because the information (I) is not material and (II) would be competitively harmful if publicly disclosed.",
			},
			pid:        1,
			errMessage: errors.New("description must not exceed 1024 characters"),
		},
		{
			name: "PROJECT_CODE_LENGTH_ERROR",
			data: &Project{
				Name:           "tttttt",
				SubscriptionID: 1,
				ProjectCode:    "HELLO123",
			},
			pid: 1,

			errMessage: errors.New("project code must not exceed 5 characters"),
		},
		{
			name: "REGION_LENGTH_LESS_THAN_64",
			data: &Project{
				Name:           "test-pro",
				Region:         "ttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt",
				SubscriptionID: 1,
				ProjectCode:    "1234",
			},
			pid:        1,
			errMessage: errors.New("region must not exceed 64 characters"),
		},
		{
			name: "BASEDOMAIN_LENGTH_LESS_THAN_64",
			data: &Project{
				Name:           "test-pro",
				BaseDomain:     "ttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttttt",
				SubscriptionID: 1,
				ProjectCode:    "1234",
			},
			pid:        1,
			errMessage: errors.New("base domain must not exceed 64 characters"),
		},
		{
			name: "INVALID_BASEDOMAIN",
			data: &Project{
				Name:           "test-pro",
				BaseDomain:     "t???**&.com",
				SubscriptionID: 1,
				ProjectCode:    "1234",
			},
			pid:        1,
			errMessage: errors.New("invalid domain"),
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

func TestProjectCreate(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewIProject()
	saved, err := data.Save(server.DB, tProject)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.Equal(t, saved.UserID, tProject.UserID)
}

func TestFind(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
	}
	tProject.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "project_test"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewIProject()
	project, err := data.Find(server.DB, uint64(tProject.ID))
	if err != nil {
		t.Errorf("this is the error getting one project: %v\n", err)
		return
	}
	assert.NotNil(t, project)
	assert.NoError(t, err)
}

func TestIsNameExists(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
	}
	tProject.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), "project_test"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewIProject()
	isExist := data.IsNameExists(server.DB, uint(tProject.UserID), "project_test", tProject)
	if err != nil {
		t.Errorf("this is the error getting exist name: %v\n", err)
		return
	}
	assert.NotNil(t, isExist)
	assert.NoError(t, err)
}

func TestFindAllProject(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
		OrganizationId: 1,
	}
	tProject.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "organization_id"}).AddRow(1, time.Now(), time.Now(), "project_test", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewIProject()
	project, err := data.FindAll(server.DB, tProject)
	if err != nil {
		t.Errorf("this is the error getting one project: %v\n", err)
		return
	}
	assert.Equal(t, len(*project), 1)
}

func TestFindAllByUser(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
		OrganizationId: 1,
	}
	tProject.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "organization_id"}).AddRow(1, time.Now(), time.Now(), "project_test", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewIProject()
	projects, err := data.FindAllByUser(server.DB, uint(tProject.UserID), tProject)
	if err != nil {
		t.Errorf("this is the error getting all users project: %v\n", err)
		return
	}
	assert.Equal(t, len(projects), 1)
}

func TestFindUserProjectOnly(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
		OrganizationId: 1,
	}
	tProject.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "organization_id"}).AddRow(1, time.Now(), time.Now(), "project_test", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewIProject()
	projects, err := data.FindUserProjectOnly(server.DB, uint(tProject.UserID), uint(tProject.OrganizationId))
	if err != nil {
		t.Errorf("this is the error getting users only  project: %v\n", err)
		return
	}
	assert.Equal(t, len(projects), 1)
}

// func TestFindSearchProject(t *testing.T) {
// 	tProject := &Project{
// 		Name:           "project_test",
// 		Description:    "project_test is purposed to test",
// 		SubscriptionID: 1,
// 		UserID:         1,
// 		Active:         true,
// 		OrganizationId: 1,
// 	}
// 	tProject.ID = 1
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "user_id"}).AddRow(1, time.Now(), time.Now(), "project_test", 1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewIProject()
// 	projects, err := data.SearchProject(server.DB, uint(tProject.UserID), "test")
// 	if err != nil {
// 		t.Errorf("this is the error getting users only  project: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, len(projects), 1)
// }

func TestFindOrgProject(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
		OrganizationId: 1,
	}
	tProject.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "organization_id"}).AddRow(1, time.Now(), time.Now(), "project_test", 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewIProject()
	project, err := data.FindProject(server.DB, uint(tProject.OrganizationId))
	if err != nil {
		t.Errorf("this is the error getting one project: %v\n", err)
		return
	}
	assert.Equal(t, len(*project), 1)
}

func TestChangeIsActive(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
		OrganizationId: 1,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tProject.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewIProject()
	err := data.ChangeIsActive(server.DB, tProject, false)
	if err != nil {
		t.Errorf("this is the error  project status change: %v\n", err)
		return
	}
	assert.NoError(t, err)
}

func TestActiveDeactiveAl(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
		OrganizationId: 1,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tProject.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewIProject()
	err := data.ActiveDeactiveAll(server.DB, tProject, uint64(tProject.UserID), false)
	if err != nil {
		t.Errorf("this is the error project active and deactive: %v\n", err)
		return
	}
	assert.NoError(t, err)
}

func TestDeleteProject(t *testing.T) {
	tProject := &Project{
		Name:           "project_test",
		Description:    "project_test is purposed to test",
		SubscriptionID: 1,
		UserID:         1,
		Active:         true,
		OrganizationId: 1,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, tProject.Name))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewIProject()
	row, err := data.DeleteProjectOwner(server.DB, uint64(tProject.UserID), tProject.OrganizationId)
	if err != nil {
		t.Errorf("this is the error deleting project: %v\n", err)
		return
	}
	assert.NotNil(t, row)
}
