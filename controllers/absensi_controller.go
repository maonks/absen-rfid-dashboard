package webcontroller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/utils"
	"gorm.io/gorm"
)

type HitungRow struct {
	Nama        string `json:"nama"`
	Tanggal     string `json:"tanggal"`
	Masuk       string `json:"masuk"`
	Pulang      string `json:"pulang"`
	Terlambat   string `json:"terlambat"`
	PulangCepat string `json:"pulang_cepat"`
	Lembur      string `json:"lembur"`
	JamKerja    string `json:"jam_kerja"`
}

func AbsensiPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var rows []HitungRow

		db.Raw(`
			WITH daily AS (
			  SELECT
			    s.nama AS nama,
			    DATE(a.waktu) AS tanggal,
			    MIN(a.waktu) AS masuk,
			    MAX(a.waktu) AS pulang
			  FROM absens a
			  JOIN kartus k ON k.uid = a.uid
			  JOIN siswas s ON s.id = k.siswa_id
			  GROUP BY s.nama, DATE(a.waktu)
			)
			SELECT
			  nama,
			  tanggal::text AS tanggal,
			  to_char(masuk, 'HH24:MI:SS') AS masuk,
			  to_char(pulang, 'HH24:MI:SS') AS pulang,

			  CASE
			    WHEN masuk::time > '07:00:00'
			      THEN to_char(masuk::time - '07:00:00'::time, 'HH24:MI:SS')
			    ELSE '-'
			  END AS terlambat,

			  CASE
			    WHEN pulang::time < '16:00:00'
			      THEN to_char('16:00:00'::time - pulang::time, 'HH24:MI:SS')
			    ELSE '-'
			  END AS pulang_cepat,

			  CASE
			    WHEN pulang::time > '16:00:00'
			      THEN to_char(pulang::time - '16:00:00'::time, 'HH24:MI:SS')
			    ELSE '-'
			  END AS lembur,

			  to_char(pulang - masuk, 'HH24:MI:SS') AS jam_kerja
			FROM daily
			ORDER BY tanggal DESC, nama
		`).Scan(&rows)

		return utils.Render(c, "pages/absensi_page", fiber.Map{
			"Title": "Perhitungan Kehadiran",
			"Rows":  rows,
		}, "layouts/main")
	}
}
