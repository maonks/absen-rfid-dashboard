// webcontroller/absensi_bulanan.go
package webcontroller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/maonks/absen-rfid-backend/models"
	"github.com/maonks/absen-rfid-backend/utils"
)

func AbsensiBulananPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		now := time.Now()

		// ==== AMBIL QUERY PARAM ====
		month := c.QueryInt("bulan", int(now.Month()))
		year := c.QueryInt("tahun", now.Year())

		// ==== JUMLAH HARI DALAM BULAN ====
		lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()

		days := make([]int, lastDay)
		for i := 1; i <= lastDay; i++ {
			days[i-1] = i
		}

		// ==== RAW DATA ====
		type rawRow struct {
			Nama   string
			Tgl    int
			Masuk  time.Time
			Pulang time.Time
		}

		var raws []rawRow

		db.Raw(`
			SELECT
			  s.nama,
			  EXTRACT(DAY FROM a.waktu)::int AS tgl,
			  MIN(a.waktu) AS masuk,
			  MAX(a.waktu) AS pulang
			FROM absens a
			JOIN kartus k ON k.uid = a.uid
			JOIN siswas s ON s.id = k.siswa_id
			WHERE EXTRACT(MONTH FROM a.waktu) = ?
			  AND EXTRACT(YEAR FROM a.waktu) = ?
			GROUP BY s.nama, tgl
			ORDER BY s.nama ASC
		`, month, year).Scan(&raws)

		// ==== MAPPING KE TABLE ====
		rowMap := map[string]*models.AbsensiRow{}

		for _, r := range raws {

			if _, ok := rowMap[r.Nama]; !ok {
				rowMap[r.Nama] = &models.AbsensiRow{
					Nama: r.Nama,
					Hari: map[int]*models.HariCell{},
				}
			}

			status := "OK"
			if r.Masuk.Hour() > 7 || (r.Masuk.Hour() == 7 && r.Masuk.Minute() > 30) {
				status = "LATE"
			}

			rowMap[r.Nama].Hari[r.Tgl] = &models.HariCell{
				Masuk:  r.Masuk.Format("15:04"),
				Pulang: r.Pulang.Format("15:04"),
				Status: status,
			}
		}

		var rows []models.AbsensiRow
		for _, r := range rowMap {
			rows = append(rows, *r)
		}

		//=======STATUS ROW =====

		type statusRow struct {
			Nama   string
			Tgl    int
			Status string
		}

		var statuses []statusRow

		db.Raw(`
			SELECT
			s.nama,
			EXTRACT(DAY FROM ast.tanggal)::int AS tgl,
			ast.status
			FROM absensi_statuses ast
			JOIN siswas s ON s.id = ast.siswa_id
			WHERE EXTRACT(MONTH FROM ast.tanggal) = ?
			AND EXTRACT(YEAR FROM ast.tanggal) = ?
		`, month, year).Scan(&statuses)

		// ==== KIRIM KE VIEW ====
		return utils.Render(c, "pages/absensi_bulanan", fiber.Map{
			"Days":          days,
			"Rows":          rows,
			"SelectedMonth": month,
			"SelectedYear":  year,
			"Months":        utils.MonthList(),
			"Years":         utils.YearList(),
		}, "layouts/main")
	}
}

