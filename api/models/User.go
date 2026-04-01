package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	gorm.Model
	FirstName      string         `gorm:"size:255;not null;" json:"first_name,omitempty"`
	LastName       string         `gorm:"size:255;not null;" json:"last_name"`
	Email          string         `gorm:"size:255;not null;unique" json:"email,omitempty"`
	Image          string         `gorm:"size:255;null;" json:"image,omitempty"`
	Company        string         `gorm:"size:255;null;" json:"company,omitempty"`
	Designation    string         `gorm:"size:255;null;" json:"designation,omitempty"`
	Password       string         `gorm:"size:100;not null;" json:"password"`
	EmailVerified  bool           `gorm:"not null;default:false" json:"email_verified"`
	Active         bool           `gorm:"not null;default:true" json:"active"`
	IsAdmin        bool           `gorm:"not null;default:false" json:"is_admin,omitempty"`
	AddressUpdated bool           `gorm:"not null;default:false" json:"address_updated,omitempty"`
	Quotas         postgres.Jsonb `gorm:"null" json:"quotas"`
	UsedDemo       bool           `gorm:"default:false" json:"used_demo"`
	Reference      string         `gorm:"null" json:"reference"`
}
type Quotas struct {
	UserProject      int `json:"user_project"`
	UserOrganization int `json:"user_organization"`
}
type Config struct {
	Quotas               Quotas `json:"quotas"`
	PaymentThresholdDays int    `json:"payment_threshold_days"`
	ProjectThresholdDays int    `json:"project_threshold_days"`
}

type Session struct {
	gorm.Model
	UserId     uint   `json:"user_id"`
	Active     bool   `gorm:"null;default:false" json:"active"`
	Action     string `json:"action"` // login/logout/logout-all/deactivate
	DeviceName string `json:"device_name"`
	OS         string `json:"os"`
	IP         string `json:"ip"`
	Browser    string `json:"browser"`
	Location   string `json:"location"`
}

type Campaign struct {
	gorm.Model
	Name           string `json:"name"`
	User           *User  `json:"user"`
	PluginId       uint   `json:"plugin_id"`
	SubscriptionId uint   `json:"subscription_id"`
	OrganizationId uint   `json:"organization_id"`
	Reference      string `json:"reference"`
	CampaignType   string `json:"campaign_type"`
}

func (data *User) ToJson() (map[string]interface{}, error) {
	msg, _ := json.Marshal(data)
	res := map[string]interface{}{}
	err = json.Unmarshal(msg, &res)
	return res, err
}

func Hash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}

func isHashed(password string) bool {
	return len(password) == 60
}

func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (u *User) BeforeSave() error {
	if !isHashed(u.Password) {
		hashedPassword, err := Hash(u.Password)
		if err != nil {
			return err
		}
		u.Password = string(hashedPassword)
	}
	return nil
}

func (u *User) Prepare() {
	u.ID = 0
	u.FirstName = html.EscapeString(strings.TrimSpace(u.FirstName))
	u.LastName = html.EscapeString(strings.TrimSpace(u.LastName))
	u.Company = html.EscapeString(strings.TrimSpace(u.Company))
	u.Designation = html.EscapeString(strings.TrimSpace(u.Designation))
	u.Email = html.EscapeString(strings.TrimSpace(u.Email))
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	u.IsAdmin = false
}

func (u *User) Validate(action string) error {
	switch strings.ToLower(action) {
	case "update":
		if u.FirstName == "" {
			return errors.New("required FirstName")
		}
		if u.LastName == "" {
			return errors.New("required LastName")
		}
		if u.Password == "" {
			return errors.New("required Password")
		}
		if u.Email == "" {
			return errors.New("required Email")
		}
		if err := CheckEmail(u.Email); err != nil {
			return errors.New("invalid email")
		}
		if !ValidateName(u.FirstName) {
			return errors.New("allowed alphanumeric, underscore, hyphen and space only")
		}
		return nil
	case "login":
		if u.Password == "" {
			return errors.New("required Password")
		}
		if u.Email == "" {
			return errors.New("required Email")
		}
		if err := CheckEmail(u.Email); err != nil {
			return errors.New("invalid email")
		}
		return nil
	case "forgotpassword":
		if u.Email == "" {
			return errors.New("required Email")
		}
		if u.Email != "" {
			if CheckEmail(u.Email) != nil {
				return errors.New("invalid email")
			}
		}
		return nil
	default:
		if u.FirstName == "" {
			return errors.New("required FirstName")
		}
		if u.LastName == "" {
			return errors.New("required LastName")
		}
		if u.Password == "" {
			return errors.New("required Password")
		}
		if u.Email == "" {
			return errors.New("required Email")
		}
		if err := CheckEmail(u.Email); err != nil {
			return errors.New("invalid email")
		}
		return nil
	}
}

