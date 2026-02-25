package user

import (
	"math/rand"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kigongo-vincent/my-broker-backend/core"
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
		return c.Status(400).JSON(fiber.Map{"msg": "failed to retrieve OTP" + err.Error()})
	}

	// check for otp
	if tmpUser.OTP == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "Invalid OTP format"})
	}

	// check for the user with the phone number  and otp
	if err := db.Where("phone_number = ? OR email = ?", tmpUser.PhoneNumber, tmpUser.Email).Where("otp = ?", tmpUser.OTP).First(&foundUser).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid credentials submitted" + err.Error()})
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

		var user = getParticipant(room.Users, UserID)

		var LastMessage Message
		var NewMessages int64
		var messages []Message

		if err := db.Where("room_id = ?", room.ID).Last(&LastMessage).Error; err != nil {
			// return c.Status(400).JSON(fiber.Map{"msg": "failed to get room last messages"})
		}

		if err := db.Where("room_id = ? AND is_read = ?", room.ID, false).Find(&messages).Count(&NewMessages).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get room messages count"})
		}

		chats = append(chats, Chat{
			Id:          room.ID,
			User:        user,
			LastMessage: core.Trim(LastMessage.Text, 30),
			NewMessages: uint(NewMessages),
		})
	}

	return c.JSON(fiber.Map{"data": chats})

}

func GetOrCreateRoomByParticipants(c *fiber.Ctx, db *gorm.DB, UserID uint, ParticipantID uint) error {

	// Step 1: Try to find a room with exactly these two participants
	var room Room
	err := db.Model(&Room{}).
		Joins("JOIN user_rooms ur1 ON ur1.room_id = rooms.id AND ur1.user_id = ?", UserID).
		Joins("JOIN user_rooms ur2 ON ur2.room_id = rooms.id AND ur2.user_id = ?", ParticipantID).
		Where("id IN (?)",
			db.Model(&UserRoom{}).
				Select("room_id").
				Group("room_id").
				Having("COUNT(user_id) = ?", 2),
		).
		Preload("Users").
		First(&room).Error

	// Step 2: If not found, create new room and add both users
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			newRoom := Room{}
			if err := db.Create(&newRoom).Error; err != nil {
				return core.ThrowNewError(c, "failed to create room")
			}

			rows := []map[string]any{
				{"room_id": newRoom.ID, "user_id": UserID},
				{"room_id": newRoom.ID, "user_id": ParticipantID},
			}

			if err := db.Table("user_rooms").Create(rows).Error; err != nil {
				return core.ThrowNewError(c, "failed to assign users to room")
			}

			room = newRoom
		} else {
			return core.ThrowNewError(c, "failed to fetch room")
		}
	}

	// Step 3: Fetch messages
	var messages []Message
	if err := db.Where("room_id = ?", room.ID).
		Preload("Post").
		Preload("Post.User").
		Order("id ASC").
		Find(&messages).Error; err != nil {
		// return core.ThrowNewError(c, "failed to get room messages")
	}

	// Prepare response
	type ChatContent struct {
		Text  string `json:"text"`
		Image string `json:"image"`
		Post  Post   `json:"post"`
	}

	type ChatMessage struct {
		Id                 uint        `json:"id"`
		Content            ChatContent `json:"content"`
		CreatedAt          time.Time   `json:"createdAt"`
		SeenByRecipient    bool        `json:"seenByRecipient"`
		IsOwnedByRecipient bool        `json:"isOwnedByRecipient"`
	}

	var chatMessages []ChatMessage
	if len(messages) != 0 {
		for _, m := range messages {
			chatMessages = append(chatMessages, ChatMessage{
				Id: m.ID,
				Content: ChatContent{
					Text:  m.Text,
					Image: m.Attachment,
					Post:  m.Post,
				},
				CreatedAt:          m.CreatedAt,
				IsOwnedByRecipient: m.UserID == UserID,
				SeenByRecipient:    m.IsRead,
			})
		}
	}

	chat := struct {
		Id       uint          `json:"id"`
		User     User          `json:"user"`
		Messages []ChatMessage `json:"messages"`
	}{
		Id:       room.ID,
		User:     getParticipant(room.Users, UserID),
		Messages: chatMessages,
	}

	return c.JSON(fiber.Map{"data": chat})
}

func getParticipant(users []User, UserID uint) User {
	for _, user := range users {
		if user.ID != UserID {
			return user
		}
	}
	return User{}
}

func UpdateProfile(c *fiber.Ctx, db *gorm.DB) error {

	type UserRequest struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Photo    string `json:"photo"`
	}

	var body UserRequest
	if err := c.BodyParser(&body); err != nil {
		return core.ThrowNewError(c, "failed to parse user data")
	}

	var user User
	if err := db.First(&user, body.ID).Error; err != nil {
		return core.ThrowNewError(c, "user not found")
	}

	if body.Username != "" {
		if err := db.Model(&user).Update("name", body.Username).Error; err != nil {
			return core.ThrowNewError(c, "failed to update username")
		}
	}

	if body.Photo != "" {
		if err := db.Model(&user).Update("photo", body.Photo).Error; err != nil {
			return core.ThrowNewError(c, "failed to update photo")
		}
	}

	return c.SendStatus(202)

}

func GetChatsAndFavourites(c *fiber.Ctx, db *gorm.DB, UserID uint) error {
	var favourites int
	var user User

	var chats []Chat
	if err := db.Preload("Rooms").Preload("Rooms.Users").Preload("Liked").Preload("Liked.Likers").First(&user, UserID).Error; err != nil {

		return c.Status(400).JSON(fiber.Map{"msg": "failed to get rooms"})
	}

	for _, room := range user.Rooms {

		var user = getParticipant(room.Users, UserID)

		var LastMessage Message
		var NewMessages int64
		var messages []Message

		if err := db.Where("room_id = ?", room.ID).Last(&LastMessage).Error; err != nil {
			// return c.Status(400).JSON(fiber.Map{"msg": "failed to get room last messages"})
		}

		if err := db.Where("room_id = ? AND is_read = ?", room.ID, false).Find(&messages).Count(&NewMessages).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": "failed to get room messages count"})
		}

		chats = append(chats, Chat{
			Id:          room.ID,
			User:        user,
			LastMessage: core.Trim(LastMessage.Text, 30),
			NewMessages: uint(NewMessages),
		})
	}

	favourites = len(user.Liked)
	data := map[string]any{
		"chats":      getUnreadMessagesCountForRooms(chats),
		"favourites": favourites,
	}

	return c.JSON(fiber.Map{"data": data})

}

func getUnreadMessagesCountForRooms(chats []Chat) int {
	var count int
	for _, chat := range chats {
		count += int(chat.NewMessages)
	}
	return count
}

func UpdateID(c *fiber.Ctx, db *gorm.DB) error {

	type Req struct {
		ID    uint   `json:"id"`
		Type  string `json:"type"` // phone or email
		Value string `json:"value"`
	}

	var req Req
	if err := c.BodyParser(&req); err != nil {
		return core.ThrowNewError(c, "failed to parse body")
	}

	var user User
	var column string
	if req.Type == "phone" {
		column = "phone_number"
	} else if req.Type == "email" {
		column = "email"
	}

	if err := db.First(&user, req.ID).Error; err != nil {
		return core.ThrowNewError(c, "failed to get user")
	}

	if err := db.Model(&user).Update("otp", rand.Intn(9000)+1000).Update(column, req.Value).Error; err != nil {
		return core.ThrowNewError(c, "failed to generate OTP")
	}

	return c.JSON(fiber.Map{"msg": "otp generated successfully"})

}
