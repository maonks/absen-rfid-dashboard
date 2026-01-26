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

type BulanItem struct {
	Value int
	Label string
}

func getBulanList() []BulanItem {
	return []BulanItem{
		{1, "Januari"},
		{2, "Februari"},
		{3, "Maret"},
		{4, "April"},
		{5, "Mei"},
		{6, "Juni"},
		{7, "Juli"},
		{8, "Agustus"},
		{9, "September"},
		{10, "Oktober"},
		{11, "November"},
		{12, "Desember"},
	}
}

func getTahunList() []int {
	now := time.Now().Year()
	var years []int
	for i := 0; i < 10; i++ {
		years = append(years, now-i)
	}
	return years
}

func ReportPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var kelas []models.Kelas
		db.Find(&kelas)

		return utils.Render(c, "pages/report_page", fiber.Map{
			"kelas": kelas,
		}, "layouts/main")
	}
}

func ReportBulananKelasPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		kelasID := c.QueryInt("kelas_id")
		jurusanID := c.QueryInt("jurusan_id")
		bulan := c.QueryInt("bulan")
		tahun := c.QueryInt("tahun")

		var rows []models.ReportRow

		if kelasID != 0 && bulan != 0 && tahun != 0 {

			query := `
				SELECT 
					s.nama AS nama,
					k.nama AS kelas,
					j.nama AS jurusan,

					SUM(CASE WHEN a.status = 'OK' THEN 1 ELSE 0 END) AS hadir,
					SUM(CASE WHEN a.status = 'LATE' THEN 1 ELSE 0 END) AS telat,
					SUM(CASE WHEN a.status = 'SAKIT' THEN 1 ELSE 0 END) AS sakit,
					SUM(CASE WHEN a.status = 'IJIN' THEN 1 ELSE 0 END) AS izin,
					SUM(CASE WHEN a.status = 'ALPA' THEN 1 ELSE 0 END) AS alpa

				FROM absensi_statuses a
				JOIN siswas s ON s.id = a.siswa_id
				JOIN kelas k ON k.id = s.kelas_id
				JOIN jurusans j ON j.id = s.jurusan_id

				WHERE s.kelas_id = ?
				  AND EXTRACT(MONTH FROM a.tanggal) = ?
				  AND EXTRACT(YEAR FROM a.tanggal) = ?
			`

			args := []interface{}{kelasID, bulan, tahun}

			if jurusanID != 0 {
				query += " AND s.jurusan_id = ? "
				args = append(args, jurusanID)
			}

			query += `
				GROUP BY s.nama, k.nama, j.nama
				ORDER BY s.nama
			`

			db.Raw(query, args...).Scan(&rows)
		}

		var kelas []models.Kelas
		var jurusan []models.Jurusan
		db.Find(&kelas)
		db.Find(&jurusan)

		if tahun == 0 {
			tahun = time.Now().Year()
		}

		return utils.Render(c, "pages/report_bulanan_kelas", fiber.Map{
			"Kelas":     kelas,
			"Jurusan":   jurusan,
			"Rows":      rows,
			"KelasID":   kelasID,
			"JurusanID": jurusanID,
			"Bulan":     bulan,
			"Tahun":     tahun,
			"BulanList": getBulanList(),
			"TahunList": getTahunList(),
		}, "layouts/main")
	}
}

func ExportReportBulananKelasExcel(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		kelasID := c.QueryInt("kelas_id")
		jurusanID := c.QueryInt("jurusan_id")
		bulan := c.QueryInt("bulan")
		tahun := c.QueryInt("tahun")

		var rows []models.ReportRow

		query := `
			SELECT 
				s.nama AS nama,
				k.nama AS kelas,
				j.nama AS jurusan,

				SUM(CASE WHEN a.status = 'OK' THEN 1 ELSE 0 END) AS hadir,
				SUM(CASE WHEN a.status = 'LATE' THEN 1 ELSE 0 END) AS telat,
				SUM(CASE WHEN a.status = 'SAKIT' THEN 1 ELSE 0 END) AS sakit,
				SUM(CASE WHEN a.status = 'IJIN' THEN 1 ELSE 0 END) AS izin,
				SUM(CASE WHEN a.status = 'ALPA' THEN 1 ELSE 0 END) AS alpa

			FROM absensi_statuses a
			JOIN siswas s ON s.id = a.siswa_id
			JOIN kelas k ON k.id = s.kelas_id
			JOIN jurusans j ON j.id = s.jurusan_id

			WHERE s.kelas_id = ?
			  AND EXTRACT(MONTH FROM a.tanggal) = ?
			  AND EXTRACT(YEAR FROM a.tanggal) = ?
		`

		args := []interface{}{kelasID, bulan, tahun}

		if jurusanID != 0 {
			query += " AND s.jurusan_id = ? "
			args = append(args, jurusanID)
		}

		query += `
			GROUP BY s.nama, k.nama, j.nama
			ORDER BY s.nama
		`

		db.Raw(query, args...).Scan(&rows)

		f := excelize.NewFile()
		sheet := "Report"
		f.SetSheetName("Sheet1", sheet)

		headers := []string{"Nama", "Kelas", "Jurusan", "Hadir", "Telat", "Sakit", "Izin", "Alpa"}
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
		}

		for i, row := range rows {
			r := i + 2
			f.SetCellValue(sheet, "A"+strconv.Itoa(r), row.Nama)
			f.SetCellValue(sheet, "B"+strconv.Itoa(r), row.Kelas)
			f.SetCellValue(sheet, "C"+strconv.Itoa(r), row.Jurusan)
			f.SetCellValue(sheet, "D"+strconv.Itoa(r), row.Hadir)
			f.SetCellValue(sheet, "E"+strconv.Itoa(r), row.Telat)
			f.SetCellValue(sheet, "F"+strconv.Itoa(r), row.Sakit)
			f.SetCellValue(sheet, "G"+strconv.Itoa(r), row.Izin)
			f.SetCellValue(sheet, "H"+strconv.Itoa(r), row.Alpa)
		}

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename=report_bulanan_kelas.xlsx")
		return f.Write(c)
	}
}
