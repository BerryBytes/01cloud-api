package models

import (
	"errors"
	"html"
	"strings"

	"github.com/jinzhu/gorm"
)

type Storage struct {
	gorm.Model
	EnvironmentID uint    `json:"environment_id"`
	Name          string  `gorm:"size:255;not null;" json:"name"`
	VolumeName    *string `gorm:"size:255;null;" json:"volume_name"`
	AccessModes   *string `gorm:"size:255;null;" json:"access_modes"`
	MountPath     *string `gorm:"size:255;null;" json:"mount_path"`
	StorageType   *string `gorm:"size:255;null;" json:"storage_type"`
	AttachedTo    *string `gorm:"size:255;null;" json:"attached_to"`
	Capacity      uint64  `gorm:"not null;" json:"capacity"`
	UsedStorage   *uint64 `gorm:"default 0;" json:"used_storage"`
	Active        bool    `gorm:"default true;" json:"active"`
	User          *User   `gorm:"foreignkey:UserID;null" json:"user,omitempty"`
	UserID        uint    `gorm:"default:0;" json:"user_id"`
}

type StorageResponse struct {
	EnvironmentID uint64    `json:"environment_id"`
	StorageList   []Storage `json:"storage"`
}

func (d *StorageType) SaveStorage(db *gorm.DB, data Storage) (*Storage, error) {
	err := db.Model(&Storage{}).Save(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (data *Storage) Prepare(temp *Storage) {
	if temp != nil {
		if temp.Name != "" {
			data.Name = temp.Name
		}
		if temp.AttachedTo != nil {
			data.AttachedTo = temp.AttachedTo
		}
		if temp.AccessModes != nil {
			data.AccessModes = temp.AccessModes
		}
		if temp.Active != data.Active {
			data.Active = temp.Active
		}
		if temp.Capacity != 0 {
			data.Capacity = temp.Capacity
		}
		if temp.MountPath != nil {
			data.MountPath = temp.MountPath
		}
		if temp.StorageType != nil {
			data.StorageType = temp.StorageType
		}
		if temp.UsedStorage != nil {
			data.UsedStorage = temp.UsedStorage
		}
		if temp.VolumeName != nil {
			data.VolumeName = temp.VolumeName
		}
	}
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
}

func (data *Storage) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	return nil
}

func (d *StorageType) Update(db *gorm.DB, data Storage) (*Storage, error) {
	err := db.Model(&data).Update(&data).Take(&data).Error
	if err != nil {
		return &Storage{}, err
	}
	return &data, nil
}

func (data *Storage) FindAllWithFilters(db *gorm.DB, eid uint, page uint64, size uint64) (*[]Storage, error) {
	dataList := []Storage{}
	err := db.Model(&Storage{}).Where(&Storage{EnvironmentID: eid}).Limit(size).Offset(size * (page - 1)).Find(&dataList).Error
	if err != nil {
		return &[]Storage{}, err
	}
	return &dataList, nil
}

func (d *StorageType) Find(db *gorm.DB, pid uint64) (*Storage, error) {
	var err error
	data := Storage{}
	err = db.Model(&Storage{}).
		Where("id = ?", pid).
		Preload("User").
		Take(&data).Error
	if err != nil {
		return &Storage{}, err
	}
	return &data, nil
}
func (d *StorageType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&Storage{}).Where("id = ?", id).Take(&Storage{}).Delete(&Storage{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Resource not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