func (d *UserType) SaveUser(db *gorm.DB, u *User) (*User, error) {
	err := db.Create(&u).Error
	if err != nil {
		return &User{}, err
	}
	return u, nil
}

func (d *UserType) FindAllUsers(db *gorm.DB) (*[]User, error) {
	var err error
	users := []User{}
	err = db.Model(&User{}).Order("id desc").Find(&users).Error
	if err != nil {
		return &[]User{}, err
	}
	return &users, err
}

func (d *UserType) FindAllUsersWithFilters(db *gorm.DB, page uint64, size uint64, q string, sortColumn string, sortDirection string) (*[]User, int, error) {
	var err error
	users := []User{}
	count := 0
	err = db.Model(&User{}).Where("first_name like ? OR last_name like ? OR email like ? OR company like ?",
		"%"+q+"%", "%"+q+"%", "%"+q+"%", "%"+q+"%").
		Order(sortColumn + " " + sortDirection).Limit(size).Offset(size * (page - 1)).Find(&users).Error
	if err != nil {
		return &[]User{}, 0, err
	}
	db.Model(&User{}).Where("first_name like ? OR last_name like ? OR email like ? OR company like ?",
		"%"+q+"%", "%"+q+"%", "%"+q+"%", "%"+q+"%").Count(&count)
	return &users, count, err
}

func (d *UserType) FindUserByID(db *gorm.DB, uid uint) (*User, error) {
	u := User{}
	err := db.Model(User{}).Where("id = ?", uid).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User Not Found")
	}
	return &u, err
}

func (d *UserType) FindUserByEmail(db *gorm.DB, email string) (*User, error) {
	u := User{}
	err := db.Model(User{}).Where("email = ?", email).Take(&u).Error
	if err != nil {
		return &User{}, err
	}
	if gorm.IsRecordNotFoundError(err) {
		return &User{}, errors.New("User Not Found")
	}
	return &u, err
}

func (d *UserType) UpdateAUser(db *gorm.DB, uid uint, data *User) (*User, error) {
	var err error
	var app = User{
		Active:   data.Active,
		UsedDemo: data.UsedDemo,
	}
	if data.FirstName != "" {
		app.FirstName = data.FirstName
	}
	if data.LastName != "" {
		app.LastName = data.LastName
	}
	if data.Company != "" {
		app.Company = data.Company
	}
	if data.Designation != "" {
		app.Designation = data.Designation
	}
	if data.Image != "" {
		app.Image = data.Image
	}
	if data.Reference != "" {
		app.Reference = data.Reference
	}
	if data.Password != "" {
		err := data.BeforeSave()
		if err != nil {
			log.Warn(err)
		}
		app.Password = data.Password
	}
	err = db.Model(&User{}).Where("id = ?", uid).Updates(app).Error
	if err != nil {
		return &User{}, err
	}
	return data, nil
}
func (d *UserType) UpdateUserQuota(db *gorm.DB, data *User) (*User, error) {
	mp := map[string]interface{}{}
	if string(data.Quotas.RawMessage) != "" {
		mp["quotas"] = data.Quotas
	}
	err = db.Model(&User{}).Where("id = ?", data.ID).UpdateColumns(mp).Error
	if err != nil {
		return &User{}, err
	}
	return data, nil

}
func (d *UserType) VerifyEmail(db *gorm.DB, uid uint) (*User, error) {
	data := User{}
	var err error
	err = db.Model(User{}).Where("id = ?", uid).Take(&data).Error
	if err != nil {
		return &data, err
	}
	if data.EmailVerified {
		return &data, errors.New("email already verified")
	}
	var app = User{Active: true, EmailVerified: true}
	err = db.Model(&User{}).Where("id = ?", uid).Updates(app).Error
	if err != nil {
		return &User{}, err
	}
	return &app, nil
}