// ================= EDIT MODE TABLE =================
func AbsensiBulananTableEdit(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

		month := c.QueryInt("bulan", int(now.Month()))
		year := c.QueryInt("tahun", now.Year())

		// ==== DAYS ====
		lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()
		days := make([]int, lastDay)
		for i := 1; i <= lastDay; i++ {
			days[i-1] = i
		}

		// ==== LIBUR MAP (GLOBAL) ====
		liburMap := getLiburMap(db, month, year)

		// ==== RAW TAP DATA ====
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
			GROUP BY s.id, s.nama, tgl
			ORDER BY s.nama ASC
		`, month, year).Scan(&raws)

		rowMap := map[uint]*models.AbsensiRow{}

		// ==== HASIL TAP → OK / LATE (BELUM FINAL) ====
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
				Status: status,
			}
		}

		// ==== STATUS FINAL PER HARI ====
		for _, row := range rowMap {
			for _, d := range days {

				// 🔥 1. LIBUR GLOBAL (PALING KUAT)
				if liburMap[d] {
					row.Hari[d] = &models.HariCell{Status: "LIBUR"}
					continue
				}

				// 2. Sudah ada tap → pakai hasil tap
				if _, ok := row.Hari[d]; ok {
					continue
				}

				tgl := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.Local)

				// 3. Hari lewat & tidak tap → ALPA
				if tgl.Before(today) {
					row.Hari[d] = &models.HariCell{Status: "ALPA"}
				} else {
					// 4. Hari belum lewat → PENDING
					row.Hari[d] = &models.HariCell{Status: ""}
				}
			}
		}

		// ==== MANUAL OVERRIDE PER SISWA (KECUALI LIBUR) ====
		type statusRow struct {
			SiswaID uint
			Tgl     int
			Status  string
		}

		var statuses []statusRow
		db.Raw(`
			SELECT siswa_id, EXTRACT(DAY FROM tanggal)::int AS tgl, status
			FROM absensi_statuses
			WHERE EXTRACT(MONTH FROM tanggal) = ?
			  AND EXTRACT(YEAR FROM tanggal) = ?
		`, month, year).Scan(&statuses)

		for _, st := range statuses {
			if row, ok := rowMap[st.SiswaID]; ok {
				if row.Hari[st.Tgl].Status != "LIBUR" {
					row.Hari[st.Tgl].Status = st.Status
				}
			}
		}

		// ==== MAP → SLICE ====
		var rows []models.AbsensiRow
		for _, r := range rowMap {
			rows = append(rows, *r)
		}

		return utils.Render(c, "partials/absensi_bulanan_table_edit", fiber.Map{
			"Days":          days,
			"Rows":          rows,
			"SelectedMonth": month,
			"SelectedYear":  year,
		})
	}
}

// ini untuk simpan perubahan
func AbsensiUpdate(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		type Req struct {
			UserID  uint   `form:"user_id"`
			Tanggal int    `form:"tanggal"`
			Status  string `form:"status"`
		}

		var req Req
		if err := c.BodyParser(&req); err != nil {
			return c.SendStatus(400)
		}

		now := time.Now()
		tgl := time.Date(
			now.Year(),
			now.Month(),
			req.Tanggal,
			0, 0, 0, 0,
			time.Local,
		)

		// =======================
		// 🔥 JIKA STATUS = LIBUR
		// =======================
		if req.Status == "LIBUR" {

			// simpan hari libur (GLOBAL)
			var hari models.AbsensiHari
			err := db.Where("tanggal = ?", tgl).First(&hari).Error

			if err == gorm.ErrRecordNotFound {
				db.Create(&models.AbsensiHari{
					Tanggal: tgl,
					Status:  "LIBUR",
				})
			}

			// hapus semua status siswa di tanggal itu
			db.Where("tanggal = ?", tgl).Delete(&models.AbsensiStatus{})

			return c.SendStatus(200)
		}

		// =======================
		// NORMAL (NON-LIBUR)
		// =======================

		var rec models.AbsensiStatus
		err := db.Where("siswa_id = ? AND tanggal = ?", req.UserID, tgl).
			First(&rec).Error

		if err == gorm.ErrRecordNotFound {
			rec = models.AbsensiStatus{
				SiswaID: req.UserID,
				Tanggal: tgl,
				Status:  req.Status,
			}
			db.Create(&rec)
		} else {
			db.Model(&rec).Update("status", req.Status)
		}

		return c.SendStatus(200)
	}
}

// ini untuk websocket auto refresh
func AbsensiBulananTable(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		now := time.Now()
		today := time.Date(
			now.Year(), now.Month(), now.Day(),
			0, 0, 0, 0, time.Local,
		)

		month := c.QueryInt("bulan", int(now.Month()))
		year := c.QueryInt("tahun", now.Year())

		// ==== DAYS ====
		lastDay := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.Local).Day()
		days := make([]int, lastDay)
		for i := 1; i <= lastDay; i++ {
			days[i-1] = i
		}

		// ==== LIBUR GLOBAL ====
		liburMap := getLiburMap(db, month, year)

		// ==== RAW TAP DATA ====
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
			GROUP BY s.id, s.nama, tgl
			ORDER BY s.nama ASC
		`, month, year).Scan(&raws)

		// ==== MAP SISWA ====
		rowMap := map[uint]*models.AbsensiRow{}

		// TAP → OK / LATE
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

		// ==== STATUS FINAL PER HARI ====
		for _, row := range rowMap {
			for _, d := range days {

				// 🔥 1. LIBUR GLOBAL (PALING KUAT)
				if liburMap[d] {
					row.Hari[d] = &models.HariCell{Status: "LIBUR"}
					continue
				}

				// 2. Sudah ada tap → pakai hasil tap
				if _, ok := row.Hari[d]; ok {
					continue
				}

				tgl := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.Local)

				// 3. Hari lewat & tidak tap → ALPA
				if tgl.Before(today) {
					row.Hari[d] = &models.HariCell{Status: "ALPA"}
				} else {
					// 4. Hari belum lewat → pending
					row.Hari[d] = &models.HariCell{Status: ""}
				}
			}
		}

		// ==== MANUAL OVERRIDE PER SISWA (KECUALI LIBUR) ====
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

		// ==== MAP → SLICE ====
		var rows []models.AbsensiRow
		for _, r := range rowMap {
			rows = append(rows, *r)
		}

		return utils.Render(c, "partials/absensi_bulanan_table", fiber.Map{
			"Days":          days,
			"Rows":          rows,
			"SelectedMonth": month,
			"SelectedYear":  year,
		})
	}
}

func SetHariLibur(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		type Req struct {
			Tanggal string `form:"tanggal"` // YYYY-MM-DD
			Status  string `form:"status"`  // LIBUR | AKTIF
		}

		var req Req
		if err := c.BodyParser(&req); err != nil {
			return c.SendStatus(400)
		}

		tgl, err := time.Parse("2006-01-02", req.Tanggal)
		if err != nil {
			return c.SendStatus(400)
		}

		var rec models.AbsensiHari
		err = db.Where("tanggal = ?", tgl).First(&rec).Error

		if err == gorm.ErrRecordNotFound {
			rec = models.AbsensiHari{
				Tanggal: tgl,
				Status:  req.Status,
			}
			db.Create(&rec)
		} else {
			db.Model(&rec).Update("status", req.Status)
		}

		return c.SendStatus(200)
	}
}

func getLiburMap(db *gorm.DB, month, year int) map[int]bool {

	type row struct {
		Tgl int
	}

	var rows []row
	db.Raw(`
		SELECT EXTRACT(DAY FROM tanggal)::int AS tgl
		FROM absensi_haris
		WHERE status = 'LIBUR'
		  AND EXTRACT(MONTH FROM tanggal) = ?
		  AND EXTRACT(YEAR FROM tanggal) = ?
	`, month, year).Scan(&rows)

	libur := map[int]bool{}
	for _, r := range rows {
		libur[r.Tgl] = true
	}
	return libur
}
