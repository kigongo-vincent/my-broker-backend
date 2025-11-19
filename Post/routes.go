package post

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	posts := app.Group("/posts")
	posts.Post("/", func(c *fiber.Ctx) error {
		return CreatePost(c, db)
	})
}
