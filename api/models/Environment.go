package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/sirupsen/logrus"
)

type EnvironmentInterface interface {
	UpdateFileManagerStatus(db *gorm.DB, data *Environment) (*Environment, error)
	GetEnvironmentByName(db *gorm.DB, name string, app_id uint64) (Environment, error)
}

type EnvironmentType struct {
}

func NewEnvironment() EnvironmentInterface {
	return &EnvironmentType{}
}

func (data *EnvironmentType) GetEnvironmentByName(db *gorm.DB, name string, app_id uint64) (env Environment, err error) {
	err = db.Model(&Environment{}).Where("lower(name) = ? and application_id = ?", strings.ToLower(name), app_id).First(&env).Error
	return
}

type Environment struct {
	gorm.Model
	Name               string            `gorm:"size:255;not null;" json:"name"`
	Application        *Application      `gorm:"foreignkey:ApplicationID" json:"application" `
	ApplicationID      uint64            `gorm:"default:0" json:"application_id"`
	Resource           *Resource         `gorm:"foreignkey:ResourceID" json:"resource"`
	ResourceID         uint64            `gorm:"not null;" json:"resource_id"`
	PluginVersion      *PluginVersion    `gorm:"foreignkey:PluginVersionID" json:"plugin_version"`
	PluginVersionID    uint64            `gorm:"not null;" json:"plugin_version_id"`
	Replicas           uint16            `gorm:"not null;" json:"replicas"`
	GitUrl             string            `gorm:"size:255;null;" json:"git_url"`
	GitRepository      *GitRepo          `sql:"-" json:"git_repository_info"`
	GitBranch          string            `gorm:"size:255;null;" json:"git_branch"`
	ImageTag           string            `gorm:"size:255;null;" json:"image_tag"`
	ImageUrl           string            `gorm:"size:255;null;" json:"image_url"`
	ServiceType        int               `gorm:"default:0;" json:"service_type"` // template/git/image
	Variables          postgres.Jsonb    `sql:"json" json:"variables"`
	Version            postgres.Jsonb    `sql:"json" json:"version"`
	OtherVersion       postgres.Jsonb    `json:"other_version"`
	UserVariables      postgres.Jsonb    `sql:"json" json:"user_variables"`
	Active             bool              `gorm:"not null;" json:"active"`
	ApplyImmediately   bool              `gorm:"default:false;" json:"apply_immediately"`
	Attributes         postgres.Jsonb    `sql:"json" json:"attributes"`
	RepositoryImage    *RepositoryImage  `sql:"-" json:"repository_image"`
	CiRequest          *CIRequest        `sql:"-" json:"ci_request"`
	AutoScaler         postgres.Jsonb    `json:"auto_scaler"`
	Storage            []*Storage        `gorm:"foreignkey:EnvironmentID;association_foreignkey:ID"`
	CronJob            []*CronJob        `gorm:"foreignkey:EnvironmentID;association_foreignkey:ID"`
	InitContainers     []*InitContainer  `gorm:"foreignkey:EnvironmentID;association_foreignkey:ID"`
	LoadBalancer       *LoadBalancer     `gorm:"foreignkey:LoadBalancerID" json:"load_balancer" `
	DeploymentStrategy postgres.Jsonb    `sql:"json" json:"deployment_strategy"`
	LoadBalancerID     uint64            `gorm:"default:0" json:"load_balancer_id"`
	Parent             *Environment      `gorm:"foreignkey:ParentID" json:"parent" `
	ParentID           uint64            `gorm:"default:0" json:"parent_id"`
	Action             string            `sql:"-" json:"action"`
	ExternalSecret     postgres.Jsonb    `json:"external_secret"`
	Scripts            postgres.Jsonb    `json:"scripts"`
	OperatorPayload    postgres.Jsonb    `gorm:"null" json:"operator_payload"`
	Schedules          postgres.Jsonb    `gorm:"null" json:"schedules"`
	FileManagerEnabled *time.Time        `gorm:"default:null;" json:"file_manager_enabled"`
	CloneEnvironment   *CloneEnvironment `sql:"-" json:"clone_environment"`
	ExternalLogging    postgres.Jsonb    `json:"external_logging"`
	ErrorMessage       postgres.Jsonb    `json:"error_message"`
	Setting            postgres.Jsonb    `json:"setting"`
	ExternalURL        bool              `json:"external_url"`
	WhitelistedIPs     string            `json:"whitelisted_ips"`
	ManagedService     string            `json:"managed_service"`
}

