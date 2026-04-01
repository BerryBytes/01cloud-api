package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type PluginInterface interface {
	Save(db *gorm.DB, data *Plugin) (*Plugin, error)
	FindBySupportCi(db *gorm.DB, ci, isManagedService bool, oid uint, page uint64, size uint64, q string, sortColumn string, sortDirection string) (*[]Plugin, error)
	FindPluginByIds(db *gorm.DB, ids []string) ([]*Plugin, error)
	FindAll(db *gorm.DB) (*[]Plugin, error)
	FindAllActiveWithFilters(db *gorm.DB, page uint64, size uint64, q string, sortColumn string, sortDirection string) (*[]Plugin, error)
	FindAllWithFilters(db *gorm.DB, page uint64, size uint64, q string, sortColumn string, sortDirection string) (*[]Plugin, int, error)
	Find(db *gorm.DB, pid uint64) (*Plugin, error)
	FindAddOns(db *gorm.DB, pid uint64, query string, catid []string) ([]*Plugin, error)
	Update(db *gorm.DB, data *Plugin) (*Plugin, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
}
type PluginType struct {
}

func NewPlugin() *PluginType {
	return &PluginType{}
}

type PluginVersionInterface interface {
	Save(db *gorm.DB, data *PluginVersion) (*PluginVersion, error)
	FindAll(db *gorm.DB) (*[]PluginVersion, error)
	FindAllWithInactive(db *gorm.DB) (*[]PluginVersion, error)
	FindAllByPlugin(db *gorm.DB, pid uint64) (*[]PluginVersion, error)
	Find(db *gorm.DB, pid uint64) (*PluginVersion, error)
	FindLatestPluginVersion(db *gorm.DB, pid uint64) (*PluginVersion, error)
	Update(db *gorm.DB, data *PluginVersion) (*PluginVersion, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
}

type PluginVersionType struct{}

func NewPluginVersion() *PluginVersionType {
	return &PluginVersionType{}
}

type Plugin struct {
	gorm.Model
	Name             string            `gorm:"size:255;not null;" json:"name"`
	Description      string            `gorm:"size:1024;not null;" json:"description"`
	SourceUrl        string            `gorm:"size:255;not null;" json:"source_url"`
	Image            string            `gorm:"size:255;null;" json:"image"`
	Active           bool              `gorm:"not null;" json:"active"`
	SupportCi        bool              `gorm:"default:false" json:"support_ci"`
	IsManagedService bool              `gorm:"default:false" json:"is_managed_service"`
	MinCpu           uint64            `gorm:"default:250" json:"min_cpu"`
	MinMemory        uint64            `gorm:"default:256" json:"min_memory"`
	IsAddOn          bool              `gorm:"default:false" json:"is_add_on"`
	ServiceDetail    postgres.Jsonb    `json:"service_detail"`
	AddOns           []*Plugin         `gorm:"many2many:add_ons;association_jointable_foreignkey:addon_id"`
	Attributes       string            `gorm:"size:1024;null;" json:"attributes"`
	Categories       []*PluginCategory `gorm:"many2many:plugin_category_pivots"`
}

type PluginVersion struct {
	gorm.Model
	Plugin      *Plugin                  `gorm:"foreignkey:PluginId"  json:"plugin,omitempty"`
	PluginID    uint64                   `gorm:"not null;" json:"plugin_id"`
	Version     string                   `gorm:"size:1024;not null;" json:"version"`
	Url         string                   `gorm:"size:255;null;" json:"url"`
	ChangeLogs  string                   `gorm:"size:1024;not null;" json:"change_logs"`
	Attributes  string                   `gorm:"size:1024;null;" json:"attributes"`
	ReleaseDate time.Time                `gorm:"default:CURRENT_TIMESTAMP" json:"released"`
	Active      bool                     `gorm:"not null;" json:"active"`
	Versions    []map[string]interface{} `sql:"-" json:"versions"`
	Upgradable  bool                     `gorm:"default:true" json:"upgradable"`
}

func (data *Plugin) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Description = html.EscapeString(strings.TrimSpace(data.Description))
	data.Image = html.EscapeString(strings.TrimSpace(data.Image))
	data.SourceUrl = html.EscapeString(strings.TrimSpace(data.SourceUrl))
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *Plugin) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	if data.Description == "" {
		return errors.New("required description")
	}
	if data.SourceUrl == "" {
		return errors.New("required source url")
	}
	if data.MinCpu == 0 {
		return errors.New("required minimum cpu")
	}
	if data.MinMemory == 0 {
		return errors.New("required minimum memory")
	}
	if data.IsManagedService {
		if data.SupportCi {
			return errors.New("managed service cannot support ci")
		}
	}
	return nil
}

func (d *PluginType) Save(db *gorm.DB, data *Plugin) (*Plugin, error) {
	count := 0
	db.Model(&Plugin{}).Where("name = ?", data.Name).Count(&count)
	if count > 0 {
		return nil, errors.New("name " + data.Name + " exists.")
	}
	err := db.Model(&Plugin{}).Create(&data).Error
	if err != nil {
		return &Plugin{}, err
	}
	return data, nil
}

