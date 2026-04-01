package models

import (
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestOrgPlanPrepare(t *testing.T) {
	testCases := []struct {
		name string
		data *OrganizationPlan
	}{
		{
			name: "ORG_PLAN_PREPARE_CASE",
			data: &OrganizationPlan{
				Name:     "test",
				Cluster:  50,
				Memory:   5000,
				Cores:    1000,
				NoOfUser: 5,
				Price:    10,
				Weight:   10,
				Active:   true,
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			tc.data.Prepare()
		})
	}
}

func TestValidateOrganizationPlan(t *testing.T) {
	testcases := []struct {
		Name    string
		OrgPlan *OrganizationPlan
		Err     error
	}{
		{
			Name: "OK",
			OrgPlan: &OrganizationPlan{
				Name:     "test",
				Cluster:  50,
				Memory:   5000,
				Cores:    1000,
				NoOfUser: 5,
				Price:    10,
				Weight:   10,
				Active:   true,
			},
			Err: nil,
		},
		{
			Name:    "EMPTY_NAME",
			OrgPlan: &OrganizationPlan{},
			Err:     errors.New("required Name"),
		},
		{
			Name: "EMPTY_NAME",
			OrgPlan: &OrganizationPlan{
				Name:    "&*tybg",
				Cluster: 78,
			},
			Err: errors.New("allowed alphanumeric, underscore, hyphen and space only"),
		},
		{
			Name: "CLUSTER REQUIRED",
			OrgPlan: &OrganizationPlan{
				Name: "test",
			},
			Err: errors.New("required number of Clusters"),
		},
		{
			Name: "NUMBER_OF_USER_REQUIRED",
			OrgPlan: &OrganizationPlan{
				Name:    "test",
				Cluster: 40,
			},
			Err: errors.New("required No of Users"),
		},
		{
			Name: "MEMORY_REQUIRED",
			OrgPlan: &OrganizationPlan{
				Name:     "test",
				Cluster:  40,
				NoOfUser: 10,
			},
			Err: errors.New("required Memory"),
		},
		{
			Name: "CORE_REQUIRED",
			OrgPlan: &OrganizationPlan{
				Name:     "test",
				Cluster:  40,
				NoOfUser: 10,
				Memory:   20000,
			},
			Err: errors.New("required Cores"),
		},
	}
	for _, tc := range testcases {
		err := tc.OrgPlan.Validate()
		assert.Equal(t, err, tc.Err)
	}

}

// func TestOrganizationPlanCreate(t *testing.T) {
// 	orgPlan := OrganizationPlan{
// 		Name:    "aa",
// 		Cluster: 50,
// 		Memory:  2000,
// 		Cores:   1000,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.ExpectBegin()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewOrganizationPlan()
// 	saved, err := data.Save(server.DB, &orgPlan)
// 	if err != nil {
// 		t.Errorf("this is the error getting the Organization plan: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, saved.Name, orgPlan.Name)
// }

// func TestOrgPlanFind(t *testing.T) {
// 	orgPlan := OrganizationPlan{
// 		Name:    "aa",
// 		Cluster: 50,
// 		Memory:  2000,
// 		Cores:   1000,
// 	}
// 	orgPlan.ID = 1
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), orgPlan.Name))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewOrganizationPlan()
// 	organizationPlan, err := data.Find(server.DB, uint64(orgPlan.ID))
// 	if err != nil {
// 		t.Errorf("this is the error getting one organization plan: %v\n", err)
// 		return
// 	}
// 	assert.NotNil(t, organizationPlan)
// 	assert.NoError(t, err)
// }

// func TestFindAllOrgPlan(t *testing.T) {
// 	orgPlan := OrganizationPlan{
// 		Name:    "aa",
// 		Cluster: 50,
// 		Memory:  2000,
// 		Cores:   1000,
// 	}
// 	orgPlan.ID = 1
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), orgPlan.Name))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewOrganizationPlan()
// 	organizationPlan, err := data.FindAll(server.DB)
// 	if err != nil {
// 		t.Errorf("this is the error getting organization plan %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, len(*organizationPlan), 1)
// }

// func TestFindAllWithInActiveOrgPlan(t *testing.T) {
// 	orgPlan := OrganizationPlan{
// 		Name:    "aa",
// 		Cluster: 50,
// 		Memory:  2000,
// 		Cores:   1000,
// 	}
// 	orgPlan.ID = 1
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "name"}).AddRow(1, time.Now(), time.Now(), orgPlan.Name))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewOrganizationPlan()
// 	organizationPlan, err := data.FindAllWithInActive(server.DB)
// 	if err != nil {
// 		t.Errorf("this is the error getting active organization plan %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, len(*organizationPlan), 1)
// }

func TestUpdateOrganizationPlan(t *testing.T) {
	orgPlan := OrganizationPlan{
		Name:    "aa",
		Cluster: 50,
		Memory:  2000,
		Cores:   1000,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "cluster"}).AddRow(1, orgPlan.Name, orgPlan.Cluster))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description"}).AddRow(1, "aa", "good"))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	orgPlan.ID = 1
	data := NewOrganizationPlan()
	updatedorganization, err := data.Update(server.DB, &orgPlan)
	if err != nil {
		t.Errorf("this is the error updating the organization plan: %v\n", err)
		return
	}
	assert.Equal(t, updatedorganization.Name, orgPlan.Name)
}

// func TestDeleteOrganizationPlan(t *testing.T) {
// 	orgPlan := OrganizationPlan{
// 		Name:    "aa",
// 		Cluster: 50,
// 		Memory:  2000,
// 		Cores:   1000,
// 	}
// 	orgPlan.ID = 1
// 	server.Mock.ExpectBegin()
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(
// 		`SELECT`)).WillReturnRows(sqlmock.NewRows([]string{"id", "name", "cluster"}).AddRow(1, orgPlan.Name, orgPlan.Cluster))
// 	server.Mock.ExpectExec(regexp.QuoteMeta(
// 		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	data := NewOrganizationPlan()
// 	id, err := data.Delete(server.DB, uint64(orgPlan.ID))
// 	if err != nil {
// 		t.Errorf("this is the error updating the user: %v\n", err)
// 		return
// 	}
// 	assert.Equal(t, int64(id), int64(1))
// }
