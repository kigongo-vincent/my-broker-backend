package core

import "github.com/gofiber/fiber/v2"

func ThrowNewError(c *fiber.Ctx, msg string) error {
	return c.Status(400).JSON(fiber.Map{"msg": msg})
}

func Trim(s string, limit int) string {
	if len(s) > limit {
		return s[:limit-3] + "..."
	}
	return s
}
