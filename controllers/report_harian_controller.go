package webcontroller

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"github.com/maonks/absen-rfid-backend/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func ReportHargaKelasPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		kelasID := c.QueryInt("kelas_id")
		jurusanID := c.QueryInt("jurusan_id")

		startDate := c.Query("start_date") // format: YYYY-MM-DD
		endDate := c.Query("end_date")

		var rows []models.ReportRow

		if kelasID != 0 && startDate != "" && endDate != "" {

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
				  AND a.tanggal BETWEEN ? AND ?
			`

			args := []interface{}{kelasID, startDate, endDate}

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

		var kelasName, jurusanName string

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

		return utils.Render(c, "pages/report_harian_kelas", fiber.Map{
			"Kelas":       kelas,
			"Jurusan":     jurusan,
			"Rows":        rows,
			"KelasID":     kelasID,
			"JurusanID":   jurusanID,
			"KelasName":   kelasName,
			"JurusanName": jurusanName,
			"StartDate":   startDate,
			"EndDate":     endDate,
		}, "layouts/main")
	}
}

func ExportReportHarianKelasExcel(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		kelasID := c.QueryInt("kelas_id")
		jurusanID := c.QueryInt("jurusan_id")
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

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
			  AND a.tanggal BETWEEN ? AND ?
		`

		args := []interface{}{kelasID, startDate, endDate}

		if jurusanID != 0 {
			query += " AND s.jurusan_id = ? "
			args = append(args, jurusanID)
		}

		query += `
			GROUP BY s.nama, k.nama, j.nama
			ORDER BY s.nama
		`

		db.Raw(query, args...).Scan(&rows)

		var kelasName, jurusanName string

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

		f := excelize.NewFile()
		sheet := "Report"
		f.SetSheetName("Sheet1", sheet)

		title := "Report Harian"
		if kelasName != "" {
			title += " Kelas " + kelasName
		}
		if jurusanName != "" {
			title += " - " + jurusanName
		}
		title += " (" + startDate + " s/d " + endDate + ")"

		f.SetCellValue(sheet, "A1", title)
		f.MergeCell(sheet, "A1", "H1")

		headers := []string{"Nama", "Kelas", "Jurusan", "Hadir", "Telat", "Sakit", "Izin", "Alpa"}

		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 2)
			f.SetCellValue(sheet, cell, h)
		}

		for i, row := range rows {
			r := i + 3
			f.SetCellValue(sheet, "A"+strconv.Itoa(r), row.Nama)
			f.SetCellValue(sheet, "B"+strconv.Itoa(r), row.Kelas)
			f.SetCellValue(sheet, "C"+strconv.Itoa(r), row.Jurusan)
			f.SetCellValue(sheet, "D"+strconv.Itoa(r), row.Hadir)
			f.SetCellValue(sheet, "E"+strconv.Itoa(r), row.Telat)
			f.SetCellValue(sheet, "F"+strconv.Itoa(r), row.Sakit)
			f.SetCellValue(sheet, "G"+strconv.Itoa(r), row.Izin)
			f.SetCellValue(sheet, "H"+strconv.Itoa(r), row.Alpa)
		}

		filename := "report_harian_" + startDate + "_to_" + endDate + ".xlsx"

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename="+filename)

		return f.Write(c)
	}
}
