package apicontroller

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Login(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("layouts/login_page", nil)
	}
}
