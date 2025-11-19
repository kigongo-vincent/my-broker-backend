package main

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/gofiber/fiber/v2"
// 	"gorm.io/driver/postgres"
// 	"gorm.io/gorm"
// )

// // Models
// type Author struct {
// 	ID    uint   `gorm:"primaryKey" json:"id"`
// 	Name  string `json:"name"`
// 	Email string `json:"email" gorm:"unique"`
// }

// type Post struct {
// 	ID       uint   `gorm:"primaryKey" json:"id"`
// 	Title    string `json:"title"`
// 	Content  string `json:"content"`
// 	AuthorID uint   `json:"author_id"`
// 	Author   Author `json:"author" gorm:"foreignKey:AuthorID"`
// }

// var db *gorm.DB

// func main() {
// 	// Initialize database
// 	dsn := fmt.Sprintf(
// 		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
// 		getEnv("DB_HOST", "localhost"),
// 		getEnv("DB_USER", "postgres"),
// 		getEnv("DB_PASSWORD", "postgres"),
// 		getEnv("DB_NAME", "postsdb"),
// 		getEnv("DB_PORT", "5432"),
// 	)

// 	var err error
// 	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("Failed to connect to database:", err)
// 	}

// 	// Auto migrate
// 	db.AutoMigrate(&Author{}, &Post{})

// 	// Initialize Fiber app
// 	app := fiber.New()

// 	// Author routes
// 	app.Post("/authors", createAuthor)
// 	app.Get("/authors", getAuthors)
// 	app.Get("/authors/:id", getAuthor)
// 	app.Put("/authors/:id", updateAuthor)
// 	app.Delete("/authors/:id", deleteAuthor)

// 	// Post routes
// 	app.Post("/posts", createPost)
// 	app.Get("/posts", getPosts)
// 	app.Get("/posts/:id", getPost)
// 	app.Put("/posts/:id", updatePost)
// 	app.Delete("/posts/:id", deletePost)

// 	// Start server
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "3000"
// 	}
// 	log.Fatal(app.Listen(":" + port))
// }

// // Author handlers
// func createAuthor(c *fiber.Ctx) error {
// 	author := new(Author)
// 	if err := c.BodyParser(author); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	if err := db.Create(&author).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.Status(201).JSON(author)
// }

// func getAuthors(c *fiber.Ctx) error {
// 	var authors []Author
// 	db.Find(&authors)
// 	return c.JSON(authors)
// }

// func getAuthor(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var author Author

// 	if err := db.First(&author, id).Error; err != nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Author not found"})
// 	}

// 	return c.JSON(author)
// }

// func updateAuthor(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var author Author

// 	if err := db.First(&author, id).Error; err != nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Author not found"})
// 	}

// 	if err := c.BodyParser(&author); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	db.Save(&author)
// 	return c.JSON(author)
// }

// func deleteAuthor(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	if err := db.Delete(&Author{}, id).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.SendStatus(204)
// }

// // Helper function
// func getEnv(key, defaultValue string) string {
// 	value := os.Getenv(key)
// 	if value == "" {
// 		return defaultValue
// 	}
// 	return value
// }

// // Post handlers
// func createPost(c *fiber.Ctx) error {
// 	post := new(Post)
// 	if err := c.BodyParser(post); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	// Verify author exists
// 	var author Author
// 	if err := db.First(&author, post.AuthorID).Error; err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": "Author not found"})
// 	}

// 	if err := db.Create(&post).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	db.Preload("Author").First(&post, post.ID)
// 	return c.Status(201).JSON(post)
// }

// func getPosts(c *fiber.Ctx) error {
// 	var posts []Post
// 	db.Preload("Author").Find(&posts)
// 	return c.JSON(posts)
// }

// func getPost(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var post Post

// 	if err := db.Preload("Author").First(&post, id).Error; err != nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Post not found"})
// 	}

// 	return c.JSON(post)
// }

// func updatePost(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var post Post

// 	if err := db.First(&post, id).Error; err != nil {
// 		return c.Status(404).JSON(fiber.Map{"error": "Post not found"})
// 	}

// 	if err := c.BodyParser(&post); err != nil {
// 		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	db.Save(&post)
// 	db.Preload("Author").First(&post, post.ID)
// 	return c.JSON(post)
// }

// func deletePost(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	if err := db.Delete(&Post{}, id).Error; err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
// 	}

// 	return c.SendStatus(204)
// }
