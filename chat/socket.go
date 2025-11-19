package chat

import (
	"log"
	"strconv"

	socketio "github.com/googollee/go-socket.io"
)

func RegisterSocket(server *socketio.Server, svc *Service) {
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		log.Println("connected:", s.ID())
		return nil
	})

	// Join room
	server.OnEvent("/", "join_room", func(s socketio.Conn, roomID uint) {
		s.Join(strconv.Itoa(int(roomID)))
		log.Println("User joined room:", roomID)
	})

	// Send message
	server.OnEvent("/", "send_message", func(s socketio.Conn, roomID uint, senderID uint, content string) {
		msg := &Message{
			RoomID:   roomID,
			SenderID: senderID,
			Content:  content,
		}

		if err := svc.CreateMessage(msg); err != nil {
			log.Println("Error saving message:", err)
			return
		}

		// Broadcast to all clients in room
		server.BroadcastToRoom("/", strconv.Itoa(int(roomID)), "new_message", msg)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		log.Println("disconnected:", s.ID(), reason)
	})
}
