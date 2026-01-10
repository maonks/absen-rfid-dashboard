package webcontroller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"gorm.io/gorm"
)

func SiswaPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var siswa []models.Siswa
		db.Find(&siswa)

		return c.Render("pages/siswa_page", fiber.Map{
			"Siswa": siswa,
		}, "layouts/main")
	}
}

// // GET /siswa/create
func CreateSiswa(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var kartuKosong []models.Kartu

		db.Where("siswa_id IS NULL").Find(&kartuKosong)

		return c.Render("components/tambah_siswa", fiber.Map{
			"KartuKosong": kartuKosong,
		})
	}
}

// POST /siswa/store
func StoreSiswa(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		tgl, _ := time.Parse("2006-01-02", c.FormValue("tanggal_lahir"))

		siswa := models.Siswa{
			NIS:          c.FormValue("nis"),
			Nama:         c.FormValue("nama"),
			JenisKelamin: c.FormValue("jenis_kelamin"),
			TempatLahir:  c.FormValue("tempat_lahir"),
			TanggalLahir: tgl,
			Kelas:        c.FormValue("kelas"),
			Jurusan:      c.FormValue("jurusan"),
			Alamat:       c.FormValue("alamat"),
			NamaWali:     c.FormValue("nama_wali"),
			NoHP:         c.FormValue("no_hp"),
			Status:       c.FormValue("status"),
		}

		if err := db.Create(&siswa).Error; err != nil {
			return c.Status(400).SendString("Gagal menyimpan siswa")
		}

		// 🔥 jika kartu dipilih → update kartu
		kartuID := c.FormValue("kartu_id")
		if kartuID != "" {
			db.Model(&models.Kartu{}).
				Where("id = ?", kartuID).
				Update("siswa_id", siswa.ID)
		}
		// Tutup modal & reload halaman siswa
		return c.SendString(`
		<script>
			closeModal();
			location.reload();
		</script>
	`)
	}
}

// GET /siswa/:id/edit
func EditSiswa(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var siswa models.Siswa
		if err := db.Preload("Kartu").First(&siswa, id).Error; err != nil {
			return c.Status(404).SendString("Siswa tidak ditemukan")
		}

		// ambil kartu FREE
		var kartuFree []models.Kartu
		db.Where("siswa_id IS NULL").Find(&kartuFree)

		// gabungkan: kartu aktif + kartu free
		var kartuPilihan []models.Kartu

		if siswa.Kartu != nil {
			kartuPilihan = append(kartuPilihan, *siswa.Kartu)
		}
		kartuPilihan = append(kartuPilihan, kartuFree...)

		return c.Render("components/edit_siswa", fiber.Map{
			"Siswa":        siswa,
			"KartuPilihan": kartuPilihan,
			"KartuAktif":   siswa.Kartu,
		})
	}
}

// POST /siswa/:id/update
func UpdateSiswa(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var siswa models.Siswa
		if err := db.Preload("Kartu").First(&siswa, id).Error; err != nil {
			return c.Status(404).SendString("Siswa tidak ditemukan")
		}

		kartuID := c.FormValue("kartu_id")

		// 1️⃣ lepaskan kartu lama
		if siswa.Kartu != nil {
			db.Model(&models.Kartu{}).
				Where("id = ?", siswa.Kartu.ID).
				Update("siswa_id", nil)
		}

		// 2️⃣ assign kartu baru (jika ada)
		if kartuID != "" {
			db.Model(&models.Kartu{}).
				Where("id = ?", kartuID).
				Update("siswa_id", siswa.ID)
		}

		// 3️⃣ update data siswa
		db.Model(&siswa).Updates(map[string]interface{}{
			"nis":           c.FormValue("nis"),
			"nama":          c.FormValue("nama"),
			"jenis_kelamin": c.FormValue("jenis_kelamin"),
			"kelas":         c.FormValue("kelas"),
			"jurusan":       c.FormValue("jurusan"),
			"alamat":        c.FormValue("alamat"),
			"nama_wali":     c.FormValue("nama_wali"),
			"no_hp":         c.FormValue("no_hp"),
			"status":        c.FormValue("status"),
		})

		return c.SendString(`
			<div class="p-4 text-center text-green-600">
				✅ Data siswa berhasil diperbarui
			</div>
			<script>
				setTimeout(() => location.reload(), 800)
			</script>
		`)
	}
}
