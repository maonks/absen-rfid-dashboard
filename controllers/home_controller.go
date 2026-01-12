package webcontroller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/utils"
	"gorm.io/gorm"
)

type HadirRow struct {
	ID     uint
	Uid    string
	Nama   string
	Masuk  string
	Pulang *string
	Status string
}

type TanpaRow struct {
	ID   uint
	Uid  string
	Nama string
}

func HomePage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var hadir []HadirRow
		var tanpa []TanpaRow

		// ================= SISWA HADIR =================
		db.Raw(`
			WITH daily AS (
			  SELECT
			    s.id AS id,
			    k.uid AS uid,
			    s.nama AS nama,
			    MIN(a.waktu) AS masuk,
			    MAX(
			      CASE
			        WHEN a.waktu::time >= '14:00:00'
			        THEN a.waktu
			      END
			    ) AS pulang
			  FROM absens a
			  JOIN kartus k ON k.uid = a.uid
			  JOIN siswas s ON s.id = k.siswa_id
			  WHERE DATE(a.waktu) = CURRENT_DATE
			  GROUP BY s.id, k.uid, s.nama
			)
			SELECT
			  id,
			  uid,
			  nama,
			  to_char(masuk, 'HH24:MI:SS') AS masuk,
			  CASE
			    WHEN pulang IS NOT NULL
			      THEN to_char(pulang, 'HH24:MI:SS')
			    ELSE NULL
			  END AS pulang,
			  CASE
			    WHEN pulang IS NULL THEN 'MASUK'
			    ELSE 'PULANG'
			  END AS status
			FROM daily
			ORDER BY nama
		`).Scan(&hadir)

		// ================= TANPA KETERANGAN =================
		db.Raw(`
			SELECT
			  s.id,
			  k.uid,
			  s.nama
			FROM siswas s
			LEFT JOIN kartus k ON k.siswa_id = s.id
			LEFT JOIN absens a
			  ON a.uid = k.uid
			  AND DATE(a.waktu) = CURRENT_DATE
			WHERE a.uid IS NULL
			ORDER BY s.nama
		`).Scan(&tanpa)

		return utils.Render(c, "pages/home_page", fiber.Map{
			"Hadir":           hadir,
			"TanpaKeterangan": tanpa,
		}, "layouts/main")
	}
}

func HomeRow(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid := c.Params("uid")

		var row HadirRow

		db.Raw(`
			WITH daily AS (
			SELECT
				k.uid,
				s.nama,
				MIN(a.waktu) AS masuk,
				MAX(
				CASE
					WHEN a.waktu::time >= '14:00:00'
					THEN a.waktu
				END
				) AS pulang
			FROM absens a
			JOIN kartus k
				ON k.uid = a.uid
			LEFT JOIN siswas s
				ON s.id = k.siswa_id
			WHERE a.waktu >= date_trunc('day', now())
				AND a.waktu < date_trunc('day', now()) + interval '1 day'
				AND k.uid = ?
			GROUP BY k.uid, s.nama
			)
			SELECT
			uid,
			COALESCE(nama, 'Belum Terdaftar') AS nama,
			to_char(masuk, 'HH24:MI:SS') AS masuk,
			CASE
				WHEN pulang IS NOT NULL
				THEN to_char(pulang, 'HH24:MI:SS')
				ELSE NULL
			END AS pulang,
			CASE
				WHEN pulang IS NULL THEN 'MASUK'
				ELSE 'PULANG'
			END AS status
			FROM daily

		`, uid).Scan(&row)

		// jika row belum ada (tap pertama)
		if row.Uid == "" {
			return c.SendStatus(fiber.StatusNoContent)
		}

		return utils.Render(c, "partials/home_row", fiber.Map{
			"No":     0, // optional: bisa dihitung di client
			"Uid":    row.Uid,
			"Nama":   row.Nama,
			"Masuk":  row.Masuk,
			"Pulang": row.Pulang,
			"Status": row.Status,
		})
	}
}
