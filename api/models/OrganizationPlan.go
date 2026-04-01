package models

import (
	"errors"
	"html"
	"strings"

	"github.com/jinzhu/gorm"
)

type OrganizationPlan struct {
	gorm.Model
	Name       string `gorm:"size:255;not null;" json:"name"`
	Cluster    uint32 `gorm:"not null;" json:"cluster"`
	Memory     uint32 `gorm:"not null;" json:"memory"`
	Cores      uint32 `gorm:"not null;" json:"cores"`
	NoOfUser   uint32 `gorm:"not null;" json:"no_of_user"`
	Price      uint32 `gorm:"not null;" json:"price"`
	Weight     uint32 `gorm:"default:10;" json:"weight"`
	Attributes string `gorm:"size:1024;null;" json:"attributes"`
	Active     bool   `gorm:"not null;" json:"active"`
}

type OrganizationPlanInterface interface {
	Save(db *gorm.DB, data *OrganizationPlan) (*OrganizationPlan, error)
	FindAll(db *gorm.DB) (*[]OrganizationPlan, error)
	FindAllWithInActive(db *gorm.DB) (*[]OrganizationPlan, error)
	Find(db *gorm.DB, pid uint64) (*OrganizationPlan, error)
	Update(db *gorm.DB, data *OrganizationPlan) (*OrganizationPlan, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
}

type OrganizationPlanRepo struct{}

func NewOrganizationPlan() OrganizationPlanInterface {
	return &OrganizationPlanRepo{}
}

func (data *OrganizationPlan) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Active = true
}

func (data *OrganizationPlan) Validate() error {
	if data.Name == "" {
		return errors.New("required Name")
	}
	if data.Cluster == 0 {
		return errors.New("required number of Clusters")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	if data.NoOfUser == 0 {
		return errors.New("required No of Users")
	}
	if data.Memory == 0 {
		return errors.New("required Memory")
	}
	if data.Cores == 0 {
		return errors.New("required Cores")
	}
	return nil
}

func (r OrganizationPlanRepo) Save(db *gorm.DB, data *OrganizationPlan) (*OrganizationPlan, error) {
	err = db.Model(&OrganizationPlan{}).Create(&data).Error
	if err != nil {
		return &OrganizationPlan{}, err
	}
	return data, nil
}

func (r OrganizationPlanRepo) FindAll(db *gorm.DB) (*[]OrganizationPlan, error) {
	datas := []OrganizationPlan{}
	err = db.Model(&OrganizationPlan{}).Where("active = ? ", true).Order("weight desc").Limit(100).Find(&datas).Error
	if err != nil {
		return &[]OrganizationPlan{}, err
	}
	return &datas, nil
}

func (r OrganizationPlanRepo) FindAllWithInActive(db *gorm.DB) (*[]OrganizationPlan, error) {
	datas := []OrganizationPlan{}
	err = db.Model(&OrganizationPlan{}).Order("weight desc").Find(&datas).Error
	if err != nil {
		return &[]OrganizationPlan{}, err
	}
	return &datas, nil
}

func (r OrganizationPlanRepo) Find(db *gorm.DB, pid uint64) (*OrganizationPlan, error) {
	data := &OrganizationPlan{}
	err = db.Model(&OrganizationPlan{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &OrganizationPlan{}, err
	}
	return data, nil
}

func (r OrganizationPlanRepo) Update(db *gorm.DB, data *OrganizationPlan) (*OrganizationPlan, error) {
	mp := map[string]interface{}{
		"active": data.Active,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Price != 0 {
		mp["price"] = data.Price
	}
	if data.Cores != 0 {
		mp["cores"] = data.Cores
	}
	if data.Memory != 0 {
		mp["memory"] = data.Memory
	}
	if data.Attributes != "" {
		mp["attributes"] = data.Attributes
	}
	if data.Weight != 0 {
		mp["weight"] = data.Weight
	}
	if data.Cluster != 0 {
		mp["cluster"] = data.Cluster
	}
	if data.NoOfUser != 0 {
		mp["no_of_user"] = data.NoOfUser
	}

	err = db.Model(&OrganizationPlan{}).Where("id = ?", data.ID).UpdateColumn(mp).Error
	if err != nil {
		return &OrganizationPlan{}, err
	}
	return data, nil
}

func (r OrganizationPlanRepo) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&OrganizationPlan{}).Where("id = ?", id).Take(&OrganizationPlan{}).Delete(&OrganizationPlan{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("OrganizationPlan not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
