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

	pakaijwt.Get("/kelas/create", webcontroller.CreateKelas(db))
	pakaijwt.Post("/kelas/store", webcontroller.StoreKelas(db))
	pakaijwt.Get("/kelas/:id/row", webcontroller.KelasRow(db))
	pakaijwt.Get("/kelas/:id/edit", webcontroller.KelasEditRow(db))
	pakaijwt.Post("/kelas/:id/update", webcontroller.KelasUpdate(db))
	pakaijwt.Delete("/kelas/:id", webcontroller.KelasDelete(db))

	pakaijwt.Get("/jurusan/create", webcontroller.CreateJurusan(db))
	pakaijwt.Post("/jurusan/store", webcontroller.StoreJurusan(db))
	pakaijwt.Get("/jurusan/:id/row", webcontroller.JurusanRow(db))
	pakaijwt.Get("/jurusan/:id/edit", webcontroller.JurusanEditRow(db))
	pakaijwt.Post("/jurusan/:id/update", webcontroller.JurusanUpdate(db))
	pakaijwt.Delete("/jurusan/:id", webcontroller.JurusanDelete(db))

	pakaijwt.Get("/kartu", webcontroller.KartuPage(db))

	// Absensi Bulanan
	pakaijwt.Get("/absensi/bulanan", webcontroller.AbsensiBulananPage(db))
	pakaijwt.Get("/absensi/bulanan/table", webcontroller.AbsensiBulananTable(db))
	pakaijwt.Post("/absensi/bulanan/update", webcontroller.AbsensiUpdate(db))
	pakaijwt.Get("/absensi/bulanan/table-edit", webcontroller.AbsensiBulananTableEdit(db))
	pakaijwt.Post("/absensi/hari/libur", webcontroller.SetHariLibur(db))

	// Report

	pakaijwt.Get("/report", webcontroller.ReportPage(db))
	pakaijwt.Get("/report/bulanan/kelas", webcontroller.ReportBulananKelasPage(db))
	pakaijwt.Get("/report/bulanan/kelas/excel", webcontroller.ExportReportBulananKelasExcel(db))

	pakaijwt.Get("/report/bulanan/kelas-detail", webcontroller.ReportBulananDetailPage(db))
	pakaijwt.Get("/report/bulanan/kelas-detail/excel", webcontroller.ExportReportDetailBulananKelasExcel(db))

	pakaijwt.Get("/report/detail-siswa", webcontroller.ReportDetailSiswaPage(db))
	pakaijwt.Get("/report/detail-siswa/excel", webcontroller.ExportDetailSiswaExcel(db))

}
