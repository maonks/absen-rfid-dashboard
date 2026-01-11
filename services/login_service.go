package services

import (
	"github.com/maonks/absen-rfid-backend/models"
	"gorm.io/gorm"
)

func CekUsername(db *gorm.DB) func(user string) (models.User, error) {
	return func(user string) (models.User, error) {

		var userx models.User

		err := db.Where("username=?", user).First(&userx).Error

		return userx, err
	}
}
