package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	socketio "github.com/googollee/go-socket.io"
	"github.com/joho/godotenv"
	post "github.com/kigongo-vincent/my-broker-backend/Post"
	user "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/chat"
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

	// routes
	user.RegisterRoutes(app, DB)
	post.RegisterRoutes(app, DB)

	chatSvc := chat.NewService(DB)

	// REST routes (optional fallback for chat history)
	chat.RegisterRoutes(app, chatSvc)

	// Socket.IO
	server := socketio.NewServer(nil)
	chat.RegisterSocket(server, chatSvc)

	go func() {
		log.Fatal(server.Serve())
	}()
	defer server.Close()

	app.Listen(":3000")
}
