package models

import "github.com/jinzhu/gorm"

type UserInviteInterface interface {
	Save(db *gorm.DB, data UserInvite) (*UserInvite, error)
	FindAll(db *gorm.DB) (*[]UserInvite, error)
	Find(db *gorm.DB, pid uint64) (*UserInvite, error)
	FindByEmail(db *gorm.DB, email string) (*UserInvite, error)
	FindByToken(db *gorm.DB, token string) (*UserInvite, error)
	Update(db *gorm.DB, data UserInvite) (*UserInvite, error)
	Delete(db *gorm.DB, uid uint64) (int64, error)
}
type UserInviteType struct {
}

func NewUserInvite() UserInviteInterface {
	return &UserInviteType{}
}
