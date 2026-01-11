package middlewares

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func PakaiJWT() fiber.Handler {
	return func(c *fiber.Ctx) error {

		tokenString := c.Cookies("access_token")
		if tokenString == "" {
			return c.Redirect("/login")
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return c.Redirect("/login")
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := uint(claims["sub"].(float64))

		c.Locals("user_id", userID)

		return c.Next()
	}
}
