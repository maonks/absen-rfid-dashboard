package routes

import (
	"github.com/gofiber/fiber/v2"
	apicontroller "github.com/maonks/absen-rfid-backend/controllers/api_controller"
	"github.com/maonks/absen-rfid-backend/middlewares"
	"gorm.io/gorm"
)

func AdminRoute(app *fiber.App, db *gorm.DB) {

	admin := app.Group("/admin",
		middlewares.PakaiJWT(),
		middlewares.LoadUser(db))

	admin.Get("/users", apicontroller.AdminUserPage(db))

	admin.Get("/users/create", apicontroller.AdminUserCreate(db))
	admin.Post("/users/store", apicontroller.AdminUserStore(db))

	admin.Get("/users/:id/edit", apicontroller.AdminUserEdit(db))
	admin.Post("/users/:id/update", apicontroller.AdminUserUpdate(db))

	admin.Get("/users/:id/delete", apicontroller.AdminUserDelete(db))
	admin.Post("/users/:id/destroy", apicontroller.AdminUserDestroy(db))

}
