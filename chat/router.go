package chat

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, svc *Service) {
	chatRouter := app.Group("/chat")

	chatRouter.Get("/rooms/:id/messages", func(c *fiber.Ctx) error {
		roomID, _ := strconv.Atoi(c.Params("id"))
		msgs, err := svc.GetRoomMessages(uint(roomID))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(msgs)
	})
}
