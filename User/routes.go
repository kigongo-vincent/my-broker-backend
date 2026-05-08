package user

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
	"github.com/kigongo-vincent/my-broker-backend/middleware"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, DB *gorm.DB) {

	user := app.Group("/auth")

	user.Post("/signin", func(c *fiber.Ctx) error {
		return CheckPhoneForPin(c, DB)
	})
	user.Post("/verify-otp", func(c *fiber.Ctx) error {
		return CompletePhonePin(c, DB)
	})
	user.Get("/verification-status", middleware.AuthMiddlewareJSON(), func(c *fiber.Ctx) error {
		return GetVerificationStatusJSON(c, DB)
	})
	user.Post("/id-verification", middleware.AuthMiddlewareJSON(), func(c *fiber.Ctx) error {
		return SubmitIDVerificationJSON(c, DB)
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
			return fbcodec.SendError(c, 400, "failed to get participant id")
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
	user.Post("/sync-session", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return SyncSession(c, DB)
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
	user.Post("/chat/delete-room", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return DeleteChatRoom(c, DB)
	})

	user.Get("/admin/id-verifications/pending", middleware.AuthMiddlewareJSON(), func(c *fiber.Ctx) error {
		return ListPendingIDVerificationsForAdmin(c, DB)
	})
	user.Get("/admin/users", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return ListUsersForAdmin(c, DB)
	})
	user.Post("/admin/users/update", middleware.AuthMiddlewareJSON(), func(c *fiber.Ctx) error {
		return AdminUpdateUserJSON(c, DB)
	})
	user.Post("/admin/users/delete", middleware.AuthMiddlewareJSON(), func(c *fiber.Ctx) error {
		return AdminDeleteUserJSON(c, DB)
	})
	user.Post("/admin/approve-id", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return ApproveUserID(c, DB)
	})
	user.Post("/admin/reject-id", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return RejectUserID(c, DB)
	})

}
