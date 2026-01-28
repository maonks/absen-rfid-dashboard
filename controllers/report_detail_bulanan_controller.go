package webcontroller

import (
	"sort"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"github.com/maonks/absen-rfid-backend/utils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func ReportBulananDetailPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		now := time.Now()

		kelasID := c.QueryInt("kelas_id")
		jurusanID := c.QueryInt("jurusan_id")
		month := c.QueryInt("bulan")
		year := c.QueryInt("tahun", now.Year())

		lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()
		days := make([]int, lastDay)
		for i := 1; i <= lastDay; i++ {
			days[i-1] = i
		}

		liburMap := getLiburMap(db, month, year)

		type rawRow struct {
			ID    uint
			Nama  string
			Tgl   int
			Masuk time.Time
		}

		var raws []rawRow

		db.Raw(`
			SELECT
			  s.id,
			  s.nama,
			  EXTRACT(DAY FROM a.waktu)::int AS tgl,
			  MIN(a.waktu) AS masuk
			FROM absens a
			JOIN kartus k ON k.uid = a.uid
			JOIN siswas s ON s.id = k.siswa_id
			WHERE EXTRACT(MONTH FROM a.waktu) = ?
			  AND EXTRACT(YEAR FROM a.waktu) = ?
			  AND (? = 0 OR s.kelas_id = ?)
			  AND (? = 0 OR s.jurusan_id = ?)
			GROUP BY s.id, s.nama, tgl
			ORDER BY s.nama ASC
		`, month, year, kelasID, kelasID, jurusanID, jurusanID).Scan(&raws)

		rowMap := map[uint]*models.AbsensiRow{}

		for _, r := range raws {
			if _, ok := rowMap[r.ID]; !ok {
				rowMap[r.ID] = &models.AbsensiRow{
					ID:   r.ID,
					Nama: r.Nama,
					Hari: map[int]*models.HariCell{},
				}
			}

			status := ""
			if !r.Masuk.IsZero() {
				if r.Masuk.Hour() > 7 || (r.Masuk.Hour() == 7 && r.Masuk.Minute() > 0) {
					status = "LATE"
				} else {
					status = "OK"
				}
			}

			rowMap[r.ID].Hari[r.Tgl] = &models.HariCell{
				Masuk:  r.Masuk.Format("15:04"),
				Status: status,
			}
		}

		for _, row := range rowMap {
			for _, d := range days {

				if liburMap[d] {
					row.Hari[d] = &models.HariCell{Status: "LIBUR"}
					continue
				}

				if _, ok := row.Hari[d]; ok {
					continue
				}

				tgl := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.Local)
				if tgl.Before(time.Now()) {
					row.Hari[d] = &models.HariCell{Status: "ALPA"}
				} else {
					row.Hari[d] = &models.HariCell{Status: ""}
				}
			}
		}

		type statusRow struct {
			SiswaID uint
			Tgl     int
			Status  string
		}

		var statuses []statusRow
		db.Raw(`
			SELECT siswa_id,
			       EXTRACT(DAY FROM tanggal)::int AS tgl,
			       status
			FROM absensi_statuses
			WHERE EXTRACT(MONTH FROM tanggal) = ?
			  AND EXTRACT(YEAR FROM tanggal) = ?
		`, month, year).Scan(&statuses)

		for _, st := range statuses {
			if row, ok := rowMap[st.SiswaID]; ok {
				if cell, ok := row.Hari[st.Tgl]; ok {
					if cell.Status != "LIBUR" {
						cell.Status = st.Status
					}
				}
			}
		}

		var rows []models.AbsensiRow
		for _, r := range rowMap {
			rows = append(rows, *r)
		}

		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Nama < rows[j].Nama
		})

		var kelas []models.Kelas
		var jurusan []models.Jurusan
		db.Find(&kelas)
		db.Find(&jurusan)

		//untuk info report excel n print
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
			if b.Value == month {
				bulanLabel = b.Label
				break
			}
		}

		return utils.Render(c, "pages/report_detail_bulanan_kelas", fiber.Map{
			"Days":        days,
			"Rows":        rows,
			"Kelas":       kelas,
			"Jurusan":     jurusan,
			"KelasID":     kelasID,
			"KelasName":   kelasName,
			"JurusanID":   jurusanID,
			"JurusanName": jurusanName,
			"Bulan":       month,
			"Tahun":       year,
			"BulanLabel":  bulanLabel,
			"BulanList":   getBulanList(),
			"TahunList":   getTahunList(),
		}, "layouts/main")
	}
}

