package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestLoadBalancer_Find(t *testing.T) {
	loadbalancer := LoadBalancer{
		Name:         "good",
		CustomDomain: "aaa",
	}
	loadbalancer.ID = 1
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "remarks"}).AddRow(1, time.Now(), time.Now(), "good"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewLoadBalancer()
	fpc, err := data.Find(server.DB, uint64(loadbalancer.ID))
	if err != nil {
		t.Errorf("this is the error getting one deduction: %v\n", err)
		return
	}
	assert.NotNil(t, fpc)
	assert.NoError(t, err)
}

// func TestLoadBalancerCreate(t *testing.T) {
// 	loadbalancer := &LoadBalancer{
// 		Name:         "good",
// 		CustomDomain: "aaa",
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	//saved, err := server.store.CreateDeduction(deduction)
// 	data := NewLoadBalancer()
// 	saved, err := data.Save(server.DB, loadbalancer)
// 	if err != nil {
// 		t.Errorf("this is the error getting the LoadBalancer: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Name, loadbalancer.Name)
// 	assert.Equal(t, saved.CustomDomain, loadbalancer.CustomDomain)
// }

func TestFindAllByProject(t *testing.T) {
	var testloadbalancer *[]LoadBalancer
	var err error
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "custom_domain"}).AddRow(1, "dd", "ff"))

	data := NewLoadBalancer()
	testloadbalancer, err = data.FindAllByProject(server.DB, 1)
	if err != nil {
		t.Errorf("this is the error getting the users: %v\n", err)
		return
	}
	assert.Equal(t, len(*testloadbalancer), 1)
}

//	func TestFindAllLoadBalancerByProject(t *testing.T) {
//		loadbalancer := LoadBalancer{
//			Action:    "aa",
//			Module:    "dd",
//			Remarks:   "Nepal",
//			ProjectID: 1,
//			UserID:    1,
//		}
//		loadbalancer.ID = 1
//		var testloadbalancer *[]LoadBalancer
//		var err error
//		server.Mock.ExpectQuery(regexp.QuoteMeta(
//			`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, "aa", "dd", "Nepal", 1, 1))
//		server.Mock.ExpectQuery(regexp.QuoteMeta(
//			`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "dd", "ff"))
//		server.Mock.ExpectQuery(regexp.QuoteMeta(
//			`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name", "last_name"}).AddRow(1, "dd", "ff"))
//		data := NewLoadBalancer()
//		testloadbalancer, err = data.FindAllLoadBalancerByProject(server.DB, loadbalancer.ProjectID, 20, 2, loadbalancer.Action, "", "", int(loadbalancer.UserID))
//		if err != nil {
//			t.Errorf("this is the error getting the users: %v\n", err)
//			return
//		}
//		assert.Equal(t, len(*testloadbalancer), 1)
//	}
func TestUpdateLoadBalancer(t *testing.T) {
	loadbalancer := &LoadBalancer{
		Name:         "good",
		CustomDomain: "aaa",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "custom_domain"}).AddRow(1, loadbalancer.Name, loadbalancer.CustomDomain))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "custom_domain"}).AddRow(1, "good", "db"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	loadbalancer.ID = 1
	data := NewLoadBalancer()
	updatedloadbalancer, err := data.Update(server.DB, loadbalancer)
	if err != nil {
		t.Errorf("this is the error updating the loadbalancer: %v\n", err)
		return
	}
	assert.Equal(t, updatedloadbalancer.Name, loadbalancer.Name)
	assert.Equal(t, updatedloadbalancer.CustomDomain, loadbalancer.CustomDomain)
}

func TestDeleteLoadBalancer(t *testing.T) {
	loadbalancer := &LoadBalancer{
		Name:         "good",
		CustomDomain: "aaa",
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "custom_domain"}).AddRow(1, loadbalancer.Name, loadbalancer.CustomDomain))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)

	data := NewLoadBalancer()
	id, err := data.Delete(server.DB, uint64(loadbalancer.ID))
	if err != nil {
		t.Errorf("this is the error updating the user: %v\n", err)
		return
	}
	assert.Equal(t, int64(id), int64(1))
}

// func TestInvoice_Find(t *testing.T) {
// 	loadbalancer := LoadBalancer{
// 		Action:    "aa",
// 		Module:    "dd",
// 		Remarks:   "Nepal",
// 		ProjectID: 1,
// 		UserID:    1,
// 	}
// 	loadbalancer.ID = 1
// 	project := Project{
// 		Name: "bbb",
// 	}
// 	project.ID = 1
// 	user := User{
// 		FirstName: "bish",
// 	}
// 	user.ID = 1
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "action", "module", "remarks", "project_id", "user_id"}).AddRow(1, loadbalancer.Action, loadbalancer.Module, loadbalancer.Remarks, loadbalancer.ProjectID, loadbalancer.UserID))
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, project.Name))
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "first_name"}).AddRow(1, user.FirstName))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewLoadBalancer()
// 	fpc, err := data.FindAllLoadBalancerByProject(server.DB, loadbalancer.ProjectID, 20, 2, loadbalancer.Action, "", "", int(loadbalancer.UserID))
// 	if err != nil {
// 		t.Errorf("this is the error getting the users: %v\n", err)
// 		return
// 	}
// 	assert.NotNil(t, fpc)
// 	assert.NoError(t, err)
// }
