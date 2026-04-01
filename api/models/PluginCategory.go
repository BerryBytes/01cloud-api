package models

import (
	"errors"
	"html"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
)

type PluginCategory struct {
	gorm.Model
	Name        string    `gorm:"size:255;not null;" json:"name"`
	Description string    `gorm:"size:1024;not null;" json:"description"`
	IsAddOn     bool      `gorm:"default:false" json:"is_add_on"`
	Plugins     []*Plugin `gorm:"many2many:plugin_category_pivots"`
}
type PluginCategoryInterface interface {
	Save(db *gorm.DB, data *PluginCategory) (*PluginCategory, error)
	FindAll(db *gorm.DB, IsAddOn string, query string) (*[]PluginCategory, error)
	Find(db *gorm.DB, id uint64) (*PluginCategory, error)
	Update(db *gorm.DB, data *PluginCategory) (*PluginCategory, error)
	Delete(db *gorm.DB, id int64) (int64, error)
	AddPlugins(db *gorm.DB, plugin *Plugin) error
	DeletePlugins(db *gorm.DB, plugin *Plugin) error
	FindCategoryByIds(db *gorm.DB, ids []string) ([]*PluginCategory, error)
}

type PluginCategoryType struct {
}

func NewPluginCategory() PluginCategoryInterface {
	return &PluginCategoryType{}
}

func (data *PluginCategory) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
}

func (data *PluginCategory) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	return nil
}

func (d *PluginCategoryType) Save(db *gorm.DB, data *PluginCategory) (*PluginCategory, error) {
	err := db.Model(&PluginCategory{}).Create(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (data *PluginCategoryType) FindAll(db *gorm.DB, IsAddOn string, query string) (*[]PluginCategory, error) {
	var err error
	var dataList []PluginCategory
	err = db.Model(&PluginCategory{}).
		Preload("Plugins").
		Where("lower(name) LIKE lower(?) and is_add_on = ?", "%"+query+"%", IsAddOn).
		Find(&dataList).Error
	if err != nil {
		return &[]PluginCategory{}, err
	}
	return &dataList, nil
}

func (d *PluginCategoryType) Find(db *gorm.DB, id uint64) (*PluginCategory, error) {
	var err error
	var data = &PluginCategory{}
	err = db.Model(&PluginCategory{}).
		Where("id = ?", id).
		Preload("Plugins").
		Take(&data).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (data *PluginCategoryType) FindCategoryByIds(db *gorm.DB, ids []string) ([]*PluginCategory, error) {
	datas := []*PluginCategory{}
	err := db.Debug().Model(&PluginCategory{}).Where("id in (?)", ids).Order("id desc").Find(&datas).Error
	if err != nil {
		return []*PluginCategory{}, err
	}
	return datas, nil
}

func (d *PluginCategoryType) Update(db *gorm.DB, data *PluginCategory) (*PluginCategory, error) {
	var err error
	mp := map[string]interface{}{}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Description != "" {
		mp["description"] = data.Description
	}
	mp["is_plugin"] = data.IsAddOn

	err = db.Model(&PluginCategory{}).
		Where("id = ?", data.ID).
		UpdateColumn(mp).Error
	if err != nil {
		return &PluginCategory{}, err
	}
	return data, nil
}

func (dat *PluginCategoryType) Delete(db *gorm.DB, id int64) (int64, error) {
	var data = &PluginCategory{}
	db = db.Model(&data).
		Where("id = ?", id).
		Take(&PluginCategory{}).
		Delete(&PluginCategory{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Plugin not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (da *PluginCategoryType) AddPlugins(db *gorm.DB, plugin *Plugin) error {
	var err error
	var data = &PluginCategory{}
	findPlugin := *plugin
	err = db.Model(&data).Association("Plugins").Find(&findPlugin).Error
	if err == nil {
		return errors.New("plugin already added")
	}
	err = db.Model(&data).Association("Plugins").Append([]Plugin{*plugin}).Error
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (d *PluginCategoryType) DeletePlugins(db *gorm.DB, plugin *Plugin) error {
	var data = &PluginCategory{}
	err := db.Model(&data).Association("Plugins").Delete(&plugin).Error
	if err != nil {
		return err
	}
	return nil
}
