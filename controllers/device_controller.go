package controllers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"gorm.io/gorm"
)

func DeviceStatus(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var devices []models.Device
		db.Find(&devices)

		type Status struct {
			DeviceId string `json:"device_id"`
			Online   bool   `json:"online"`
		}

		now := time.Now()
		var res []Status

		for _, d := range devices {
			online := now.Sub(d.LastSeen) <= 30*time.Second
			res = append(res, Status{
				DeviceId: d.DeviceId,
				Online:   online,
			})
		}

		return c.Render("device_status", res)

	}
}