func (d *PluginType) FindBySupportCi(db *gorm.DB, ci bool, isManagedService bool, oid uint, page uint64, size uint64, q string, sortColumn string, sortDirection string) (*[]Plugin, int, error) {
	var err error
	datas := []Plugin{}
	count := 0
	if oid > 0 {
		err = db.
			Model(&Plugin{}).
			Where("id in (select plugin_id from organization_plugins where organization_id = ? )", oid).
			Where("support_ci = ? and is_managed_service = ? and active = ? and is_add_on = ?", ci, isManagedService, true, false).
			Where("name LIKE ? OR description LIKE ?", "%"+q+"%", "%"+q+"%").
			Preload("Categories").
			Order(sortColumn + " " + sortDirection).
			Limit(size).Offset(size * (page - 1)).
			Find(&datas).Count(&count).Error
		if err != nil {
			return &[]Plugin{}, 0, err
		}
		return &datas, count, nil
	}

	err = db.
		Model(&Plugin{}).
		Where("support_ci = ? and is_managed_service = ? and active = ? and is_add_on = ?", ci, isManagedService, true, false).
		Where("name LIKE ? OR description LIKE ?", "%"+q+"%", "%"+q+"%").
		Preload("Categories").
		Order(sortColumn + " " + sortDirection).
		Limit(size).Offset(size * (page - 1)).
		Find(&datas).Count(&count).Error
	if err != nil {
		return &[]Plugin{}, 0, err
	}
	return &datas, count, nil
}

func (d *PluginType) FindPluginByIds(db *gorm.DB, ids []string) ([]*Plugin, error) {
	var err error
	datas := []*Plugin{}
	err = db.Model(&Plugin{}).Where("id in (?) and active = ?", ids, true).Order("id desc").Find(&datas).Error
	if err != nil {
		return []*Plugin{}, err
	}
	return datas, nil
}

func (d *PluginType) FindAll(db *gorm.DB) (*[]Plugin, error) {
	var err error
	datas := []Plugin{}
	err = db.Model(&Plugin{}).Order("id desc").Preload("AddOns").Find(&datas).Error
	if err != nil {
		return &[]Plugin{}, err
	}
	return &datas, nil
}

func (d *PluginType) FindAllActiveWithFilters(db *gorm.DB, page uint64, size uint64, q string, sortColumn string, sortDirection string) (*[]Plugin, int, error) {
	var err error
	datas := []Plugin{}
	count := 0
	err = db.Model(&Plugin{}).
		Where("active = ?", true).
		Where("is_add_on = ?", false).
		Where("name LIKE ? OR description LIKE ?", "%"+q+"%", "%"+q+"%").
		Preload("Categories").
		Order(sortColumn + " " + sortDirection).
		Limit(size).
		Offset(size * (page - 1)).
		Find(&datas).Count(&count).Error
	if err != nil {
		return &[]Plugin{}, 0, err
	}
	return &datas, count, nil
}

func (d *PluginType) FindAllWithFilters(db *gorm.DB, page uint64, size uint64, q string, sortColumn string, sortDirection string) (*[]Plugin, int, error) {
	var err error
	datas := []Plugin{}
	count := 0
	err = db.Model(&Plugin{}).Where("name LIKE ? OR description LIKE ?", "%"+q+"%", "%"+q+"%").
		Preload("Categories").Order(sortColumn + " " + sortDirection).
		Limit(size).Offset(size * (page - 1)).Find(&datas).Error
	if err != nil {
		return &[]Plugin{}, 0, err
	}
	db.Model(&Plugin{}).Where("name LIKE ? OR description LIKE ?", "%"+q+"%", "%"+q+"%").Count(&count)
	return &datas, count, nil
}

func (d *PluginType) Find(db *gorm.DB, pid uint64) (*Plugin, error) {
	data := &Plugin{}
	var err error
	print(pid)
	err = db.Model(&Plugin{}).Where("id = ?", pid).Preload("AddOns").
		Preload("Categories").Preload("AddOns.Categories").Take(&data).Error
	if err != nil {
		return &Plugin{}, err
	}
	return data, nil
}
func checkExists(data *Plugin, list []*Plugin) bool {
	for _, d := range list {
		if d.ID == data.ID {
			return true
		}
	}
	return false
}

func (d *PluginType) FindAddOns(db *gorm.DB, pid uint64, query string, catid []string) ([]*Plugin, error) {
	data := &Plugin{}
	err := db.Model(&Plugin{}).Where("id = ?", pid).Preload("AddOns").Preload("AddOns.Categories").Take(&data).Error
	if err != nil {
		return nil, err
	}
	result := []*Plugin{}
	catdat := []*Plugin{}
	if len(catid) != 0 {
		err := db.Model(&Plugin{}).Where("id in(select plugin_id from plugin_category_pivots where plugin_category_id in (?))", catid).Find(&catdat).Error
		if err != nil {
			return nil, err
		}
	}

	for _, addon := range data.AddOns {
		if (len(catid) == 0 || checkExists(addon, catdat)) && strings.Contains(addon.Name, query) {
			result = append(result, addon)
		}
	}
	return result, nil
}

