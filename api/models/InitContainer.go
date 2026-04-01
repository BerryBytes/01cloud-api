package models

import (
	"errors"
	"html"
	"strings"

	"github.com/jinzhu/gorm"
)

type InitContainerInterface interface {
	SaveInitContainer(db *gorm.DB, pid uint64, data *InitContainer) (*InitContainer, error)
	FindAllWithFilters(db *gorm.DB, eid uint, page uint64, size uint64) (*[]InitContainer, error)
	Update(db *gorm.DB, data *InitContainer) (*InitContainer, error)
	Find(db *gorm.DB, pid uint64) (*InitContainer, error)
}

type InitContainerType struct{}

func NewInitContainer() InitContainerInterface {
	return &InitContainerType{}
}

type InitContainer struct {
	gorm.Model
	EnvironmentID uint
	Image         string `gorm:"size:255;not null;" json:"image"`
	Name          string `gorm:"size:255;not null;" json:"name"`
	Command       string `gorm:"not null;" json:"command"`
}

func (d *InitContainerType) SaveInitContainer(db *gorm.DB, pid uint64, data *InitContainer) (*InitContainer, error) {
	var err error
	data.EnvironmentID = uint(pid)
	err = db.Model(&InitContainer{}).Save(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (data *InitContainer) Prepare() {
	//d, _ := json.Marshal(map[string]string{"app/name": data.Name})
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Command = html.EscapeString(strings.TrimSpace(data.Command))
}

func (data *InitContainer) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	if data.Command == "" {
		return errors.New("required command")
	}
	return nil
}

func (d *InitContainerType) Update(db *gorm.DB, data *InitContainer) (*InitContainer, error) {
	err := db.Model(&data).Update(&data).Error
	if err != nil {
		return &InitContainer{}, err
	}
	return data, nil
}

func (d *InitContainerType) FindAllWithFilters(db *gorm.DB, eid uint, page uint64, size uint64) (*[]InitContainer, error) {
	dataList := []InitContainer{}
	err := db.Model(&InitContainer{}).Where(&InitContainer{EnvironmentID: eid}).Limit(size).Offset(size * (page - 1)).Find(&dataList).Error
	if err != nil {
		return &[]InitContainer{}, err
	}
	return &dataList, nil
}

func (d *InitContainerType) Find(db *gorm.DB, pid uint64) (*InitContainer, error) {
	var err error
	var data = InitContainer{}
	print(pid)
	err = db.Model(&InitContainer{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &InitContainer{}, err
	}
	return &data, nil
}
