package models

import (
	"errors"

	"time"

	"github.com/jinzhu/gorm"
)

type Activity struct {
	gorm.Model
	User           *User        `gorm:"foreignkey:UserID" json:"user"`
	UserID         uint         `gorm:"not null" json:"user_id"`
	Project        *Project     `gorm:"foreignkey:ProjectID" json:"project"`
	ProjectID      uint         `gorm:"not null" json:"project_id"`
	Application    *Application `gorm:"foreignkey:ApplicationID" json:"application"`
	ApplicationID  uint         `gorm:"null;DEFAULT:0" json:"application_id"`
	Environment    *Environment `gorm:"foreignkey:EnvironmentID" json:"environment"`
	EnvironmentID  uint         `gorm:"null;DEFAULT:0" json:"environment_id"`
	OrganizationID uint         `gorm:"default:0" json:"organization_id"`
	Action         string       `gorm:"size:255;null;" json:"action"`
	Module         string       `gorm:"size:255;null;" json:"module"`
	Active         bool         `gorm:"not null;" json:"active"`
	Remarks        string       `gorm:"size:1024;null;" json:"remarks"`
	Extras         string       `gorm:"size:1024;null;" json:"extras"`
}

func (data *Activity) Prepare() {
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *Activity) Validate() error {
	if data.OrganizationID == 0 {
		if data.ProjectID == 0 {
			return errors.New("required Project Id")
		}
	}
	if data.Action == "" {
		return errors.New("required Action")
	}
	if data.Module == "" {
		return errors.New("required Module")
	}
	return nil
}

func (d *ActivityType) Save(db *gorm.DB, data Activity) (*Activity, error) {
	err = db.Model(&Activity{}).Create(&data).Error
	if err != nil {
		return &Activity{}, err
	}
	return &data, nil
}

func (d *ActivityType) FindAll(db *gorm.DB) (*[]Activity, error) {
	var err error
	datas := []Activity{}
	err = db.Model(&Activity{}).Preload("Project").Preload("User").Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]Activity{}, err
	}
	return &datas, nil
}

func (d *ActivityType) FindAllActivityByProject(db *gorm.DB, pid uint, limit int, offset int, action, module, startdate, enddate string, userid int) (*[]Activity, int, error) {
	count := 0
	datas := []Activity{}
	a := db.Model(&Activity{}).Where("project_id=$1", pid)
	if len(action) != 0 {
		if action == "share" {
			a = a.Where("action=? or action=?", "share", "unshare")
		} else {
			a = a.Where("action=?", action)
		}
	}
	if len(module) != 0 {
		a = a.Where("module=?", module)
	}
	if len(startdate) != 0 && len(enddate) != 0 {
		a = a.Where("created_at BETWEEN ?::timestamp AND ?::timestamp", startdate, enddate)
	}
	if userid != 0 {
		a = a.Where("user_id=?", userid)
	}
	err := a.Preload("Project").Preload("User").Preload("Application").Preload("Environment").Order("id desc").Limit(limit).Offset(offset).Find(&datas).Error
	if err != nil {
		return &[]Activity{}, count, err
	}
	a.Count(&count)
	return &datas, count, nil
}

func (d *ActivityType) FindAllActivityByOrganization(db *gorm.DB, oid uint, limit int, offset int, action, module, startdate, enddate string, userid int) (*[]Activity, int, error) {
	count := 0
	datas := []Activity{}
	a := db.Model(&Activity{}).Where("organization_id=$1", oid)
	if len(action) != 0 {
		if action == "share" {
			a = a.Where("action=? or action=?", "share", "unshare")
		} else {
			a = a.Where("action=?", action)
		}
	}
	if len(module) != 0 {
		a = a.Where("module=?", module)
	}
	if len(startdate) != 0 && len(enddate) != 0 {
		a = a.Where("created_at BETWEEN ?::timestamp AND ?::timestamp", startdate, enddate)
	}
	if userid != 0 {
		a = a.Where("user_id=?", userid)
	}
	err := a.Preload("User").Order("id desc").Limit(limit).Offset(offset).Find(&datas).Error
	if err != nil {
		return &[]Activity{}, count, err
	}
	a.Count(&count)
	return &datas, count, nil
}

func (data *ActivityType) Find(db *gorm.DB, pid uint64) (*Activity, error) {
	var err error
	datas := Activity{}
	err = db.Model(&Activity{}).Preload("Project").Where("id = ?", pid).Take(&datas).Error
	if err != nil {
		return &Activity{}, err
	}
	return &datas, nil
}

func (d *ActivityType) Update(db *gorm.DB, data Activity) (*Activity, error) {
	var err error
	var app = Activity{Active: data.Active}
	if data.ApplicationID != 0 {
		app.ApplicationID = data.ApplicationID
	}
	if data.ProjectID != 0 {
		app.ProjectID = data.ProjectID
	}
	if data.EnvironmentID != 0 {
		app.EnvironmentID = data.EnvironmentID
	}
	if data.Action != "" {
		app.Action = data.Action
	}
	if data.Module != "" {
		app.Module = data.Module
	}
	if data.Remarks != "" {
		app.Remarks = data.Remarks
	}
	if data.Extras != "" {
		app.Extras = data.Extras
	}
	err = db.Model(&Activity{}).Where("id = ?", data.ID).Updates(app).Error
	if err != nil {
		return &Activity{}, err
	}
	return &data, nil
}

func (d *ActivityType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&Activity{}).Where("id = ?", id).Take(&Activity{}).Delete(&Activity{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Activity not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
