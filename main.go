package main

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/websocket/v2"
	"github.com/joho/godotenv"
	post "github.com/kigongo-vincent/my-broker-backend/Post"
	user "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/core"
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

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"msg": "connected"})
	})
	app.Static("/web", "./web")
	app.Get("/ws", func(c *fiber.Ctx) error {
		token := c.Query("token")
		if strings.TrimSpace(token) == "" {
			return c.Status(401).JSON(fiber.Map{"msg": "missing token"})
		}
		if _, err := core.ParseJWT(token); err != nil {
			return c.Status(401).JSON(fiber.Map{"msg": "invalid token"})
		}
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}, websocket.New(func(conn *websocket.Conn) {
		defer conn.Close()
		for {
			messageType, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if err = conn.WriteMessage(messageType, msg); err != nil {
				break
			}
		}
	}))

	// routes
	user.RegisterRoutes(app, DB)
	post.RegisterRoutes(app, DB)

	app.Listen(":3000")
}
