package webcontroller

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"github.com/maonks/absen-rfid-backend/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type SiswaDetailRow struct {
	ID           uint
	NIS          string
	Nama         string
	JenisKelamin string
	TempatLahir  string
	TanggalLahir time.Time
	Alamat       string
	NamaWali     string
	NoHP         string
	Status       string
	Kelas        string
	Jurusan      string
}

func ReportDetailSiswaPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		now := time.Now()

		kelasID := c.QueryInt("kelas_id")
		jurusanID := c.QueryInt("jurusan_id")
		bulan := c.QueryInt("bulan", int(now.Month()))
		tahun := c.QueryInt("tahun", now.Year())

		var rows []SiswaDetailRow

		query := `
		SELECT 
		  s.id,
		  s.nis,
		  s.nama,
		  s.jenis_kelamin,
		  s.tempat_lahir,
		  s.tanggal_lahir,
		  s.alamat,
		  s.nama_wali,
		  s.no_hp,
		  s.status,
		  k.nama AS kelas,
		  j.nama AS jurusan
		FROM siswas s
		JOIN kelas k ON k.id = s.kelas_id
		JOIN jurusans j ON j.id = s.jurusan_id
		WHERE s.kelas_id = ?
		  AND s.jurusan_id = ?
		ORDER BY s.nama
		`
		db.Raw(query, kelasID, jurusanID).Scan(&rows)

		var kelas []models.Kelas
		var jurusan []models.Jurusan

		db.Find(&kelas)
		db.Find(&jurusan)

		var kelasName, jurusanName, bulanLabel string

		if kelasID != 0 {
			var k models.Kelas
			if err := db.First(&k, kelasID).Error; err == nil {
				kelasName = k.Nama
			}
		}

		if jurusanID != 0 {
			var j models.Jurusan
			if err := db.First(&j, jurusanID).Error; err == nil {
				jurusanName = j.Nama
			}
		}

		for _, b := range getBulanList() {
			if b.Value == bulan {
				bulanLabel = b.Label
				break
			}
		}

		return utils.Render(c, "pages/report_detail_siswa", fiber.Map{
			"Kelas":       kelas,
			"Jurusan":     jurusan,
			"Rows":        rows,
			"KelasID":     kelasID,
			"JurusanID":   jurusanID,
			"KelasName":   kelasName,
			"JurusanName": jurusanName,
			"Bulan":       bulan,
			"BulanLabel":  bulanLabel,
			"Tahun":       tahun,
		}, "layouts/main")
	}
}

func ExportDetailSiswaExcel(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		now := time.Now()

		kelasID := c.QueryInt("kelas_id")
		jurusanID := c.QueryInt("jurusan_id")
		bulan := c.QueryInt("bulan", int(now.Month()))
		tahun := c.QueryInt("tahun", now.Year())

		var rows []SiswaDetailRow

		query := `
		SELECT 
		  s.nis,
		  s.nama,
		  s.jenis_kelamin,
		  s.tempat_lahir,
		  s.tanggal_lahir,
		  s.alamat,
		  s.nama_wali,
		  s.no_hp,
		  s.status,
		  k.nama AS kelas,
		  j.nama AS jurusan
		FROM siswas s
		JOIN kelas k ON k.id = s.kelas_id
		JOIN jurusans j ON j.id = s.jurusan_id
		WHERE s.kelas_id = ?
		  AND s.jurusan_id = ?
		ORDER BY s.nama
		`

		db.Raw(query,
			kelasID,
			jurusanID,
		).Scan(&rows)

		var kelasName, jurusanName string

		if kelasID != 0 {
			var k models.Kelas
			db.First(&k, kelasID)
			kelasName = k.Nama
		}

		if jurusanID != 0 {
			var j models.Jurusan
			db.First(&j, jurusanID)
			jurusanName = j.Nama
		}

		bulanLabel := ""
		for _, b := range getBulanList() {
			if b.Value == bulan {
				bulanLabel = b.Label
				break
			}
		}

		f := excelize.NewFile()
		sheet := "Detail"
		f.SetSheetName("Sheet1", sheet)

		title := "Detail Siswa"
		if kelasName != "" {
			title += " Kelas " + kelasName
		}
		if jurusanName != "" {
			title += " - " + jurusanName
		}
		title += " (" + bulanLabel + " " + strconv.Itoa(tahun) + ")"

		f.SetCellValue(sheet, "A1", title)
		f.MergeCell(sheet, "A1", "K1")

		headers := []string{
			"NIS", "Nama", "JK", "Tempat Lahir", "Tanggal Lahir",
			"Alamat", "Wali", "No HP", "Status", "Kelas", "Jurusan",
		}

		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 2)
			f.SetCellValue(sheet, cell, h)
		}

		for i, row := range rows {
			r := i + 3
			f.SetCellValue(sheet, "A"+strconv.Itoa(r), row.NIS)
			f.SetCellValue(sheet, "B"+strconv.Itoa(r), row.Nama)
			f.SetCellValue(sheet, "C"+strconv.Itoa(r), row.JenisKelamin)
			f.SetCellValue(sheet, "D"+strconv.Itoa(r), row.TempatLahir)
			f.SetCellValue(sheet, "E"+strconv.Itoa(r), row.TanggalLahir.Format("2006-01-02"))
			f.SetCellValue(sheet, "F"+strconv.Itoa(r), row.Alamat)
			f.SetCellValue(sheet, "G"+strconv.Itoa(r), row.NamaWali)
			f.SetCellValue(sheet, "H"+strconv.Itoa(r), row.NoHP)
			f.SetCellValue(sheet, "I"+strconv.Itoa(r), row.Status)
			f.SetCellValue(sheet, "J"+strconv.Itoa(r), row.Kelas)
			f.SetCellValue(sheet, "K"+strconv.Itoa(r), row.Jurusan)
		}

		filename := "detail_siswa_" + bulanLabel + "_" + strconv.Itoa(tahun) + ".xlsx"

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename="+filename)

		return f.Write(c)
	}
}
