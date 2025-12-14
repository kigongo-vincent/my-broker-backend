package post

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	user "github.com/kigongo-vincent/my-broker-backend/User"
	"gorm.io/gorm"
)

func CreatePost(ctx *fiber.Ctx, db *gorm.DB) error {

	var post user.Post
	post.UserID = 1

	if err := ctx.BodyParser(&post); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	if err := db.Create(&post).First(&post).Error; err != nil {
		return ctx.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	return ctx.Status(201).JSON(fiber.Map{"msg": "post created successfully", "data": post})

}

func GetPaginatedPosts(c *fiber.Ctx, db *gorm.DB) error {

	var posts []user.Post
	var page int = 1
	var limit int = 10
	var total int64

	// get the controls
	page, pageErr := strconv.Atoi(c.Query("page"))
	limit, limitErr := strconv.Atoi(c.Query("limit"))

	if pageErr != nil || limitErr != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "page or limit missing"})
	}

	if err := db.Where("is_approved = ?", true).Limit(limit).Offset((page - 1) * limit).Preload("User").Find(&posts).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	if len(posts) == 0 {
		return c.Status(202).JSON(fiber.Map{"msg": "no posts found"})
	}

	if err := db.Model(&user.Post{}).Where("is_approved = ?", true).Count(&total).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	return c.JSON(fiber.Map{"data": posts, "total": total, "page": page, "limit": limit})
}
