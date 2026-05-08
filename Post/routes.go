package post

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kigongo-vincent/my-broker-backend/middleware"
	"gorm.io/gorm"
)

func RegisterRoutes(app *fiber.App, db *gorm.DB) {
	posts := app.Group("/posts")
	posts.Get("/favourites", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)
		return GetFavourites(c, db, userID)
	})
	posts.Get("/new-app", func(c *fiber.Ctx) error {
		return c.SendStatus(202)
	})
	posts.Post("/", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return CreatePost(c, db)
	})
	posts.Get("/", func(c *fiber.Ctx) error {
		return GetPaginatedPosts(c, db)
	})
	posts.Get("/update-like-status", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return UpdateLikerStatus(c, db)
	})
	posts.Get("/creator", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		userID := c.Locals("userID").(uint)
		return GetPostsByCreator(c, db, userID)
	})
	posts.Get("/post", middleware.OptionalAuthMiddleware(), func(c *fiber.Ctx) error {
		return GetPostDetails(c, db)
	})
	posts.Post("/mine/delete", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return DeleteMyPost(c, db)
	})
	posts.Post("/mine/update", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return UpdateMyPost(c, db)
	})
	posts.Get("/mine/availability", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return SetMyPostAvailability(c, db)
	})
	posts.Get("/posts-by-location", func(c *fiber.Ctx) error {
		return GetPostLocations(c, db)
	})
	posts.Get("/admin/pending", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return GetPendingPostsForAdmin(c, db)
	})
	posts.Post("/admin/approve", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return ApprovePostByAdmin(c, db)
	})
	posts.Post("/admin/reject", middleware.AuthMiddleware(), func(c *fiber.Ctx) error {
		return RejectPostByAdmin(c, db)
	})

}
