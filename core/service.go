package core

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
)

func ThrowNewError(c *fiber.Ctx, msg string) error {
	return fbcodec.SendError(c, 400, msg)
}

func Trim(s string, limit int) string {
	if len(s) > limit {
		return s[:limit-3] + "..."
	}
	return s
}
