package models

import "github.com/jinzhu/gorm"

type IDNS interface {
	Save(db *gorm.DB, data *DNS) (*DNS, error)
	IsNameExists(db *gorm.DB, oid uint64, name string) bool
	FindAllByOrganization(db *gorm.DB, orgId uint) (*[]DNS, error)
	FindAll(db *gorm.DB) (*[]DNS, error)
	FindAllWithInactive(db *gorm.DB) (*[]DNS, error)
	Find(db *gorm.DB, pid uint64) (*DNS, error)
	Update(db *gorm.DB, data *DNS) (*DNS, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
}

var err error

type DNSType struct {
}

func NewDNS() IDNS {
	return &DNSType{}
}
