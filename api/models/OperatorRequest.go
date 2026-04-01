package models

import (
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type OperatorRequestInterface interface {
	Save(db *gorm.DB, data OperatorRequest) (*OperatorRequest, error)
	FindAll(db *gorm.DB, data OperatorRequest) (*[]OperatorRequest, error)
	Find(db *gorm.DB, pid uint64) (*OperatorRequest, error)
	FindByNameAndClusterID(db *gorm.DB, cid uint64, packageName string) (*OperatorRequest, error)
	Update(db *gorm.DB, data OperatorRequest) (*OperatorRequest, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
}

type OperatorRequestType struct{}

func NewOperatorRequest() OperatorRequestInterface {
	return &OperatorRequestType{}
}

type OperatorRequest struct {
	gorm.Model
	PackageName         string          `gorm:"size:255;not null;" json:"package_name"`
	CsvValue            string          `gorm:"null;" json:"csv_value"`
	Cluster             *Cluster        `sql:"-" json:"cluster,omitempty"`
	ClusterRequest      *ClusterRequest `gorm:"foreignkey:ClusterRequestID" json:"cluster_request,omitempty"`
	ClusterRequestID    uint64          `gorm:"default:0" json:"cluster_request_id"`
	InstallPlanApproval string          `gorm:"size:255;not null;" json:"install_plan_approval"`
	Channel             string          `json:"channel"`
	GlobalOperator      bool            `gorm:"null;default:false" json:"global_operator"`
	InstallationMode    string          `gorm:"size:255;not null;" json:"installation_mode"`
	OperatorDetails     postgres.Jsonb  `gorm:"null" json:"operator_details"`
	Organization        *Organization   `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationID      uint64          `gorm:"default:0" json:"organization_id"`
}

func (data *OperatorRequest) Prepare() {
	data.ID = 0
	data.PackageName = html.EscapeString(strings.TrimSpace(data.PackageName))
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *OperatorRequest) Validate() error {
	if data.PackageName == "" {
		return errors.New("required package name")
	}
	if data.InstallationMode == "" {
		return errors.New("required installationMode")
	}
	if data.InstallPlanApproval == "" {
		return errors.New("required installPlanApproval")
	}
	return nil
}

func (d *OperatorRequestType) Save(db *gorm.DB, data OperatorRequest) (*OperatorRequest, error) {
	err := db.Model(&OperatorRequest{}).Create(&data).Error
	if err != nil {
		return &OperatorRequest{}, err
	}
	return &data, nil
}

func (d *OperatorRequestType) FindAll(db *gorm.DB, data OperatorRequest) (*[]OperatorRequest, error) {
	var err error
	datas := []OperatorRequest{}
	err = db.Model(&OperatorRequest{}).
		Where("cluster_request_id = ?", data.ClusterRequestID).
		Limit(100).
		Find(&datas).Error
	if err != nil {
		return &[]OperatorRequest{}, err
	}
	return &datas, nil
}

func (d *OperatorRequestType) Find(db *gorm.DB, pid uint64) (*OperatorRequest, error) {
	data := OperatorRequest{}
	err := db.Model(&OperatorRequest{}).
		Where("id = ?", pid).
		Preload("ClusterRequest").
		Preload("ClusterRequest.Cluster").
		Take(&data).Error
	if err != nil {
		return &OperatorRequest{}, err
	}
	return &data, nil
}

func (d *OperatorRequestType) FindByNameAndClusterID(db *gorm.DB, cid uint64, packageName string) (*OperatorRequest, error) {
	data := OperatorRequest{}
	err := db.Model(&OperatorRequest{}).
		Where("cluster_request_id = ?", cid).
		Where("package_name = ?", packageName).
		Preload("ClusterRequest").
		Preload("ClusterRequest.Cluster").
		Take(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (d *OperatorRequestType) Update(db *gorm.DB, data OperatorRequest) (*OperatorRequest, error) {
	var err error
	//var app = OperatorRequest{Active: data.Active}
	mp := map[string]interface{}{}
	if data.PackageName != "" {
		mp["package_name"] = data.PackageName
	}
	if data.InstallationMode != "" {
		mp["installation_mode"] = data.InstallationMode
	}
	if data.InstallPlanApproval != "" {
		mp["install_plan_approval"] = data.InstallPlanApproval
	}
	if data.CsvValue != "" {
		mp["csv_value"] = data.CsvValue
	}
	if data.Channel != "" {
		mp["channel"] = data.Channel
	}

	err = db.Model(&OperatorRequest{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &OperatorRequest{}, err
	}
	return &data, nil
}

func (data *OperatorRequestType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&OperatorRequest{}).Where("id = ?", id).Take(&OperatorRequest{}).Delete(&OperatorRequest{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("OperatorRequest not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
