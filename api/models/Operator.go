package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type Operator struct {
	gorm.Model
	Name                      string         `json:"name"`
	PackageName               string         `gorm:"unique" json:"packageName"`
	DisplayName               string         `json:"displayName"`
	Provider                  string         `json:"provider"`
	ThumbUrl                  string         `json:"thumbUrl"`
	Version                   string         `json:"version"`
	VersionForCompare         string         `json:"versionForCompare"`
	K8sMinVersion             string         `json:"k8sMinVersion"`
	K8sMaxVersion             string         `json:"k8sMaxVersion"`
	Replaces                  string         `json:"replaces"`
	CapabilityLevel           string         `json:"capabilityLevel"`
	Repository                string         `json:"repository"`
	Description               string         `json:"description"`
	ContainerImage            string         `json:"containerImage"`
	Channel                   string         `json:"channel"`
	GlobalOperator            bool           `json:"globalOperator"`
	Active                    bool           `gorm:"default:false" json:"active"`
	Channels                  postgres.Jsonb `json:"channels"`
	Links                     postgres.Jsonb `json:"links"`
	CustomResourceDefinitions postgres.Jsonb `json:"customResourceDefinitions"`
	Categories                postgres.Jsonb `json:"categories"`
	Keywords                  postgres.Jsonb `json:"keywords"`
	OperaterCreatedAt         string         `json:"createdAt"`
}

type OperatorReq struct {
	Data Operator `json:"operator"`
}

type OperatorRepo struct{}

type OperatorInterface interface {
	Save(db *gorm.DB, data *Operator) (*Operator, error)
	FindAllOperator(db *gorm.DB, isActive bool, limit uint64, page uint64) (map[string]interface{}, error)
	FindOperator(db *gorm.DB, packageName string) (*Operator, error)
	UpdateOperator(db *gorm.DB, data *Operator) error
	EnableDisable(db *gorm.DB, packageName string, isActive bool) error
	SyncOperator(db *gorm.DB, data *Operator) error
}

func NewOperatorRepo() OperatorInterface {
	return &OperatorRepo{}
}

func (r *OperatorRepo) Save(db *gorm.DB, data *Operator) (*Operator, error) {
	if data == nil {
		return nil, errors.New("required fields")
	}
	err = db.Model(&Operator{}).Create(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *OperatorRepo) EnableDisable(db *gorm.DB, packageName string, isActive bool) error {
	err := db.Model(&Operator{}).Where("package_name = ?", packageName).Take(&Operator{}).UpdateColumns(
		map[string]interface{}{
			"active": isActive,
		},
	).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *OperatorRepo) UpdateOperator(db *gorm.DB, data *Operator) error {
	if data == nil {
		return errors.New("required fileds")
	}
	err = db.Model(&Operator{}).Where("package_name = ?", data.PackageName).Take(&Operator{}).Update(data).UpdateColumn(map[string]interface{}{
		"active":          data.Active,
		"global_operator": data.GlobalOperator,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (r *OperatorRepo) FindAllOperator(db *gorm.DB, isActive bool, size, page uint64) (map[string]interface{}, error) {
	count := 0
	type Operator struct {
		Name              string `json:"name"`
		PackageName       string `gorm:"unique" json:"packageName"`
		DisplayName       string `json:"displayName"`
		Provider          string `json:"provider"`
		ThumbUrl          string `json:"thumbUrl"`
		Version           string `json:"version"`
		Repository        string `json:"repository"`
		Channel           string `json:"channel"`
		ContainerImage    string `json:"containerImage"`
		GlobalOperator    bool   `json:"globalOperator"`
		Active            bool   `json:"active"`
		OperaterCreatedAt string `json:"createdAt"`
	}
	datas := &[]Operator{}
	op := db.Model(&Operator{})
	if isActive {
		op = op.Where("active=?", isActive)
	}
	err = op.Order("package_name").Limit(size).Offset(size * (page - 1)).Find(datas).Error
	op.Count(&count)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"operators": datas,
		"count":     count,
	}, nil
}

func (r *OperatorRepo) FindOperator(db *gorm.DB, packageName string) (*Operator, error) {
	data := &Operator{}
	err = db.Model(&Operator{}).Where("package_name = ?", packageName).Take(data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *OperatorRepo) SyncOperator(db *gorm.DB, data *Operator) error {
	_, err := r.FindOperator(db, data.PackageName)
	if err == nil {
		err = r.UpdateOperator(db, data)
		if err != nil {
			return err
		}
	} else {
		_, err = r.Save(db, data)
		if err != nil {
			return err
		}
	}
	return nil
}
