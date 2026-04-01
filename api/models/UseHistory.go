package models

import (
	"errors"

	"github.com/jinzhu/gorm"
)

type IOperation interface {
	ProjectUsageHistory(db *gorm.DB) error
}

type UsageHistory struct {
	gorm.Model
	UserId         uint `gorm:"user_id" json:"user_id"`
	OrganizationId uint `gorm:"organization_id" json:"organization_id"`
	ProjectId      uint `gorm:"project_id" json:"project_id"`
	IsOperational  bool `gorm:"is_operational" json:"is_operational"`
}

func (operation *UsageHistory) ProjectUsageHistory(db *gorm.DB) error {
	if operation == nil {
		return errors.New("operation is nil")
	}
	return db.Model(&UsageHistory{}).Create(operation).Error
}
