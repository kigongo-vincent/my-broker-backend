package user

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, DB *gorm.DB) {
	user := app.Group("/auth")
	user.Post("/signin", func(c *fiber.Ctx) error {
		return RequestOTP(c, DB)
	})
	user.Post("/verify-otp", func(c *fiber.Ctx) error {
		return VerifyOTP(c, DB)
	})
}
