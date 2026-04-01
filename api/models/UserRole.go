package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type UserRole struct {
	gorm.Model
	Name        string `gorm:"size:255;not null;" json:"name"`
	Code        uint64 `gorm:"not null;" json:"code"`
	Description string `gorm:"size:1024;null" json:"description"`
	Active      bool   `gorm:"not null;" json:"active"`
}

func (data *UserRole) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Description = html.EscapeString(strings.TrimSpace(data.Description))
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *UserRole) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if data.Description == "" {
		return errors.New("required description")
	}
	if data.Code == 0 {
		return errors.New("required code")
	}
	return nil
}

func (d *UserRoleType) Save(db *gorm.DB, data *UserRole) (*UserRole, error) {
	err := db.Model(&UserRole{}).Create(&data).Error
	if err != nil {
		return &UserRole{}, err
	}
	return data, nil
}

func (d *UserRoleType) FindAll(db *gorm.DB) (*[]UserRole, error) {
	datas := []UserRole{}
	err := db.Model(&UserRole{}).Order("id").Find(&datas).Error
	if err != nil {
		return &[]UserRole{}, err
	}
	return &datas, nil
}

func (d *UserRoleType) Find(db *gorm.DB, pid uint64) (*UserRole, error) {
	data := UserRole{}
	err := db.Model(&UserRole{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &UserRole{}, err
	}
	return &data, nil
}

func (d *UserRoleType) Update(db *gorm.DB, data *UserRole) (*UserRole, error) {
	err := db.Model(&UserRole{}).Where("id = ?", data.ID).Updates(UserRole{
		Name:        data.Name,
		Code:        data.Code,
		Description: data.Description,
		Active:      data.Active,
	}).Error
	if err != nil {
		return &UserRole{}, err
	}
	return data, nil
}

func (d *UserRoleType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&UserRole{}).Where("id = ?", id).Take(&UserRole{}).Delete(&UserRole{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("UserRole not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
