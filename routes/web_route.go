package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	webcontroller "github.com/maonks/absen-rfid-backend/controllers"
	apicontroller "github.com/maonks/absen-rfid-backend/controllers/api_controller"
	"gorm.io/gorm"
)

func WebRoutes(app *fiber.App, db *gorm.DB) {

	app.Get("/websocket", websocket.New(apicontroller.WebsocketHandler))

	app.Get("/login", apicontroller.Login(db))
	app.Get("/", webcontroller.HomePage(db))

	app.Get("/absensi", webcontroller.AbsensiPage(db))

	app.Get("/monitor", webcontroller.MonitorAbsen(db))

	app.Get("/siswa", webcontroller.SiswaPage(db))
	app.Get("/siswa/create", webcontroller.CreateSiswa(db))
	app.Post("/siswa/store", webcontroller.StoreSiswa(db))

	app.Get("/siswa/:id/edit", webcontroller.EditSiswa(db))
	app.Post("/siswa/:id/update", webcontroller.UpdateSiswa(db))

	app.Get("/kartu", webcontroller.KartuPage(db))

}
