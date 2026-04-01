package models

import (
	"encoding/json"
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type HelmEnvironment struct {
	gorm.Model
	Name             string         `gorm:"size:255;not null;" json:"name"`
	Application      *Application   `gorm:"foreignkey:ApplicationID" json:"application" `
	ApplicationID    uint64         `gorm:"default:0" json:"application_id"`
	ChartVersion     *ChartVersion  `gorm:"foreignkey:ChartVersionID" json:"chartVersion"`
	ChartVersionID   string         `gorm:"not null;" json:"chart_version_id"`
	Values           string         `gorm:"not null;type:text" json:"values"`
	Version          postgres.Jsonb `sql:"json" json:"version"`
	Active           bool           `gorm:"not null;" json:"active"`
	ApplyImmediately bool           `gorm:"default:false;" json:"apply_immediately"`
	Attributes       postgres.Jsonb `sql:"json" json:"attributes"`
	CiRequest        *CIRequest     `sql:"-" json:"ci_request"`
	Action           string         `sql:"-" json:"action"`
	Scripts          postgres.Jsonb `json:"scripts"`
	Schedules        postgres.Jsonb `gorm:"null" json:"schedules"`
}

func (data *HelmEnvironment) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Application = &Application{}
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *HelmEnvironment) ToJson() (map[string]interface{}, error) {
	msg, _ := json.Marshal(data)
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}

func (data *HelmEnvironment) Validate(isEnvironment bool) error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if !ValidateName(data.Name) && isEnvironment {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	if data.ApplicationID == 0 && isEnvironment {
		return errors.New("required application")
	}

	return nil
}

func (data *HelmEnvironment) Save(db *gorm.DB) (*HelmEnvironment, error) {
	err := db.Model(&HelmEnvironment{}).Create(&data).Error
	if err != nil {
		return &HelmEnvironment{}, err
	}
	return data, nil
}

