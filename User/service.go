package user

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
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
	normalizedPhone, err := core.NormalizeUGPhoneNumber(user.PhoneNumber)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}
	user.PhoneNumber = normalizedPhone

	var existing User
	lookup := db.Where("phone_number IN ?", core.UGPhoneCandidates(user.PhoneNumber)).Limit(1).Find(&existing)
	if lookup.Error != nil {
		return c.Status(500).JSON(fiber.Map{"error": lookup.Error.Error()})
	}

	if lookup.RowsAffected == 0 {
		// New user -> create, then send OTP
		if err := db.Create(user).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		if err := core.SendOTP(user.PhoneNumber, user.OTP); err != nil {
			return c.Status(201).JSON(fiber.Map{
				"msg":     "OTP generated for " + user.PhoneNumber,
				"warning": "sms delivery failed",
				"otp":     user.OTP,
			})
		}
		if os.Getenv("SMS_DRY_RUN") == "true" {
			return c.Status(201).JSON(fiber.Map{"msg": "OTP has been send to " + user.PhoneNumber, "otp": user.OTP})
		}
		return c.Status(201).JSON(fiber.Map{"msg": "OTP has been send to " + user.PhoneNumber})
	}

	// Existing user -> only regenerate OTP
	if existing.PhoneNumber != user.PhoneNumber {
		if err := db.Model(&existing).Update("phone_number", user.PhoneNumber).Error; err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
	}
	existing.OTP = user.OTP
	if err := db.Model(&existing).Update("otp", existing.OTP).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	if err := core.SendOTP(user.PhoneNumber, existing.OTP); err != nil {
		return c.Status(202).JSON(fiber.Map{
			"msg":     "OTP generated for " + user.PhoneNumber,
			"warning": "sms delivery failed",
			"otp":     existing.OTP,
		})
	}
	if os.Getenv("SMS_DRY_RUN") == "true" {
		return c.Status(202).JSON(fiber.Map{"msg": "OTP has been send to " + user.PhoneNumber, "otp": existing.OTP})
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

	lookup := db.Where("otp = ?", tmpUser.OTP)
	if tmpUser.PhoneNumber != "" {
		normalizedPhone, err := core.NormalizeUGPhoneNumber(tmpUser.PhoneNumber)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
		}
		lookup = lookup.Where("phone_number IN ?", core.UGPhoneCandidates(normalizedPhone))
	} else if tmpUser.Email != "" {
		lookup = lookup.Where("email = ?", tmpUser.Email)
	} else {
		return c.Status(400).JSON(fiber.Map{"msg": "phone number or email is required"})
	}

	// check for the user with the provided id (phone/email) and otp
	if err := lookup.First(&foundUser).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid credentials submitted" + err.Error()})
	}

	token, err := core.IssueJWT(foundUser.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to issue auth token"})
	}

	// update the found user (only after token was issued successfully)
	foundUser.OTP = 0
	if err := db.Save(&foundUser).Error; err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "Failed to clear the otp"})
	}

	return c.Status(200).JSON(fiber.Map{
		"msg":   "otp verified successfully",
		"token": token,
		"user":  foundUser,
	})
}

func GoogleSignin(c *fiber.Ctx, db *gorm.DB) error {
	type Req struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil || (req.AccessToken == "" && req.IDToken == "") {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid google payload"})
	}

	type GoogleProfile struct {
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	var profile GoogleProfile
	if req.AccessToken != "" {
		httpReq, _ := http.NewRequest(http.MethodGet, "https://www.googleapis.com/oauth2/v2/userinfo", nil)
		httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
		res, err := http.DefaultClient.Do(httpReq)
		if err != nil || res.StatusCode != http.StatusOK {
			return c.Status(401).JSON(fiber.Map{"msg": "failed to verify google access token"})
		}
		defer res.Body.Close()
		if err = json.NewDecoder(res.Body).Decode(&profile); err != nil || profile.Email == "" {
			return c.Status(401).JSON(fiber.Map{"msg": "invalid google profile response"})
		}
	}

	var foundUser User
	if err := db.Where("email = ?", profile.Email).First(&foundUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			foundUser = User{
				Name:        profile.Name,
				Email:       profile.Email,
				Photo:       profile.Picture,
				PhoneNumber: fmt.Sprintf("g_%d", time.Now().UnixNano()),
				Status:      "user",
			}
			if createErr := db.Create(&foundUser).Error; createErr != nil {
				return c.Status(500).JSON(fiber.Map{"msg": "failed to create google user"})
			}
		} else {
			return c.Status(500).JSON(fiber.Map{"msg": "failed to lookup user"})
		}
	}

	token, err := core.IssueJWT(foundUser.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to issue auth token"})
	}
	return c.Status(200).JSON(fiber.Map{"msg": "google signin successful", "token": token, "user": foundUser})
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
	_ = db.Model(&Message{}).
		Where("room_id = ? AND user_id <> ? AND is_read = ?", room.ID, UserID, false).
		Update("is_read", true).Error

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
		ID          uint    `json:"id"`
		Username    string  `json:"username"`
		Photo       string  `json:"photo"`
		ShowContact *bool   `json:"show_contact"`
		IsBroker    *bool   `json:"is_broker"`
		BrokerFees  *string `json:"broker_fees"`
		AcceptedPS  *bool   `json:"accepted_ps"`
	}

	var body UserRequest
	if err := c.BodyParser(&body); err != nil {
		return core.ThrowNewError(c, "failed to parse user data")
	}

	var user User
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 || authUserID != body.ID {
		return c.Status(403).JSON(fiber.Map{"msg": "forbidden"})
	}
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
	if body.ShowContact != nil {
		if err := db.Model(&user).Update("show_contact", *body.ShowContact).Error; err != nil {
			return core.ThrowNewError(c, "failed to update contact visibility")
		}
	}
	if body.IsBroker != nil {
		if err := db.Model(&user).Update("is_broker", *body.IsBroker).Error; err != nil {
			return core.ThrowNewError(c, "failed to update broker status")
		}
	}
	if body.BrokerFees != nil {
		if err := db.Model(&user).Update("broker_fees", *body.BrokerFees).Error; err != nil {
			return core.ThrowNewError(c, "failed to update broker fees")
		}
	}
	if body.AcceptedPS != nil {
		if err := db.Model(&user).Update("accepted_ps", *body.AcceptedPS).Error; err != nil {
			return core.ThrowNewError(c, "failed to update terms acceptance")
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
		normalizedPhone, err := core.NormalizeUGPhoneNumber(req.Value)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
		}
		req.Value = normalizedPhone
	} else if req.Type == "email" {
		column = "email"
	}

	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 || authUserID != req.ID {
		return c.Status(403).JSON(fiber.Map{"msg": "forbidden"})
	}

	if err := db.First(&user, req.ID).Error; err != nil {
		return core.ThrowNewError(c, "failed to get user")
	}

	if err := db.Model(&user).Update("otp", rand.Intn(9000)+1000).Update(column, req.Value).Error; err != nil {
		return core.ThrowNewError(c, "failed to generate OTP")
	}

	return c.JSON(fiber.Map{"msg": "otp generated successfully"})

}

