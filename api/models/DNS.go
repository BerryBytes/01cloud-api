package models

import (
	"encoding/json"
	"errors"
	"html"
	"strings"
	"time"

	"01cloud-api/api/utils/constants"

	"github.com/jinzhu/gorm"
)

type DNS struct {
	gorm.Model
	Name           string        `gorm:"size:255;not null;" json:"name"`
	Provider       string        `gorm:"not null;" json:"provider"`
	ProjectId      string        `gorm:"default:null;" json:"project_id"`
	Credential     string        `gorm:"default:null;" json:"credentials"`
	AccessKey      string        `json:"access_key"`
	SecretKey      string        `json:"secret_key"`
	Region         string        `gorm:"size:255;default:null;" json:"region"`
	ZoneID         string        `gorm:"size:255;default:null;" json:"zone_id"`
	TLS            string        `gorm:"size:255;default:null;" json:"tls"`
	BaseDomain     string        `gorm:"default:null;" json:"base_domain"`
	Active         bool          `gorm:"not null;" json:"active"`
	Organization   *Organization `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationID uint64        `gorm:"default:0" json:"organization_id"`
}

func (data *DNS) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Provider = html.EscapeString(strings.TrimSpace(data.Provider))
	data.ProjectId = html.EscapeString(strings.TrimSpace(data.ProjectId))
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	if data.ProjectId == constants.CloudFlare {
		data.ZoneID = strings.TrimSuffix(data.BaseDomain, ".")
	}

}

func (data *DNS) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if data.Provider == "" {
		return errors.New("required provider")
	}

	if data.BaseDomain == "" {
		return errors.New("required base domain")
	}
	if !strings.HasSuffix(data.BaseDomain, ".") {
		return errors.New("base domain should end with . ")
	}
	if data.Provider == constants.GCP {
		if data.Credential == "" {
			return errors.New("required credential file")
		}
		if data.ProjectId == "" {
			return errors.New("required project id")
		}
	}
	if data.Provider == constants.EKS {
		if data.AccessKey == "" {
			return errors.New("required access key")
		}
		if data.SecretKey == "" {
			return errors.New("required secret key")
		}
		if data.Region == "" {
			return errors.New("required aws region")
		}
	}
	if data.ProjectId != constants.CloudFlare {
		if data.ZoneID == "" {
			return errors.New("required zone id")
		}
	}

	return nil
}

func (data *DNS) ValidateCredentials() error {
	if data.Provider == "" {
		return errors.New("required provider name")
	}
	if data.Provider == constants.GCP && data.Credential == "" {
		return errors.New("required credentials file")
	}
	if data.Provider == constants.EKS && data.AccessKey == "" && data.SecretKey == "" && data.Region == "" {
		return errors.New("required access key and secret key and region")
	}
	return nil
}
func (data *DNS) ToJson() (map[string]interface{}, error) {
	msg, _ := json.Marshal(data)
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}

func (dns *DNSType) Save(db *gorm.DB, data *DNS) (*DNS, error) {
	err = db.Model(&DNS{}).Create(data).Error
	if err != nil {
		return &DNS{}, err
	}
	return data, nil
}

func (dns *DNSType) IsNameExists(db *gorm.DB, oid uint64, name string) bool {
	count := 0
	db.Model(&DNS{}).Where("organization_id=? and lower(name)=?", oid, strings.ToLower(name)).Count(&count)
	return count > 0
}

func (dns *DNSType) FindAllByOrganization(db *gorm.DB, orgId uint) (*[]DNS, error) {
	var datas []DNS
	err = db.Model(&DNS{}).
		Where("organization_id = ?", orgId).
		Find(&datas).Error
	if err != nil {
		return &[]DNS{}, err
	}
	return &datas, nil
}

func (data *DNSType) FindAll(db *gorm.DB) (*[]DNS, error) {
	var datas []DNS
	err = db.Model(&DNS{}).Where("active = ? ", true).Find(&datas).Error
	if err != nil {
		return &[]DNS{}, err
	}
	return &datas, nil
}

func (dns *DNSType) FindAllWithInactive(db *gorm.DB) (*[]DNS, error) {
	var datas []DNS
	err = db.Model(&DNS{}).Find(&datas).Error
	if err != nil {
		return &[]DNS{}, err
	}
	return &datas, nil
}

func (dns *DNSType) Find(db *gorm.DB, pid uint64) (*DNS, error) {
	data := &DNS{}
	err = db.Model(&DNS{}).Where("id = ?", pid).Take(data).Error
	if err != nil {
		return &DNS{}, err
	}
	return data, nil
}

func (dns *DNSType) Update(db *gorm.DB, data *DNS) (*DNS, error) {
	//var app = DNS{Active: data.Active}
	mp := map[string]interface{}{
		"active": data.Active,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Provider != "" {
		mp["provider"] = data.Provider
	}
	if data.ZoneID != "" {
		mp["zone_id"] = data.ZoneID
	}
	if data.ProjectId != "" {
		mp["project_id"] = data.ProjectId
	}
	if data.Region != "" {
		mp["region"] = data.Region
	}
	if data.AccessKey != "" {
		mp["access_key"] = data.AccessKey
	}
	if data.SecretKey != "" {
		mp["secret_key"] = data.SecretKey
	}
	if data.BaseDomain != "" {
		mp["base_domain"] = data.BaseDomain
	}
	if data.Credential != "" {
		mp["credential"] = data.Credential
	}
	if data.TLS != "" {
		mp["tls"] = data.TLS
	}
	err = db.Model(&DNS{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &DNS{}, err
	}
	return data, nil
}

func (dns *DNSType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&DNS{}).Where("id = ?", id).Take(&DNS{}).Delete(&DNS{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("DNS not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
