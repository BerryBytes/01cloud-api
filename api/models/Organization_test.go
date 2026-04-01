package models

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// func TestOrganization_Find(t *testing.T) {
// 	organization := Organization{
// 		Name:   "aa",
// 		Domain: "ben.com",
// 		UserID: 1,
// 		Members: []*OrganizationMembers{
// 			{

// 				UserID:         1,
// 				OrganizationID: 1,
// 			},
// 		},
// 		Plugins: []*Plugin{
// 			{
// 				Name:        "plu",
// 				Description: "plugins all",
// 				IsAddOn:     false,
// 				AddOns: []*Plugin{
// 					{
// 						Name: "plu",
// 					},
// 				},
// 			},
// 		},
// 	}
// 	organization.ID = 1
// 	organization.Members[0].ID = 1
// 	organization.Plugins[0].ID = 1

// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name", "domain"}).AddRow(1, time.Now(), time.Now(), "aa", "ben.com"))
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "user_id", "organization_id"}).AddRow(1, time.Now(), time.Now(), 1, 1))
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "organizatio_id", "plugin_id"}).AddRow(1, time.Now(), time.Now(), 1, 1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewOrganization()
// 	fpc, err := data.Find(server.DB, organization.ID)
// 	if err != nil {
// 		t.Errorf("this is the error getting one organization: %v\n", err)
// 		return
// 	}
// 	assert.NotNil(t, fpc)
// 	assert.NoError(t, err)
// }

// func TestOrganizationCreate(t *testing.T) {
// 	organization := Organization{
// 		Name:        "aa",
// 		Description: "good",
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewOrganization()
// 	saved, err := data.Save(server.DB, &organization)
// 	if err != nil {
// 		t.Errorf("this is the error getting the Organization: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Name, organization.Name)
// 	assert.Equal(t, saved.Description, organization.Description)
// }

// func TestFindAllOrganization(t *testing.T) {
// 	var testorganization *[]Organization
// 	var err error
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "dd", "good"))

// 	data := NewOrganization()
// 	testorganization, err = data.FindAll(server.DB, uint(1), "")
// 	if err != nil {
// 		t.Errorf("this is the error getting the users: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, len(*testorganization), 1)
// }

// // // func TestFindAllOrganizationByProject(t *testing.T) {
// // // 	organization := Organization{
// // // 		Action:    "aa",
// // // 		Module:    "dd",
// // // 		Remarks:   "Nepal",
// // // 		ProjectID: 1,
// // // 		UserID:    1,
// // // 	}
// // // 	organization.ID = 1
// // // 	var testorganization *[]Organization
// // // 	var err error
// // // 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// // // 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, "dd", "ff", "gg", 1, 1))
// // // 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// // // 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, "dd", "ff", "gg", 1, 1))
// // // 	data := NewOrganization(server.DB)
// // // 	testorganization, err = data.FindAllOrganizationByProject(organization.ProjectID, 20, 2, organization.Action, "", "", int(organization.UserID))
// // // 	if err != nil {
// // // 		t.Errorf("this is the error getting the users: %v\n", err)
// // // 		return
// // // 	}
// // // 	assert.Equal(t, len(*testorganization), 1)
// // // }
func TestUpdateOrganization(t *testing.T) {
	organization := Organization{
		Name:        "aa",
		Description: "good",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, organization.Name, organization.Description))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "aa", "good"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	organization.ID = 1
	data := NewOrganization()
	updatedorganization, err := data.Update(server.DB, &organization)
	if err != nil {
		t.Errorf("this is the error updating the organization: %v\n", err)
		return
	}
	assert.Equal(t, updatedorganization.Name, organization.Name)
	assert.Equal(t, updatedorganization.Description, organization.Description)
}

// func TestDeleteOrganization(t *testing.T) {
// 	organization := Organization{
// 		Name:        "aa",
// 		Description: "good",
// 	}
// 	server.Mock.ExpectBegin()
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, organization.Name, organization.Description))
// 	server.Mock.ExpectExec(regexp.QuoteMeta(
// 		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)

// 	data := NewOrganization()
// 	id, err := data.Delete(server.DB, int64(organization.ID))
// 	if err != nil {
// 		t.Errorf("this is the error updating the organization: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, int64(id), int64(1))
// }

// // func TestInvoice_Find(t *testing.T) {
// // 	organization := Organization{
// // 		Name:      "aa",
// // 		Memory:    512,
// // 		DiskSpace: 200,
// // 		Apps:      1,
// // 	}
// // 	organization.ID = 1
// // 	user.ID = 1
// // 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, organization.Action, organization.Module, organization.Remarks, organization.ProjectID, organization.UserID))
// // 	server.Mock.ExpectCommit()
// // 	server.Mock.MatchExpectationsInOrder(false)
// // 	data := NewOrganization(server.DB)
// // 	fpc, err := data.FindAllOrganizationByProject(organization.ProjectID, 20, 2, organization.Action, "", "", int(organization.UserID))
// // 	if err != nil {
// // 		t.Errorf("this is the error getting the users: %v\n", err)
// // 		return
// // 	}
// // 	assert.NotNil(t, fpc)
// // 	assert.NoError(t, err)
// // }