func (data *HelmEnvironment) SearchEnvironment(db *gorm.DB, uid uint, query string) ([]*HelmEnvironment, error) {
	dataList := []*HelmEnvironment{}
	err := db.Model(&HelmEnvironment{}).
		Preload("Application").
		Preload("Application.Project").
		Preload("Application.Project.Organization").
		Preload("ChartVersion").
		Where(
			" lower(helmenvironmets.name) LIKE lower(?) and"+
				" (helm_environments.id IN ("+
				"     select distinct env.id from helm_environments env"+
				"     inner join applications app on env.application_id = app.id and app.deleted_at is null"+
				"     inner join projects prj on app.project_id = prj.id and prj.user_id =? and prj.deleted_at is null"+
				"     where env.deleted_at is null) or"+
				" helm_environments.id IN ("+
				"     select distinct env.id from helm_environments env"+
				"     inner join applications app on env.application_id = app.id and app.deleted_at is null"+
				"     inner join projects prj on prj.id = app.project_id and prj.deleted_at is null"+
				"     inner join authorizations auth on auth.project_id = prj.id and auth.application_id = 0 and prj.deleted_at is null"+
				"     where (auth.user_id = ? or auth.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and env.deleted_at is null) or"+
				" helm_environments.id IN ("+
				"     select distinct env.id from helm_environments env"+
				"     inner join applications app on env.application_id = app.id and app.deleted_at is null"+
				"     inner join authorizations auth on app.id = auth.application_id and auth.application_id <> 0 and auth.environment_id = 0 and auth.deleted_at is null"+
				"     where (auth.user_id = ? or auth.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and env.deleted_at is null) or"+
				" helm_environments.id IN (select environment_id from authorizations"+
				"     where (user_id = ? or authorizations.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and environment_id <> 0 and deleted_at is null))"+
				" and helm_environments.deleted_at is null and helm_environments.active = true",
			"%"+query+"%", uid, uid, uid, uid, uid, uid, uid).
		Order("helm_environments.name, helm_environments.id desc").
		Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (data *HelmEnvironment) FindAllByApplication(db *gorm.DB, aid uint64, preload bool, uid uint64) ([]*HelmEnvironment, error) {

	dataList := []*HelmEnvironment{}
	query := db.Model(&HelmEnvironment{})
	if preload {
		query = query.Preload("Application").
			Preload("ChartVersion")
	}
	err :=
		query.
			Where("helm_environments.id IN ("+
				"     select distinct env.id from helm_environments env"+
				"     inner join applications app on env.application_id = app.id and app.id = ? and app.deleted_at is null"+
				"     inner join projects prj on app.project_id = prj.id and (prj.organization_id in (select id from organizations where user_id=?) or prj.organization_id in (select organization_id from organization_members where user_id= ? and user_role=1)) and prj.deleted_at is null"+
				"     where env.deleted_at is null) or"+
				" (helm_environments.id IN ("+
				"     select distinct env.id from helm_environments env"+
				"     inner join applications app on env.application_id = app.id and app.id = ? and app.deleted_at is null"+
				"     inner join projects prj on app.project_id = prj.id and prj.user_id =? and prj.deleted_at is null"+
				"     where env.deleted_at is null) or"+
				" helm_environments.id IN ("+
				"     select distinct env.id from helm_environments env"+
				"     inner join applications app on env.application_id = app.id and app.id = ? and app.deleted_at is null"+
				"     inner join projects prj on prj.id = app.project_id and prj.deleted_at is null"+
				"     inner join authorizations auth on auth.project_id = prj.id and auth.application_id = 0 and auth.deleted_at is null"+
				"     where (auth.user_id = ? or auth.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and env.deleted_at is null) or"+
				" helm_environments.id IN ("+
				"     select distinct env.id from helm_environments env"+
				"     inner join applications app on env.application_id = app.id and app.id = ? and app.deleted_at is null"+
				"     inner join authorizations auth on app.id = auth.application_id and auth.application_id <> 0 and auth.environment_id = 0 and auth.deleted_at is null"+
				"     where (auth.user_id = ? or auth.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and env.deleted_at is null) or"+
				" helm_environments.id IN (select environment_id from authorizations"+
				"     where (user_id = ? or authorizations.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and application_id = ? and environment_id <> 0 and deleted_at is null))"+
				" and helm_environments.deleted_at is null",
				aid, uid, uid, aid, uid, aid, uid, uid, aid, uid, uid, uid, uid, aid).
			Order("helm_environments.name, helm_environments.id desc").
			Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (data *HelmEnvironment) FindAllByApplicationAdmin(db *gorm.DB, aid uint64) (*[]HelmEnvironment, error) {
	var err error
	datas := []HelmEnvironment{}
	err = db.Model(&HelmEnvironment{}).Preload("Application").Preload("Application.Project").Preload("ChartVersion").Where(&HelmEnvironment{ApplicationID: aid}).Order("name, id desc").Find(&datas).Error
	if err != nil {
		return &[]HelmEnvironment{}, err
	}
	return &datas, nil
}
func (data *HelmEnvironment) FindAllEnvironmentCountByApplication(db *gorm.DB, aid uint64) int {
	count := 0
	db.Model(&HelmEnvironment{}).Where(&HelmEnvironment{ApplicationID: aid}).Count(&count)
	return count
}

func (data *HelmEnvironment) IsNameExists(db *gorm.DB, aid uint, name string) bool {
	count := 0
	db.Model(&HelmEnvironment{}).Where(&HelmEnvironment{ApplicationID: uint64(aid), Name: name}).Count(&count)
	return count > 0
}

func (data *HelmEnvironment) Find(db *gorm.DB, pid int64) (*HelmEnvironment, error) {
	err := db.Model(&HelmEnvironment{}).
		Preload("Application").
		Preload("Application.Project").
		Preload("Application.Project.Organization").
		Preload("Application.Project.Subscription").
		Preload("Application.Project.User").
		Preload("Application.Cluster").
		Preload("Application.Cluster.Organization").
		Preload("Application.Cluster.DNS").
		Preload("Application.Cluster.ImageRegistry").
		Preload("Application.Plugin").
		Preload("Application.Chart").
		Preload("ChartVersion").
		//Preload("AddOns").
		Where("ID = ?", pid).Take(&data).Error

	if err != nil {
		return &HelmEnvironment{}, err
	}
	return data, nil
}
func (data *HelmEnvironment) UpdateStatus(db *gorm.DB, id uint64, status bool) (*HelmEnvironment, error) {
	err := db.Model(&HelmEnvironment{}).
		Where("id = ?", id).UpdateColumns(
		map[string]interface{}{
			"active": status,
		}).Error
	if err != nil {
		return &HelmEnvironment{}, err
	}
	return data, nil
}

func (data *HelmEnvironment) Update(db *gorm.DB) (*HelmEnvironment, error) {
	var err error
	var app = HelmEnvironment{Active: data.Active}
	if data.Name != "" {
		app.Name = data.Name
	}
	if data.ApplicationID != 0 {
		app.ApplicationID = data.ApplicationID
	}
	if data.ChartVersionID != "" {
		app.ChartVersionID = data.ChartVersionID
	}
	if data.Attributes.RawMessage != nil {
		app.Attributes = data.Attributes
	}
	if data.Scripts.RawMessage != nil {
		app.Scripts = data.Scripts
	}
	if data.Version.RawMessage != nil {
		app.Version = data.Version
	}
	if data.CiRequest != nil {
		app.CiRequest = data.CiRequest
	}
	err = db.Model(&HelmEnvironment{}).Where("ID = ?", data.ID).Updates(app).Error
	if err != nil {
		return &HelmEnvironment{}, err
	}
	return data, nil
}
func (data *HelmEnvironment) ChangeIsActive(db *gorm.DB, isActive bool) error {
	db = db.Model(&HelmEnvironment{}).Where("id = ?", data.ID).Take(&HelmEnvironment{}).UpdateColumns(
		map[string]interface{}{
			"active": isActive,
		},
	)
	if db.Error != nil {
		return db.Error
	}
	return nil
}
func (data *HelmEnvironment) Delete(db *gorm.DB, id uint64) (int64, error) {
	err := db.Model(&Authorization{}).Delete(&Authorization{}, "environment_id = ?", id).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return 0, err
	}

	db = db.Model(&HelmEnvironment{}).Where("id = ?", id).Take(&HelmEnvironment{}).Delete(&HelmEnvironment{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Environment not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (data *HelmEnvironment) UpdateHelmScheudle(db *gorm.DB, id uint) (*HelmEnvironment, error) {
	if data.Schedules.RawMessage != nil {
		err = db.Model(&HelmEnvironment{}).Where("id = ?", id).UpdateColumns(map[string]interface{}{
			"schedules": data.Schedules,
		}).Error
		if err != nil {
			return &HelmEnvironment{}, err
		}
	}
	return data, nil
}
