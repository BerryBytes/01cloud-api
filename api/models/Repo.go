package models

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Repo struct {
	ID             string         `gorm:"type:uuid;primaryKey"`
	Name           string         `gorm:"column:name;type:text"`
	URL            string         `gorm:"column:url;type:text"`
	RepoType       string         `gorm:"column:repo_type;type:text"`
	Filter         string         `gorm:"column:filter;type:text"`
	OrganizationId int64          `gorm:"column:organization_id"`
	Authorization  postgres.Jsonb `sql:"json" json:"authorization"`
}

type RepoInterface interface {
	BeforeCreate(tx *gorm.DB, u *Repo) (err error)
	BeforeDelete(tx *gorm.DB, u *Repo) (err error)
}
type RepoType struct {
}

func NewRepo() RepoInterface {
	return &RepoType{}
}

func (d *RepoType) BeforeCreate(tx *gorm.DB, u *Repo) (err error) {
	u.ID = uuid.New().String()
	return
}
func (d *RepoType) BeforeDelete(tx *gorm.DB, u *Repo) (err error) {
	var count int64
	err = tx.Model(&Chart{}).Where(&Chart{RepoID: u.ID}).Count(&count).Error
	if err != nil {
		return errors.New("repository contains catalogues, please delete them first")
	}
	return
}
