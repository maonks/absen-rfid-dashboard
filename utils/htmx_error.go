package utils

import "github.com/gofiber/fiber/v2"

func HtmxError(c *fiber.Ctx, msg string) error {
	return c.SendString(`
	<div class="text-red-600 text-sm mt-2">
		` + msg + `
	</div>
	`)
}
