package models

import (
	"encoding/json"
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type Authorization struct {
	gorm.Model
	Email         string       `gorm:"size:255;null;" json:"email"`
	User          *User        `gorm:"foreignkey:UserID;null" json:"user"`
	UserID        uint64       `gorm:"default:0;" json:"user_id"`
	UserRole      *UserRole    `gorm:"foreignkey:UserRoleID" json:"user_role"`
	UserRoleID    uint64       `gorm:"not null;" json:"user_role_id"`
	Project       *Project     `gorm:"foreignkey:ProjectID" json:"project"`
	ProjectID     uint64       `gorm:"not null;" json:"project_id"`
	Application   *Application `gorm:"foreignkey:ProjectID" json:"application"`
	ApplicationID uint64       `gorm:"not null;default:0" json:"application_id"`
	Environment   *Environment `gorm:"foreignkey:EnvironmentID;null" json:"environment"`
	EnvironmentID uint64       `gorm:"not null;default:0;" json:"environment_id"`
	Active        bool         `gorm:"not null;" json:"active"`
	Attributes    string       `gorm:"null;" json:"attributes"`
	Group         *Group       `gorm:"foreignkey:GroupID;null" json:"group"`
	GroupID       uint64       `gorm:"default:0;" json:"group_id"`
}

type AuthorizationInterface interface {
	Save(db *gorm.DB, data *Authorization) (*Authorization, error)
	Find(db *gorm.DB, pid uint64) (*Authorization, error)
	FindAll(db *gorm.DB) (*[]Authorization, error)
	Update(db *gorm.DB, data *Authorization) (*Authorization, error)
	IsUserExists(db *gorm.DB, data *Authorization) bool
	Delete(db *gorm.DB, id uint64) (int64, error)
	FindAllInProject(db *gorm.DB, pid uint64) (*[]Authorization, error)
	FindAllInEnv(db *gorm.DB, eid uint64) (*[]Authorization, error)

	GetRoleProject(db *gorm.DB, uid uint64, pid uint64) (Authorization, error)
	IsAuthorizedProject(db *gorm.DB, uid uint64, pid uint64) bool
	IsWriteAuthorizedProject(db *gorm.DB, uid uint64, pid uint64) bool
	IsAdminOfProject(db *gorm.DB, uid uint64, pid uint64) bool

	GetRoleApp(db *gorm.DB, uid uint64, pid uint64, aid uint64) (Authorization, error)
	IsAuthorizedApplication(db *gorm.DB, uid uint64, aid uint64, pid uint64) bool
	IsAdminOfApplication(db *gorm.DB, uid uint64, aid uint64, pid uint64) bool
	IsWriteAuthorizedApplication(db *gorm.DB, uid uint64, aid uint64, pid uint64) bool

	GetRoleEnv(db *gorm.DB, uid uint64, pid uint64, aid uint64, eid uint64) (Authorization, error)
	IsAdminOfEnvironment(db *gorm.DB, uid uint64, env *Environment) bool
	IsWriteAuthorizedEnvironment(db *gorm.DB, uid uint64, env *Environment) bool
	IsAuthorizedEnvironment(db *gorm.DB, uid uint64, env *Environment) bool

	IsAdminOfHelmEnvironment(db *gorm.DB, uid uint64, env *HelmEnvironment) bool
	IsWriteAuthorizedHelmEnvironment(db *gorm.DB, uid uint64, env *HelmEnvironment) bool
	IsAuthorizedHelmEnvironment(db *gorm.DB, uid uint64, env *HelmEnvironment) bool

	IsAuthorizedOrganization(db *gorm.DB, uid uint, oid uint) bool
	DeleteOrganizationMember(db *gorm.DB, id uint64, oid uint64) (int64, error)
}

type AuthorizationRepo struct{}

func NewAuthorizationRepo() AuthorizationInterface {
	return &AuthorizationRepo{}
}

func (data *Authorization) ToJson() (map[string]interface{}, error) {
	msg, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}

func (data *Authorization) Prepare() {
	data.ID = 0
	data.Email = html.EscapeString(strings.TrimSpace(data.Email))
	data.Active = true
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *Authorization) Validate() error {
	if data.Email == "" && data.GroupID == 0 {
		return errors.New("required Email or Group Id")
	}
	if data.UserRoleID == 0 {
		return errors.New("required UserRole")
	}
	if data.ProjectID == 0 {
		return errors.New("required Project")
	}
	return nil
}

func (d *AuthorizationRepo) Save(db *gorm.DB, data *Authorization) (*Authorization, error) {
	err = db.Model(&Authorization{}).Create(&data).Error
	if err != nil {
		return &Authorization{}, err
	}
	return data, nil
}

func (d *AuthorizationRepo) Find(db *gorm.DB, pid uint64) (*Authorization, error) {
	data := &Authorization{}
	err = db.Model(&Authorization{}).
		Preload("Application").
		Preload("Application.Project").
		Preload("Application.Cluster").
		//Preload("Resource").
		//Preload("PluginVersion").
		Preload("Group").
		Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &Authorization{}, err
	}
	return data, nil
}

