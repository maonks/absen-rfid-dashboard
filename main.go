package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"
	"github.com/maonks/absen-rfid-backend/configs"
	"github.com/maonks/absen-rfid-backend/routes"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println(".ENV Tidak di temukan")
	}

	engine := html.New("./views", ".html")

	engine.AddFunc("add", func(a, b int) int {
		return a + b
	})

	engine.AddFunc("seq", func(from, to int) []int {
		var s []int
		for i := from; i <= to; i++ {
			s = append(s, i)
		}
		return s
	})

	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Use(cors.New())

	db, err := configs.KonekDB()

	if err != nil {
		log.Fatal("Gagal Konek DB", err)
	}

	routes.DeviceRoute(app, db)
	routes.AbsenRoute(app, db)
	routes.WebRoutes(app, db)
	routes.AdminRoute(app, db)

	app.Listen("" + os.Getenv("APP_HOST") + ":" + os.Getenv("APP_PORT") + "")
}
