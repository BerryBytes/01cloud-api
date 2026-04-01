package models

import (
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Resource struct {
	gorm.Model
	Name           string        `gorm:"size:255;not null;" json:"name"`
	Cores          uint64        `gorm:"not null;" json:"cores"`
	Memory         uint64        `gorm:"not null;" json:"memory"`
	Active         bool          `gorm:"not null;" json:"active"`
	Weight         uint32        `gorm:"default:10;" json:"weight"`
	Attributes     string        `gorm:"size:1024;null;" json:"attributes"`
	Organization   *Organization `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationID uint64        `gorm:"default:0" json:"organization_id"`
}

func (data *Resource) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *Resource) Validate() error {
	if data.Name == "" {
		return errors.New("required Name")
	}
	if data.Cores == 0 {
		return errors.New("required Cores")
	}
	if data.Memory == 0 {
		return errors.New("required Memory")
	}
	return nil
}

func (d *ResourceType) Save(db *gorm.DB, data *Resource) (*Resource, error) {
	err = db.Model(&Resource{}).Create(&data).Error
	if err != nil {
		return &Resource{}, err
	}
	keys := []string{
		fmt.Sprintf("resource-list-active-%d", data.OrganizationID),
		fmt.Sprintf("resource-list-inactive-%d", data.OrganizationID),
	}
	d.Cache.DeleteMulti(keys)
	return data, nil
}

// This methods checks duplicate entries while creating the Resource according to orgination id and name.
func (d *ResourceType) IsNameExists(db *gorm.DB, oId uint, name string) bool {
	count := 0
	db.Model(&Resource{}).Where("organization_id=? and lower(name)=?", oId, strings.ToLower(name)).Count(&count)
	return count > 0
}

// This methods checks duplicate resources entries while creating the Resource according core memory
func (d *Resource) IsResourceExist(db *gorm.DB, oid uint, cores, memory uint64) bool {
	count := 0
	db.Model(&Resource{}).Where("cores=? and memory=? and organization_id=?", cores, memory, oid).Count(&count)
	return count > 0
}

func (d *ResourceType) FindAll(db *gorm.DB, data *Resource) (*[]Resource, error) {
	key := fmt.Sprintf("resource-list-active-%d", data.OrganizationID)
	var value interface{}
	var datas []Resource
	if ok := d.Cache.Get(key, &value); ok {
		_ = ConvertType(value, &datas)
		return &datas, nil
	}
	err := db.Model(&Resource{}).
		Where("organization_id = ?", data.OrganizationID).
		Where("active = ? ", true).
		Order("weight desc").
		Limit(100).
		Find(&datas).Error
	if err != nil {
		return &[]Resource{}, err
	}
	d.Cache.Set(key, datas)
	return &datas, nil
}

func (d *ResourceType) FindAllWithInactive(db *gorm.DB, data *Resource) (*[]Resource, error) {
	key := fmt.Sprintf("resource-list-inactive-%d", data.OrganizationID)
	var value interface{}
	var datas []Resource
	if ok := d.Cache.Get(key, &value); ok {
		_ = ConvertType(value, &datas)
		return &datas, nil
	}
	err := db.Model(&Resource{}).
		Where("organization_id = ?", data.OrganizationID).
		Order("weight desc").
		Limit(100).Find(&datas).Error
	if err != nil {
		return &[]Resource{}, err
	}
	d.Cache.Set(key, datas)
	return &datas, nil
}

func (d *ResourceType) Find(db *gorm.DB, pid uint64) (*Resource, error) {
	key := fmt.Sprintf("resource-%d", pid)
	var value interface{}
	var data Resource
	if ok := d.Cache.Get(key, &value); ok {
		_ = ConvertType(value, &data)
		return &data, nil
	}
	err := db.Model(&Resource{}).
		Where("id = ?", pid).
		Take(&data).Error
	if err != nil {
		return &Resource{}, err
	}
	d.Cache.Set(key, data)
	return &data, nil
}

func (d *ResourceType) Update(db *gorm.DB, data *Resource) (*Resource, error) {
	var err error
	//var app = Resource{Active: data.Active}
	mp := map[string]interface{}{
		"active": data.Active,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Cores != 0 {
		mp["cores"] = data.Cores
	}
	if data.Memory != 0 {
		mp["memory"] = data.Memory
	}
	if data.Weight != 0 {
		mp["weight"] = data.Weight
	}
	if data.Attributes != "" {
		mp["attributes"] = data.Attributes
	}

	err = db.Model(&Resource{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &Resource{}, err
	}
	keys := []string{
		fmt.Sprintf("resource-%d", data.ID),
		fmt.Sprintf("resource-list-active-%d", data.OrganizationID),
		fmt.Sprintf("resource-list-inactive-%d", data.OrganizationID),
	}
	d.Cache.DeleteMulti(keys)
	return data, nil
}

func (data *ResourceType) Delete(db *gorm.DB, id uint64) (int64, error) {
	resource := &Resource{}
	db = db.Model(&Resource{}).Where("id = ?", id).Take(resource).Delete(&Resource{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Resource not found")
		}
		return 0, db.Error
	}
	keys := []string{
		fmt.Sprintf("resource-%d", id),
		fmt.Sprintf("resource-list-active-%d", resource.OrganizationID),
		fmt.Sprintf("resource-list-inactive-%d", resource.OrganizationID),
	}
	data.Cache.DeleteMulti(keys)
	return db.RowsAffected, nil
}

func (d *ResourceType) CheckResourceLimit(db *gorm.DB, org *Organization, data *Resource) bool {
	//sub, err := data.FindAllWithInactive(db)
	//if err != nil {
	//	return false
	//}
	cores := data.Cores
	memory := data.Memory
	//for _, d := range *sub {
	//	if d.ID != data.ID {
	//		cores += d.Cores
	//		memory += d.Memory
	//	}
	//}

	if cores > uint64(org.OrganizationPlan.Cores) {
		return false
	}
	if memory > uint64(org.OrganizationPlan.Memory) {
		return false
	}
	return true
}
