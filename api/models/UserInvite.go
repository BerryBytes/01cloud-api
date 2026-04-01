package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type UserInvite struct {
	gorm.Model
	FirstName   string `gorm:"size:255;null;" json:"first_name"`
	LastName    string `gorm:"size:255;null;" json:"last_name"`
	Email       string `gorm:"size:255;not null;unique" json:"email,omitempty"`
	Company     string `gorm:"size:255;null;" json:"company"`
	Designation string `gorm:"size:255;null;" json:"designation"`
	EmailSent   bool   `gorm:"not null;default:false" json:"email_sent"`
	Token       string `gorm:"size:255;null;" json:"token"`
	Remarks     string `gorm:"size:255;null;" json:"remarks"`
}

func (data *UserInvite) Prepare() {
	data.ID = 0
	data.Email = html.EscapeString(strings.TrimSpace(data.Email))
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *UserInvite) Validate() error {
	if data.Email == "" {
		return errors.New("required email ")
	}
	if err := CheckEmail(data.Email); err != nil {
		return errors.New("invalid email ")
	}

	return nil
}

func (d *UserInviteType) Save(db *gorm.DB, data UserInvite) (*UserInvite, error) {
	err := db.Model(&UserInvite{}).Create(&data).Error
	if err != nil {
		return &UserInvite{}, err
	}
	return &data, nil
}

func (d *UserInviteType) FindAll(db *gorm.DB) (*[]UserInvite, error) {
	var err error
	datas := []UserInvite{}
	err = db.Model(&UserInvite{}).Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]UserInvite{}, err
	}
	return &datas, nil
}

func (d *UserInviteType) FindByEmail(db *gorm.DB, email string) (*UserInvite, error) {
	var err error
	data := UserInvite{}
	err = db.Model(&UserInvite{}).Where("email = ?", email).Take(&data).Error
	if err != nil {
		return &UserInvite{}, err
	}
	return &data, nil
}

func (d *UserInviteType) FindByToken(db *gorm.DB, token string) (*UserInvite, error) {
	var err error
	data := UserInvite{}
	err = db.Model(&UserInvite{}).Where("token = ?", token).Take(&data).Error
	if err != nil {
		return &UserInvite{}, err
	}
	return &data, nil
}

func (d *UserInviteType) Find(db *gorm.DB, pid uint64) (*UserInvite, error) {
	var err error
	data := UserInvite{}
	err = db.Model(&UserInvite{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &UserInvite{}, err
	}
	return &data, nil
}

func (d *UserInviteType) Update(db *gorm.DB, data UserInvite) (*UserInvite, error) {
	var err error
	user := UserInvite{}
	if data.FirstName != "" {
		user.FirstName = data.FirstName
	}
	if data.LastName != "" {
		user.LastName = data.LastName
	}
	if data.Company != "" {
		user.Company = data.Company
	}
	if data.Designation != "" {
		user.Designation = data.Designation
	}
	if data.EmailSent {
		user.EmailSent = data.EmailSent
	}
	if data.Token != "" {
		user.Token = data.Token
	}
	err = db.Model(&UserInvite{}).Where("id = ?", data.ID).Updates(user).Error
	if err != nil {
		return &UserInvite{}, err
	}
	return &data, nil
}

func (d *UserInviteType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&UserInvite{}).Where("id = ?", id).Take(&UserInvite{}).Delete(&UserInvite{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("UserInvite not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
