package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type ImageRegistryInterface interface {
	Save(db *gorm.DB, data *ImageRegistry) (*ImageRegistry, error)
	IsNameExists(db *gorm.DB, oId uint, name string) bool
	FindAllWithInactive(db *gorm.DB, data *ImageRegistry) (*[]ImageRegistry, error)
	Find(db *gorm.DB, pid uint64) (*ImageRegistry, error)
	Update(db *gorm.DB, data *ImageRegistry) (*ImageRegistry, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
	FindAll(db *gorm.DB, data *ImageRegistry) (*[]ImageRegistry, error)
}
type ImageRegistryType struct{}

func NewImageRegistry() ImageRegistryInterface {
	return &ImageRegistryType{}
}

type ImageRegistry struct {
	gorm.Model
	Name           string         `gorm:"size:255;not null;" json:"name"`
	Service        string         `gorm:"null;" json:"service"`
	Provider       string         `gorm:"null;default:null" json:"provider"`
	ProjectName    string         `gorm:"null;" json:"project_name"`
	UserName       string         `gorm:"null;" json:"user_name"`
	Password       string         `gorm:"null;" json:"password"`
	Credentials    postgres.Jsonb `json:"credentials"`
	Active         bool           `gorm:"not null;" json:"active"`
	Organization   *Organization  `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationID uint64         `gorm:"default:0" json:"organization_id"`
}
type Credentials struct {
	Server      string `json:"docker_registry_server"`
	UserName    string `json:"docker_username"`
	Password    string `json:"docker_password"`
	ProjectName string `json:"project_name"`
}

func (data *ImageRegistry) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Provider = html.EscapeString(strings.TrimSpace(data.Provider))
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *ImageRegistry) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	//if data.Service == "" {
	//	return errors.New("Required service name")
	//}
	if data.Provider == "" {
		return errors.New("required provider")
	}

	return nil
}

func (d *ImageRegistryType) Save(db *gorm.DB, data *ImageRegistry) (*ImageRegistry, error) {
	err := db.Model(&ImageRegistry{}).Create(&data).Error
	if err != nil {
		return &ImageRegistry{}, err
	}
	return data, nil
}

// This methods checks duplicate entries while creating the ImageRegistry according to orgination id and name.
func (d *ImageRegistryType) IsNameExists(db *gorm.DB, oId uint, name string) bool {
	count := 0
	db.Model(&ImageRegistry{}).Where("organization_id=? and lower(name)=?", oId, strings.ToLower(name)).Count(&count)
	return count > 0
}

func (d *ImageRegistryType) FindAll(db *gorm.DB, data *ImageRegistry) (*[]ImageRegistry, error) {
	var err error
	datas := []ImageRegistry{}
	err = db.Model(&ImageRegistry{}).
		Where("organization_id = ?", data.OrganizationID).
		Where("active = ? ", true).
		Limit(100).
		Find(&datas).Error
	if err != nil {
		return &[]ImageRegistry{}, err
	}
	return &datas, nil
}

func (d *ImageRegistryType) FindAllWithInactive(db *gorm.DB, data *ImageRegistry) (*[]ImageRegistry, error) {
	var err error
	datas := []ImageRegistry{}
	err = db.Model(&ImageRegistry{}).
		Where("organization_id = ?", data.OrganizationID).
		Limit(100).Find(&datas).Error
	if err != nil {
		return &[]ImageRegistry{}, err
	}
	return &datas, nil
}

func (d *ImageRegistryType) Find(db *gorm.DB, pid uint64) (*ImageRegistry, error) {
	var err error
	var data = ImageRegistry{}
	err = db.Model(&ImageRegistry{}).
		Where("id = ?", pid).
		Take(&data).Error
	if err != nil {
		return &ImageRegistry{}, err
	}
	return &data, nil
}

func (d *ImageRegistryType) Update(db *gorm.DB, data *ImageRegistry) (*ImageRegistry, error) {
	var err error
	//var app = ImageRegistry{Active: data.Active}
	mp := map[string]interface{}{
		"active": data.Active,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.UserName != "" {
		mp["user_name"] = data.UserName
	}
	if data.Password != "" {
		mp["password"] = data.Password
	}
	if data.Credentials.RawMessage != nil {
		mp["credentials"] = data.Credentials
	}
	if data.Provider != "" {
		mp["provider"] = data.Provider
	}
	if data.ProjectName != "" {
		mp["project_name"] = data.ProjectName
	}
	if data.Service != "" {
		mp["service"] = data.Service
	}

	err = db.Model(&ImageRegistry{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &ImageRegistry{}, err
	}
	return data, nil
}

func (data *ImageRegistryType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&ImageRegistry{}).Where("id = ?", id).Take(&ImageRegistry{}).Delete(&ImageRegistry{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("ImageRegistry not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
