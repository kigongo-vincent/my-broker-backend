package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// JWTValidator is an interface to allow dependency injection
type JWTValidator interface {
	ValidateToken(token string) (userID uint, err error)
}

// AuthMiddleware returns a Fiber middleware function
func AuthMiddleware(jwt JWTValidator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Get("Authorization") // expecting "Bearer <token>"
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
		}

		// strip "Bearer " prefix
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		userID, err := jwt.ValidateToken(token)
		if err != nil {
			log.Println("JWT validation failed:", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
		}

		// store userID in context
		c.Locals("userID", userID)
		return c.Next()
	}
}
