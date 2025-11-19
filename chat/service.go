package chat

import (
	"gorm.io/gorm"
)

type Service struct {
	DB *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{DB: db}
}

// Create message in DB
func (s *Service) CreateMessage(msg *Message) error {
	return s.DB.Create(msg).Error
}

// Get all messages for a room
func (s *Service) GetRoomMessages(roomID uint) ([]Message, error) {
	var messages []Message
	err := s.DB.Where("room_id = ?", roomID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

// Get or create 1-to-1 room between two users
func (s *Service) GetOrCreateRoom(userIDs []uint) (*Room, error) {
	var room Room
	err := s.DB.Where("members @> ?", userIDs).First(&room).Error
	if err == gorm.ErrRecordNotFound {
		room = Room{Members: userIDs}
		err = s.DB.Create(&room).Error
	}
	return &room, err
}
