package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type CronImageInterface interface {
	Save(db *gorm.DB, data *CronImage) (*CronImage, error)
	FindAll(db *gorm.DB) (*[]CronImage, error)
	FindAllWithInactive(db *gorm.DB) (*[]CronImage, error)
	Find(db *gorm.DB, pid uint64) (*CronImage, error)
	Update(db *gorm.DB, data *CronImage) (*CronImage, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
}
type CronImageType struct {
}

func NewCronImage() CronImageInterface {
	return &CronImageType{}
}

type CronImage struct {
	gorm.Model
	Name      string `gorm:"size:255;not null;" json:"name"`
	ImageName string `gorm:"not null;" json:"image_name"`
	Version   string `gorm:"null;" json:"version"`
	Active    bool   `gorm:"not null;" json:"active"`
}

func (data *CronImage) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.ImageName = html.EscapeString(strings.TrimSpace(data.ImageName))
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *CronImage) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if data.ImageName == "" {
		return errors.New("required image name")
	}
	return nil
}

func (cron *CronImageType) Save(db *gorm.DB, data *CronImage) (*CronImage, error) {
	err := db.Model(&CronImage{}).Create(&data).Error
	if err != nil {
		return &CronImage{}, err
	}
	return data, nil
}

func (cron *CronImageType) FindAll(db *gorm.DB) (*[]CronImage, error) {
	var err error
	var datas []CronImage
	err = db.Model(&CronImage{}).Where("active = ? ", true).Find(&datas).Error
	if err != nil {
		return &[]CronImage{}, err
	}
	return &datas, nil
}

func (cron *CronImageType) FindAllWithInactive(db *gorm.DB) (*[]CronImage, error) {
	var err error
	var datas []CronImage
	err = db.Model(&CronImage{}).Find(&datas).Error
	if err != nil {
		return &[]CronImage{}, err
	}
	return &datas, nil
}

func (cron *CronImageType) Find(db *gorm.DB, pid uint64) (*CronImage, error) {
	var err error
	data := CronImage{}
	err = db.Model(&CronImage{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &CronImage{}, err
	}
	return &data, nil
}

func (cron *CronImageType) Update(db *gorm.DB, data *CronImage) (*CronImage, error) {
	var err error
	//var app = CronImage{Active: data.Active}
	mp := map[string]interface{}{
		"active": data.Active,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.ImageName != "" {
		mp["image_name"] = data.ImageName
	}

	err = db.Model(&CronImage{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &CronImage{}, err
	}
	return data, nil
}

func (cron *CronImageType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&CronImage{}).Where("id = ?", id).Take(&CronImage{}).Delete(&CronImage{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("CronImage not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
