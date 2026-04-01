package models

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestOrganizationAddMember(t *testing.T) {
	orgColumns := []string{"id", "user_id", "organization_id", "CreatedAt", "UpdatedAt"}
	user := &User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FirstName: "test",
		LastName:  "test",
	}
	org := &Organization{
		Name:   "test",
		UserID: 1,
	}
	data := &OrganizationMembers{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:         1,
		User:           user,
		Organization:   org,
		OrganizationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(orgColumns).AddRow(data.ID, data.UserID, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectExec(regexp.QuoteMeta(`INSERT`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	orgrepo := NewOrganizationMember()
	err = orgrepo.AddMember(server.DB, data)
	assert.Error(t, err)

}

func TestUpdateOrganizationMembers(t *testing.T) {
	orgColumns := []string{"id", "user_id", "organization_id", "CreatedAt", "UpdatedAt"}
	users := &User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FirstName: "test",
		LastName:  "test",
	}
	org := &Organization{
		Name:   "test",
		UserID: 1,
	}
	data := &OrganizationMembers{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Organization:   org,
		OrganizationID: 1,
		User:           users,
		UserID:         uint64(users.ID),
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(orgColumns).AddRow(data.ID, data.UserID, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(orgColumns).AddRow(data.ID, data.UserID, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewOrganizationMember()
	role := RequestObject{}
	err := d.UpdateMember(server.DB, data.UserID, data.OrganizationID, role.Role)
	if err != nil {
		t.Errorf("this is the error updating the organization members: %v\n", err)
		return
	}
}

// func TestCheckOrganizationMember(t *testing.T) {
// 	orgColumns := []string{"id", "user_id", "organization_id", "CreatedAt", "UpdatedAt"}
// 	user := &User{
// 		Model: gorm.Model{
// 			ID:        1,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		FirstName: "test",
// 		LastName:  "test",
// 	}
// 	org := &Organization{
// 		Name:   "test",
// 		UserID: 1,
// 	}
// 	data := &OrganizationMembers{
// 		Model: gorm.Model{
// 			ID:        1,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		UserID:         uint64(user.ID),
// 		User:           user,
// 		Organization:   org,
// 		OrganizationID: 1,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
// 		WillReturnRows(server.Mock.NewRows(orgColumns).AddRow(data.ID, data.UserID, data.OrganizationID, time.Now(), time.Now()))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	d := NewOrganizationMember()
// 	check := d.CheckMember(server.DB, data)
// 	assert.NotNil(t, check)
// 	assert.NoError(t, err)
// }

// func TestCheckLimitOrganizationMember(t *testing.T) {
// 	orgColumns := []string{"id", "user_id", "organization_id", "CreatedAt", "UpdatedAt"}
// 	user := &User{
// 		Model: gorm.Model{
// 			ID:        1,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		FirstName: "test",
// 		LastName:  "test",
// 	}
// 	org := &Organization{
// 		Name:   "test",
// 		UserID: 1,
// 		OrganizationPlan: &OrganizationPlan{
// 			Model: gorm.Model{
// 				ID:        1,
// 				CreatedAt: time.Now(),
// 				UpdatedAt: time.Now(),
// 			},
// 			Name: "testOrg",
// 		},
// 		OrganizationPlanID: 1,
// 	}
// 	data := &OrganizationMembers{
// 		Model: gorm.Model{
// 			ID:        1,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		UserID:         uint64(user.ID),
// 		User:           user,
// 		Organization:   org,
// 		OrganizationID: 1,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
// 		WillReturnRows(server.Mock.NewRows(orgColumns).AddRow(data.ID, data.UserID, data.OrganizationID, time.Now(), time.Now()))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	d := NewOrganizationMember()
// 	checkLimit := d.CheckLimit(server.DB, data.Organization.OrganizationPlan, data.Organization.ID)
// 	assert.NotNil(t, checkLimit)
// 	assert.NoError(t, err)
// }

// func TestCheckRoleOrganizationMember(t *testing.T) {
// 	orgColumns := []string{"id", "user_id", "organization_id", "CreatedAt", "UpdatedAt"}
// 	user := &User{
// 		Model: gorm.Model{
// 			ID:        1,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		FirstName: "test",
// 		LastName:  "test",
// 	}
// 	org := &Organization{
// 		Name:   "test",
// 		UserID: 1,
// 		OrganizationPlan: &OrganizationPlan{
// 			Model: gorm.Model{
// 				ID:        1,
// 				CreatedAt: time.Now(),
// 				UpdatedAt: time.Now(),
// 			},
// 			Name: "testOrg",
// 		},
// 		OrganizationPlanID: 1,
// 	}
// 	data := &OrganizationMembers{
// 		Model: gorm.Model{
// 			ID:        1,
// 			CreatedAt: time.Now(),
// 			UpdatedAt: time.Now(),
// 		},
// 		UserID:         uint64(user.ID),
// 		User:           user,
// 		Organization:   org,
// 		OrganizationID: 1,
// 	}
// 	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
// 		WillReturnRows(server.Mock.NewRows(orgColumns).AddRow(data.ID, data.UserID, data.OrganizationID, time.Now(), time.Now()))
// 	server.Mock.ExpectCommit()
// 	server.Mock.MatchExpectationsInOrder(false)
// 	d := NewOrganizationMember()
// 	role := RequestObject{}
// 	checkRole := d.IsRoleExist(server.DB, uint(data.OrganizationID), data.User.ID, role.Role)
// 	assert.NotNil(t, checkRole)
// 	assert.NoError(t, err)
// }

func TestDeleteOrganizationMembers(t *testing.T) {
	orgColumns := []string{"id", "user_id", "organization_id", "CreatedAt", "UpdatedAt"}
	user := &User{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FirstName: "test",
		LastName:  "test",
	}
	org := &Organization{
		Name:   "test",
		UserID: 1,
	}
	org.ID = 1
	data := &OrganizationMembers{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		UserID:         uint64(user.ID),
		User:           user,
		Organization:   org,
		OrganizationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(orgColumns).AddRow(data.ID, data.UserID, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	orgRepo := NewOrganizationMember()
	err = orgRepo.DeleteMember(server.DB, data.UserID, data.OrganizationID)
	assert.NoError(t, err)
}
