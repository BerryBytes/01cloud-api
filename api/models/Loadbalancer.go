package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type LoadBalancer struct {
	gorm.Model
	Name         string         `gorm:"size:255;not null;" json:"name"`
	CustomDomain string         `gorm:"size:255;null;" json:"custom_domain"`
	Cluster      *Cluster       `gorm:"foreignkey:ClusterID" json:"cluster,omitempty"`
	ClusterID    uint64         `gorm:"default:0" json:"cluster_id"`
	Attributes   postgres.Jsonb `gorm:"size:1024;null;" json:"attributes"`
	Project      *Project       `gorm:"foreignkey:ProjectID" json:"project,omitempty"`
	ProjectID    uint64         `gorm:"default:0" json:"project_id"`
}

func (data *LoadBalancer) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *LoadBalancer) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if data.ClusterID == 0 {
		return errors.New("required cluster")
	}
	return nil
}

func (data *LoadBalancer) IsLoadbalancerUsed(db *gorm.DB) bool {
	count := 0
	err := db.Model(&Environment{}).Where(&Environment{LoadBalancerID: uint64(data.ID)}).Count(&count).Error
	if err != nil {
		return false
	}
	return count > 0
}

func (d *LoadBalancerType) Save(db *gorm.DB, data *LoadBalancer) (*LoadBalancer, error) {
	err := db.Model(&LoadBalancer{}).Create(&data).Error
	if err != nil {
		return &LoadBalancer{}, err
	}
	return data, nil
}

func (d *LoadBalancerType) FindAllByProject(db *gorm.DB, id uint64) (*[]LoadBalancer, error) {
	var err error
	datas := []LoadBalancer{}
	err = db.Model(&LoadBalancer{}).
		Where("project_id = ?", id).
		Find(&datas).Error
	if err != nil {
		return &[]LoadBalancer{}, err
	}
	return &datas, nil
}

func (data *LoadBalancer) FindAllWithInactive(db *gorm.DB) (*[]LoadBalancer, error) {
	var err error
	datas := []LoadBalancer{}
	err = db.Model(&LoadBalancer{}).Preload("Project").Preload("Cluster").
		Preload("Project.User").Joins("left join projects on projects.id=load_balancers.project_id").
		Where("projects.organization_id=0 and projects.deleted_at is null").Find(&datas).Error
	if err != nil {
		return &[]LoadBalancer{}, err
	}
	return &datas, nil
}
func (data *Environment) FindEnvironmentByLoadBalancer(db *gorm.DB, id uint64) (*[]Environment, error) {
	datas := []Environment{}
	err := db.Model(&Environment{}).
		Where("load_balancer_id = ?", id).
		Find(&datas).Error
	if err != nil {
		return &[]Environment{}, err
	}
	return &datas, nil
}

func (d *LoadBalancerType) Find(db *gorm.DB, pid uint64) (*LoadBalancer, error) {
	var err error
	data := LoadBalancer{}
	err = db.Model(&LoadBalancer{}).
		Where("id = ?", pid).
		Preload("Cluster").
		Preload("Cluster.DNS").
		Preload("Project").
		Take(&data).Error
	if err != nil {
		return &LoadBalancer{}, err
	}
	return &data, nil
}

func (d *LoadBalancerType) Update(db *gorm.DB, data *LoadBalancer) (*LoadBalancer, error) {
	var err error
	mp := map[string]interface{}{}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.CustomDomain != "" {
		mp["custom_domain"] = data.CustomDomain
	}

	err = db.Model(&LoadBalancer{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &LoadBalancer{}, err
	}
	return data, nil
}

func (d *LoadBalancerType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&LoadBalancer{}).Where("id = ?", id).Take(&LoadBalancer{}).Delete(&LoadBalancer{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("LoadBalancer not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
