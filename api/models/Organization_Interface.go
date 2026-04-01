package models

import "github.com/jinzhu/gorm"

type OrganizationInterface interface {
	Save(db *gorm.DB, data *Organization) (*Organization, error)
	FindAll(db *gorm.DB, userID uint, active bool, query string) (*[]Organization, error)
	Find(db *gorm.DB, oid uint) (*Organization, error)
	Update(db *gorm.DB, data *Organization) (*Organization, error)
	Delete(db *gorm.DB, id int64) (int64, error)
	CountOrginationByUser(db *gorm.DB, uid uint) (int, error)
	ActiveDeactiveOrganization(db *gorm.DB, oid uint, active bool) (*Organization, error)
}
type OrganizationType struct {
}

func NewOrganization() OrganizationInterface {
	return &OrganizationType{}
}
