package post

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func CreatePost(ctx *fiber.Ctx, db *gorm.DB) error {

	var post Post
	post.UserID = 1

	if err := ctx.BodyParser(&post); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	if err := db.Create(&post).First(&post).Error; err != nil {
		return ctx.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	return ctx.Status(201).JSON(fiber.Map{"msg": "post created successfully", "data": post})

}