func ExportReportDetailBulananKelasExcel(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		kelasID := c.QueryInt("kelas_id")
		jurusanID := c.QueryInt("jurusan_id")
		bulan := c.QueryInt("bulan")
		tahun := c.QueryInt("tahun")

		if kelasID == 0 || bulan == 0 || tahun == 0 {
			return c.SendStatus(400)
		}

		// ======================
		// HITUNG JUMLAH HARI
		// ======================
		lastDay := time.Date(tahun, time.Month(bulan)+1, 0, 0, 0, 0, 0, time.Local).Day()
		var days []int
		for i := 1; i <= lastDay; i++ {
			days = append(days, i)
		}

		// ======================
		// AMBIL DATA TAP
		// ======================
		type rawRow struct {
			ID    uint
			Nama  string
			Tgl   int
			Masuk time.Time
		}

		query := `
			SELECT
			  s.id,
			  s.nama,
			  EXTRACT(DAY FROM a.waktu)::int AS tgl,
			  MIN(a.waktu) AS masuk
			FROM absens a
			JOIN kartus k ON k.uid = a.uid
			JOIN siswas s ON s.id = k.siswa_id
			WHERE s.kelas_id = ?
			  AND EXTRACT(MONTH FROM a.waktu) = ?
			  AND EXTRACT(YEAR FROM a.waktu) = ?
		`

		args := []interface{}{kelasID, bulan, tahun}

		if jurusanID != 0 {
			query += " AND s.jurusan_id = ? "
			args = append(args, jurusanID)
		}

		query += `
			GROUP BY s.id, s.nama, tgl
			ORDER BY s.nama
		`

		var raws []rawRow
		db.Raw(query, args...).Scan(&raws)

		rowMap := map[uint]*models.AbsensiRow{}

		for _, r := range raws {
			if _, ok := rowMap[r.ID]; !ok {
				rowMap[r.ID] = &models.AbsensiRow{
					ID:   r.ID,
					Nama: r.Nama,
					Hari: map[int]*models.HariCell{},
				}
			}

			status := "OK"
			if r.Masuk.Hour() > 7 || (r.Masuk.Hour() == 7 && r.Masuk.Minute() > 0) {
				status = "LATE"
			}

			rowMap[r.ID].Hari[r.Tgl] = &models.HariCell{
				Masuk:  r.Masuk.Format("15:04"),
				Status: status,
			}
		}

		// ======================
		// OVERRIDE STATUS MANUAL
		// ======================
		type statusRow struct {
			SiswaID uint
			Tgl     int
			Status  string
		}

		var statuses []statusRow
		db.Raw(`
			SELECT siswa_id,
			       EXTRACT(DAY FROM tanggal)::int AS tgl,
			       status
			FROM absensi_statuses
			WHERE EXTRACT(MONTH FROM tanggal) = ?
			  AND EXTRACT(YEAR FROM tanggal) = ?
		`, bulan, tahun).Scan(&statuses)

		for _, st := range statuses {
			if row, ok := rowMap[st.SiswaID]; ok {
				if cell, ok := row.Hari[st.Tgl]; ok {
					cell.Status = st.Status
				} else {
					row.Hari[st.Tgl] = &models.HariCell{Status: st.Status}
				}
			}
		}

		// ======================
		// LIBUR MAP
		// ======================
		liburMap := getLiburMap(db, bulan, tahun)

		today := time.Date(
			time.Now().Year(),
			time.Now().Month(),
			time.Now().Day(),
			0, 0, 0, 0,
			time.Local,
		)

		// ======================
		// ISI ALPA / LIBUR
		// ======================
		for _, row := range rowMap {
			for _, d := range days {

				// LIBUR global
				if liburMap[d] {
					row.Hari[d] = &models.HariCell{Status: "LIBUR"}
					continue
				}

				// sudah ada data
				if _, ok := row.Hari[d]; ok {
					continue
				}

				tgl := time.Date(tahun, time.Month(bulan), d, 0, 0, 0, 0, time.Local)

				// hari lewat tanpa tap
				if tgl.Before(today) {
					row.Hari[d] = &models.HariCell{Status: "ALPA"}
				} else {
					row.Hari[d] = &models.HariCell{Status: ""}
				}
			}
		}

		var rows []models.AbsensiRow
		for _, r := range rowMap {
			rows = append(rows, *r)
		}

		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Nama < rows[j].Nama
		})

		// ======================
		// AMBIL NAMA FILTER
		// ======================
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

		bulanLabel := ""
		for _, b := range getBulanList() {
			if b.Value == bulan {
				bulanLabel = b.Label
				break
			}
		}

		// ======================
		// BUAT FILE EXCEL
		// ======================
		f := excelize.NewFile()
		sheet := "Rekap"
		f.SetSheetName("Sheet1", sheet)

		title := "Rekap Detail Absensi"
		if kelasName != "" {
			title += " Kelas " + kelasName
		}
		if jurusanName != "" {
			title += " - " + jurusanName
		}
		title += " (" + bulanLabel + " " + strconv.Itoa(tahun) + ")"

		lastCol, _ := excelize.ColumnNumberToName(len(days) + 1)

		f.SetCellValue(sheet, "A1", title)
		f.MergeCell(sheet, "A1", lastCol+"1")

		// HEADER
		f.SetCellValue(sheet, "A2", "Nama")
		for i, d := range days {
			cell, _ := excelize.CoordinatesToCellName(i+2, 2)
			f.SetCellValue(sheet, cell, d)
		}

		// DATA
		for i, row := range rows {
			r := i + 3
			f.SetCellValue(sheet, "A"+strconv.Itoa(r), row.Nama)

			for j, d := range days {
				cell, _ := excelize.CoordinatesToCellName(j+2, r)

				if cdata, ok := row.Hari[d]; ok {
					if cdata.Status == "OK" || cdata.Status == "LATE" {
						f.SetCellValue(sheet, cell, cdata.Masuk)
					} else {
						f.SetCellValue(sheet, cell, cdata.Status)
					}
				} else {
					f.SetCellValue(sheet, cell, "-")
				}
			}
		}

		filename := "Report_detail_" + bulanLabel + "_" + strconv.Itoa(tahun)
		if kelasName != "" {
			filename += "_kelas_" + kelasName
		}
		if jurusanName != "" {
			filename += "_" + jurusanName
		}
		filename += ".xlsx"

		c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Set("Content-Disposition", "attachment; filename="+filename)

		return f.Write(c)
	}
}