type WhitelistedIP struct {
	WhitelistedIPs string `json:"whitelisted_ips"`
}

type ExternalURLRequest struct {
	ExternalURL bool `json:"external_url"`
}

type AddonExternalURLRequest struct {
	ExternalURL bool `json:"external_url"`
}

type Script struct {
	Build       string                 `json:"build"`
	Run         string                 `json:"run"`
	SubDir      string                 `json:"sub_dir"`
	Dockerfile  string                 `json:"dockerfile"`
	CiSteps     interface{}            `json:"ci_steps"`
	CIVariables map[string]interface{} `json:"ci_variables"`
}

type RenameRequest struct {
	Name string `json:"name"`
}

type CloneEnvironment struct {
	Name                  string `json:"name"`
	EnvId                 int64  `json:"id"`
	OldEnvId              int64  `json:"old_env_id"`
	IsResource            bool   `json:"is_resource"`
	IsUserPermission      bool   `json:"is_user_permission"`
	IsCronJob             bool   `json:"is_cron_job"`
	IsCIConfig            bool   `json:"is_ci_config"`
	IsCDConfig            bool   `json:"is_cd_config"`
	IsBackupSetting       bool   `json:"is_backup_setting"`
	IsAddon               bool   `json:"is_addon"`
	IsPVC                 bool   `json:"is_pvc"`
	IsBuildAndRunScript   bool   `json:"is_build_and_run_scrpt"`
	IsHPASetting          bool   `json:"is_hpa_setting"`
	IsEnvironmentVariable bool   `json:"is_environment_variable"`
	IsScheduler           bool   `json:"is_scheduler"`
}

