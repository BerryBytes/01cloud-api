package models

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type ChartVersion struct {
	ID         string         `gorm:"type:uuid;primaryKey"`
	Chart      *Chart         `gorm:"foreignkey:ChartID"  json:"chart,omitempty"`
	ChartID    string         `json:"chart_id"`
	Version    string         `gorm:"column:version;size:255;null;" json:"version"`
	AppVersion string         `gorm:"column:app_version;size:255;null;" json:"app_version"`
	Digest     string         `json:"digest"`
	URLs       postgres.Jsonb `sql:"json" json:"urls"`
	Readme     string         `gorm:"column:readme;type:text" json:"readme"`
	Values     string         `gorm:"column:values;type:text" json:"values"`
	Schema     string         `gorm:"column:schema;type:text" json:"schema"`
}

type ChartVersionInterface interface {
	Find(*gorm.DB, string) (*ChartVersion, error)
}

type ChartVersionType struct {
}

func NewChartVersion() ChartVersionInterface {
	return &ChartVersionType{}
}

func (u *ChartVersion) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}

func (u *ChartVersionType) Find(db *gorm.DB, pid string) (*ChartVersion, error) {
	data := ChartVersion{}
	err = db.Model(&ChartVersion{}).
		Where("ID = ?", pid).
		Preload("Chart").
		Take(&data).Error
	if err != nil {
		return &ChartVersion{}, err
	}
	return &data, nil
}
