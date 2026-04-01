package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
)

type Token struct {
	gorm.Model
	UserId     uint64     `gorm:"default:0" json:"user_id"`
	Name       string     `gorm:"size:255;not null;" json:"name,omitempty"`
	Token      string     `gorm:"size:255;not null;" json:"token,omitempty"`
	ExpiryDate *time.Time `gorm:"default:null;" json:"expiry_date,omitempty"`
}

type TokenInterface interface {
	CreateToken(db *gorm.DB, token *Token) (*Token, error)
	IsTokenExistsByName(db *gorm.DB, token *Token) bool
	FindAllByUserId(db *gorm.DB, uid uint64) (*[]Token, error)
	RevokeAllToken(db *gorm.DB, uid uint64) error
	RevokeSingleToken(db *gorm.DB, tokenId uint) error
	GetToken(db *gorm.DB, token string) (*Token, error)
}

type TokenType struct{}

func NewToken() TokenInterface {
	return &TokenType{}
}
func (t *TokenType) CreateToken(db *gorm.DB, token *Token) (*Token, error) {
	err = db.Model(&Token{}).Create(&token).Error
	if err != nil {
		return &Token{}, err
	}
	return token, nil
}
func (t *TokenType) GetToken(db *gorm.DB, token string) (*Token, error) {
	data := Token{}
	err = db.Model(&Token{}).Where("token=?", token).Take(&data).Error
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// IsTokenExistsByName This method checks duplicate entries while creating the Token according to user id.
func (t *TokenType) IsTokenExistsByName(db *gorm.DB, token *Token) bool {
	count := 0
	db.Model(&Token{}).Where("name=? and user_id=?", token.Name, token.UserId).Count(&count)
	return count > 0
}

func (t *TokenType) RevokeAllToken(db *gorm.DB, uid uint64) error {
	err := db.Where("user_id = ?", uid).Delete(&Token{}).Error
	if err != nil {
		return errors.New("token not found")
	}
	return nil
}

func (t *TokenType) RevokeSingleToken(db *gorm.DB, tokenId uint) error {
	err := db.Where("id = ?", tokenId).Delete(&Token{}).Error
	if err != nil {
		return errors.New("token not found")
	}
	return nil
}

func (t *TokenType) FindAllByUserId(db *gorm.DB, uid uint64) (*[]Token, error) {
	var token []Token
	err := db.Model(&Token{}).Where("user_id = ?", uid).Order("created_at DESC").Find(&token).Error
	if err != nil {
		return &[]Token{}, errors.New("token not found")
	}
	return &token, nil
}
