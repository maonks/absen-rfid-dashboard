package apicontroller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/services"
	"github.com/maonks/absen-rfid-backend/utils"
	"gorm.io/gorm"
)

type ReqLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(db *gorm.DB) fiber.Handler {

	var cekUsername = services.CekUsername(db)

	return func(c *fiber.Ctx) error {

		var req ReqLogin
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Request tidak valid")
		}

		user, err := cekUsername(req.Username)
		if err != nil {
			return c.Status(401).SendString("Username tidak ditemukan")
		}

		if !utils.CekPassword(user.Password, req.Password) {
			return c.Status(401).SendString("Password salah")
		}

		token, err := utils.BuatJWT(user.ID)
		if err != nil {
			return c.Status(500).SendString("Gagal membuat token")
		}

		// ✅ SET COOKIE (HTTP ONLY)
		c.Cookie(&fiber.Cookie{
			Name:     "access_token",
			Value:    token,
			HTTPOnly: true,
			Secure:   false, // true jika HTTPS
			SameSite: "Lax",
			Expires:  time.Now().Add(30 * time.Minute),
			Path:     "/",
		})

		// 🔁 UNTUK HTMX
		if c.Get("HX-Request") == "true" {
			c.Set("HX-Redirect", "/")
			return c.SendStatus(204)
		}

		return c.Redirect("/")
	}
}

func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token", // ⬅️ HARUS SAMA
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   false, // true jika HTTPS
	})

	return c.SendStatus(fiber.StatusOK)
}
