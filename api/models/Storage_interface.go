package models

import "github.com/jinzhu/gorm"

type StorageInterface interface {
	SaveStorage(db *gorm.DB, data Storage) (*Storage, error)
	Find(db *gorm.DB, pid uint64) (*Storage, error)
	Update(db *gorm.DB, data Storage) (*Storage, error)
	Delete(db *gorm.DB, uid uint64) (int64, error)
}
type StorageType struct {
}

func NewStorage() StorageInterface {
	return &StorageType{}
}
