package user

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, DB *gorm.DB) {

	user := app.Group("/auth")

	user.Get("/room/:id", func(c *fiber.Ctx) error {
		// return c.SendStatus(202)
		user_id, err := strconv.ParseUint(c.Query("user_id"), 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get user id"})
		}
		room_id, err := strconv.ParseUint(c.Params("id"), 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get room id"})
		}
		return GetRoomDetails(c, DB, uint(room_id), uint(user_id))
	})

	user.Post("/signin", func(c *fiber.Ctx) error {
		return RequestOTP(c, DB)
	})
	user.Post("/verify-otp", func(c *fiber.Ctx) error {
		return VerifyOTP(c, DB)
	})
	user.Get("/rooms", func(c *fiber.Ctx) error {
		user_id, err := strconv.ParseUint(c.Query("user_id"), 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get user id"})
		}
		return GetRooms(c, DB, uint(user_id))
	})

}
