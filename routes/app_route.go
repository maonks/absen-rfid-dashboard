package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/controllers"
	"gorm.io/gorm"
)

func AbsenRoute(app *fiber.App, db *gorm.DB) {

	app.Post("/api/absen", controllers.CreateAbsen(db))
	app.Get("/api/absen/table", controllers.SearchAbsen(db))
	app.Post("/api/kartu/:uid", controllers.UpdateKartu(db))
	app.Get("/api/kartu/:uid/edit", controllers.EditModal(db))

}
