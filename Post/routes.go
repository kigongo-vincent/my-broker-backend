package post

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kigongo-vincent/my-broker-backend/core"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	posts := app.Group("/posts")
	posts.Get("/favourites", func(c *fiber.Ctx) error {
		UserID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
		if UserID == 0 {
			return core.ThrowNewError(c, "failed to get user")
		}
		return GetFavourites(c, db, uint(UserID))
	})
	posts.Get("/new-app", func(c *fiber.Ctx) error {
		return c.SendStatus(202)
	})
	posts.Post("/", func(c *fiber.Ctx) error {
		return CreatePost(c, db)
	})
	posts.Get("/", func(c *fiber.Ctx) error {
		return GetPaginatedPosts(c, db)
	})
	posts.Get("/update-like-status", func(c *fiber.Ctx) error {
		return UpdateLikerStatus(c, db)
	})
	posts.Get("/creator", func(c *fiber.Ctx) error {
		return GetPostsByCreator(c, db)
	})
	posts.Get("/post", func(c *fiber.Ctx) error {
		return GetPostDetails(c, db)
	})
	posts.Get("/posts-by-location", func(c *fiber.Ctx) error {
		return GetPostLocations(c, db)
	})

}
