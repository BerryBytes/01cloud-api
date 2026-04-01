package models

import (
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/jinzhu/gorm"
)

type Subscription struct {
	gorm.Model
	Name           string         `gorm:"size:255;not null;" json:"name"`
	Apps           uint32         `gorm:"not null;default:1" json:"apps"`
	DiskSpace      uint32         `gorm:"not null;" json:"disk_space"`
	Memory         uint32         `gorm:"not null;" json:"memory"`
	Cores          uint32         `gorm:"not null;" json:"cores"`
	DataTransfer   uint32         `gorm:"not null;" json:"data_transfer"`
	Price          uint32         `gorm:"not null;" json:"price"`
	CiBuild        uint32         `gorm:"default:50;" json:"ci_build"`
	Weight         uint32         `gorm:"default:10;" json:"weight"`
	Attributes     string         `gorm:"size:1024;null;" json:"attributes"`
	Active         bool           `gorm:"not null;" json:"active"`
	CronJob        uint64         `gorm:"default:1;" json:"cron_job"`
	Backups        uint64         `gorm:"default:5;" json:"backups"`
	ResourceList   postgres.Jsonb `sql:"json" json:"resource_list"`
	Organization   *Organization  `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationID uint64         `gorm:"default:0" json:"organization_id"`
	UserID         uint64         `gorm:"default:0" json:"user_id"`
	LoadBalancer   uint64         `gorm:"default:0;" json:"load_balancer"`
	PriceList      postgres.Jsonb `json:"price_list"`
	Validity       uint64         `gorm:"default:0;" json:"validity"`
}

func (data *Subscription) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
}

func (data *Subscription) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if data.Apps == 0 {
		return errors.New("required number of apps")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	if data.DiskSpace == 0 {
		return errors.New("required disk space")
	}
	if data.Memory == 0 {
		return errors.New("required memory")
	}
	if data.Cores == 0 {
		return errors.New("required cores")
	}
	if data.DataTransfer == 0 {
		return errors.New("required data transfer")
	}
	return nil
}

func (d *SubscriptionType) Save(db *gorm.DB, data Subscription) (*Subscription, error) {
	err := db.Model(&Subscription{}).Create(&data).Error
	if err != nil {
		return &Subscription{}, err
	}
	return &data, nil
}

func (d *SubscriptionType) FindAll(db *gorm.DB, orgid uint64) (*[]Subscription, error) {
	key := fmt.Sprintf("subscription-list-active-%d", orgid)
	var value interface{}
	var datas []Subscription
	if ok := d.Cache.Get(key, &value); ok {
		_ = ConvertType(value, &datas)
		return &datas, nil
	}
	err := db.Model(&Subscription{}).
		Where("organization_id = ?", orgid).
		Where("active = ? ", true).
		Order("weight desc").
		Limit(100).
		Find(&datas).Error
	if err != nil {
		return &[]Subscription{}, err
	}
	d.Cache.Set(key, datas)
	return &datas, nil
}

func (d *SubscriptionType) FindAllByUserID(db *gorm.DB, uid uint64) (*[]Subscription, error) {
	key := fmt.Sprintf("subscription-user-list-%d", uid)
	var value interface{}
	var datas []Subscription
	if ok := d.Cache.Get(key, &value); ok {
		_ = ConvertType(value, &datas)
		return &datas, nil
	}
	err := db.Model(&Subscription{}).
		Where("active = ? ", true).
		Where("organization_id = 0").
		Where("user_id = ? or user_id = 0", uid).
		Order("weight desc").
		Limit(100).
		Find(&datas).Error
	if err != nil {
		return &[]Subscription{}, err
	}
	d.Cache.Set(key, datas)
	return &datas, nil
}

func (d *SubscriptionType) FindByUserID(db *gorm.DB, uid uint64) (*[]Subscription, error) {
	key := fmt.Sprintf("subscription-list-user-%d", uid)
	var value interface{}
	var datas []Subscription
	if ok := d.Cache.Get(key, &value); ok {
		_ = ConvertType(value, &datas)
		return &datas, nil
	}
	err := db.Model(&Subscription{}).
		Where("active = ? ", true).
		Where("organization_id = 0").
		Where("user_id = ?", uid).
		Order("weight desc").
		Limit(100).
		Find(&datas).Error
	if err != nil {
		return &[]Subscription{}, err
	}
	d.Cache.Set(key, datas)
	return &datas, nil
}

func (data *Subscription) FindAllWithInActive(db *gorm.DB) (*[]Subscription, error) {
	datas := []Subscription{}
	err := db.Model(&Subscription{}).
		Where("organization_id = ?", data.OrganizationID).
		Order("weight desc").Find(&datas).Error
	if err != nil {
		return &[]Subscription{}, err
	}
	return &datas, nil
}

// This methods checks duplicate entries while creating the Subscription according to orgination id and name.
func (d *Subscription) IsNameExists(db *gorm.DB, oId uint, name string) bool {
	count := 0
	db.Model(&Subscription{}).Where("organization_id=? and lower(name)=?", oId, strings.ToLower(name)).Count(&count)
	return count > 0
}

func (d *SubscriptionType) Find(db *gorm.DB, pid uint64) (*Subscription, error) {
	key := fmt.Sprintf("subscription-%d", pid)
	var value interface{}
	var data Subscription
	if ok := d.Cache.Get(key, &value); ok {
		_ = ConvertType(value, &data)
		return &data, nil
	}
	err := db.Model(&Subscription{}).
		Where("id = ?", pid).
		Take(&data).Error
	if err != nil {
		return &Subscription{}, err
	}
	d.Cache.Set(key, data)
	return &data, nil
}

func (d *SubscriptionType) Update(db *gorm.DB, data Subscription) (*Subscription, error) {
	mp := map[string]interface{}{
		"active": data.Active,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Apps != 0 {
		mp["apps"] = data.Apps
	}
	if data.Price != 0 {
		mp["price"] = data.Price
	}
	if data.CiBuild != 0 {
		mp["ci_build"] = data.CiBuild
	}
	if data.DataTransfer != 0 {
		mp["data_transfer"] = data.DataTransfer
	}
	if data.Cores != 0 {
		mp["cores"] = data.Cores
	}
	if data.Memory != 0 {
		mp["memory"] = data.Memory
	}
	if data.DiskSpace != 0 {
		mp["disk_space"] = data.DiskSpace
	}
	if data.Attributes != "" {
		mp["attributes"] = data.Attributes
	}
	if data.Weight != 0 {
		mp["weight"] = data.Weight
	}
	if data.CronJob != 0 {
		mp["cronjob"] = data.CronJob
	}

	if data.ResourceList.RawMessage != nil {
		mp["resource_list"] = data.ResourceList
	}
	if data.PriceList.RawMessage != nil {
		mp["price_list"] = data.PriceList
	}
	if data.Backups != 0 {
		mp["backups"] = data.Backups
	}
	if data.Validity != 0 {
		mp["validity"] = data.Validity
	}

	err := db.Model(&Subscription{}).
		Where("id = ?", data.ID).
		UpdateColumn(mp).Error
	if err != nil {
		return &Subscription{}, err
	}
	keys := []string{
		fmt.Sprintf("subscription-%d", data.ID),
		fmt.Sprintf("subscription-list-user-%d", data.UserID),
		fmt.Sprintf("subscription-list-active-%d", data.OrganizationID),
		fmt.Sprintf("subscription-user-list-%d", data.UserID),
	}
	d.Cache.DeleteMulti(keys)
	return &data, nil
}

func (d *SubscriptionType) Delete(db *gorm.DB, id uint64) (int64, error) {
	data := &Subscription{}
	db = db.Model(&Subscription{}).
		Where("id = ?", id).
		Take(&data).Delete(&Subscription{})

	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Subscription not found")
		}
		return 0, db.Error
	}
	keys := []string{
		fmt.Sprintf("subscription-%d", id),
		fmt.Sprintf("subscription-list-user-%d", data.UserID),
		fmt.Sprintf("subscription-list-active-%d", data.OrganizationID),
		fmt.Sprintf("subscription-user-list-%d", data.UserID),
	}
	d.Cache.DeleteMulti(keys)
	return db.RowsAffected, nil
}

func (data *Subscription) CheckSubscriptionResourceLimit(db *gorm.DB, org *Organization) bool {
	//sub, err := data.FindAllWithInActive(db)
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

	if cores > org.OrganizationPlan.Cores {
		return false
	}
	if memory > org.OrganizationPlan.Memory {
		return false
	}
	return true
}
