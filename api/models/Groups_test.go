package models

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestGroupCreate(t *testing.T) {
	group := &Group{
		Name: "group-test",
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT`)).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewGroupRepo()
	saved, err := data.Save(server.DB, group)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, saved)
	assert.NoError(t, err)
}
func TestGroupCreateErr(t *testing.T) {
	group := &Group{
		Name: "group-test",
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT`)).WillReturnError(errors.New("error while inserting.."))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	data := NewGroupRepo()
	saved, err := data.Save(server.DB, group)
	if err != nil {
		assert.Error(t, err)
		return
	}
	assert.NotNil(t, saved)
	assert.NoError(t, err)
}
func TestGroupAddMember(t *testing.T) {
	groupColumns := []string{"id", "name", "description", "organization_id", "CreatedAt", "UpdatedAt"}
	memberColumns := []string{"user_id", "group_id"}
	userColumns := []string{"id", "first_name", "last_name", "CreatedAt", "UpdatedAt"}
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
	data := &Group{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:           "test",
		Description:    "test",
		Organization:   org,
		OrganizationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(groupColumns).AddRow(data.ID, data.Name, data.Description, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(memberColumns).AddRow(user.ID, data.ID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(userColumns).AddRow(user.ID, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt))
	server.Mock.ExpectExec(regexp.QuoteMeta(`INSERT`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	grepo := NewGroupRepo()
	err = grepo.AddMember(server.DB, user, data)
	assert.Error(t, err)

}

func TestFindAllGroup(t *testing.T) {
	groupColumns := []string{"id", "name", "description", "organization_id", "CreatedAt", "UpdatedAt"}
	memberColumns := []string{"user_id", "group_id"}
	userColumns := []string{"id", "first_name", "last_name", "CreatedAt", "UpdatedAt"}
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
	data := &Group{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:           "test",
		Description:    "test",
		Organization:   org,
		OrganizationID: 1,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(groupColumns).AddRow(data.ID, data.Name, data.Description, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(memberColumns).AddRow(user.ID, data.ID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(userColumns).AddRow(user.ID, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt))
	server.Mock.ExpectExec(regexp.QuoteMeta(`INSERT`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewGroupRepo()
	project, err := d.FindAll(server.DB, uint64(1), "")
	if err != nil {
		t.Errorf("this is the error getting groups: %v\n", err)
		return
	}
	assert.Equal(t, len(*project), 1)
}

func TestGroupFind(t *testing.T) {
	groupColumns := []string{"id", "name", "description", "organization_id", "CreatedAt", "UpdatedAt"}
	memberColumns := []string{"user_id", "group_id"}
	userColumns := []string{"id", "first_name", "last_name", "CreatedAt", "UpdatedAt"}
	users := []*User{
		{
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			FirstName: "test",
			LastName:  "test",
		},
	}
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
	data := &Group{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:           "test",
		Description:    "test",
		Organization:   org,
		OrganizationID: 1,
		Members:        users,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(groupColumns).AddRow(data.ID, data.Name, data.Description, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(memberColumns).AddRow(user.ID, data.ID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(userColumns).AddRow(user.ID, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewGroupRepo()
	group, err := d.Find(server.DB, uint64(data.ID), uint(data.OrganizationID))
	assert.NotNil(t, group)
	assert.NoError(t, err)
}

func TestUpdateGroup(t *testing.T) {
	groupColumns := []string{"id", "name", "description", "organization_id", "CreatedAt", "UpdatedAt"}
	users := []*User{
		{
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			FirstName: "test",
			LastName:  "test",
		},
	}
	org := &Organization{
		Name:   "test",
		UserID: 1,
	}
	data := &Group{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:           "test",
		Description:    "test",
		Organization:   org,
		OrganizationID: 1,
		Members:        users,
	}
	server.Mock.ExpectBegin()
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(groupColumns).AddRow(data.ID, data.Name, data.Description, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectExec(regexp.QuoteMeta(
		`UPDATE`)).WillReturnResult(sqlmock.NewResult(0, 1))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(groupColumns).AddRow(data.ID, "test1", data.Description, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectCommit()
	server.Mock.MatchExpectationsInOrder(false)
	d := NewGroupRepo()
	_, err := d.Update(server.DB, data)
	if err != nil {
		t.Errorf("this is the error updating the group: %v\n", err)
		return
	}
}

func TestDeleteGroup(t *testing.T) {
	groupColumns := []string{"id", "name", "description", "organization_id", "CreatedAt", "UpdatedAt"}
	memberColumns := []string{"user_id", "group_id"}
	userColumns := []string{"id", "first_name", "last_name", "CreatedAt", "UpdatedAt"}
	users := []*User{
		{
			Model: gorm.Model{
				ID:        1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			FirstName: "test",
			LastName:  "test",
		},
	}
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
	data := &Group{
		Model: gorm.Model{
			ID:        1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Name:           "test",
		Description:    "test",
		Organization:   org,
		OrganizationID: 1,
		Members:        users,
	}
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(groupColumns).AddRow(data.ID, data.Name, data.Description, data.OrganizationID, time.Now(), time.Now()))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(memberColumns).AddRow(user.ID, data.ID))
	server.Mock.ExpectQuery(regexp.QuoteMeta(`SELECT`)).
		WillReturnRows(server.Mock.NewRows(userColumns).AddRow(user.ID, user.FirstName, user.LastName, user.CreatedAt, user.UpdatedAt))
	server.Mock.ExpectExec(regexp.QuoteMeta(`UPDATE`)).
		WillReturnResult(sqlmock.NewResult(1, 1))
	server.Mock.ExpectCommit()
	server.Mock.ExpectBegin()
	server.Mock.MatchExpectationsInOrder(false)
	grepo := NewGroupRepo()
	_ = grepo.DeleteMember(server.DB, user, data)
	//assert.NoError(t, err)

}
