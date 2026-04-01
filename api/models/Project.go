package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type ClusterScopeEnum int

const (
	SHARED_SCOPE ClusterScopeEnum = iota
	ORGANIZATION_SCOPE
)

var types = [...]string{
	"Shared Cluster",
	"Organization Cluster",
}

func (c ClusterScopeEnum) String() string {
	return types[c]
}

func (c *ClusterScopeEnum) Scan(value interface{}) error {
	*c = ClusterScopeEnum(value.(int64))
	return nil
}
func (c ClusterScopeEnum) Value() (driver.Value, error) { return int64(c), nil }

type Project struct {
	gorm.Model
	Name           string           `gorm:"size:255;not null;" json:"name"`
	Description    string           `gorm:"size:1024;null;" json:"description"`
	ProjectCode    string           `gorm:"size:5;null;" json:"project_code"`
	Tags           postgres.Jsonb   `gorm:"null;" json:"tags"`
	ClusterScope   ClusterScopeEnum `gorm:"null;" json:"cluster_scope"`   // "Shared Cluster" / "Organization Cluster"
	Region         string           `gorm:"size:127;null;" json:"region"` // Global or Regional
	Logging        postgres.Jsonb   `json:"logging"`                      // blank for default Logging
	Monitoring     postgres.Jsonb   `json:"monitoring"`                   // blank for default Monitoring
	BaseDomain     string           `gorm:"size:127;null;" json:"base_domain"`
	DedicatedLb    bool             `gorm:"default:false;" json:"dedicated_lb"`
	OptimizeCost   bool             `gorm:"default:false;" json:"optimize_cost"`
	Active         bool             `gorm:"not null;" json:"active"`
	Attributes     postgres.Jsonb   `json:"attributes"`
	Subscription   *Subscription    `gorm:"foreignkey:SubscriptionID" json:"subscription,omitempty"`
	SubsUpdated    *time.Time       `gorm:"null" json:"subscription_updated"`
	SubscriptionID uint64           `gorm:"not null" json:"subscription_id"`
	Image          string           `gorm:"null" json:"image,omitempty"`
	Variables      postgres.Jsonb   `sql:"json" json:"variables"`
	User           *User            `gorm:"foreignkey:UserID" json:"user,omitempty"`
	UserID         uint64           `gorm:"not null" json:"user_id"`
	Organization   *Organization    `gorm:"foreignkey:OrganizationID" json:"organization,omitempty"`
	OrganizationId uint64           `gorm:"default:0" json:"organization_id"`
	UsageQuota     postgres.Jsonb   `json:"usage_quota"`
}

type Usage struct {
	Memory       uint64                  `json:"memory"`
	Disk         uint64                  `json:"disk"`
	Core         uint64                  `json:"core"`
	DataTransfer *map[string]interface{} `json:"data_transfer"`
	TotalCiBuild uint64                  `json:"total_ci_build"`
	TotalCronJob uint64                  `json:"total_cron_job"`
	Apps         int                     `json:"apps"`
}

