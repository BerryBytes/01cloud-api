package seed

import (
	"01cloud-api/api/models"

	"github.com/jinzhu/gorm"

	log "github.com/sirupsen/logrus"
)

var users = []models.User{
	{
		FirstName:     "admin",
		LastName:      "admin",
		Email:         "admin@admin.com",
		Password:      "01cl0ud@2o2o",
		IsAdmin:       true,
		Active:        true,
		EmailVerified: true,
	},
}

var roles = []models.UserRole{
	{
		Code:   1,
		Name:   "admin",
		Active: true,
	},
	{
		Code:   2,
		Name:   "write",
		Active: true,
	},
	{
		Code:   3,
		Name:   "read",
		Active: true,
	},
}

func Load(db *gorm.DB) {
	for i := range users {
		err := db.Debug().Model(&models.User{}).FirstOrCreate(&users[i], &models.User{
			Email: users[i].Email,
		}).Error
		if err != nil {
			log.Fatalf("cannot seed users table: %v", err)
		}
	}

	for i := range roles {
		err := db.Debug().Model(&models.UserRole{}).FirstOrCreate(&roles[i], &models.UserRole{
			Code: roles[i].Code,
		}).Error
		if err != nil {
			log.Fatalf("cannot seed users table: %v", err)
		}
	}
}
