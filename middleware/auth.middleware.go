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

// AuthMiddlewareJSON is the same as AuthMiddleware but returns JSON errors for JSON-only routes.
func AuthMiddlewareJSON() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "missing authorization token"})
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		userID, err := core.ParseJWT(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "invalid token"})
		}

		c.Locals("userID", userID)
		return c.Next()
	}
}

func OptionalAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization")
		if token == "" {
			return c.Next()
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		userID, err := core.ParseJWT(token)
		if err == nil && userID != 0 {
			c.Locals("userID", userID)
		}
		return c.Next()
	}
}
