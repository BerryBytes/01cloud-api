package models

import (
	"encoding/json"
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

type EventType string

const (
	All    EventType = "all"
	Normal EventType = "normal"
	Error  EventType = "error"
)

type CiConfigInterface interface {
	Save(db *gorm.DB, data *CiConfig) (*CiConfig, error)
	FindAll(db *gorm.DB) (*[]CiConfig, error)
	Find(db *gorm.DB, pid uint64) (*CiConfig, error)
	FindByEnvironment(db *gorm.DB, eid uint) (*CiConfig, error)
	Update(db *gorm.DB, data *CiConfig) (*CiConfig, error)
	Delete(db *gorm.DB, id uint64) (int64, error)
}

type CiConfig struct {
	gorm.Model
	WebhookUrl          string       `gorm:"null;" json:"webhook_url"`
	WebhookToken        string       `gorm:"null;" json:"webhook_token"`
	SlackWebhookUrl     string       `gorm:"null;" json:"slack_webhook_url"`
	Emails              string       `sql:"json" json:"emails"`
	HookId              string       `gorm:"null;" json:"hook_id"`
	Events              string       `sql:"json" json:"events"`
	EmailNotification   bool         `gorm:"default:false;" json:"email_notification"`
	SlackNotification   bool         `gorm:"default:false;" json:"slack_notification"`
	WebhookNotification bool         `gorm:"default:false;" json:"webhook_notification"`
	Environment         *Environment `gorm:"foreignkey:EnvironmentID" json:"environment,omitempty"`
	EnvironmentID       uint         `gorm:"not null" json:"environment_id"`
	EventType           EventType    `gorm:"default:'all'" json:"event_type"`
}

type CiConfigRepo struct{}

func NewCiConfigRepo() CiConfigInterface {
	return &CiConfigRepo{}
}

func (data *CiConfig) ToJson() (map[string]interface{}, error) {
	msg, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}

func (data *CiConfig) Prepare() {
	data.ID = 0
	data.WebhookUrl = html.EscapeString(strings.TrimSpace(data.WebhookUrl))
	data.SlackWebhookUrl = html.EscapeString(strings.TrimSpace(data.SlackWebhookUrl))
	data.WebhookToken = html.EscapeString(strings.TrimSpace(data.WebhookToken))
	data.CreatedAt = time.Now()
	data.UpdatedAt = time.Now()
}

func (data *CiConfig) Validate() {
	if data.EventType == Normal {
		data.EventType = Normal
	} else if data.EventType == Error {
		data.EventType = Error
	} else {
		data.EventType = All
	}
}

func (r *CiConfigRepo) Save(db *gorm.DB, data *CiConfig) (*CiConfig, error) {
	err = db.Model(&CiConfig{}).Create(&data).Error
	if err != nil {
		return &CiConfig{}, err
	}
	return data, nil
}

func (r *CiConfigRepo) FindAll(db *gorm.DB) (*[]CiConfig, error) {
	datas := []CiConfig{}
	err = db.Model(&CiConfig{}).Order("id desc").Find(&datas).Error
	if err != nil {
		return &[]CiConfig{}, err
	}
	return &datas, nil
}

func (r *CiConfigRepo) Find(db *gorm.DB, pid uint64) (*CiConfig, error) {
	data := &CiConfig{}
	err = db.Model(&CiConfig{}).Where("id = ?", pid).Preload("Environment").Take(&data).Error
	if err != nil {
		return &CiConfig{}, err
	}
	return data, nil
}

func (r *CiConfigRepo) FindByEnvironment(db *gorm.DB, eid uint) (*CiConfig, error) {
	data := &CiConfig{}
	err = db.Model(&CiConfig{}).Where(&CiConfig{EnvironmentID: eid}).Preload("Environment").Take(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (r *CiConfigRepo) Update(db *gorm.DB, data *CiConfig) (*CiConfig, error) {
	mp := map[string]interface{}{}
	mp["slack_webhook_url"] = data.SlackWebhookUrl
	mp["webhook_url"] = data.WebhookUrl
	mp["webhook_token"] = data.WebhookToken
	mp["emails"] = data.Emails
	mp["events"] = data.Events
	mp["email_notification"] = data.EmailNotification
	mp["slack_notification"] = data.SlackNotification
	mp["webhook_notification"] = data.WebhookNotification
	mp["hook_id"] = data.HookId
	if data.EventType != "" {
		mp["event_type"] = data.EventType
	}
	err = db.Model(&CiConfig{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &CiConfig{}, err
	}

	return data, nil
}

func (r *CiConfigRepo) Delete(db *gorm.DB, id uint64) (int64, error) {
	db = db.Model(&CiConfig{}).Where("id = ?", id).Take(&CiConfig{}).Delete(&CiConfig{})
	if db.Error != nil {
		if gorm.IsRecordNotFoundError(db.Error) {
			return 0, errors.New("CiConfig not found")
		}
		return 0, db.Error
	}
	return db.RowsAffected, nil
}