func (d *AuthorizationRepo) FindAll(db *gorm.DB) (*[]Authorization, error) {
	datas := []Authorization{}
	err = db.Model(&Authorization{}).Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]Authorization{}, err
	}
	return &datas, nil
}

func (d *AuthorizationRepo) FindAllInProject(db *gorm.DB, pid uint64) (*[]Authorization, error) {
	datas := []Authorization{}
	err = db.Model(&Authorization{}).
		Preload("UserRole").
		Preload("User").
		Preload("Group").
		Where("project_id = ? and environment_id = ? and application_id = ? and user_id in (select id from users where active = ?)", pid, 0, 0, true).
		Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]Authorization{}, err
	}
	return &datas, nil
}

func (d *AuthorizationRepo) FindAllInEnv(db *gorm.DB, eid uint64) (*[]Authorization, error) {
	datas := []Authorization{}
	err = db.Model(&Authorization{}).
		Preload("UserRole").
		Preload("User").
		Preload("Group").
		Where(&Authorization{EnvironmentID: eid}).
		Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]Authorization{}, err
	}
	return &datas, nil
}

func (d *AuthorizationRepo) GetRoleProject(db *gorm.DB, uid uint64, pid uint64) (Authorization, error) {
	datas := []Authorization{}
	err = db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) and project_id = ?", uid, uid, pid).
		Preload("UserRole").Order("user_role_id").
		Find(&datas).Error
	if err != nil {
		return Authorization{}, err
	}
	if len(datas) == 0 {
		return Authorization{}, errors.New("not Authorized")
	}
	for _, data := range datas {
		if data.ApplicationID == 0 {
			return data, nil
		}
	}
	return Authorization{
		UserRole: &UserRole{
			Name: "Read",
			Code: 3,
		},
	}, nil
}

func (d *AuthorizationRepo) GetRoleApp(db *gorm.DB, uid uint64, pid uint64, aid uint64) (Authorization, error) {
	datas := []Authorization{}
	err = db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND (application_id = ? OR application_id = ?)", uid, uid, pid, aid, 0).
		Preload("UserRole").Order("user_role_id").Find(&datas).Error
	if err != nil {
		return Authorization{}, err
	}
	if len(datas) == 0 {
		return Authorization{}, errors.New("not Authorized")
	}
	for _, data := range datas {
		if data.EnvironmentID == 0 {
			return data, nil
		}
	}
	return Authorization{
		UserRole: &UserRole{
			Name: "Read",
			Code: 3,
		},
	}, nil
}

func (d *AuthorizationRepo) GetRoleEnv(db *gorm.DB, uid uint64, pid uint64, aid uint64, eid uint64) (Authorization, error) {
	datas := []Authorization{}
	err = db.
		Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND ((application_id = ? AND environment_id = ?) OR (application_id = ? AND environment_id = ?))", uid, uid, pid, aid, eid, 0, 0).
		Preload("UserRole").Order("user_role_id").Limit(1).Find(&datas).Error
	if err != nil {
		return Authorization{}, err
	}
	if len(datas) == 0 {
		return Authorization{}, errors.New("not Authorized")
	}
	return datas[0], nil
}

