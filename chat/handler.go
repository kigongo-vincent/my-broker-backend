package chat

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	Service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{Service: s}
}

// GET /rooms/:id/messages
func (h *Handler) GetRoomMessages(c *fiber.Ctx) error {
	roomIDParam := c.Params("id")
	roomID, err := strconv.Atoi(roomIDParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid room ID"})
	}

	messages, err := h.Service.GetRoomMessages(uint(roomID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(messages)
}

// POST /rooms/:id/messages
func (h *Handler) SendMessage(c *fiber.Ctx) error {
	roomIDParam := c.Params("id")
	roomID, _ := strconv.Atoi(roomIDParam)

	var payload struct {
		SenderID uint   `json:"sender_id"`
		Content  string `json:"content"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	msg := Message{
		RoomID:   uint(roomID),
		SenderID: payload.SenderID,
		Content:  payload.Content,
	}

	if err := h.Service.CreateMessage(&msg); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(msg)
}
