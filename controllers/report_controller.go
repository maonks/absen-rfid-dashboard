package webcontroller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"github.com/maonks/absen-rfid-backend/utils"
	"gorm.io/gorm"
)

func ReportPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var kelas []models.Kelas
		db.Find(&kelas)

		return utils.Render(c, "pages/report_page", fiber.Map{
			"kelas": kelas,
		}, "layouts/main")
	}
}
