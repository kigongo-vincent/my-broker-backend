package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	post "github.com/kigongo-vincent/my-broker-backend/Post"
	user "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/chat"
	"github.com/kigongo-vincent/my-broker-backend/config"
	"github.com/kigongo-vincent/my-broker-backend/db"
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
)

func main() {
	// Load backend/.env (or cwd .env) before db, S3, SMS, JWT, etc. read the environment.
	config.LoadDotenv()

	app := fiber.New()

	// enable cors
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	DB := db.ConnectToDB()

	app.Get("/health", func(c *fiber.Ctx) error {
		return fbcodec.SendEmpty(c, 200, "connected")
	})
	app.Static("/web", "./web")
	chat.RegisterWebSocket(app, DB)

	// routes
	user.RegisterRoutes(app, DB)
	post.RegisterRoutes(app, DB)

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "3000"
	}
	port = strings.TrimPrefix(port, ":")
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
