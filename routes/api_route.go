package routes

import (
	"github.com/gofiber/fiber/v2"
	webcontroller "github.com/maonks/absen-rfid-backend/controllers"
	"gorm.io/gorm"
)

func AbsenRoute(app *fiber.App, db *gorm.DB) {

	//endpoin yang di hit sama device
	app.Post("/api/absen", webcontroller.CreateAbsen(db))

	//ini yang nampilkan hasil tap
	app.Get("/api/absen/table", webcontroller.RealtimeAbsen(db))

	app.Get("/api/home/row/:uid", webcontroller.HomeRow(db))

}
