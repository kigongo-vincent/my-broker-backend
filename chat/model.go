package chat

import (
	"time"
)

type Room struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `json:"name,omitempty"`           // optional for 1-to-1
	Members   []uint    `gorm:"type:json" json:"members"` // list of user IDs
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RoomID    uint      `json:"room_id"`
	SenderID  uint      `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
