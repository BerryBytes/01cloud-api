package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Application struct {
	gorm.Model
	Name                string   `gorm:"size:255;not null;" json:"name"`
	Project             *Project `gorm:"foreignkey:ProjectID" json:"project"`
	ProjectID           uint64   `gorm:"not null" json:"project_id"`
	Plugin              *Plugin  `gorm:"foreignkey:PluginID" json:"plugin"`
	PluginID            uint64   `gorm:"null;default:0;" json:"plugin_id"`
	Cluster             *Cluster `gorm:"foreignkey:ClusterID" json:"cluster"`
	ClusterID           uint64   `gorm:"null;DEFAULT:0" json:"cluster_id,omitempty"`
	Chart               *Chart   `gorm:"foreignkey:ChartID" json:"chart"`
	ChartID             string   `gorm:"null;DEFAULT:null" json:"chart_id,omitempty"`
	Owner               *User    `gorm:"foreignkey:OwnerId" json:"owner"`
	OwnerId             uint64   `gorm:"null;DEFAULT:0" json:"owner_id,omitempty"`
	GitRepository       *GitRepo `sql:"-" json:"git_repository_info"`
	GitUrl              string   `gorm:"size:255;null;" json:"git_repository"`
	GitRepoUrl          *string  `gorm:"size:255;null;" json:"git_repo_url"`
	GitToken            string   `gorm:"size:255;null;" json:"git_token"`
	GitService          string   `gorm:"size:255;null;" json:"git_service"`
	ImageUrl            string   `gorm:"size:255;null;" json:"image_url"`
	ImageNamespace      string   `gorm:"size:255;null;" json:"image_namespace"`
	ImageRepo           string   `gorm:"size:255;null;" json:"image_repo"`
	ImageService        string   `gorm:"size:255;null;" json:"image_service"`
	ManagedService      string   `gorm:"size:255;null;" json:"managed_service"`
	ServiceType         int      `gorm:"default:0;" json:"service_type"` // 0-template/1-git/2-image/3-helm/4-operator/5-managedDB
	Variables           string   `gorm:"null;" json:"variables"`
	Active              bool     `gorm:"default:true" json:"active"`
	Attributes          string   `gorm:"size:1024;null;" json:"attributes"`
	OperatorPackageName string   `gorm:"size:255;null" json:"operator_package_name"`
}

type ApplicationInterface interface {
	Save(db *gorm.DB, data *Application) (*Application, error)
	SearchApplication(db *gorm.DB, uid uint, query string) ([]*Application, error)
	FindAll(db *gorm.DB) (*[]Application, error)
	FindAllApplicationByProject(db *gorm.DB, pid uint64, uid uint64, preload bool) ([]*Application, error)
	FindAllApplicationByProjectForAdmin(db *gorm.DB, pid uint64, uid uint64) (*[]Application, error)
	IsNameExists(db *gorm.DB, pid uint64, name string) bool
	FindAllApplicationCountByProject(db *gorm.DB, pid uint64, uid uint64) int
	Find(db *gorm.DB, pid uint64) (*Application, error)
	Update(db *gorm.DB, data *Application) (*Application, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
	ChangeIsActive(db *gorm.DB, isActive bool, id uint) error
	RenameApp(db *gorm.DB, id uint, name string) error
	FindAllApplicationByInactiveProject(db *gorm.DB, pid uint64) (*[]Application, error)
	CountAllApplicationByProject(db *gorm.DB, pid uint64) int
	GetApplicationByName(db *gorm.DB, name string, proj_id uint64) (Application, error)
}

type ApplicationType struct{}

func NewApplication() ApplicationInterface {
	return &ApplicationType{}
}

func (data *ApplicationType) GetApplicationByName(db *gorm.DB, name string, proj_id uint64) (app Application, err error) {
	err = db.Model(&Application{}).Where("lower(name) = ? and project_id = ?", strings.ToLower(name), proj_id).First(&app).Error
	fmt.Printf("err: %v\n", err)
	return
}

func (data *Application) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.GitUrl = html.EscapeString(strings.TrimSpace(data.GitUrl))
	data.Plugin = &Plugin{}
	data.Project = &Project{}
	data.GitToken = html.EscapeString(strings.TrimSpace(data.GitToken))
	data.GitService = html.EscapeString(strings.TrimSpace(data.GitService))
	data.ManagedService = html.EscapeString(strings.TrimSpace(data.ManagedService))
	serviceType := 0
	if len(data.GitService) > 0 {
		serviceType = 1
	} else if len(data.ImageService) > 0 {
		serviceType = 2
	} else if len(data.ChartID) > 0 {
		serviceType = 3
	} else if len(data.OperatorPackageName) > 0 {
		serviceType = 4
	} else if len(data.ManagedService) > 0 {
		serviceType = 5
	}
	data.ServiceType = serviceType
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	data.Active = true
}

