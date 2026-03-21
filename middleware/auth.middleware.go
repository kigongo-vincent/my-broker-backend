package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kigongo-vincent/my-broker-backend/core"
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
)

func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return fbcodec.SendError(c, fiber.StatusUnauthorized, "missing authorization token")
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		userID, err := core.ParseJWT(token)
		if err != nil {
			return fbcodec.SendError(c, fiber.StatusUnauthorized, "invalid token")
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}
