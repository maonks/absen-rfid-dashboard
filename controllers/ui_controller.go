package controllers

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func EditModal(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.Render("edit_modal", fiber.Map{
			"Uid": c.Params("uid"),
		})
	}
}
