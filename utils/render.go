package utils

import "github.com/gofiber/fiber/v2"

func Render(
	c *fiber.Ctx,
	view string,
	data fiber.Map,
	layout ...string,
) error {

	if data == nil {
		data = fiber.Map{}
	}

	// inject user login ke semua halaman
	if user := c.Locals("auth_user"); user != nil {
		data["AuthUser"] = user
	}

	// jika layout dikirim
	if len(layout) > 0 {
		return c.Render(view, data, layout[0])
	}

	return c.Render(view, data)
}