func (data *Application) ToJson() (map[string]interface{}, error) {
	msg, _ := json.Marshal(data)
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}

func (data *Application) Validate() error {
	if data.Name == "" {
		return errors.New("required Name")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	if data.ProjectID == 0 {
		return errors.New("required Project Id")
	}
	if data.PluginID == 0 && data.ServiceType < 2 {
		return errors.New("required Plugin")
	}
	if data.ClusterID == 0 {
		return errors.New("region Plugin")
	}
	return nil
}

func (data *Application) GetAvailableResource(db *gorm.DB, pid uint64) map[string]interface{} {
	usage, _ := Environment{}.GetUsedResource(db, pid, false)
	return map[string]interface{}{
		"cpu":    uint64(data.Project.Subscription.Cores) - usage.Core,
		"memory": uint64(data.Project.Subscription.Memory) - usage.Memory,
		"disk":   uint64(data.Project.Subscription.DiskSpace) - usage.Disk,
	}

}

func (r *ApplicationType) Save(db *gorm.DB, data *Application) (*Application, error) {
	err = db.Model(&Application{}).Create(&data).Error
	if err != nil {
		return &Application{}, err
	}
	return data, nil
}

func (r *ApplicationType) SearchApplication(db *gorm.DB, uid uint, query string) ([]*Application, error) {
	dataList := []*Application{}
	err := db.Model(&Application{}).
		Preload("Cluster").
		Preload("Project").
		Preload("Chart").
		Preload("Plugin").
		Preload("Project.Organization").
		Where(
			" lower(applications.name) LIKE lower(?) and"+
				" (applications.id IN ("+
				"     select distinct applications.id from applications"+
				"     inner join projects p on applications.project_id = p.id"+
				"     where p.user_id = ?"+
				"     and p.deleted_at is null and applications.deleted_at is null) or"+
				" applications.id IN ("+
				"     select distinct applications.id from applications"+
				"     inner join projects on applications.project_id = projects.id and projects.deleted_at is null"+
				"     inner join authorizations on authorizations.project_id = projects.id"+
				"     and authorizations.application_id = 0 and authorizations.deleted_at is null"+
				"     where (authorizations.user_id = ? or authorizations.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and applications.deleted_at is null) or"+
				" applications.id IN (select application_id from authorizations"+
				"     where (user_id = ? or authorizations.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and application_id <> 0 and deleted_at is null))"+
				" and applications.deleted_at is null",
			"%"+query+"%", uid, uid, uid, uid, uid).
		Order("applications.name,applications.id desc").
		Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (r *ApplicationType) FindAll(db *gorm.DB) (*[]Application, error) {
	datas := []Application{}
	err = db.Model(&Application{}).Order("name, id desc").Find(&datas).Error
	if err != nil {
		return &[]Application{}, err
	}
	return &datas, nil
}

func (r *ApplicationType) FindAllApplicationByProject(db *gorm.DB, pid uint64, uid uint64, preload bool) ([]*Application, error) {
	dataList := []*Application{}
	query := db.Model(&Application{})
	if preload {
		query = query.Preload("Cluster").
			Preload("Project")
	}
	err := query.
		Preload("Plugin").
		Preload("Owner").
		Preload("Chart").
		Where("applications.id IN ("+
			"     select distinct applications.id from applications"+
			"     inner join projects p on applications.project_id = p.id"+
			"     where p.id = ? and (p.organization_id in(select id from organizations where user_id=?) or p.organization_id in(select organization_id from organization_members where user_id=? and user_role=1))"+
			"     and p.deleted_at is null and applications.deleted_at is null) or"+
			" (applications.id IN ("+
			"     select distinct applications.id from applications"+
			"     inner join projects p on applications.project_id = p.id"+
			"     where p.id = ? and p.user_id = ?"+
			"     and p.deleted_at is null and applications.deleted_at is null) or"+
			" applications.id IN ("+
			"     select distinct applications.id from applications"+
			"     inner join projects on applications.project_id = projects.id and projects.id = ? and projects.deleted_at is null"+
			"     inner join authorizations on authorizations.project_id = projects.id"+
			"     and authorizations.application_id = 0 and authorizations.deleted_at is null"+
			"     where (authorizations.user_id = ? or authorizations.group_id"+
			"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
			"     where gm.user_id = ? and groups.deleted_at is null))"+
			"     and applications.deleted_at is null) or"+
			" applications.id IN (select application_id from authorizations"+
			"     where (user_id = ? or authorizations.group_id"+
			"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
			"     where gm.user_id = ? and groups.deleted_at is null))"+
			"     and project_id = ? and application_id <> 0 and deleted_at is null))"+
			" and applications.deleted_at is null ", pid, uid, uid, pid, uid, pid, uid, uid, uid, uid, pid).
		Order("applications.name, applications.id desc").
		Find(&dataList).Error

	if err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (r *ApplicationType) FindAllApplicationByProjectForAdmin(db *gorm.DB, pid uint64, uid uint64) (*[]Application, error) {
	datas := []Application{}
	err = db.Model(&Application{}).Preload("Project").Preload("Plugin").Where(&Application{ProjectID: pid}).Order("name, id desc").Find(&datas).Error
	if err != nil {
		return &[]Application{}, err
	}
	return &datas, nil
}
func (r *ApplicationType) FindAllApplicationByInactiveProject(db *gorm.DB, pid uint64) (*[]Application, error) {
	datas := []Application{}
	err = db.Model(&Application{}).Where(&Application{ProjectID: pid}).Find(&datas).Error
	if err != nil {
		return &[]Application{}, err
	}
	return &datas, nil
}

func (r *ApplicationType) IsNameExists(db *gorm.DB, pid uint64, name string) bool {
	count := 0
	db.Model(&Application{}).Where(&Application{ProjectID: pid, Name: name}).Count(&count)
	return count > 0
}

func (r *ApplicationType) FindAllApplicationCountByProject(db *gorm.DB, pid uint64, uid uint64) int {
	count := 0
	db.Model(&Application{}).Where(&Application{ProjectID: pid}).Count(&count)
	return count
}

func (r *ApplicationType) CountAllApplicationByProject(db *gorm.DB, pid uint64) int {
	count := 0
	db.Model(&Application{}).Where(&Application{ProjectID: pid}).Count(&count)
	return count
}

func (r *ApplicationType) Find(db *gorm.DB, pid uint64) (*Application, error) {
	data := &Application{}
	err = db.Model(&Application{}).
		Preload("Project").
		Preload("Project.Subscription").
		Preload("Project.User").
		Preload("Plugin").
		Preload("Cluster").
		Preload("Chart").
		Preload("Owner").
		Preload("Chart.Repo").
		Preload("Cluster.Organization").
		Preload("Cluster.DNS").
		Preload("Cluster.ImageRegistry").
		Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &Application{}, err
	}
	return data, nil
}

func (r *ApplicationType) Update(db *gorm.DB, data *Application) (*Application, error) {
	var app = Application{Active: data.Active}
	if data.Name != "" {
		app.Name = data.Name
	}
	if data.PluginID != 0 {
		app.PluginID = data.PluginID
	}
	if data.ProjectID != 0 {
		app.ProjectID = data.ProjectID
	}
	if data.ClusterID != 0 {
		app.ClusterID = data.ClusterID
	}
	if data.ChartID != "" {
		app.ChartID = data.ChartID
	}
	if data.GitService != "" {
		app.GitService = data.GitService
	}
	if data.GitUrl != "" {
		app.GitUrl = data.GitUrl
	}
	if data.GitToken != "" {
		app.GitToken = data.GitToken
	}
	if data.Variables != "" {
		app.Variables = data.Variables
	}
	if data.Attributes != "" {
		app.Attributes = data.Attributes
	}
	err = db.Model(&Application{}).Where("id = ?", data.ID).Updates(app).Error
	if err != nil {
		return &Application{}, err
	}
	return data, nil
}

func (r *ApplicationType) Delete(db *gorm.DB, id uint64) (int64, error) {
	err := db.Model(&Authorization{}).Delete(&Authorization{}, "application_id = ?", id).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return 0, err
	}

	db = db.Model(&Application{}).Where("id = ?", id).Take(&Application{}).Delete(&Application{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Application not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (r *ApplicationType) ChangeIsActive(db *gorm.DB, isActive bool, id uint) error {
	db = db.Model(&Application{}).Where("id = ?", id).Take(&Application{}).UpdateColumns(
		map[string]interface{}{
			"active": isActive,
		},
	)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func (r *ApplicationType) RenameApp(db *gorm.DB, id uint, name string) error {
	err = db.Model(&Application{}).Where("id = ?", id).Take(&Application{}).UpdateColumns(
		map[string]interface{}{
			"name": name,
		},
	).Error
	if err != nil {
		return err
	}
	return nil
}