type RequestSchedule struct {
	Enabled   bool      `json:"enabled"`
	TimeZone  string    `json:"time_zone"`
	Start     string    `json:"start"`
	Stop      string    `json:"stop"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type ErrorMessage struct {
	ClusterId     int       `json:"cluster_id"`
	EnvironmentId int       `json:"environment_id"`
	Code          int       `json:"code"`
	Message       string    `json:"message"`
	Source        string    `json:"source"`
	Time          time.Time `json:"time"`
}

func (data *Environment) ToJson() (map[string]interface{}, error) {
	msg, _ := json.Marshal(data)
	res := map[string]interface{}{}
	err := json.Unmarshal(msg, &res)
	return res, err
}

func (data *Environment) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Application = &Application{}
	data.Resource = &Resource{}
	data.GitBranch = strings.TrimSpace(data.GitBranch)
	data.ManagedService = strings.TrimSpace(data.ManagedService)
	data.PluginVersion = &PluginVersion{}
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
	serviceType := 0
	if len(data.GitBranch) > 0 {
		serviceType = 1
	} else if len(data.ImageTag) > 0 {
		serviceType = 2
	} else if len(data.ManagedService) > 0 {
		serviceType = 5
	} else if data.OperatorPayload.RawMessage != nil {
		serviceType = 4
	}
	data.ServiceType = serviceType
	data.CronJob = nil
	data.Schedules.RawMessage = nil
	data.FileManagerEnabled = nil
}

func (data *WhitelistedIP) PrepareWhiteListedIP() {
	data.WhitelistedIPs = html.EscapeString(strings.TrimSpace(data.WhitelistedIPs))
}

func (data *WhitelistedIP) ValidateWhiteListedIP() error {
	if data != nil {
		if data.WhitelistedIPs != "" {
			var errorIps []string
			ipList := strings.Split(data.WhitelistedIPs, ",")
			for _, ip := range ipList {
				ip = strings.TrimSpace(ip)
				// Try parsing as CIDR
				_, _, err := net.ParseCIDR(ip)
				if err == nil {
					continue // Valid CIDR notation
				}
				// Try parsing as a normal IP
				parsedIP := net.ParseIP(ip)
				if parsedIP == nil {
					logrus.Errorf("error validating IP address: %s", ip)
					errorIps = append(errorIps, ip)
				}
			}
			if len(errorIps) > 0 {
				return fmt.Errorf("invalid IP addresses: %s", strings.Join(errorIps, ", "))
			}
		}
	}
	return nil
}

func (e *Environment) UpdatePrepareScript(req postgres.Jsonb) error {
	data := Script{}
	script := Script{}
	if req.RawMessage != nil {
		err := json.Unmarshal(req.RawMessage, &data)
		if err != nil {
			return err
		}
	}
	err := json.Unmarshal(e.Scripts.RawMessage, &script)
	if err != nil {
		return err
	}
	if script.Build != "" {
		data.Build = script.Build
	}
	if script.Run != "" {
		data.Run = script.Run
	}
	if script.Dockerfile != "" {
		data.Dockerfile = script.Dockerfile
	}
	if script.SubDir != "" {
		data.SubDir = script.SubDir
	}
	if len(script.CIVariables) > 0 {
		data.CIVariables = script.CIVariables
	}
	if script.CiSteps != nil {
		data.CiSteps = script.CiSteps
	}
	scriptBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	e.Scripts.RawMessage = scriptBytes
	return nil
}

func (data *Environment) PrepareAddon(parent *Environment) {
	data.ID = 0
	data.Name = "Plugin-" + string(rune(rand.Int()))
	data.ApplicationID = 0
	if data.ResourceID == 0 {
		data.ResourceID = parent.ResourceID
	}
	data.Replicas = 1
	if data.Variables.RawMessage != nil {
		var currentVariableMap map[string]interface{}
		err := json.Unmarshal(data.Variables.RawMessage, &currentVariableMap)
		if err == nil {
			if _, ok := currentVariableMap["slave"]; ok {
				if d, ok := currentVariableMap["slave"].(map[string]interface{})["replicas"]; ok {
					data.Replicas = uint16(d.(float64))
				}
			}
		}
	}
	data.Resource = &Resource{}
	data.PluginVersion = &PluginVersion{}
	data.Active = true
	data.ApplyImmediately = false
	data.ParentID = uint64(parent.ID)
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *Environment) Validate(isEnvironment bool) error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if !ValidateName(data.Name) && isEnvironment {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	if data.ResourceID == 0 {
		return errors.New("required resource")
	}
	if data.ApplicationID == 0 && isEnvironment {
		return errors.New("required application")
	}
	if data.PluginVersionID == 0 && data.ServiceType < 2 {
		return errors.New("required plugin version")
	}
	if data.Replicas == 0 {
		return errors.New("required replicas")
	}
	if data.Version.RawMessage == nil && data.ServiceType < 2 {
		return errors.New("required version detail")
	}
	return nil
}
func (data *Environment) ValidateAttribute() error {
	if data.Attributes.RawMessage == nil {
		return errors.New("attribute not in format, must be an array")
	}
	return nil
}

func (data *RequestSchedule) Validate() error {
	if !data.Enabled {
		return nil
	}
	if data.Start == "" {
		return errors.New("required start cron expression")
	}
	if data.Stop == "" {
		return errors.New("required stop cron expression")
	}
	if data.TimeZone == "" {
		return errors.New("required time zone")
	}
	return nil
}

func (data *Environment) Save(db *gorm.DB) (*Environment, error) {
	err := db.Model(&Environment{}).Create(&data).Error
	if err != nil {
		return &Environment{}, err
	}
	return data, nil
}

func (data *Environment) CheckDublicateDNS(db *gorm.DB, clients_fqdn string) error {
	count := 0
	err = db.Model(&Environment{}).Where("lower(variables::text)::json->> lower('clients_fqdn')::text = ?", strings.ToLower(clients_fqdn)).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("domain already in use")
	}
	return nil
}
func (data *Environment) InstallAddOn(db *gorm.DB, addOn *Plugin) (*Environment, error) {
	err := db.Model(&data).Association("AddOns").Append([]Plugin{*addOn}).Error
	if err != nil {
		return data, err
	}
	return data, nil
}

func (data *Environment) UninstallAddOn(db *gorm.DB, addOn *Plugin) (*Environment, error) {
	err := db.Model(&data).Association("AddOns").Delete(&addOn).Error
	if err != nil {
		return &Environment{}, err
	}
	return nil, nil
}

func (data *Environment) SearchEnvironment(db *gorm.DB, uid uint, query string) ([]*Environment, error) {
	dataList := []*Environment{}
	err := db.Model(&Environment{}).
		Preload("Application").
		Preload("Application.Project").
		Preload("Application.Project.Organization").
		Preload("PluginVersion").
		Preload("PluginVersion.Plugin").
		Preload("Resource").
		Where(
			" lower(environments.name) LIKE lower(?) and"+
				" (environments.id IN ("+
				"     select distinct env.id from environments env"+
				"     inner join applications app on env.application_id = app.id and app.deleted_at is null"+
				"     inner join projects prj on app.project_id = prj.id and prj.user_id =? and prj.deleted_at is null"+
				"     where env.deleted_at is null) or"+
				" environments.id IN ("+
				"     select distinct env.id from environments env"+
				"     inner join applications app on env.application_id = app.id and app.deleted_at is null"+
				"     inner join projects prj on prj.id = app.project_id and prj.deleted_at is null"+
				"     inner join authorizations auth on auth.project_id = prj.id and auth.application_id = 0 and prj.deleted_at is null"+
				"     where (auth.user_id = ? or auth.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and env.deleted_at is null) or"+
				" environments.id IN ("+
				"     select distinct env.id from environments env"+
				"     inner join applications app on env.application_id = app.id and app.deleted_at is null"+
				"     inner join authorizations auth on app.id = auth.application_id and auth.application_id <> 0 and auth.environment_id = 0 and auth.deleted_at is null"+
				"     where (auth.user_id = ? or auth.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and env.deleted_at is null) or"+
				" environments.id IN (select environment_id from authorizations"+
				"     where (user_id = ? or authorizations.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and environment_id <> 0 and deleted_at is null))"+
				" and environments.deleted_at is null and environments.active = true",
			"%"+query+"%", uid, uid, uid, uid, uid, uid, uid).
		Order("environments.name, environments.id desc").
		Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (data *Environment) FindAllByApplication(db *gorm.DB, aid uint64, preload bool, uid uint64) ([]*Environment, error) {

	dataList := []*Environment{}
	query := db.Model(&Environment{}).Preload("CronJob")
	if preload {
		query = query.Preload("Application").
			Preload("PluginVersion").
			Preload("Application.Cluster").
			Preload("PluginVersion.Plugin").
			Preload("Resource")
	}
	err :=
		query.
			Where("environments.id IN ("+
				"     select distinct env.id from environments env"+
				"     inner join applications app on env.application_id = app.id and app.id = ? and app.deleted_at is null"+
				"     inner join projects prj on app.project_id = prj.id and (prj.organization_id in (select id from organizations where user_id=?) or prj.organization_id in (select organization_id from organization_members where user_id= ? and user_role=1)) and prj.deleted_at is null"+
				"     where env.deleted_at is null) or"+
				" (environments.id IN ("+
				"     select distinct env.id from environments env"+
				"     inner join applications app on env.application_id = app.id and app.id = ? and app.deleted_at is null"+
				"     inner join projects prj on app.project_id = prj.id and prj.user_id =? and prj.deleted_at is null"+
				"     where env.deleted_at is null) or"+
				" environments.id IN ("+
				"     select distinct env.id from environments env"+
				"     inner join applications app on env.application_id = app.id and app.id = ? and app.deleted_at is null"+
				"     inner join projects prj on prj.id = app.project_id and prj.deleted_at is null"+
				"     inner join authorizations auth on auth.project_id = prj.id and auth.application_id = 0 and auth.deleted_at is null"+
				"     where (auth.user_id = ? or auth.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and env.deleted_at is null) or"+
				" environments.id IN ("+
				"     select distinct env.id from environments env"+
				"     inner join applications app on env.application_id = app.id and app.id = ? and app.deleted_at is null"+
				"     inner join authorizations auth on app.id = auth.application_id and auth.application_id <> 0 and auth.environment_id = 0 and auth.deleted_at is null"+
				"     where (auth.user_id = ? or auth.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and env.deleted_at is null) or"+
				" environments.id IN (select environment_id from authorizations"+
				"     where (user_id = ? or authorizations.group_id"+
				"     IN(select groups.id from groups inner join group_members gm on groups.id = gm.group_id"+
				"     where gm.user_id = ? and groups.deleted_at is null))"+
				"     and application_id = ? and environment_id <> 0 and deleted_at is null))"+
				" and environments.deleted_at is null",
				aid, uid, uid, aid, uid, aid, uid, uid, aid, uid, uid, uid, uid, aid).
			Order("environments.name, environments.id desc").
			Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	return dataList, nil
}
func (data *Environment) FindAllByApplicationAdmin(db *gorm.DB, aid uint64) (*[]Environment, error) {
	var err error
	datas := []Environment{}
	err = db.Model(&Environment{}).Preload("Application").Preload("Application.Project").Preload("Resource").Preload("PluginVersion").Where(&Environment{ApplicationID: aid}).Order("name, id desc").Find(&datas).Error
	if err != nil {
		return &[]Environment{}, err
	}
	return &datas, nil
}
func (data *Environment) FindAllEnvironmentCountByApplication(db *gorm.DB, aid uint64) int {
	count := 0
	db.Model(&Environment{}).Where(&Environment{ApplicationID: aid}).Count(&count)
	return count
}

func (data *Environment) IsNameExists(db *gorm.DB, aid uint, name string) bool {
	count := 0
	db.Model(&Environment{}).Where(&Environment{ApplicationID: uint64(aid), Name: name}).Count(&count)
	return count > 0
}

func (data *Environment) IsValidResource(db *gorm.DB, aid uint64, env *Environment, resource *Resource, additionalStorage uint64) (string, bool) {

	var err error
	application := NewApplication()
	app, err := application.Find(db, aid)
	if err != nil {
		return "Server Error", false
	}
	usage, _ := data.GetUsedResource(db, app.ProjectID, false)

	tempEnv := Environment{}
	if env != nil {
		tempEnv = *env
		if tempEnv.Resource != nil {
			usage.Core = usage.Core - uint64(int(data.Resource.Cores)*int(data.Replicas))
			usage.Memory = usage.Memory - uint64(int(data.Resource.Memory)*int(data.Replicas))
		}
	} else {
		tempEnv = *data
	}

	if usage.Disk+additionalStorage > uint64(app.Project.Subscription.DiskSpace) {
		return "DiskSpace quota limit exceeds", false
	}

	usage.Core += resource.Cores * uint64(tempEnv.Replicas)
	usage.Memory += resource.Memory * uint64(tempEnv.Replicas)

	return "Invalid resource parameter, quota limit exceeds", usage.Core <= uint64(app.Project.Subscription.Cores) && usage.Memory <= uint64(app.Project.Subscription.Memory)

}

func (data *Environment) Find(db *gorm.DB, pid uint64) (*Environment, error) {
	err := db.Model(&Environment{}).
		Preload("Application").
		Preload("Application.Project").
		Preload("Application.Project.Organization").
		Preload("Application.Project.Subscription").
		Preload("Application.Project.User").
		Preload("Application.Cluster").
		Preload("Application.Cluster.Organization").
		Preload("Application.Cluster.DNS").
		Preload("Application.Cluster.ImageRegistry").
		Preload("Resource").
		Preload("Application.Plugin").
		Preload("PluginVersion").
		Preload("PluginVersion.Plugin").
		Preload("CronJob").
		Preload("Storage").
		Preload("LoadBalancer").
		Preload("CronJob.User").
		Preload("InitContainers").
		//Preload("AddOns").
		Where("id = ?", pid).Take(&data).Error

	if err != nil {
		return &Environment{}, err
	}
	return data, nil
}

func (data *Environment) FindWithProApp(db *gorm.DB, pid uint64) (*Environment, error) {
	err := db.Model(&Environment{}).Preload("Application").Preload("Application.Project").Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &Environment{}, err
	}
	return data, nil
}

func (data *Environment) UpdateStatus(db *gorm.DB, id uint64, status bool) (*Environment, error) {
	err := db.Model(&Environment{}).
		Where("id = ?", id).UpdateColumns(
		map[string]interface{}{
			"active":     status,
			"updated_at": time.Now(),
		}).Error
	if err != nil {
		return &Environment{}, err
	}
	return data, nil
}
func (d *EnvironmentType) UpdateFileManagerStatus(db *gorm.DB, data *Environment) (*Environment, error) {
	err := db.Model(&Environment{}).
		Where("id = ?", uint64(data.ID)).UpdateColumns(
		map[string]interface{}{
			"file_manager_enabled": data.FileManagerEnabled,
		}).Error
	if err != nil {
		return &Environment{}, err
	}
	return data, nil
}
func (data *Environment) Update(db *gorm.DB) (*Environment, error) {
	var err error
	var app = Environment{Active: data.Active, ApplyImmediately: data.ApplyImmediately}
	if data.Name != "" {
		app.Name = data.Name
	}
	if len(data.Setting.RawMessage) > 0 {
		app.Setting.RawMessage = data.Setting.RawMessage
	}
	if data.ApplicationID != 0 {
		app.ApplicationID = data.ApplicationID
	}
	if data.ResourceID != 0 {
		app.ResourceID = data.ResourceID
	}
	if data.PluginVersionID != 0 {
		app.PluginVersionID = data.PluginVersionID
	}
	if data.Replicas != 0 {
		app.Replicas = data.Replicas
	}
	if data.GitBranch != "" {
		app.GitBranch = data.GitBranch
	}
	if data.GitUrl != "" {
		app.GitUrl = data.GitUrl
	}
	if data.ImageTag != "" {
		app.ImageTag = data.ImageTag
	}
	if data.Variables.RawMessage != nil {
		app.Variables = data.Variables
	}
	if data.Attributes.RawMessage != nil {
		app.Attributes = data.Attributes
	}
	if data.Scripts.RawMessage != nil {
		app.Scripts = data.Scripts
	}
	if data.OtherVersion.RawMessage != nil {
		app.OtherVersion = data.OtherVersion
	}
	if data.Version.RawMessage != nil {
		app.Version = data.Version
	}
	if data.DeploymentStrategy.RawMessage != nil {
		app.DeploymentStrategy = data.DeploymentStrategy
	}
	if data.AutoScaler.RawMessage != nil {
		app.AutoScaler = data.AutoScaler
	}
	if data.CiRequest != nil {
		app.CiRequest = data.CiRequest
	}
	if data.RepositoryImage != nil {
		app.RepositoryImage = data.RepositoryImage
	}
	if data.OperatorPayload.RawMessage != nil {
		app.OperatorPayload = data.OperatorPayload
	}
	err = db.Model(&Environment{}).Where("id = ?", data.ID).Updates(app).Update("external_url", data.ExternalURL).Error
	if err != nil {
		return &Environment{}, err
	}
	return data, nil
}

func (data *Environment) UpdateWhiteListedIPs(db *gorm.DB, id uint64, whitelistedIps string) (*Environment, error) {
	env := &Environment{}
	err := db.Model(&Environment{}).
		Where("id = ?", id).Update("whitelisted_ips", whitelistedIps).Take(env).Error
	if err != nil {
		return &Environment{}, err
	}
	return env, nil
}

func (data *Environment) ChangeIsActive(db *gorm.DB, isActive bool) error {
	db = db.Model(&Environment{}).Where("id = ?", data.ID).Take(&Environment{}).UpdateColumns(
		map[string]interface{}{
			"active": isActive,
		},
	)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func (data *Environment) UpdateVariables(db *gorm.DB) (*Environment, error) {
	err := db.Model(&Environment{}).Where("id = ?", data.ID).UpdateColumns(map[string]interface{}{
		"variables":         data.Variables,
		"user_variables":    data.UserVariables,
		"apply_immediately": data.ApplyImmediately,
	}).Error

	if err != nil {
		return &Environment{}, err
	}
	return data, nil
}

func (data *Environment) Delete(db *gorm.DB, id uint64) (int64, error) {
	err := db.Model(&Authorization{}).Delete(&Authorization{}, "environment_id = ?", id).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return 0, err
	}

	db = db.Model(&Environment{}).Where("id = ?", id).Take(&Environment{}).Delete(&Environment{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Environment not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (data *Environment) FindAllAddons(db *gorm.DB, id uint64) ([]*Environment, error) {
	dataList := []*Environment{}
	err := db.Model(&Environment{}).
		Preload("PluginVersion").
		Preload("PluginVersion.Plugin").
		Preload("Resource").
		Preload("LoadBalancer").
		Where("parent_id = ?", id).
		Order("environments.name, environments.id desc").
		Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (data *Environment) FindAddon(db *gorm.DB, pid uint64, parentId uint64) (*Environment, error) {
	err := db.Model(&Environment{}).
		Preload("Resource").
		Preload("PluginVersion").
		Preload("PluginVersion.Plugin").
		//Preload("AddOns").
		Where("id = ? and parent_id = ?", pid, parentId).Take(&data).Error

	if err != nil {
		return &Environment{}, err
	}
	return data, nil
}

func (data *Environment) FindAllEnvironmentByProject(db *gorm.DB, pid uint64) ([]*Environment, error) {
	dataList := []*Environment{}
	err := db.Model(&Environment{}).
		Preload("Application").
		Preload("Application.Cluster").
		Preload("Parent").
		Preload("Resource").
		Preload("Storage").
		Preload("CronJob").
		Preload("LoadBalancer").
		Joins("inner join applications on applications.id = environments.application_id and applications.project_id = ? and applications.deleted_at is null", pid).
		Where("environments.deleted_at is null").
		Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	addonEnv := []*Environment{}
	for _, d := range dataList {
		addons := []*Environment{}
		err = db.Model(&Environment{}).
			Preload("Resource").
			Preload("PluginVersion").
			Preload("PluginVersion.Plugin").
			Where(&Environment{ParentID: uint64(d.ID)}).
			Find(&addons).Error
		if err != nil {
			return nil, err
		}
		addonEnv = append(addonEnv, addons...)
	}
	if len(addonEnv) > 0 {
		dataList = append(dataList, addonEnv...)
	}
	return dataList, nil
}
func (data *Environment) FindEnvironmentsByProject(db *gorm.DB, pid uint64, startdate, enddate string) ([]*Environment, error) {
	t := time.Now()
	dateend := t
	datestart := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	if startdate != "" {
		datestart, _ = time.Parse("2006-01-02", startdate)
	}
	if enddate != "" {
		dateend, _ = time.Parse("2006-01-02", enddate)
	}
	dataList, app, cluster := []*Environment{}, &Application{}, &Cluster{}
	err = db.Raw("select * from environments WHERE  environments.application_id IN (select id from applications where applications.project_id=?) and (environments.deleted_at is null or environments.deleted_at BETWEEN ?::timestamp AND ?::timestamp)", pid, datestart, dateend).Scan(&dataList).Error
	if err != nil {
		return nil, err
	}
	mp := map[uint64]interface{}{}
	for _, j := range dataList {
		d, ok := mp[j.ApplicationID]
		if ok {
			app.Cluster = d.(map[string]interface{})["cluster"].(*Cluster)
			j.Application = d.(map[string]interface{})["app"].(*Application)
			continue
		}
		err = db.Raw("select * from applications WHERE id=?", j.ApplicationID).Scan(&app).Error
		if err != nil {
			return nil, err
		}
		err = db.Raw("select * from clusters WHERE  id=?", app.ClusterID).Scan(&cluster).Error
		if err != nil {
			return nil, err
		}
		mp[j.ApplicationID] = map[string]interface{}{
			"cluster": cluster,
			"app":     app,
		}
		app.Cluster = cluster
		j.Application = app
	}
	return dataList, nil
}

func (data Environment) GetUsedResource(db *gorm.DB, pid uint64, isCluster bool) (Usage, error) {
	memory, core, cronjob := uint64(0), uint64(0), uint64(0)
	var disk uint64
	var err error
	var env []*Environment
	if isCluster {
		env, err = data.FindAllEnvironmentByCluster(db, pid)
	} else {
		env, err = data.FindAllEnvironmentByProject(db, pid)
	}
	if err == nil {
		for _, d := range env {
			memory += d.Resource.Memory * uint64(d.Replicas)
			core += d.Resource.Cores * uint64(d.Replicas)
			for range d.CronJob {
				memory += d.Resource.Memory
				core += d.Resource.Cores
			}
			for range d.InitContainers {
				memory += d.Resource.Memory
				core += d.Resource.Cores
			}
			for _, st := range d.Storage {
				disk += st.Capacity
			}
			cronjob += uint64(len(d.CronJob))
		}
	}
	return Usage{
		Core:         core,
		Memory:       memory,
		Disk:         disk,
		TotalCronJob: cronjob,
	}, nil
}

func (data *Environment) FindAllEnvironmentByCluster(db *gorm.DB, cid uint64) ([]*Environment, error) {
	dataList := []*Environment{}
	err := db.Model(&Environment{}).
		Preload("Application").
		Preload("Parent").
		Preload("Storage").
		Preload("Resource").
		Joins("inner join applications on applications.id = environments.application_id and applications.cluster_id = ? and applications.deleted_at is null", cid).
		Where("environments.deleted_at is null").
		Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	addonEnv := []*Environment{}
	for _, d := range dataList {
		addons := []*Environment{}
		err = db.Model(&Environment{}).
			Preload("Resource").
			Preload("PluginVersion").
			Preload("PluginVersion.Plugin").
			Where(&Environment{ParentID: uint64(d.ID)}).
			Find(&addons).Error
		if err != nil {
			return nil, err
		}
		addonEnv = append(addonEnv, addons...)
	}
	if len(addonEnv) > 0 {
		dataList = append(dataList, addonEnv...)
	}
	return dataList, nil
}

func (data *Environment) UpdateScheudle(db *gorm.DB, id uint) (*Environment, error) {
	if data.Schedules.RawMessage != nil {
		err = db.Model(&Environment{}).Where("id = ?", id).UpdateColumns(map[string]interface{}{
			"schedules": data.Schedules,
		}).Error
		if err != nil {
			return &Environment{}, err
		}
	}
	return data, nil
}

func (data *Environment) UpdateExternalSecret(db *gorm.DB) (*Environment, error) {
	if data.ExternalSecret.RawMessage != nil {
		err = db.Model(&Environment{}).Where("id = ?", data.ID).UpdateColumns(map[string]interface{}{
			"external_secret": data.ExternalSecret,
		}).Error
		if err != nil {
			return &Environment{}, err
		}
	}
	return data, nil
}
func (data *Environment) UpdateLogging(db *gorm.DB) (*Environment, error) {
	if data.ExternalLogging.RawMessage != nil {
		err = db.Model(&Environment{}).Where("id = ?", data.ID).UpdateColumns(map[string]interface{}{
			"external_logging": data.ExternalLogging,
		}).Error
		if err != nil {
			return &Environment{}, err
		}
	}
	return data, nil
}

func (data *Environment) UpdateErrorMessage(db *gorm.DB) error {
	return db.Model(&Environment{}).Where("id = ?", data.ID).UpdateColumns(map[string]interface{}{
		"error_message": data.ErrorMessage,
	}).Error
}

func (data *Environment) RenameEnvironment(db *gorm.DB, id uint, name string) error {
	err = db.Model(&Environment{}).Where("id = ?", id).Take(&Environment{}).UpdateColumns(
		map[string]interface{}{
			"name": name,
		},
	).Error
	if err != nil {
		return err
	}
	return nil
}

func (data *Environment) UpdateEnvironmentScript(db *gorm.DB) error {
	return db.Model(&Environment{}).Where("id = ?", data.ID).UpdateColumns(map[string]interface{}{
		"scripts": data.Scripts,
	}).Error
}
