package user

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, DB *gorm.DB) {

	user := app.Group("/auth")

	user.Get("/room-by-participants", func(c *fiber.Ctx) error {
		// return c.SendStatus(202)
		user_id, err := strconv.ParseUint(c.Query("user_id"), 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get user id"})
		}
		participant_id, err := strconv.ParseUint(c.Query("participant_id"), 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get participant id"})
		}
		return GetOrCreateRoomByParticipants(c, DB, uint(user_id), uint(participant_id))
	})

	user.Get("/chats-and-favourites", func(c *fiber.Ctx) error {
		UserID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
		return GetChatsAndFavourites(c, DB, uint(UserID))
	})

	user.Post("update-id", func(c *fiber.Ctx) error {
		return UpdateID(c, DB)
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

	user.Post("/update-profile", func(c *fiber.Ctx) error {
		return UpdateProfile(c, DB)
	})

}