//func (u *User) UpdateAUser(db *gorm.DB, uid uint) (*User, error) {
//	// To hash the password
//	err := u.BeforeSave()
//	if err != nil {
//		log.Warnf(err)
//	}
//	db = db.Model(&User{}).Where("id = ?", uid).Take(&User{}).UpdateColumns(
//		map[string]interface{}{
//			"password":    u.Password,
//			"first_name":   u.FirstName,
//			"last_name":    u.LastName,
//			"company":     u.Company,
//			"designation": u.Designation,
//			"email":       u.Email,
//			"update_at":   time.Now(),
//		},
//	)
//	if db.Error != nil {
//		return &User{}, db.Error
//	}
//	// This is the display the updated user
//	err = db.Model(&User{}).Where("id = ?", uid).Take(&u).Error
//	if err != nil {
//		return &User{}, err
//	}
//	return u, nil
//}

func (d *UserType) DeleteAUser(db *gorm.DB, uid uint) (int64, error) {

	db = db.Model(&User{}).Where("id = ?", uid).Take(&User{}).Delete(&User{})

	if db.Error != nil {
		return 0, db.Error
	}
	return db.RowsAffected, nil
}

func (d *UserType) UpdatePassword(db *gorm.DB, u *User) error {

	err := u.BeforeSave()
	if err != nil {
		log.Warn(err)
	}

	db = db.Model(&User{}).Where("email = ?", u.Email).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"password":       u.Password,
			"email_verified": true,
		},
	)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func (u *User) ChangeActive(db *gorm.DB, active bool) error {
	db = db.Model(&User{}).Where("email = ?", u.Email).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"active": active,
		},
	)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func (u *User) ChangeAdmin(db *gorm.DB, isAdmin bool) error {
	db = db.Model(&User{}).Where("email = ?", u.Email).Take(&User{}).UpdateColumns(
		map[string]interface{}{
			"is_admin": isAdmin,
		},
	)
	if db.Error != nil {
		return db.Error
	}
	return nil
}

func CheckEmail(e string) error {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(e) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

func (d *UserType) SaveSession(db *gorm.DB, u *Session) (*Session, error) {
	err := db.Model(&Session{}).Create(&u).Error
	if err != nil {
		return &Session{}, err
	}
	return u, nil
}

func (d *UserType) GetSessionById(db *gorm.DB, id uint) (*Session, error) {
	data := Session{}
	err = db.Model(&Session{}).Where("id = ? and active = ?", id, true).Find(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func (d *UserType) GetAllSessionByUserId(db *gorm.DB, userid uint, limit, offset int) (*[]Session, int, error) {
	count := 0
	datas := []Session{}
	year := time.Now().AddDate(-1, 0, 0)
	DB := db.Model(&Session{}).Where("user_id = ? and created_at >= ?", userid, year)
	err := DB.Order("id desc").Limit(limit).Offset(offset).Find(&datas).Error
	if err != nil {
		return nil, 0, err
	}
	DB.Count(&count)
	return &datas, count, nil
}

func (d *UserType) GetActiveSessionByUserId(db *gorm.DB, userid uint) (*[]Session, int, error) {
	count := 0
	datas := []Session{}
	DB := db.Model(&Session{}).Where("user_id = ? and active = ?", userid, true)
	err := DB.Order("id desc").Find(&datas).Error
	if err != nil {
		return nil, 0, err
	}
	DB.Count(&count)
	return &datas, count, nil
}

func (d *UserType) UpdateSession(db *gorm.DB, data Session) error {
	var err error
	sessionData := map[string]interface{}{
		"active": data.Active,
	}
	if data.Action != "" {
		sessionData["action"] = data.Action
	}
	if data.DeviceName != "" {
		sessionData["device_name"] = data.DeviceName
	}
	if data.UserId != 0 {
		sessionData["user_id"] = data.UserId
	}
	if data.Action != "" {
		sessionData["action"] = data.Action
	}
	if data.OS != "" {
		sessionData["os"] = data.OS
	}
	if data.Browser != "" {
		sessionData["browser"] = data.Browser
	}
	if data.Location != "" {
		sessionData["location"] = data.Location
	}
	err = db.Model(&Session{}).Where("id = ?", data.ID).Updates(sessionData).Error
	if err != nil {
		return err
	}
	return nil
}
