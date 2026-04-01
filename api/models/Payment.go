package models

import (
	"github.com/jinzhu/gorm"
)

type PaymentHistory struct {
	gorm.Model
	Balance float64 `gorm:"default:0" json:"balance"`
}
type PaymentInterface interface {
	HasUserBalance(db *gorm.DB, uid uint) bool
	GetUserBalance(db *gorm.DB, uid uint) float64
}
type PaymentRepo struct {
}

func NewPayment() PaymentInterface {
	return &PaymentRepo{}
}

func (d *PaymentRepo) HasUserBalance(db *gorm.DB, uid uint) bool {
	data := PaymentHistory{}
	db.Model(&PaymentHistory{}).Where("user_id = ?", uid).Order("id desc").Take(&data)
	return int64(data.Balance) >= 0
}

func (d *PaymentRepo) GetUserBalance(db *gorm.DB, uid uint) float64 {
	data := PaymentHistory{}
	err = db.Model(&PaymentHistory{}).Where("user_id = ?", uid).Order("id desc").Take(&data).Error
	if err != nil {
		return 0
	}
	return data.Balance
}
