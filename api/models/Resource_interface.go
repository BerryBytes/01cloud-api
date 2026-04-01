package models

import (
	"01cloud-api/api/utils/cache"

	"github.com/jinzhu/gorm"
)

type ResourceInterface interface {
	Save(db *gorm.DB, data *Resource) (*Resource, error)
	FindAll(db *gorm.DB, data *Resource) (*[]Resource, error)
	Find(db *gorm.DB, pid uint64) (*Resource, error)
	FindAllWithInactive(db *gorm.DB, data *Resource) (*[]Resource, error)
	Update(db *gorm.DB, data *Resource) (*Resource, error)
	Delete(db *gorm.DB, uid uint64) (int64, error)
	CheckResourceLimit(db *gorm.DB, org *Organization, data *Resource) bool
	IsNameExists(db *gorm.DB, oId uint, name string) bool
}
type ResourceType struct {
	Cache cache.ICache
}

func NewResource() ResourceInterface {
	return &ResourceType{
		Cache: cache.NewCache(CacheDuration()),
	}
}
