package models

import (
	"encoding/json"
	"errors"
	"html"
	"strings"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

type Organization struct {
	gorm.Model
	Name               string                 `gorm:"size:255;not null;" json:"name"`
	Description        string                 `gorm:"null;" json:"description"`
	Domain             string                 `gorm:"null;" json:"domain"`
	Image              string                 `gorm:"null;" json:"image"`
	User               *User                  `gorm:"foreignkey:UserID;null" json:"user"`
	UserID             uint64                 `gorm:"null;" json:"user_id"`
	OrganizationPlan   *OrganizationPlan      `gorm:"foreignkey:OrganizationPlanID" json:"organization_plan,omitempty"`
	OrganizationPlanID uint64                 `gorm:"not null" json:"organization_plan_id"`
	Plugins            []*Plugin              `gorm:"many2many:organization_plugins;association_jointable_foreignkey:plugin_id"`
	Members            []*OrganizationMembers `gorm:"foreignkey:OrganizationID"`
	Active             bool                   `gorm:"default:true;" json:"active"`
}

// type OrganizationInterface interface {
// 	Save(db *gorm.DB) (*Organization, error)
// 	FindAll(db *gorm.DB, userID uint64, query string) (*[]Organization, error)
// 	Find(db *gorm.DB, gid uint64, uid uint) error
// 	Update(db *gorm.DB) (*Organization, error)
// 	Delete(db *gorm.DB, id int64) (int64, error)
// }

func (data *Organization) ToJson() (map[string]interface{}, error) {
	msg, _ := json.Marshal(data)
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}

func (data *Organization) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
}

func (data *Organization) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}

	if data.OrganizationPlanID == 0 {
		return errors.New("required organization plan")
	}
	return nil
}

func (d *OrganizationType) Save(db *gorm.DB, data *Organization) (*Organization, error) {
	err := db.Model(&Organization{}).Create(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *OrganizationType) FindAll(db *gorm.DB, userID uint, active bool, query string) (*[]Organization, error) {
	var dataList []Organization
	err := db.
		Model(&Organization{}).
		Where("deleted_at IS NULL AND"+
			" (user_id = ? or "+
			" id in (select organization_id from organization_members where user_id = ? and deleted_at IS NULL and active = ?))", userID, userID, active).
		Where("lower(name) LIKE lower(?)", "%"+query+"%").
		Preload("User").
		Preload("Members").
		Preload("OrganizationPlan").
		Preload("Members.User").
		Find(&dataList).Error
	if err != nil {
		return &[]Organization{}, err
	}
	return &dataList, nil
}

func (data *Organization) FindAllForAdmin(db *gorm.DB, size, page uint64, query string) (*[]Organization, int64, error) {
	var dataList []Organization
	count := 0
	f := db.Model(&Organization{}).
		Where("lower(name) LIKE lower(?)", "%"+query+"%").Preload("User").
		Preload("OrganizationPlan")
	err := f.Order("id desc").Limit(size).
		Offset(size * (page - 1)).Find(&dataList).Error
	if err != nil {
		return &[]Organization{}, 0, err
	}
	f.Count(&count)
	return &dataList, int64(count), nil
}

func (d *OrganizationType) CountOrginationByUser(db *gorm.DB, uid uint) (int, error) {
	count := 0
	err := db.Model(&Organization{}).Where("deleted_at IS NULL AND user_id = ?", uid).Count(&count).Error
	if err != nil {
		return count, nil
	}
	return count, nil

}

func (d *OrganizationType) Find(db *gorm.DB, oid uint) (*Organization, error) {
	var err error
	data := Organization{}
	err = db.Model(&Organization{}).
		Where("id = ?", oid).
		Preload("User").
		Preload("OrganizationPlan").
		Preload("Members").
		Preload("Members.User").
		Preload("Plugins").
		Take(&data).Error
	if err != nil {
		return &Organization{}, err
	}
	return &data, nil
}

func (d *OrganizationType) Update(db *gorm.DB, data *Organization) (*Organization, error) {
	var err error
	mp := map[string]interface{}{}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Description != "" {
		mp["description"] = data.Description
	}
	if data.OrganizationPlanID != 0 {
		mp["organization_plan_id"] = data.OrganizationPlanID
	}
	if data.Domain != "" {
		mp["domain"] = data.Domain
	}
	if data.Image != "" {
		mp["image"] = data.Image
	}
	err = db.Model(&Organization{}).
		Where("id = ?", data.ID).
		UpdateColumn(mp).Error
	if err != nil {
		return &Organization{}, err
	}
	return data, nil
}

func (d *OrganizationType) Delete(db *gorm.DB, id int64) (int64, error) {
	data := Organization{}
	db = db.Model(&data).
		Where("id = ?", id).
		Take(&Organization{}).
		Delete(&Organization{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Organization not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (data *Organization) CheckRole(db *gorm.DB, uid uint64, oid uint64) bool {
	err := db.Model(&Organization{}).
		Where("(id in ( select distinct organization_id from "+
			" organization_members where user_id = ? and organization_id = ? and deleted_at is null and user_role = ?)"+
			" or user_id = ?)"+
			" and organizations.id = ?", uid, oid, Admin, uid, oid).
		Where("deleted_at IS NULL").
		Find(&data).Error
	if err != nil {
		log.Error(err)
		return true
	}
	if data.ID > 0 {
		return false
	}
	return false
}

func (data *Organization) Verify(db *gorm.DB, uid uint64, oid uint64) bool {
	err := db.Model(&Organization{}).
		Where("(id in ( select distinct organization_id from "+
			" organization_members where user_id = ? and organization_id = ? and deleted_at is null)"+
			" or user_id = ?)", uid, oid, uid).
		Where("deleted_at IS NULL").
		Find(&data).Error
	if err != nil {
		log.Error(err)
		return false
	}
	if data.ID > 0 {
		return true
	}
	return false
}

func (data *Organization) AddPlugin(db *gorm.DB, plugin *Plugin) (*Organization, error) {
	err := db.Model(&data).Association("Plugins").Append([]Plugin{*plugin}).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (data *Organization) RemovePlugin(db *gorm.DB, plugin *Plugin) (*Organization, error) {
	err := db.Model(&data).Association("Plugins").Delete(&plugin).Error
	if err != nil {
		return &Organization{}, err
	}
	return nil, nil
}

func (d *OrganizationType) ActiveDeactiveOrganization(db *gorm.DB, oid uint, active bool) (*Organization, error) {
	data := &Organization{}
	err = db.Model(&Organization{}).
		Where("id = ?", oid).
		UpdateColumn("active", active).Take(&data).Error
	if err != nil {
		return &Organization{}, err
	}
	return data, nil
}
