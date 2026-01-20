package webcontroller

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"github.com/maonks/absen-rfid-backend/utils"
	"gorm.io/gorm"
)

func SiswaPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var siswa []models.Siswa
		db.Find(&siswa)

		return utils.Render(c, "pages/siswa_page", fiber.Map{
			"Siswa": siswa,
		}, "layouts/main")
	}
}

// GET /siswa/create
func CreateSiswa(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var kartuKosong []models.Kartu
		var Kelas []models.Kelas
		var Jurusan []models.Jurusan

		db.Where("siswa_id IS NULL").Find(&kartuKosong)
		db.Order("nama asc").Find(&Kelas)
		db.Order("nama asc").Find(&Jurusan)

		return utils.Render(c, "modals/tambah_siswa", fiber.Map{
			"KartuKosong": kartuKosong,
			"Kelas":       Kelas,
			"Jurusan":     Jurusan,
		})
	}
}

// POST /siswa/store
func StoreSiswa(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		tgl, _ := time.Parse("2006-01-02", c.FormValue("tanggal_lahir"))

		kelasStr := c.FormValue("kelas")
		jurusanStr := c.FormValue("jurusan")

		kelasID, err := strconv.ParseUint(kelasStr, 10, 64)

		if err != nil {
			return c.Status(400).SendString("Kelas Salah")
		}

		jurusanID, err := strconv.ParseUint(jurusanStr, 10, 64)

		if err != nil {
			return c.Status(400).SendString("Jurusan Salah")
		}

		siswa := models.Siswa{
			NIS:          c.FormValue("nis"),
			Nama:         c.FormValue("nama"),
			JenisKelamin: c.FormValue("jenis_kelamin"),
			TempatLahir:  c.FormValue("tempat_lahir"),
			TanggalLahir: tgl,
			KelasID:      uint(kelasID),
			JurusanID:    uint(jurusanID),
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

		return utils.Render(c, "modals/edit_siswa", fiber.Map{
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

// GET /kelas/create
func CreateKelas(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var kelas []models.Kelas

		db.Order("nama asc").Find(&kelas)

		return utils.Render(c, "modals/tambah_kelas", fiber.Map{

			"Kelas": kelas,
		})
	}
}

// POST /kelas/store
func StoreKelas(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		nama := c.FormValue("nama")
		if nama == "" {
			return c.SendStatus(400)
		}

		db.Create(&models.Kelas{
			Nama: nama,
		})

		var kelas []models.Kelas
		db.Order("nama asc").Find(&kelas)

		// RETURN HANYA LIST, BUKAN MODAL
		return utils.Render(c, "partials/kelas_list", fiber.Map{
			"Kelas": kelas,
		})
	}
}

func KelasRow(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var kelas models.Kelas
		db.First(&kelas, c.Params("id"))

		return utils.Render(c, "partials/kelas_row", fiber.Map{
			"Kelas": kelas,
			"Index": 0,
		})
	}
}

func KelasEditRow(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var kelas models.Kelas
		db.First(&kelas, c.Params("id"))

		return utils.Render(c, "partials/kelas_row_edit", fiber.Map{
			"Kelas": kelas,
		})
	}
}

func KelasUpdate(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		db.Model(&models.Kelas{}).
			Where("id = ?", c.Params("id")).
			Update("nama", c.FormValue("nama"))

		var kelas models.Kelas
		db.First(&kelas, c.Params("id"))

		return utils.Render(c, "partials/kelas_row", fiber.Map{
			"Kelas": kelas,
			"Index": 0,
		})
	}
}

func KelasDelete(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		db.Delete(&models.Kelas{}, c.Params("id"))
		return c.SendString("")
	}
}

// GET /kelas/create
func CreateJurusan(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var jurusan []models.Jurusan

		db.Order("nama asc").Find(&jurusan)

		return utils.Render(c, "modals/tambah_jurusan", fiber.Map{

			"Jurusan": jurusan,
		})
	}
}

// POST /kelas/store
func StoreJurusan(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		nama := c.FormValue("nama")
		if nama == "" {
			return c.SendStatus(400)
		}

		db.Create(&models.Jurusan{
			Nama: nama,
		})

		var jurusan []models.Jurusan
		db.Order("nama asc").Find(&jurusan)

		// RETURN HANYA LIST, BUKAN MODAL
		return utils.Render(c, "partials/jurusan_list", fiber.Map{
			"Jurusan": jurusan,
		})
	}
}

func JurusanRow(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var jurusan models.Jurusan
		db.First(&jurusan, c.Params("id"))

		return utils.Render(c, "partials/jurusan_row", fiber.Map{
			"Jurusan": jurusan,
			"Index":   0,
		})
	}
}

func JurusanEditRow(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var jurusan models.Jurusan
		db.First(&jurusan, c.Params("id"))

		return utils.Render(c, "partials/jurusan_row_edit", fiber.Map{
			"Jurusan": jurusan,
		})
	}
}

func JurusanUpdate(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		db.Model(&models.Jurusan{}).
			Where("id = ?", c.Params("id")).
			Update("nama", c.FormValue("nama"))

		var jurusan models.Jurusan
		db.First(&jurusan, c.Params("id"))

		return utils.Render(c, "partials/jurusan_row", fiber.Map{
			"Jurusan": jurusan,
			"Index":   0,
		})
	}
}

func JurusanDelete(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		db.Delete(&models.Jurusan{}, c.Params("id"))
		return c.SendString("")
	}
}
