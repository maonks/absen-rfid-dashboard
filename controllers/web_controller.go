package controllers

import "github.com/gofiber/fiber/v2"

func Dashboard(c *fiber.Ctx) error {
	return c.Render("dashboard", fiber.Map{}, "layout")
}
