package models

import (
	"errors"
)

type ExternalLogin struct {
	Service   string `json:"service_name"`
	Code      string `json:"service_code"`
	User      string `json:"service_user"`
	Url       string `json:"service_url"`
	IsOauth   bool   `json:"is_oauth"`
	AccessId  string `json:"access_id"`
	SecretKey string `json:"secret_key"`
	Region    string `json:"region"`
}

func (u *ExternalLogin) Validate() error {
	if u.Code == "" {
		return errors.New("required service code")
	}
	if u.Service == "" {
		return errors.New("required service name")
	}
	return nil
}