func UpdateLastSeen(c *fiber.Ctx, db *gorm.DB) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(401).JSON(fiber.Map{"msg": "unauthorized"})
	}
	if err := db.Model(&User{}).Where("id = ?", userID).Update("last_seen", time.Now().UTC().Format(time.RFC3339)).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to update last seen"})
	}
	return c.SendStatus(202)
}

func GetProfileByID(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Query("user_id")
	if userID == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "user_id is required"})
	}

	var profile User
	if err := db.First(&profile, userID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"msg": "user not found"})
	}

	return c.JSON(fiber.Map{"data": profile})
}

func BlockUser(c *fiber.Ctx, db *gorm.DB) error {
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return c.Status(401).JSON(fiber.Map{"msg": "unauthorized"})
	}

	type Req struct {
		BlockedUserID uint `json:"blocked_user_id"`
		Blocked       bool `json:"blocked"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil || req.BlockedUserID == 0 || req.BlockedUserID == authUserID {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid block payload"})
	}

	if req.Blocked {
		record := BlockedUser{UserID: authUserID, BlockedUserID: req.BlockedUserID}
		if err := db.Where("user_id = ? AND blocked_user_id = ?", authUserID, req.BlockedUserID).FirstOrCreate(&record).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"msg": "failed to block user"})
		}
		return c.JSON(fiber.Map{"msg": "user blocked"})
	}

	if err := db.Where("user_id = ? AND blocked_user_id = ?", authUserID, req.BlockedUserID).Delete(&BlockedUser{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to unblock user"})
	}
	return c.JSON(fiber.Map{"msg": "user unblocked"})
}

func ReportUser(c *fiber.Ctx, db *gorm.DB) error {
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return c.Status(401).JSON(fiber.Map{"msg": "unauthorized"})
	}
	type Req struct {
		ReportedID uint   `json:"reported_id"`
		Reason     string `json:"reason"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil || req.ReportedID == 0 || req.ReportedID == authUserID {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid report payload"})
	}
	report := UserReport{
		ReporterID: authUserID,
		ReportedID: req.ReportedID,
		Reason:     strings.TrimSpace(req.Reason),
	}
	if report.Reason == "" {
		report.Reason = "no reason provided"
	}
	if err := db.Create(&report).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to submit report"})
	}
	return c.JSON(fiber.Map{"msg": "report submitted"})
}

func ClearChat(c *fiber.Ctx, db *gorm.DB) error {
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return c.Status(401).JSON(fiber.Map{"msg": "unauthorized"})
	}
	type Req struct {
		RoomID uint `json:"room_id"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil || req.RoomID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid room id"})
	}

	// Ensure user belongs to this room
	var rel UserRoom
	if err := db.Where("room_id = ? AND user_id = ?", req.RoomID, authUserID).First(&rel).Error; err != nil {
		return c.Status(403).JSON(fiber.Map{"msg": "forbidden"})
	}
	if err := db.Where("room_id = ?", req.RoomID).Delete(&Message{}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to clear chat"})
	}
	return c.JSON(fiber.Map{"msg": "chat cleared"})
}

func ListUsersForAdmin(c *fiber.Ctx, db *gorm.DB) error {
	adminID := c.Locals("userID").(uint)
	var admin User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return c.Status(403).JSON(fiber.Map{"msg": "admin access required"})
	}

	var users []User
	if err := db.Order("created_at DESC").Find(&users).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to fetch users"})
	}

	return c.JSON(fiber.Map{"data": users})
}

func ApproveUserID(c *fiber.Ctx, db *gorm.DB) error {
	adminID := c.Locals("userID").(uint)
	var admin User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return c.Status(403).JSON(fiber.Map{"msg": "admin access required"})
	}

	type Req struct {
		UserID uint `json:"user_id"`
	}
	var req Req
	if err := c.BodyParser(&req); err != nil || req.UserID == 0 {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid user id"})
	}

	if err := db.Model(&User{}).Where("id = ?", req.UserID).Update("verified", "true").Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to approve user id"})
	}
	return c.JSON(fiber.Map{"msg": "user id approved"})
}
