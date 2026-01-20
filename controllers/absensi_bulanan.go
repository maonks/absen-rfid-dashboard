// webcontroller/absensi_bulanan.go
package webcontroller

import (
	"fmt"
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

func AbsensiBulananTable(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		now := time.Now()

		// ==== AMBIL QUERY PARAM ====
		month := c.QueryInt("bulan", int(now.Month()))

		year := c.QueryInt("tahun", now.Year())

		fmt.Println("bulan", month)

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
			if r.Masuk.Hour() > 7 || (r.Masuk.Hour() == 7 && r.Masuk.Minute() > 0) {
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

		// ==== KIRIM KE VIEW ====
		return utils.Render(c, "partials/absensi_bulanan_table", fiber.Map{
			"Days":          days,
			"Rows":          rows,
			"SelectedMonth": month,
			"SelectedYear":  year,
			"Months":        utils.MonthList(),
			"Years":         utils.YearList(),
		})
	}
}
