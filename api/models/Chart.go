package models

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Chart struct {
	ID              string          `gorm:"type:uuid;primaryKey"`
	Name            string          `gorm:"column:name;type:text"`
	Repo            *Repo           `gorm:"foreignkey:RepoID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"repo,omitempty"`
	RepoID          string          `gorm:"not null" json:"repo_id"`
	OrganizationID  int64           `gorm:"not null;default:0;" json:"organization_id"`
	Description     string          `gorm:"column:description;type:text"`
	Home            string          `gorm:"column:home;type:text"`
	Keywords        postgres.Jsonb  `sql:"json" json:"keywords"`
	Maintainers     postgres.Jsonb  `sql:"json" json:"maintainers"`
	Sources         postgres.Jsonb  `sql:"json" json:"sources"`
	Icon            string          `gorm:"column:icon;type:text"`
	IconContentType string          `gorm:"column:icon_content_type" json:"icon_content_type" bson:"icon_content_type,omitempty"`
	Category        string          `json:"category"`
	ChartVersions   []*ChartVersion `gorm:"foreignkey:ChartID;association_foreignkey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (u *Chart) BeforeCreate(tx *gorm.DB) (err error) {
	u.ID = uuid.New().String()
	return
}
