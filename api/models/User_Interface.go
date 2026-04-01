package models

import "github.com/jinzhu/gorm"

type UserInterface interface {
	SaveUser(db *gorm.DB, data *User) (*User, error)
	FindAllUsers(db *gorm.DB) (*[]User, error)
	FindUserByID(db *gorm.DB, uid uint) (*User, error)
	FindUserByEmail(db *gorm.DB, email string) (*User, error)
	FindAllUsersWithFilters(db *gorm.DB, page uint64, size uint64, q string, sortColumn string, sortDirection string) (*[]User, int, error)
	UpdateAUser(db *gorm.DB, uid uint, data *User) (*User, error)
	VerifyEmail(db *gorm.DB, uid uint) (*User, error)
	UpdatePassword(db *gorm.DB, u *User) error
	DeleteAUser(db *gorm.DB, uid uint) (int64, error)
	UpdateUserQuota(db *gorm.DB, data *User) (*User, error)
	SaveSession(db *gorm.DB, data *Session) (*Session, error)
	GetAllSessionByUserId(db *gorm.DB, userid uint, limit, offset int) (*[]Session, int, error)
	GetSessionById(db *gorm.DB, uid uint) (*Session, error)
	UpdateSession(db *gorm.DB, data Session) error
	GetActiveSessionByUserId(db *gorm.DB, userid uint) (*[]Session, int, error)
}
type UserType struct {
}

func NewUser() UserInterface {
	return &UserType{}
}
