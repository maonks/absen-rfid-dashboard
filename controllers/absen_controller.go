package controllers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"github.com/maonks/absen-rfid-backend/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreateAbsen(db *gorm.DB) fiber.Handler {

	type ReqAbsen struct {
		Uid      string `json:"uid"`
		DeviceId string `json:"device_id"`
		Waktu    string `json:"waktu"`
	}

	return func(c *fiber.Ctx) error {

		if !utils.CekHMAC(c.Body(), c.Get("X-Signature")) {
			return c.SendStatus(401)
		}

		var req ReqAbsen
		c.BodyParser(&req)

		log.Println("📢 ABSEN API HIT")

		var kartu models.Kartu
		nama := "PENDING"

		if err := db.Where("uid=?", req.Uid).First(&kartu).Error; err != nil {
			db.Create(&models.Kartu{
				Uid:  req.Uid,
				Nama: nama,
			})
		} else {
			nama = kartu.Nama
		}

		// PARSE STRING WAKTU → time.Time
		parsedTime, err := time.Parse(
			"2006-01-02 15:04:05",
			req.Waktu,
		)

		if err != nil {
			log.Println("❌ Gagal parse waktu:", err)
			return c.Status(400).SendString("Invalid time format")
		}

		db.Create(&models.Absen{
			Uid:      req.Uid,
			DeviceId: req.DeviceId,
			Waktu:    parsedTime, // ✅ SUDAH time.Time
		})

		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "device_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_seen"}),
		}).Create(&models.Device{
			DeviceId: req.DeviceId,
			LastSeen: time.Now(),
		})

		Broadcast(models.RealTime{
			Uid:      req.Uid,
			Nama:     nama,
			DeviceId: req.DeviceId,
			Waktu:    req.Waktu,
		})

		return c.SendStatus(200)
	}
}

func SearchAbsen(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rows []models.RealTime

		db.Raw(`
		SELECT
			absens.uid,
			kartus.nama,
			absens.device_id,
			absens.waktu
		FROM absens
		LEFT JOIN kartus ON kartus.uid = absens.uid
		ORDER BY absens.waktu DESC
		LIMIT 50
	`).Scan(&rows)

		return c.Render("table", fiber.Map{
			"Rows": rows,
		})
	}

}
