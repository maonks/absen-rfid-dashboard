package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	webcontroller "github.com/maonks/absen-rfid-backend/controllers"
	apicontroller "github.com/maonks/absen-rfid-backend/controllers/api_controller"
	"github.com/maonks/absen-rfid-backend/middlewares"
	"gorm.io/gorm"
)

func WebRoutes(app *fiber.App, db *gorm.DB) {

	app.Get("/websocket", websocket.New(apicontroller.WebsocketHandler))

	app.Get("/login", func(c *fiber.Ctx) error {
		return c.Render("layouts/login_page", nil)
	})
	app.Post("/login", apicontroller.Login(db))
	app.Post("/logout", apicontroller.Logout)

	pakaijwt := app.Group("/",
		middlewares.PakaiJWT(),
		middlewares.LoadUser(db))

	pakaijwt.Get("/", webcontroller.HomePage(db))
	pakaijwt.Static("/static", "./views/static")

	pakaijwt.Get("/api/home/realtime", webcontroller.HomeRealtime(db))

	pakaijwt.Get("/absensi", webcontroller.AbsensiPage(db))

	pakaijwt.Get("/monitor", webcontroller.MonitorAbsen(db))

	pakaijwt.Get("/siswa", webcontroller.SiswaPage(db))
	pakaijwt.Get("/siswa/create", webcontroller.CreateSiswa(db))
	pakaijwt.Post("/siswa/store", webcontroller.StoreSiswa(db))
	pakaijwt.Get("/siswa/:id/edit", webcontroller.EditSiswa(db))
	pakaijwt.Post("/siswa/:id/update", webcontroller.UpdateSiswa(db))

	pakaijwt.Get("/kartu", webcontroller.KartuPage(db))

	// web_routes.go
	app.Get(
		"/absensi/bulanan",
		webcontroller.AbsensiBulananPage(db),
	)

	app.Get(
		"/absensi/bulanan/table",
		webcontroller.AbsensiBulananTable(db),
	)

}
