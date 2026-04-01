package models

import (
	"encoding/json"
	"errors"
	"html"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"

	"github.com/robfig/cron/v3"
)

type CronJob struct {
	gorm.Model
	EnvironmentID              uint
	Image                      string         `gorm:"size:255;not null;" json:"image"`
	Name                       string         `gorm:"size:255;not null;" json:"name"`
	RestartPolicy              *string        `gorm:"size:255;null;" json:"restart_policy"`
	ConcurrentPolicy           *string        `gorm:"size:255;null;" json:"concurrent_policy"`
	Command                    string         `gorm:"not null;" json:"command"`
	Labels                     postgres.Jsonb `json:"labels"`
	Schedule                   string         `gorm:"size:255;not null;" json:"schedule"`
	StartingDeadlineSeconds    *int64         `gorm:"null;" json:"starting_deadline_seconds"`
	FailedJobsHistoryLimit     *int32         `gorm:"null;" json:"failed_jobs_history_limit"`
	SuccessfulJobsHistoryLimit *int32         `gorm:"null;" json:"successful_job_history_limit"`
	Suspend                    *bool          `gorm:"default false;" json:"suspend"`
	User                       *User          `gorm:"foreignkey:UserID;null" json:"user"`
	UserID                     uint64         `gorm:"default:0;" json:"user_id"`
}

type CronJobInterface interface {
	SaveCronJob(db *gorm.DB, data *CronJob, pid uint64) (*CronJob, error)
	Update(db *gorm.DB, data *CronJob) (*CronJob, error)
	FindAllWithFilters(db *gorm.DB, eid uint, page uint64, size uint64) (*[]CronJob, error)
	Find(db *gorm.DB, pid uint64) (*CronJob, error)
}
type CronJobType struct {
}

func NewCronJob() CronJobInterface {
	return &CronJobType{}
}

func (data *CronJob) VerifyResource(received *Environment) error {
	if uint64(len(received.CronJob)) < received.Application.Project.Subscription.CronJob {
		for _, d := range received.CronJob {
			if d.Name == data.Name {
				return errors.New("name already exists")
			}
		}
		return nil
	}
	return errors.New("cronJob quota limit exceed")
}

func (cron *CronJobType) SaveCronJob(db *gorm.DB, data *CronJob, pid uint64) (*CronJob, error) {
	var err error
	data.EnvironmentID = uint(pid)
	err = db.Model(&CronJob{}).Save(&data).Error
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (data *CronJob) Prepare() {
	successLimit := int32(1)
	failedLimit := int32(2)
	startingDeadlineSeconds := int64(60)
	data.Name = html.EscapeString(strings.TrimSpace(data.Name))
	d, _ := json.Marshal(map[string]string{"io.kubernetes.job/name": data.Name, "app.kubernetes.io/provider": "zerone"})
	data.Labels = postgres.Jsonb{RawMessage: d}
	data.SuccessfulJobsHistoryLimit = &successLimit
	if data.StartingDeadlineSeconds == nil {
		data.StartingDeadlineSeconds = &startingDeadlineSeconds
	}
	data.FailedJobsHistoryLimit = &failedLimit
}

func (data *CronJob) Validate() error {
	if data.Name == "" {
		return errors.New("required name")
	}
	if !ValidateName(data.Name) {
		return errors.New("allowed alphanumeric, underscore, hyphen and space only")
	}
	if data.Command == "" {
		return errors.New("required command")
	}
	_, err := cron.ParseStandard(data.Schedule)
	if err != nil {
		return err
	}
	return nil
}

func (cron *CronJobType) Update(db *gorm.DB, data *CronJob) (*CronJob, error) {
	err := db.Model(&data).Update(&data).Take(&data).Error
	if err != nil {
		return &CronJob{}, err
	}
	return data, nil
}

func (cron *CronJobType) FindAllWithFilters(db *gorm.DB, eid uint, page uint64, size uint64) (*[]CronJob, error) {
	dataList := []CronJob{}
	err := db.Model(&CronJob{}).Where(&CronJob{EnvironmentID: eid}).Limit(size).Offset(size * (page - 1)).Find(&dataList).Error
	if err != nil {
		return &[]CronJob{}, err
	}
	return &dataList, nil
}

func (cron *CronJobType) Find(db *gorm.DB, pid uint64) (*CronJob, error) {
	var err error
	data := &CronJob{}
	print(pid)
	err = db.Model(&CronJob{}).
		Where("id = ?", pid).
		Preload("User").
		Take(&data).Error
	if err != nil {
		return &CronJob{}, err
	}
	return data, nil
}
