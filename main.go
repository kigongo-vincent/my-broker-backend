package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	post "github.com/kigongo-vincent/my-broker-backend/Post"
	user "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/db"
)

func main() {

	app := fiber.New()

	// enable cors
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	// load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load environment variables")
	}
	DB := db.ConnectToDB()

	app.Get("ps", func(c *fiber.Ctx) error {
		return c.SendStatus(202)
	})

	// routes
	user.RegisterRoutes(app, DB)
	post.RegisterRoutes(app, DB)

	app.Listen(":3000")
}
