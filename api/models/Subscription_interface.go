package models

import (
	"01cloud-api/api/utils/cache"

	"github.com/jinzhu/gorm"
)

type SubscriptionInterface interface {
	Save(db *gorm.DB, data Subscription) (*Subscription, error)
	FindAll(db *gorm.DB, orgid uint64) (*[]Subscription, error)
	Find(db *gorm.DB, pid uint64) (*Subscription, error)
	//	FindAllSubscriptionByProject(pid uint, limit int, offset int, action, startdate, enddate string, userid int) (*[]Subscription, error)
	FindByUserID(db *gorm.DB, uid uint64) (*[]Subscription, error)
	FindAllByUserID(db *gorm.DB, uid uint64) (*[]Subscription, error)
	Update(db *gorm.DB, data Subscription) (*Subscription, error)
	Delete(db *gorm.DB, uid uint64) (int64, error)
}
type SubscriptionType struct {
	Cache cache.ICache
}

func NewSubscription() SubscriptionInterface {
	return &SubscriptionType{
		Cache: cache.NewCache(CacheDuration()),
	}
}
