package main

import (
	"fmt"
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
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
	"github.com/kigongo-vincent/my-broker-backend/wschat"
)

func main() {

	app := fiber.New()

	// enable cors
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// load env
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Failed to load environment variables", err)
		log.Fatal("Failed to load environment variables")
	}
	DB := db.ConnectToDB()

	app.Get("/health", func(c *fiber.Ctx) error {
		return fbcodec.SendEmpty(c, 200, "connected")
	})
	app.Static("/web", "./web")
	app.Get("/ws", func(c *fiber.Ctx) error {
		token := c.Query("token")
		if strings.TrimSpace(token) == "" {
			return fbcodec.SendError(c, 401, "missing token")
		}
		uid, err := core.ParseJWT(token)
		if err != nil {
			return fbcodec.SendError(c, 401, "invalid token")
		}
		c.Locals("wsUserID", uid)
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	}, websocket.New(func(conn *websocket.Conn) {
		wschat.ServeChat(conn, DB)
	}))

	// routes
	user.RegisterRoutes(app, DB)
	post.RegisterRoutes(app, DB)

	app.Listen(":3000")
}
