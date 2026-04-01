package models

import (
	"html"
	"strings"

	"github.com/jinzhu/gorm"
)

type ResetPassword struct {
	gorm.Model
	Email string `gorm:"size:100;not null;" json:"email"`
	Token string `gorm:"size:255;not null;" json:"token"`
}

type ResetPasswordInterface interface {
	SaveDatails(db *gorm.DB, data *ResetPassword) (*ResetPassword, error)
	DeleteDetails(db *gorm.DB, data *ResetPassword) (int64, error)
}
type ResetPasswordType struct {
}

func NewResetPassword() ResetPasswordInterface {
	return &ResetPasswordType{}
}

func (resetPassword *ResetPassword) Prepare() {
	resetPassword.Token = html.EscapeString(strings.TrimSpace(resetPassword.Token))
	resetPassword.Email = html.EscapeString(strings.TrimSpace(resetPassword.Email))
}

func (d *ResetPasswordType) SaveDatails(db *gorm.DB, resetPassword *ResetPassword) (*ResetPassword, error) {
	err = db.Create(&resetPassword).Error
	if err != nil {
		return &ResetPassword{}, err
	}
	return resetPassword, nil
}

func (d *ResetPasswordType) DeleteDetails(db *gorm.DB, resetPassword *ResetPassword) (int64, error) {
	db = db.Model(&ResetPassword{}).Where("id = ?", resetPassword.ID).Take(&ResetPassword{}).Delete(&ResetPassword{})
	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
