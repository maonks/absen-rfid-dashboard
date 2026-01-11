package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"gorm.io/gorm"
)

func LoadUser(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		userID := c.Locals("user_id")
		if userID == nil {
			return c.Next()
		}

		var user models.User
		if err := db.First(&user, userID).Error; err == nil {
			c.Locals("auth_user", user)
		}

		return c.Next()
	}
}
