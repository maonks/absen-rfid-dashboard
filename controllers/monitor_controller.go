package webcontroller

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	apicontroller "github.com/maonks/absen-rfid-backend/controllers/api_controller"
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

		// ================= SECURITY =================
		if !utils.CekHMAC(c.Body(), c.Get("X-Signature")) {
			return c.SendStatus(401)
		}

		var req ReqAbsen
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid payload")
		}

		log.Println("📢 ABSEN API HIT", req.Uid)

		// ================= KARTU =================
		var kartu models.Kartu
		if err := db.Where("uid = ?", req.Uid).First(&kartu).Error; err != nil {
			// kartu baru (belum di-assign ke siswa)
			kartu = models.Kartu{
				UID: req.Uid,
			}
			db.Create(&kartu)
		}

		// ================= PARSE WAKTU =================
		loc, _ := time.LoadLocation("Asia/Jakarta")
		parsedTime, err := time.ParseInLocation(
			"2006-01-02 15:04:05",
			req.Waktu,
			loc,
		)
		if err != nil {
			log.Println("❌ Gagal parse waktu:", err)
			return c.Status(400).SendString("Invalid time format")
		}

		// ================= SIMPAN ABSEN =================
		db.Create(&models.Absen{
			UID:      req.Uid,
			DeviceId: req.DeviceId,
			Waktu:    parsedTime,
		})

		// ================= UPDATE DEVICE =================
		db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "device_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"last_seen"}),
		}).Create(&models.Device{
			DeviceId: req.DeviceId,
			LastSeen: time.Now(),
		})

		// ================= CARI NAMA SISWA =================
		var nama *string

		if kartu.SiswaID != nil {
			var siswa models.Siswa
			if err := db.First(&siswa, *kartu.SiswaID).Error; err == nil {
				nama = &siswa.Nama
			}
		}

		// ================= BROADCAST REALTIME =================
		apicontroller.Broadcast(models.RealTime{
			UID:      req.Uid,
			Nama:     nama, // ✅ pointer (bisa nil)
			DeviceId: req.DeviceId,
			Waktu:    req.Waktu,
		})

		return c.SendStatus(200)
	}
}

func MonitorAbsen(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rows []models.RealTime

		db.Raw(`
			SELECT
			  a.uid,
			  s.nama,
			  a.device_id,
			  to_char(a.waktu, 'YYYY-MM-DD HH24:MI:SS') AS waktu
			FROM absens a
			LEFT JOIN kartus k ON k.uid = a.uid
			LEFT JOIN siswas s ON s.id = k.siswa_id
			ORDER BY a.waktu DESC
			LIMIT 50
		`).Scan(&rows)

		return utils.Render(c, "pages/monitor_page", fiber.Map{
			"Rows": rows,
		}, "layouts/main")
	}
}

func RealtimeAbsen(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rows []models.RealTime

		db.Raw(`
			SELECT
			  a.uid,
			  s.nama,
			  a.device_id,
			  to_char(a.waktu, 'YYYY-MM-DD HH24:MI:SS') AS waktu
			FROM absens a
			LEFT JOIN kartus k ON k.uid = a.uid
			LEFT JOIN siswas s ON s.id = k.siswa_id
			ORDER BY a.waktu DESC
			LIMIT 50
		`).Scan(&rows)

		return utils.Render(c, "pages/realtime_page", fiber.Map{
			"Rows": rows,
		})
	}
}
