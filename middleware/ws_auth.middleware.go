package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/kigongo-vincent/my-broker-backend/core"
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
)

// Locals key set on the HTTP upgrade request and read by the WebSocket handler.
const WSUserIDLocal = "wsUserID"

// WebSocketUpgradeAuth validates JWT from query ?token= and allows only WebSocket upgrades.
// Compatible with browsers and mobile clients that cannot send custom headers on WebSocket.
func WebSocketUpgradeAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := strings.TrimSpace(c.Query("token"))
		if token == "" {
			return fbcodec.SendError(c, fiber.StatusUnauthorized, "missing token")
		}
		uid, err := core.ParseJWT(token)
		if err != nil {
			return fbcodec.SendError(c, fiber.StatusUnauthorized, "invalid token")
		}
		c.Locals(WSUserIDLocal, uid)
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}
}
