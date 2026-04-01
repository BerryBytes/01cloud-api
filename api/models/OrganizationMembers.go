package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type UserRoleEnum int

const (
	Admin UserRoleEnum = 1 + iota
	Members
)

type OrganizationMembers struct {
	gorm.Model
	User           *User         `gorm:"foreignkey:UserID;null" json:"user"`
	UserID         uint64        `gorm:"default:0;" json:"user_id"`
	Organization   *Organization `gorm:"foreignkey:OrganizationID;null" json:"organization"`
	OrganizationID uint64        `gorm:"default:0;" json:"organization_id"`
	UserRole       UserRoleEnum  `gorm:"default:2;" json:"user_role"`
}

type RequestObject struct {
	Email string       `json:"email"`
	Role  UserRoleEnum `json:"role"`
}

type OrganizationMemberInterface interface {
	AddMember(db *gorm.DB, data *OrganizationMembers) error
	UpdateMember(db *gorm.DB, uid uint64, oid uint64, role UserRoleEnum) error
	DeleteMember(db *gorm.DB, uid uint64, oid uint64) error
	CheckLimit(db *gorm.DB, plan *OrganizationPlan, oid uint) bool
	CheckMember(db *gorm.DB, data *OrganizationMembers) bool
	IsRoleExist(db *gorm.DB, oid, uid uint, role UserRoleEnum) bool
}

type OrganizationMemberType struct{}

func NewOrganizationMember() OrganizationMemberInterface {
	return &OrganizationMemberType{}
}

func (d *OrganizationMemberType) AddMember(db *gorm.DB, data *OrganizationMembers) error {
	member := OrganizationMembers{}
	err = db.Model(&OrganizationMembers{}).
		Where("organization_id = ?", data.OrganizationID).
		Where("user_id = ?", data.UserID).
		Take(&member).Error
	if err == nil && member.ID > 0 {
		return errors.New("member already exists")
	}
	err = db.Model(&data).Save(&data).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *OrganizationMemberType) UpdateMember(db *gorm.DB, uid uint64, oid uint64, role UserRoleEnum) error {
	var members = OrganizationMembers{UserID: uid, OrganizationID: oid, UserRole: role}
	err = db.Model(&OrganizationMembers{}).
		Where("user_id = ? and organization_id = ?", uid, oid).
		Updates(members).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *OrganizationMemberType) DeleteMember(db *gorm.DB, uid uint64, oid uint64) error {
	err = db.Model(&OrganizationMembers{}).
		Delete(&OrganizationMembers{}, "user_id = ? and organization_id = ?", uid, oid).Error
	if err != nil {
		return err
	}
	return nil
}

func (d *OrganizationMemberType) CheckLimit(db *gorm.DB, plan *OrganizationPlan, oid uint) bool {
	var count int64
	db.Model(&OrganizationMembers{}).
		Where("organization_id = ? and deleted_at is null", oid).
		Count(&count)
	return count > int64(plan.NoOfUser)
}

func (d *OrganizationMemberType) CheckMember(db *gorm.DB, data *OrganizationMembers) bool {
	member := OrganizationMembers{}
	err = db.Model(&OrganizationMembers{}).
		Where("organization_id = ?", data.OrganizationID).
		Where("user_id = ?", data.UserID).
		Take(&member).Error
	if err == nil && member.ID > 0 {
		return true
	}
	return false
}

func (d *OrganizationMemberType) IsRoleExist(db *gorm.DB, oid, uid uint, role UserRoleEnum) bool {
	member := OrganizationMembers{}
	err = db.Model(&OrganizationMembers{}).
		Where("organization_id = ?", oid).
		Where("user_id = ?", uid).
		Where("user_role = ?", role).
		Take(&member).Error
	return err == nil
}
