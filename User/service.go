package user

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RequestOTP(c *fiber.Ctx, db *gorm.DB) error {

	user := new(User)
	user.OTP = rand.Intn(9000) + 1000
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	var existing User
	// Check if user with this phone exists
	if err := db.Where("phone_number = ?", user.PhoneNumber).First(&existing).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Not found → create new user
			if err := db.Create(user).Error; err != nil {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			return c.Status(201).JSON(fiber.Map{"msg": "OTP has been send to " + user.PhoneNumber})
		}
		// Some other DB error
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Found → update OTP
	existing.OTP = user.OTP
	if err := db.Save(&existing).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(202).JSON(fiber.Map{"msg": "OTP has been send to " + user.PhoneNumber})
}

func VerifyOTP(c *fiber.Ctx, db *gorm.DB) error {
	// get user otp
	var tmpUser User
	var foundUser User
	if err := c.BodyParser(&tmpUser); err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to retrieve OTP"})
	}

	// check for otp
	if tmpUser.OTP == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "Invalid OTP format"})
	}

	// check for the user with the phone number  and otp
	if err := db.Where("phone_number = ?", tmpUser.PhoneNumber).Where("otp = ?", tmpUser.OTP).First(&foundUser).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid phone number or OTP"})
	}

	// update the found user
	foundUser.OTP = 0
	if err := db.Save(&foundUser).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "Failed to clear the otp"})
	}

	return c.Status(202).JSON(fiber.Map{"msg": "otp verified successfully"})
}

func GetRooms(c *fiber.Ctx, db *gorm.DB, UserID uint) error {

	// var rooms []Room
	// if err := db.Preload("User").Find(&rooms).Error; err != nil {
	// 	return c.Status(400).JSON(fiber.Map{"msg": "failed to get rooms"})
	// }
	var user User

	type Chat struct {
		Id          uint   `json:"id"`
		User        User   `json:"user"`
		LastMessage string `json:"lastMessage"`
		NewMessages uint   `json:"newMessages"`
	}
	var chats []Chat
	if err := db.Preload("Rooms").Preload("Rooms.Users").First(&user, UserID).Error; err != nil {

		return c.Status(400).JSON(fiber.Map{"msg": "failed to get rooms"})
	}

	for _, room := range user.Rooms {

		var user = User{
			Name:     room.Users[1].Name,
			LastSeen: room.Users[1].LastSeen,
		}

		var LastMessage Message
		var NewMessages int64
		var messages []Message

		if err := db.Where("room_id = ?", room.ID).Last(&LastMessage).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get room last messages"})
		}

		if err := db.Where("room_id = ? AND is_read = ?", room.ID, false).Find(&messages).Count(&NewMessages).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get room messages count"})
		}

		chats = append(chats, Chat{
			Id:          room.ID,
			User:        user,
			LastMessage: LastMessage.Text,
			NewMessages: uint(NewMessages),
		})
	}

	return c.JSON(fiber.Map{"data": chats})

}

func GetRoomDetails(c *fiber.Ctx, db *gorm.DB, RoomID uint, UserID uint) error {

	var room Room

	if err := db.Preload("Users").First(&room, RoomID).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get room"})
	}
	// fmt.Println(room)

	// get messages for the room
	var messsages []Message
	if err := db.Where("room_id = ?", RoomID).Preload("Post").Preload("Post.User").Find(&messsages).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "failed to get room messages"})

	}

	type ChatContent struct {
		Text  string `json:"text"`
		Image string
		Post  Post
	}

	type ChatMessage struct {
		Id                 uint        `json:"id"`
		Content            ChatContent `json:"content"`
		CreatedAt          time.Time   `json:"createdAt"`
		SeenByRecipient    bool        `json:"seenByRecipient"`
		IsOwnedByRecipient bool        `json:"isOwnedByRecipient"`
	}

	type Chat struct {
		Id       uint          `json:"id"`
		User     User          `json:"user"`
		Messages []ChatMessage `json:"messages"`
	}

	var msgs []ChatMessage

	for _, m := range messsages {

		if m.Post.ID == 0 {
			m.Post = Post{}
		}

		msgs = append(msgs, ChatMessage{
			Id: m.ID,
			Content: ChatContent{
				Text:  m.Text,
				Image: m.Attachment,
			},
			CreatedAt:          m.CreatedAt,
			IsOwnedByRecipient: m.UserID == UserID,
			SeenByRecipient:    m.IsRead,
		})
	}

	fmt.Println(room.Users[1])

	chat := Chat{
		Id:       RoomID,
		User:     room.Users[1],
		Messages: msgs,
	}

	return c.JSON(fiber.Map{"data": chat})

}
