package configs

import (
	"log"
	"os"

	"github.com/maonks/absen-rfid-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func KonekDB() (*gorm.DB, error) {

	dsn := "host=" + os.Getenv("DB_HOST") + " port=" + os.Getenv("DB_PORT") + " user=" + os.Getenv("DB_USER") + " password=" + os.Getenv("DB_PASS") + " dbname=" + os.Getenv("DB_NAME") + " sslmode=" + os.Getenv("DB_SSL") + ""

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Gagal Konek DB", err)
	}

	db.AutoMigrate(&models.User{}, &models.Absen{}, &models.Device{}, &models.Kartu{}, &models.Siswa{}, &models.AbsensiStatus{}, &models.AbsensiHari{})

	return db, err
}