type IProject interface {
	Save(db *gorm.DB, data *Project) (*Project, error)
	FindAll(db *gorm.DB, data *Project) (*[]Project, error)
	FindAllProjects(db *gorm.DB) (*[]Project, error)
	FindAllDeactivatedProjectsByThresholdDays(db *gorm.DB, thresholdDays int) (*[]Project, error)
	FindAllDeactivatedProjectsByThresholdDaysByUser(db *gorm.DB, userId uint, thresholdDays int) (*[]Project, error)
	IsNameExists(db *gorm.DB, userID uint, name string, data *Project) bool
	FindAllByUser(db *gorm.DB, userID uint, data *Project) ([]Project, error)
	FindUserProjectOnly(db *gorm.DB, userID, oid uint) ([]Project, error)
	SearchProject(db *gorm.DB, userID uint, query string) ([]Project, error)
	Find(db *gorm.DB, pid uint64) (*Project, error)
	FindProject(db *gorm.DB, oid uint) (*[]Project, error)
	Update(db *gorm.DB, data *Project) (*Project, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
	ChangeIsActive(db *gorm.DB, data *Project, isActive bool) error
	ActiveDeactiveAll(db *gorm.DB, data *Project, userId uint64, isActive bool) error
	DeleteProjectOwner(db *gorm.DB, userID uint64, oid uint64) (*Project, error)
	UpdateUsageQuota(db *gorm.DB, data *Project) (*Project, error)
	Rename(db *gorm.DB, name string, proj *Project) error
	GetProjectByName(db *gorm.DB, name string, org_id uint64) (Project, error)
}
type ProjectRepo struct {
}

func NewIProject() IProject {
	return &ProjectRepo{}
}

func (data *ProjectRepo) Rename(db *gorm.DB, name string, proj *Project) error {
	return db.Model(&Project{}).Where("id = ?", proj.ID).Update("name", name).Error
}

func (data *Project) ToJson() (map[string]interface{}, error) {
	msg, _ := json.Marshal(data)
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}

func (data *Project) Prepare() {
	data.ID = 0
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	data.Description = html.EscapeString(strings.TrimSpace(data.Description))
	data.User = &User{}
	data.Subscription = &Subscription{}
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *Project) Validate() error {
	estr := ""
	if data.ID == 0 && data.Name == "" {
		estr = "required name "
		return errors.New(estr)
	}
	if data.ID == 0 && data.SubscriptionID == 0 {
		estr = "required subscription "
		return errors.New(estr)
	}
	if len(data.Name) > 64 {
		estr = "name must not exceed 64 characters"
		return errors.New(estr)
	}
	if data.Name != "" && !ValidateName(data.Name) {
		estr = "invalid characters in name"
		return errors.New(estr)
	}
	if len(data.Description) > 1024 {
		estr = "description must not exceed 1024 characters"
		return errors.New(estr)
	}
	if len(data.ProjectCode) > 5 {
		estr = "project code must not exceed 5 characters"
		return errors.New(estr)
	}
	if len(data.Region) > 64 {
		estr = "region must not exceed 64 characters"
		return errors.New(estr)
	}
	if len(data.BaseDomain) > 64 {
		estr = "base domain must not exceed 64 characters"
		return errors.New(estr)
	}
	if len(data.BaseDomain) > 0 && !ValidateDomain(data.BaseDomain) {
		estr = "invalid domain"
		return errors.New(estr)
	}
	return nil
}

func ValidateName(name string) bool {
	valid, _ := regexp.Match("^\\w+([\\s-_]\\w+)*$", []byte(name))
	return valid
}
func ValidateDomain(name string) bool {
	valid, _ := regexp.Match("([a-z0-9|-]+\\.)*[a-z0-9|-]+\\.[a-z]+", []byte(name))
	return valid
}

func (d *ProjectRepo) Save(db *gorm.DB, data *Project) (*Project, error) {
	err = db.Model(&Project{}).Create(&data).Error
	if err != nil {
		return &Project{}, err
	}
	return data, nil
}

func (d *ProjectRepo) FindAll(db *gorm.DB, data *Project) (*[]Project, error) {
	var err error
	datas := []Project{}
	err = db.Model(&Project{}).
		Where("organization_id = ?", data.OrganizationId).
		Order("name, id desc").Find(&datas).Error
	if err != nil {
		return &[]Project{}, err
	}
	return &datas, nil
}

func (d *ProjectRepo) FindAllProjects(db *gorm.DB) (*[]Project, error) {
	datas := []Project{}
	err := db.Model(&Project{}).Preload("Subscription").Where("active = true").Find(&datas).Error
	if err != nil {
		return &[]Project{}, err
	}
	return &datas, nil
}

func (d *ProjectRepo) FindAllDeactivatedProjectsByThresholdDays(db *gorm.DB, thresholdDays int) (*[]Project, error) {
	thresholdDate := time.Now().AddDate(0, 0, -thresholdDays)
	datas := []Project{}
	err := db.Model(&Project{}).Where("active = false and deleted_at is null and updated_at <= ?", thresholdDate).Find(&datas).Error
	if err != nil {
		return &[]Project{}, err
	}
	return &datas, nil
}

func (d *ProjectRepo) FindAllDeactivatedProjectsByThresholdDaysByUser(db *gorm.DB, userId uint, thresholdDays int) (*[]Project, error) {
	thresholdDate := time.Now().AddDate(0, 0, -thresholdDays)
	datas := []Project{}
	err := db.Model(&Project{}).Where("active = false and deleted_at is null and updated_at <= ? and user_id = ?", thresholdDate, userId).Find(&datas).Error
	if err != nil {
		return &[]Project{}, err
	}
	return &datas, nil
}

func (d *ProjectRepo) IsNameExists(db *gorm.DB, userID uint, name string, data *Project) bool {
	count := 0
	db.Model(&Project{}).Where("user_id=? and organization_id=? and lower(name)=?", userID, data.OrganizationId, strings.ToLower(name)).Count(&count)
	return count > 0
}

func (d *ProjectRepo) FindAllByUser(db *gorm.DB, userID uint, data *Project) ([]Project, error) {
	dataList := []Project{}
	err := db.Model(&Project{}).
		Preload("Subscription").
		Preload("User").
		Where("organization_id = ? ", data.OrganizationId).
		Where("organization_id IN (select id from organizations where user_id =?) or"+
			" organization_id IN (select organization_id from organization_members where organization_id=? and user_id = ? and user_role = 1 and deleted_at is null) or"+
			" (projects.id IN (select distinct project_id from authorizations where (authorizations.user_id=? or authorizations.group_id IN("+
			" select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null))"+
			" and authorizations.deleted_at is null) or"+
			" projects.id IN (select id from projects where user_id=? and deleted_at is null)) and projects.deleted_at is null", userID, data.OrganizationId, userID, userID, userID, userID).
		Where(" projects.user_id in (select id from users where active = true)").
		Order("projects.name, projects.id desc").
		Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (d *ProjectRepo) FindUserProjectOnly(db *gorm.DB, userID, oid uint) ([]Project, error) {
	datas := []Project{}
	err = db.
		Model(&Project{}).
		Where("organization_id = ? ", oid).
		Where(&Project{UserID: uint64(userID)}).
		Order("name, id desc").Find(&datas).Error
	if err != nil {
		return []Project{}, err
	}
	return datas, err
}

func (data *ProjectRepo) SearchProject(db *gorm.DB, userID uint, query string) ([]Project, error) {
	dataList := []Project{}
	err := db.Model(&Project{}).
		Preload("User").
		Preload("Subscription").
		Preload("Organization").
		Where(
			" (projects.id IN (select distinct project_id from authorizations where (authorizations.user_id=? or authorizations.group_id IN("+
				" select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null))"+
				" and authorizations.deleted_at is null) or"+
				" projects.id IN (select id from projects where user_id=? and deleted_at is null)) and projects.deleted_at is null"+
				" and lower(projects.name) LIKE lower(?)",
			userID, userID, userID, "%"+query+"%").
		Order("projects.name, projects.id desc").
		Find(&dataList).Error
	if err != nil {
		return dataList, err
	}
	return dataList, nil
}

func (data *User) SearchUserName(db *gorm.DB, userID, oID uint, query string) ([]User, error) {
	dataList := []User{}
	que := strings.TrimSpace(query)
	if oID > 0 {
		if len(que) >= 2 {
			err := db.Debug().Model(&User{}).
				Joins("left join organization_members on organization_members.user_id = users.id").
				Where("organization_members.organization_id =? and active = ? and organization_members.deleted_at is null and"+
					"(CONCAT(lower(users.first_name),' ',lower(users.last_name)) LIKE lower(?)"+
					"or lower(users.email) LIKE lower(?))",
					oID,
					true, "%"+que+"%", "%"+que+"%").
				Find(&dataList).Error
			if err != nil {
				return dataList, err
			}
		} else {
			return dataList, errors.New("must contain atleast 2 characters")
		}
	} else {

		return dataList, nil
	}
	return dataList, nil
}

func (d *ProjectRepo) Find(db *gorm.DB, pid uint64) (*Project, error) {
	data := &Project{}
	err = db.Model(&Project{}).
		Where("id = ?", pid).
		Preload("User").
		Preload("Organization").
		Preload("Subscription").Take(&data).Error
	if err != nil {
		return &Project{}, err
	}
	return data, nil
}

func (data *ProjectRepo) FindProject(db *gorm.DB, oid uint) (*[]Project, error) {
	var dataList []Project
	err = db.Model(&Project{}).Preload("Subscription").
		Where("organization_id=?", oid).
		Find(&dataList).Error
	if err != nil {
		return &[]Project{}, err
	}
	return &dataList, nil
}

func (d *ProjectRepo) Update(db *gorm.DB, data *Project) (*Project, error) {
	mp := map[string]interface{}{
		"active":         data.Active,
		"dedicated_lb":   data.DedicatedLb,
		"optimized_cost": data.OptimizeCost,
		"description":    data.Description,
	}
	if data.Name != "" {
		mp["name"] = data.Name
	}
	if data.Image != "" {
		mp["image"] = data.Image
	}
	if data.SubscriptionID != 0 {
		mp["subscription_id"] = data.SubscriptionID
	}
	if data.UserID != 0 {
		mp["user_id"] = data.UserID
	}
	if data.Variables.RawMessage != nil {
		mp["variables"] = data.Variables
	}
	if len(data.BaseDomain) > 0 {
		mp["base_domain"] = data.BaseDomain
	}
	if len(data.ProjectCode) > 0 {
		mp["project_code"] = data.ProjectCode
	}
	if data.Tags.RawMessage != nil {
		mp["tags"] = data.Tags
	}
	if data.Logging.RawMessage != nil {
		mp["logging"] = data.Logging
	}
	if data.Monitoring.RawMessage != nil {
		mp["monitoring"] = data.Monitoring
	}
	if data.ClusterScope >= 0 {
		mp["cluster_scope"] = data.ClusterScope
	}
	if data.Attributes.RawMessage != nil {
		mp["attributes"] = data.Attributes
	}
	if data.SubsUpdated != nil {
		mp["subscription_updated"] = data.SubsUpdated
	}
	err := db.Model(&Project{}).Where("id = ?", data.ID).UpdateColumn(mp).Error
	if err != nil {
		return &Project{}, err
	}
	return data, nil
}

func (data *ProjectRepo) Delete(db *gorm.DB, id uint64) (int64, error) {
	err := db.Model(&Authorization{}).
		Delete(&Authorization{}, "project_id = ?", id).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return 0, err
	}

	db = db.Model(&Project{}).Where("id = ?", id).Take(&Project{}).Delete(&Project{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Project not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (d *ProjectRepo) ChangeIsActive(db *gorm.DB, data *Project, isActive bool) error {
	db = db.Model(&Project{}).
		Where("id = ?", data.ID).Take(&Project{}).UpdateColumns(
		map[string]interface{}{
			"active": isActive,
		},
	)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func (d *ProjectRepo) ActiveDeactiveAll(db *gorm.DB, data *Project, userId uint64, isActive bool) error {
	db = db.Model(&Project{}).
		Where("user_id = ?", userId).
		Where("organization_id = ? ", data.OrganizationId).
		UpdateColumns(
			map[string]interface{}{
				"active": isActive,
			},
		)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func (d *ProjectRepo) DeleteProjectOwner(db *gorm.DB, userID uint64, oid uint64) (*Project, error) {
	data := &Project{}
	err := db.Model(&Project{}).Where("user_id=? and organization_id=?", userID, oid).Update("user_id", 0).Error
	if err != nil {
		return &Project{}, err
	}
	return data, nil
}

func (d *ProjectRepo) UpdateUsageQuota(db *gorm.DB, data *Project) (*Project, error) {
	mp := map[string]interface{}{}
	if data.UsageQuota.RawMessage != nil {
		mp["usage_quota"] = data.UsageQuota
	}
	err = db.Model(&Project{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &Project{}, err
	}
	return data, nil
}

func (d *ProjectRepo) GetProjectByName(db *gorm.DB, name string, org_id uint64) (proj Project, err error) {
	err = db.Model(&Project{}).Where("lower(name) = ? and organization_id = ?", strings.ToLower(name), org_id).First(&proj).Error
	return
}
