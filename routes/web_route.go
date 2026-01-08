package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/maonks/absen-rfid-backend/controllers"
)

func WebRoutes(app *fiber.App) {

	app.Get("/", controllers.Dashboard)
	app.Get("/websocket", websocket.New(controllers.WebsocketHandler))

}
