package models

import "github.com/jinzhu/gorm"

type UserRoleInterface interface {
	Save(db *gorm.DB, data *UserRole) (*UserRole, error)
	FindAll(db *gorm.DB) (*[]UserRole, error)
	Find(db *gorm.DB, pid uint64) (*UserRole, error)
	Update(db *gorm.DB, data *UserRole) (*UserRole, error)
	Delete(db *gorm.DB, uid uint64) (int64, error)
}
type UserRoleType struct {
}

func NewUserRole() UserRoleInterface {
	return &UserRoleType{}
}