func (d *AuthorizationRepo) IsAuthorizedProject(db *gorm.DB, uid uint64, pid uint64) bool {
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in "+
			" (select groups.id from groups inner join group_members gm "+
			" on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) "+
			" AND project_id = ? AND deleted_at is null ",
			uid, uid, pid).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsAuthorizedOrganization(db *gorm.DB, uid uint, oid uint) bool {
	var count int
	db.Model(&Organization{}).
		Where("(user_id =$1 and id=$2) or "+
			"id IN (select organization_id from organization_members where organization_id=$2 and user_id = $1 and user_role = 1 and deleted_at is null)",
			uid, oid).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsWriteAuthorizedProject(db *gorm.DB, uid uint64, pid uint64) bool {
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND application_id = ? AND environment_id = ? AND (user_role_id = ? OR user_role_id = ?)", uid, uid, pid, 0, 0, 1, 2).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsAdminOfProject(db *gorm.DB, uid uint64, pid uint64) bool {
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND application_id = ? AND environment_id = ? AND user_role_id = ?", uid, uid, pid, 0, 0, 1).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsAuthorizedApplication(db *gorm.DB, uid uint64, aid uint64, pid uint64) bool {
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND (application_id = ? OR application_id = ?)", uid, uid, pid, aid, 0).
		Count(&count)
	print("count is", count)
	return count > 0
}

func (d *AuthorizationRepo) IsAdminOfApplication(db *gorm.DB, uid uint64, aid uint64, pid uint64) bool {
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND (application_id = ? OR application_id = ?) AND user_role_id = ?", uid, uid, pid, aid, 0, 1).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsWriteAuthorizedApplication(db *gorm.DB, uid uint64, aid uint64, pid uint64) bool {
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND (application_id = ? OR application_id = ?) AND (user_role_id = ? OR user_role_id = ?)", uid, uid, pid, aid, 0, 1, 2).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsAdminOfEnvironment(db *gorm.DB, uid uint64, env *Environment) bool {
	if d.IsAdminOfProject(db, uid, env.Application.ProjectID) {
		return true
	}
	var count int
	db.Model(&Authorization{}).
		Where(&Authorization{UserID: uid, ProjectID: env.Application.ProjectID, ApplicationID: env.ApplicationID, EnvironmentID: uint64(env.ID), UserRoleID: 1}).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsAdminOfHelmEnvironment(db *gorm.DB, uid uint64, env *HelmEnvironment) bool {
	if d.IsAdminOfProject(db, uid, env.Application.ProjectID) {
		return true
	}
	var count int
	db.Model(&Authorization{}).
		Where(&Authorization{UserID: uid, ProjectID: env.Application.ProjectID, ApplicationID: env.ApplicationID, EnvironmentID: uint64(env.ID), UserRoleID: 1}).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsWriteAuthorizedEnvironment(db *gorm.DB, uid uint64, env *Environment) bool {
	if d.IsWriteAuthorizedProject(db, uid, env.Application.ProjectID) {
		return true
	}
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND application_id = ? AND environment_id = ? AND (user_role_id = ? OR user_role_id = ?)", uid, uid, env.Application.ProjectID, env.ApplicationID, env.ID, 1, 2).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsWriteAuthorizedHelmEnvironment(db *gorm.DB, uid uint64, env *HelmEnvironment) bool {
	if d.IsWriteAuthorizedProject(db, uid, env.Application.ProjectID) {
		return true
	}
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND application_id = ? AND environment_id = ? AND (user_role_id = ? OR user_role_id = ?)", uid, uid, env.Application.ProjectID, env.ApplicationID, env.ID, 1, 2).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) IsAuthorizedEnvironment(db *gorm.DB, uid uint64, env *Environment) bool {
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND (application_id = ? OR application_id = ?) AND (environment_id = ? OR environment_id = ?)", uid, uid, env.Application.ProjectID, 0, env.ApplicationID, 0, env.ID).
		Count(&count)
	return count > 0
}
func (d *AuthorizationRepo) IsAuthorizedHelmEnvironment(db *gorm.DB, uid uint64, env *HelmEnvironment) bool {
	var count int
	db.Model(&Authorization{}).
		Where("(user_id = ? or group_id in (select groups.id from groups inner join group_members gm on groups.id = gm.group_id where gm.user_id = ? and groups.deleted_at is null)) AND project_id = ? AND (application_id = ? OR application_id = ?) AND (environment_id = ? OR environment_id = ?)", uid, uid, env.Application.ProjectID, 0, env.ApplicationID, 0, env.ID).
		Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) Update(db *gorm.DB, data *Authorization) (*Authorization, error) {
	var app = Authorization{Active: data.Active}
	if data.Email != "" {
		app.Email = data.Email
	}
	if data.UserID != 0 {
		app.UserID = data.UserID
	}
	if data.ProjectID != 0 {
		app.ProjectID = data.ProjectID
	}
	if data.ApplicationID != 0 {
		app.ApplicationID = data.ApplicationID
	}
	if data.UserRoleID != 0 {
		app.UserRoleID = data.UserRoleID
	}
	if data.EnvironmentID != 0 {
		app.EnvironmentID = data.EnvironmentID
	}
	if data.Attributes != "" {
		app.Attributes = data.Attributes
	}
	err = db.Model(&Authorization{}).Where("id = ?", data.ID).Updates(app).Error
	if err != nil {
		return &Authorization{}, err
	}
	return data, nil
}

func (d *AuthorizationRepo) IsUserExists(db *gorm.DB, data *Authorization) bool {
	var count int
	condition := "email=? and project_id=? and application_id=? and environment_id=? and group_id=?"
	db.Model(&Authorization{}).Where(condition, data.Email, data.ProjectID, data.ApplicationID, data.EnvironmentID, data.GroupID).Count(&count)
	return count > 0
}

func (d *AuthorizationRepo) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&Authorization{}).Where("id = ?", id).Take(&Authorization{}).Delete(&Authorization{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Authorization not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (d *AuthorizationRepo) DeleteOrganizationMember(db *gorm.DB, id uint64, oid uint64) (int64, error) {
	db = db.Model(&Authorization{}).Where("authorizations.user_id IN(select organization_members.user_id from organization_members where organization_members.user_id =?) and  authorizations.project_id IN(select projects.id from projects where projects.organization_id=?)", id, oid).Take(&Authorization{}).Delete(&Authorization{})

	//	("id = ?", id).Take(&Authorization{}).Delete(&Authorization{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("Authorization not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
