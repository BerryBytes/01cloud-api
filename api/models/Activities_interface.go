package models

import "github.com/jinzhu/gorm"

type ActivityInterface interface {
	Save(db *gorm.DB, data Activity) (*Activity, error)
	FindAll(db *gorm.DB) (*[]Activity, error)
	Find(db *gorm.DB, pid uint64) (*Activity, error)
	FindAllActivityByProject(db *gorm.DB, pid uint, limit int, offset int, action, module, startdate, enddate string, userid int) (*[]Activity, int, error)
	FindAllActivityByOrganization(db *gorm.DB, oid uint, limit int, offset int, action, module, startdate, enddate string, userid int) (*[]Activity, int, error)
	Update(db *gorm.DB, data Activity) (*Activity, error)
	Delete(db *gorm.DB, uid uint64) (int64, error)
}
type ActivityType struct {
}

func NewActivity() ActivityInterface {
	return &ActivityType{}
}
