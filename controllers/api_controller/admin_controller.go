package apicontroller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/maonks/absen-rfid-backend/models"
	"github.com/maonks/absen-rfid-backend/services"
	"github.com/maonks/absen-rfid-backend/utils"
	"gorm.io/gorm"
)

func AdminUserPage(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var users []models.User
		db.Order("id desc").Find(&users)

		return utils.Render(c, "pages/admin_user_page", fiber.Map{
			"Users": users,
		}, "layouts/main")
	}
}

func AdminUserCreate(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		return utils.Render(c, "modals/tambah_user", nil)

	}

}

func AdminUserStore(db *gorm.DB) fiber.Handler {

	cekUsername := services.CekUsername(db)

	return func(c *fiber.Ctx) error {

		var user models.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(400).SendString("Input tidak valid")
		}

		if _, err := cekUsername(user.Username); err == nil {
			return c.Status(400).SendString("Username sudah terdaftar")
		}

		// 🔐 HASH PASSWORD (PAKAI FUNGSI KAMU)
		hash, err := utils.HashPassword(user.Password)
		if err != nil {
			return c.Status(500).SendString("Gagal hash password")
		}
		user.Password = hash

		if err := db.Create(&user).Error; err != nil {
			return c.Status(500).SendString("Gagal menyimpan user")
		}

		c.Set("HX-Redirect", "/admin/users")
		return c.SendStatus(204)
	}
}

func AdminUserEdit(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var user models.User
		id := c.Params("id")
		if err := db.First(&user, id).Error; err != nil {
			return c.Status(404).SendString("user tidak ditemukan")
		}

		return utils.Render(c, "modals/edit_user", fiber.Map{
			"User": user,
		})
	}
}

func AdminUserUpdate(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var user models.User
		if err := db.First(&user, c.Params("id")).Error; err != nil {
			return c.SendStatus(404)
		}

		var input struct {
			Nama     string
			Password string
			Jabatan  string
			Role     string
		}

		c.BodyParser(&input)

		user.Nama = input.Nama
		user.Jabatan = input.Jabatan
		user.Role = input.Role

		if input.Password != "" {
			hash, _ := utils.HashPassword(input.Password)
			user.Password = hash
		}

		db.Save(&user)

		c.Set("HX-Redirect", "/admin/users")
		return c.SendStatus(204)
	}
}

func AdminUserDelete(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		var user models.User
		db.First(&user, c.Params("id"))

		return utils.Render(c, "modals/hapus_user", fiber.Map{
			"User": user,
		})
	}
}

func AdminUserDestroy(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {

		db.Delete(&models.User{}, c.Params("id"))

		c.Set("HX-Redirect", "/admin/users")
		return c.SendStatus(204)
	}
}
