package user

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kigongo-vincent/my-broker-backend/middleware"
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
	user.Post("/google", func(c *fiber.Ctx) error {
		return GoogleSignin(c, DB)
	})
	user.Get("/rooms", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)
		return GetRooms(c, DB, userID)
	})

	user.Post("/update-profile", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return UpdateProfile(c, DB)
	})

	user.Get("/room-by-participants", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)
		participantID, err := strconv.ParseUint(c.Query("participant_id"), 10, 32)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get participant id"})
		}
		return GetOrCreateRoomByParticipants(c, DB, userID, uint(participantID))
	})

	user.Get("/chats-and-favourites", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)
		return GetChatsAndFavourites(c, DB, userID)
	})

	user.Post("/update-id", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return UpdateID(c, DB)
	})
	user.Post("/last-seen", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return UpdateLastSeen(c, DB)
	})
	user.Get("/profile", func(c *fiber.Ctx) error {
		return GetProfileByID(c, DB)
	})
	user.Post("/chat/block", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return BlockUser(c, DB)
	})
	user.Post("/chat/report", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return ReportUser(c, DB)
	})
	user.Post("/chat/clear", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return ClearChat(c, DB)
	})

	user.Get("/admin/users", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return ListUsersForAdmin(c, DB)
	})
	user.Post("/admin/approve-id", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return ApproveUserID(c, DB)
	})

}
