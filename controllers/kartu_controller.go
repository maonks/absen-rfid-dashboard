package webcontroller

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type KartuRow struct {
	ID        uint
	UID       string
	NamaSiswa *string // bisa NULL
	Waktu     time.Time
	Status    string
}

func KartuPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var rows []KartuRow

		db.Raw(`
			SELECT
			  k.id,
			  k.uid,
			  s.nama AS nama_siswa,
			  CASE
			    WHEN k.siswa_id IS NULL THEN 'FREE'
			    ELSE 'TERPAKAI'
			  END AS status,			    
				k.created_at AS waktu,
				k.updated_at AS waktuupdate
			FROM kartus k
			LEFT JOIN siswas s
			  ON s.id = k.siswa_id
			ORDER BY k.created_at DESC
		`).Scan(&rows)

		return c.Render("pages/kartu_page", fiber.Map{
			"Rows": rows,
		}, "layouts/main")
	}
}