func (d *PluginType) Update(db *gorm.DB, data *Plugin) (*Plugin, error) {
	var err error
	mp := map[string]interface{}{
		"active":     data.Active,
		"support_ci": data.SupportCi,
		"is_add_on":  data.IsAddOn,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Description != "" {
		mp["description"] = data.Description
	}
	if data.Image != "" {
		mp["image"] = data.Image
	}
	if data.ServiceDetail.RawMessage != nil {
		mp["service_detail"] = data.ServiceDetail
	}
	if data.SourceUrl != "" {
		mp["source_url"] = data.SourceUrl
	}
	if data.MinMemory != 0 {
		mp["min_memory"] = data.MinMemory
	}
	if data.MinCpu != 0 {
		mp["min_cpu"] = data.MinCpu
	}
	if data.Attributes != "" {
		mp["attributes"] = data.Attributes
	}
	err = db.Model(&Plugin{}).Where("id = ?", data.ID).UpdateColumn(mp).Error
	if err != nil {
		return &Plugin{}, err
	}
	return data, nil
}

func (d *PluginType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&Plugin{}).Where("id = ?", id).Take(&Plugin{}).Delete(&Plugin{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Plugin not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

///////////////////

func (data *PluginVersion) Prepare() {
	data.ID = 0
	data.Version = html.EscapeString(strings.TrimSpace(data.Version))
	data.ChangeLogs = html.EscapeString(strings.TrimSpace(data.ChangeLogs))
	data.Url = html.EscapeString(strings.TrimSpace(data.Url))
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *PluginVersion) Validate() error {
	if data.Version == "" {
		return errors.New("required version")
	}
	if data.ChangeLogs == "" {
		return errors.New("required changelogs")
	}
	if data.Url == "" {
		return errors.New("required Url")
	}
	if data.PluginID == 0 {
		return errors.New("required plugin")
	}
	return nil
}

func (d *PluginVersionType) Save(db *gorm.DB, data *PluginVersion) (*PluginVersion, error) {
	err := db.Model(&PluginVersion{}).Create(&data).Error
	if err != nil {
		return &PluginVersion{}, err
	}
	return data, nil
}

func (data *PluginVersionType) FindAll(db *gorm.DB) (*[]PluginVersion, error) {
	datas := []PluginVersion{}
	err := db.Model(&PluginVersion{}).Where("active = ?", true).Preload("Plugin").Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]PluginVersion{}, err
	}
	return &datas, nil
}

func (data *PluginVersionType) FindAllWithInactive(db *gorm.DB) (*[]PluginVersion, error) {
	var err error
	datas := []PluginVersion{}
	err = db.Model(&PluginVersion{}).Preload("Plugin").Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]PluginVersion{}, err
	}
	return &datas, nil
}

func (data *PluginVersionType) FindAllByPlugin(db *gorm.DB, pid uint64) (*[]PluginVersion, error) {
	var err error
	datas := []PluginVersion{}
	err = db.Model(&PluginVersion{}).Preload("Plugin").Where(&PluginVersion{PluginID: pid}).Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]PluginVersion{}, err
	}
	return &datas, nil
}

func (d *PluginVersionType) Find(db *gorm.DB, pid uint64) (*PluginVersion, error) {
	data := PluginVersion{}
	err := db.Model(&PluginVersion{}).Where("id = ?", pid).Preload("Plugin").Take(&data).Error
	if err != nil {
		return &PluginVersion{}, err
	}
	return &data, nil
}
func (d *PluginVersionType) FindLatestPluginVersion(db *gorm.DB, pid uint64) (*PluginVersion, error) {
	data := PluginVersion{}
	err := db.Model(&PluginVersion{}).Where(&PluginVersion{PluginID: pid, Active: true}).Preload("Plugin").Order("id desc").First(&data).Error
	if err != nil {
		return &PluginVersion{}, err
	}
	return &data, nil
}

func (d *PluginVersionType) Update(db *gorm.DB, data *PluginVersion) (*PluginVersion, error) {
	mp := map[string]interface{}{
		"active": data.Active,
	}
	if data.Version != "" {
		mp["version"] = data.Version
	}
	if data.ChangeLogs != "" {
		mp["change_logs"] = data.ChangeLogs
	}
	if data.Url != "" {
		mp["url"] = data.Url
	}

	err := db.Model(&PluginVersion{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &PluginVersion{}, err
	}
	return data, nil
}

func (data *PluginVersionType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&PluginVersion{}).Where("id = ?", id).Take(&PluginVersion{}).Delete(&PluginVersion{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("PluginVersion not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
