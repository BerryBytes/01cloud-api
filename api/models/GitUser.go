package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

type GitUser struct {
	gorm.Model
	UserID          uint64          `gorm:"not null;" json:"user_id"`
	GitUserID       uint64          `gorm:"not null;" json:"git_user_id"`
	ServiceUserName string          `gorm:"size:255;" json:"service_user_name"`
	AccessToken     string          `gorm:"size:255;not null;" json:"access_token"`
	ServiceName     string          `gorm:"size:255;not null;" json:"service_name"`
	ServiceUrl      string          `gorm:"size:255;null" json:"service_url"`
	User            *User           `gorm:"foreignkey:UserID" json:"user"`
	SecretKey       string          `gorm:"size:255;null;" json:"secret_key"`
	Region          string          `gorm:"size:255;null;" json:"region"`
	Active          bool            `gorm:"not null;default:true" json:"active"`
	IsOauth         bool            `gorm:"not null;default:false" json:"is_oauth"`
	Token           *postgres.Jsonb `gorm:"size:255;null;" json:"token"`
}

type GitRepo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Owner    string `json:"owner"`
	HtmlUrl  string `json:"html_url"`
	CloneURL string `json:"clone_url"`
	GitURL   string `json:"git_url"`
}

type Commit struct {
	SHA     string `json:"sha"`
	Message string `json:"message"`
	Author  string `json:"author"`
	Time    string `json:"time"`
}

type GitUserInterface interface {
	Save(db *gorm.DB, data *GitUser) (*GitUser, error)
	Find(db *gorm.DB, pid uint64) (*GitUser, error)
	FindAll(db *gorm.DB) (*[]GitUser, error)
	FindByGitUserID(db *gorm.DB, pid uint64) (*GitUser, error)
	FindByUserId(db *gorm.DB, uid uint64) ([]GitUser, error)
	FindByUserIdAndService(db *gorm.DB, uid uint64, service string) (*GitUser, error)
	Update(db *gorm.DB, data *GitUser) (*GitUser, error)
	Delete(db *gorm.DB, uid uint64) (int64, error)
}
type GitUserType struct {
}

func NewGitUser() GitUserInterface {
	return &GitUserType{}
}

func (data *GitUser) Validate() error {
	if data.UserID == 0 {
		return errors.New("user id required")
	}
	if data.GitUserID == 0 {
		return errors.New("git user id required")
	}
	if data.AccessToken == "" {
		return errors.New("access token is required")
	}
	if data.ServiceName == "" {
		return errors.New("service name  is required")
	}
	return nil
}

func (d *GitUserType) Save(db *gorm.DB, data *GitUser) (*GitUser, error) {
	err := db.Model(&GitUser{}).Create(&data).Error
	if err != nil {
		return &GitUser{}, err
	}
	return data, nil
}

func (d *GitUserType) FindAll(db *gorm.DB) (*[]GitUser, error) {
	datas := []GitUser{}
	err := db.Model(&GitUser{}).Find(&datas).Error
	if err != nil {
		return &[]GitUser{}, err
	}
	return &datas, nil
}

func (d *GitUserType) Find(db *gorm.DB, pid uint64) (*GitUser, error) {
	data := &GitUser{}
	err := db.Model(&GitUser{}).Where("id = ?", pid).Take(&data).Error
	if err != nil {
		return &GitUser{}, err
	}
	return data, nil
}

func (d *GitUserType) FindByGitUserID(db *gorm.DB, pid uint64) (*GitUser, error) {
	data := &GitUser{}
	err := db.Model(&GitUser{}).Where("git_user_id = ?", pid).Take(&data).Error
	if err != nil {
		return &GitUser{}, err
	}
	return data, nil
}
func (d *GitUserType) FindByUserId(db *gorm.DB, uid uint64) ([]GitUser, error) {
	var datas []GitUser
	err := db.Model(&GitUser{}).Select("id,git_user_id,service_name,service_user_name,user_id,created_at,updated_at,active").Where(&GitUser{UserID: uid}).Find(&datas).Error
	if err != nil {
		return nil, err
	}
	return datas, nil
}

func (d *GitUserType) FindByUserIdAndService(db *gorm.DB, uid uint64, service string) (*GitUser, error) {
	data := &GitUser{}
	err := db.Model(&GitUser{}).Where(&GitUser{UserID: uid, ServiceName: service}).Take(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (d *GitUserType) Update(db *gorm.DB, data *GitUser) (*GitUser, error) {
	err := db.Model(&GitUser{}).Where("id = ?", data.ID).Updates(GitUser{
		UserID:          data.UserID,
		GitUserID:       data.GitUserID,
		AccessToken:     data.AccessToken,
		ServiceUserName: data.ServiceUserName,
		Active:          data.Active,
		IsOauth:         data.IsOauth,
		Token:           data.Token,
	}).Error
	if err != nil {
		return &GitUser{}, err
	}
	return data, nil
}

func (d *GitUserType) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&GitUser{}).Where("id = ?", id).Take(&GitUser{}).Delete(&GitUser{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("git user not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
