package chat

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/kigongo-vincent/my-broker-backend/middleware"
	"github.com/kigongo-vincent/my-broker-backend/wschat"
	"gorm.io/gorm"
)

// RegisterWebSocket mounts the real-time chat endpoint using Fiber's websocket middleware.
// Clients connect with: GET /ws?token=<JWT> (see wschat.ServeChat for message JSON shape).
func RegisterWebSocket(app *fiber.App, db *gorm.DB) {
	app.Get("/ws", middleware.WebSocketUpgradeAuth(), websocket.New(func(conn *websocket.Conn) {
		wschat.ServeChat(conn, db)
	}))
}
