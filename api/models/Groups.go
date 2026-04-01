package models

import (
	"errors"
	"html"
	"strings"

	"github.com/jinzhu/gorm"
)

type Group struct {
	gorm.Model
	Name           string        `gorm:"type:varchar(255)"`
	Description    string        `gorm:"type:varchar(255)"`
	Organization   *Organization `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationID uint64        `gorm:"default:0" json:"organization_id"`
	Members        []*User       `gorm:"many2many:group_members;association_jointable_foreignkey:user_id"`
}

type GroupRepo struct{}

type GroupInterface interface {
	Save(db *gorm.DB, data *Group) (*Group, error)
	FindAll(db *gorm.DB, userID uint64, query string) (*[]Group, error)
	Find(db *gorm.DB, gid uint64, oid uint) (*Group, error)
	Update(db *gorm.DB, data *Group) (*Group, error)
	Delete(db *gorm.DB, id int64) (int64, error)
	AddMember(db *gorm.DB, user *User, data *Group) error
	DeleteMember(db *gorm.DB, user *User, data *Group) error
}

func NewGroupRepo() GroupInterface {
	return &GroupRepo{}
}

func (data *Group) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
}

func (data *Group) Validate() error {
	if data.Name == "" {
		return errors.New("required Name")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	return nil
}

func (r *GroupRepo) Save(db *gorm.DB, data *Group) (*Group, error) {
	err = db.Model(&Group{}).Create(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *GroupRepo) FindAll(db *gorm.DB, organizationId uint64, query string) (*[]Group, error) {
	dataList := &[]Group{}
	err = db.Model(&Group{}).
		Preload("Members").
		Where("organization_id = ? ", organizationId).
		Where("lower(name) LIKE lower(?)", "%"+query+"%").
		Find(dataList).Error
	if err != nil {
		return &[]Group{}, err
	}
	return dataList, nil
}

func (r *GroupRepo) Find(db *gorm.DB, gid uint64, oid uint) (*Group, error) {
	data := &Group{}
	err = db.Model(&Group{}).
		Where("id = ?", gid).
		Where("organization_id = ?", oid).
		Preload("Members").
		Take(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *GroupRepo) Update(db *gorm.DB, data *Group) (*Group, error) {
	mp := map[string]interface{}{}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Description != "" {
		mp["description"] = data.Description
	}
	err = db.Model(&Group{}).
		Where("id = ?", data.ID).
		UpdateColumn(mp).Error
	if err != nil {
		return &Group{}, err
	}
	return data, nil
}

func (r *GroupRepo) Delete(db *gorm.DB, id int64) (int64, error) {
	data := &Group{}
	db = db.Model(&data).
		Where("id = ?", id).
		Take(&Group{}).
		Delete(&Group{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Group not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (r *GroupRepo) AddMember(db *gorm.DB, user *User, data *Group) error {
	findUser := *user
	err = db.Model(data).Association("Members").Find(&findUser).Error
	if err == nil {
		return errors.New("member already added")
	}
	err = db.Model(data).Association("Members").Append([]User{*user}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *GroupRepo) DeleteMember(db *gorm.DB, user *User, data *Group) error {
	err = db.Model(data).Association("Members").Delete(&user).Error
	if err != nil {
		return err
	}
	return nil
}
