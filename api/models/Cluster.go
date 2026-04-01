package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type ClusterInterface interface {
	Save(db *gorm.DB, data Cluster) (*Cluster, error)
	FindWithDns(db *gorm.DB, pid uint64) (*Cluster, error)
	FindAll(db *gorm.DB) (*[]Cluster, error)
	FindAllClustersByRegion(db *gorm.DB, org uint, region string) (*[]Cluster, error)
	FindAllWithInActive(db *gorm.DB) (*[]Cluster, error)
	FindRegions(db *gorm.DB, oid uint) (*[]string, error)
	FindAllRegions(db *gorm.DB, oid uint) (*[]string, error)
	Find(db *gorm.DB, pid uint64) (*Cluster, error)
	FindZeroneCluster(db *gorm.DB) (*Cluster, error)
	FindAllClusterWithOrganizationId(db *gorm.DB, oid uint64) (*[]Cluster, error)
	FindWithDetails(db *gorm.DB, pid uint64) (*Cluster, error)
	FindByRegion(db *gorm.DB, region string, oid uint) (*Cluster, error)
	FindByRegionWithDns(db *gorm.DB, region string, oid uint) (*Cluster, error)
	Update(db *gorm.DB, data Cluster) (*Cluster, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
	DeleteByClusterRequest(db *gorm.DB, id uint64) (int64, error)
	UpdateLabelsAndColor(db *gorm.DB, data Cluster) (*Cluster, error)
}
type ClusterType struct {
}

func NewCluster() ClusterInterface {
	return &ClusterType{}
}

type Cluster struct {
	gorm.Model
	Name                string         `gorm:"size:255;not null;" json:"name"`
	Context             string         `gorm:"size:255;null;" json:"context"`
	ConfigPath          string         `gorm:"size:255;not null;" json:"configPath"`
	Token               string         `gorm:"size:255;null;" json:"token"`
	Region              string         `gorm:"size:255;not null;" json:"region"`
	Provider            string         `gorm:"size:255;not null;" json:"provider"`
	ProjectName         string         `gorm:"size:255;null;default:null" json:"project_name"`
	Zone                string         `gorm:"size:255;null;default:null" json:"zone"`
	DNS                 *DNS           `gorm:"foreignkey:DNSId" json:"dns,omitempty"`
	DNSId               uint64         `gorm:"default:0" json:"dns_id"`
	Labels              string         `gorm:"null;default:null"  json:"labels"`
	Nodes               uint64         `gorm:"not null;" json:"nodes"`
	PvCapacity          uint64         `gorm:"null;default:20;" json:"pv_capacity"`
	Weight              uint32         `gorm:"default:10;" json:"weight"`
	Attributes          string         `gorm:"size:1024;null;" json:"attributes"`
	PrometheusServerUrl string         `gorm:"default:null;"`
	CloudStorage        postgres.Jsonb `json:"cloud_storage"`
	ImageRegistry       *ImageRegistry `gorm:"foreignkey:ImageRegistryID" json:"image_registry,omitempty"`
	ImageRegistryID     uint64         `gorm:"default:0" json:"image_registry_id"`
	// StorageAccessKey    string          `gorm:"default:null;" json:"-"`
	// StorageSecretKey    string          `gorm:"default:null;" json:"-"`
	Active              bool            `gorm:"not null;" json:"active"`
	Organization        *Organization   `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationID      uint64          `gorm:"default:0" json:"organization_id"`
	TotalMemory         uint64          `gorm:"default:30720" json:"total_memory"`
	ProvisionPercentage float64         `gorm:"default:0.7" json:"provision_percentage"`
	ClusterRequest      *ClusterRequest `gorm:"foreignkey:ClusterRequestID" json:"cluster_request"`
	ClusterRequestID    uint64          `gorm:"default:0" json:"cluster_request_id"`
	Color               string          `gorm:"size:255;null;" json:"color"`
	StorageClass        string          `gorm:"size:255;null;" json:"storage_class"`
	SnapshotClass       string          `gorm:"size:255;null;" json:"snapshot_class"`
}

func (data *Cluster) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.ConfigPath = html.EscapeString(strings.TrimSpace(data.ConfigPath))
	data.Context = html.EscapeString(strings.TrimSpace(data.Context))
	data.Token = html.EscapeString(strings.TrimSpace(data.Token))
	data.Region = html.EscapeString(strings.TrimSpace(data.Region))
	data.Provider = html.EscapeString(strings.TrimSpace(data.Provider))
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *Cluster) IsNameExists(db *gorm.DB) bool {
	count := 0
	db.Model(&Cluster{}).
		Where(&Cluster{Name: data.Name, OrganizationID: data.OrganizationID}).
		Count(&count)
	return count > 0
}

func (data *Cluster) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if data.ConfigPath == "" {
		return errors.New("config file is required")
	}
	if data.PrometheusServerUrl == "" {
		return errors.New("prometheus server url  is required")
	}
	if data.PvCapacity == 0 {
		return errors.New("number of pv is required")
	}
	return nil
}

func (data *Cluster) ValidateImportCluster() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if data.Region == "" {
		return errors.New(" required region")
	}
	if data.Provider == "" {
		return errors.New("required provider")
	}
	return nil
}

func (d *ClusterType) Save(db *gorm.DB, data Cluster) (*Cluster, error) {
	err := db.Model(&Cluster{}).Create(&data).Error
	if err != nil {
		return &Cluster{}, err
	}
	return &data, err
}

func (data *ClusterType) FindAll(db *gorm.DB) (*[]Cluster, error) {
	var err error
	datas := []Cluster{}
	err = db.Model(&Cluster{}).Where("active = ? ", true).Order("weight desc").Limit(100).Find(&datas).Error
	if err != nil {
		return &[]Cluster{}, err
	}
	return &datas, nil
}

func (data *ClusterType) FindAllClustersByRegion(db *gorm.DB, org uint, region string) (*[]Cluster, error) {
	datas := []Cluster{}
	dbs := db.Model(&Cluster{}).Where("active = ? ", true)
	if org != 0 {
		dbs = dbs.Where("organization_id = ? ", org)
	}
	if region != "" {
		dbs = dbs.Where("region = ? ", region)
	}
	err := dbs.Order("weight desc").Limit(100).Find(&datas).Error
	if err != nil {
		return &[]Cluster{}, err
	}
	return &datas, nil
}

func (data *ClusterType) FindAllWithInActive(db *gorm.DB) (*[]Cluster, error) {
	var err error
	datas := []Cluster{}
	err = db.Model(&Cluster{}).Order("weight desc").Find(&datas).Error
	if err != nil {
		return &[]Cluster{}, err
	}
	return &datas, nil
}

func (data *ClusterType) FindRegions(db *gorm.DB, oid uint) (*[]string, error) {
	var err error
	datas := []Cluster{}
	err = db.Model(&Cluster{}).Where("active = ? and (organization_id = ? or organization_id = ?)", true, oid, 0).Order("weight desc").Find(&datas).Error
	if err != nil {
		return &[]string{}, err
	}
	res := map[string]bool{}
	for _, cluster := range datas {
		if cluster.Name != "" {
			res[cluster.Name] = true
		}
	}
	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	return &keys, nil
}

func (data *ClusterType) FindAllRegions(db *gorm.DB, oid uint) (*[]string, error) {
	var err error
	datas := []Cluster{}
	err = db.Model(&Cluster{}).Where("active = ? and organization_id = ?", true, oid).Order("weight desc").Find(&datas).Error
	if err != nil {
		return &[]string{}, err
	}
	res := map[string]bool{}
	for _, cluster := range datas {
		if cluster.Region != "" {
			res[cluster.Region] = true
		}
	}
	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	return &keys, nil
}

func (data *ClusterType) Find(db *gorm.DB, pid uint64) (*Cluster, error) {
	cluster := Cluster{}
	err := db.Model(&Cluster{}).Where("id = ?", pid).Take(&cluster).Error
	if err != nil {
		return &Cluster{}, err
	}
	return &cluster, nil
}

func (data *Environment) FindEnvironmentByCluster(db *gorm.DB, pid uint64) (*[]Environment, error) {
	datas := []Environment{}
	error := db.Model(&Environment{}).Preload("Application").Preload("Application.Project").Where("environments.application_id in (select id from applications where applications.cluster_id=? and applications.deleted_at is null)", pid).Find(&datas).Error
	if error != nil {
		return &[]Environment{}, error
	}
	return &datas, nil
}

func (d *ClusterType) FindWithDns(db *gorm.DB, pid uint64) (*Cluster, error) {
	var err error
	data := &Cluster{}
	err = db.Model(&Cluster{}).Where("id = ?", pid).Preload("DNS").Preload("Organization").Take(&data).Error
	if err != nil {
		return &Cluster{}, err
	}
	return data, nil
}

func (data *ClusterType) FindZeroneCluster(db *gorm.DB) (*Cluster, error) {
	cluster := Cluster{}
	err := db.Model(&Cluster{}).Where("active = ? and organization_id = ?", true, 0).First(&cluster).Error
	if err != nil {
		return &Cluster{}, err
	}
	return &cluster, nil
}

func (data *ClusterType) FindAllClusterWithOrganizationId(db *gorm.DB, oid uint64) (*[]Cluster, error) {
	var err error
	datas := []Cluster{}
	err = db.Model(&Cluster{}).Preload("ClusterRequest").Where("organization_id = ?", oid).Find(&datas).Error
	if err != nil {
		return &[]Cluster{}, err
	}
	return &datas, nil
}

func (data *ClusterType) FindWithDetails(db *gorm.DB, pid uint64) (*Cluster, error) {
	cluster := Cluster{}
	err := db.Model(&Cluster{}).Where("id = ?", pid).Take(&cluster).Error
	if err != nil {
		return &Cluster{}, err
	}
	return &cluster, nil
}

func (data *ClusterType) FindByRegion(db *gorm.DB, region string, oid uint) (*Cluster, error) {
	cluster := Cluster{}
	err := db.Model(&Cluster{}).Where("region = ? and active = ? and organization_id in (0,?)", region, true, oid).Preload("DNS").Order("organization_id desc").Take(&cluster).Error
	if err != nil {
		return &Cluster{}, err
	}
	return &cluster, nil
}

func (data *ClusterType) FindByRegionWithDns(db *gorm.DB, region string, oid uint) (*Cluster, error) {
	cluster := Cluster{}
	err := db.Model(&Cluster{}).Where("name = ? and active = ? and organization_id in (0,?)", region, true, oid).Preload("DNS").Preload("Organization").Order("organization_id desc").Take(&cluster).Error
	if err != nil {
		return &Cluster{}, err
	}
	return &cluster, nil
}

func (cluster *ClusterType) Update(db *gorm.DB, data Cluster) (*Cluster, error) {
	var err error
	mp := map[string]interface{}{
		"active": data.Active,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.ConfigPath != "" {
		mp["config_path"] = data.ConfigPath
	}
	if data.Context != "" {
		mp["context"] = data.Context
	}
	if data.Token != "" {
		mp["token"] = data.Token
	}
	if data.Region != "" {
		mp["region"] = data.Region
	}
	if data.Provider != "" {
		mp["provider"] = data.Provider
	}
	if data.ProjectName != "" {
		mp["project_name"] = data.ProjectName
	}
	if data.Nodes != 0 {
		mp["nodes"] = data.Nodes
	}
	if data.PvCapacity != 0 {
		mp["pv_capacity"] = data.PvCapacity
	}
	if data.Attributes != "" {
		mp["attributes"] = data.Attributes
	}
	if data.PrometheusServerUrl != "" {
		mp["prometheus_server_url"] = data.PrometheusServerUrl
	}
	if data.ImageRegistryID != 0 {
		mp["image_registry_id"] = data.ImageRegistryID
	}
	if data.Zone != "" {
		mp["zone"] = data.Zone
	}
	if data.DNSId != 0 {
		mp["dns_id"] = data.DNSId
	}
	if data.Labels != "" {
		mp["labels"] = data.Labels
	}
	if data.CloudStorage.RawMessage != nil {
		mp["cloud_storage"] = data.CloudStorage
	}
	if data.Weight != 0 {
		mp["weight"] = data.Weight
	}
	if data.ClusterRequestID != 0 {
		mp["cluster_request_id"] = data.ClusterRequestID
	}
	err = db.Model(&Cluster{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &Cluster{}, err
	}
	return &data, nil
}

func (data *ClusterType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&Cluster{}).Where("id = ?", id).Take(&Cluster{}).Delete(&Cluster{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Cluster not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (data *ClusterType) DeleteByClusterRequest(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&Cluster{}).Where("cluster_request_id= ?", id).Take(&Cluster{}).Delete(&Cluster{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Cluster not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (cluster *ClusterType) UpdateLabelsAndColor(db *gorm.DB, data Cluster) (*Cluster, error) {
	mp := map[string]interface{}{}
	if data.Labels != "" {
		mp["labels"] = data.Labels
	}
	if data.Color != "" {
		mp["color"] = data.Color
	}
	err = db.Model(&Cluster{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &Cluster{}, err
	}
	return &data, nil
}
