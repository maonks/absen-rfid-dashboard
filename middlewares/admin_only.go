package middlewares

import (
	"github.com/gofiber/fiber/v2"
)

func AdminOnly() fiber.Handler {
	return func(c *fiber.Ctx) error {

		// role sudah diset di Locals oleh middleware JWT
		role, ok := c.Locals("role").(string)

		if !ok || role != "admin" {
			return c.Status(fiber.StatusForbidden).SendString("Akses ditolak")
		}

		return c.Next()
	}
}
