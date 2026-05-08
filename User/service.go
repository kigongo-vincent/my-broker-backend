package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/kigongo-vincent/my-broker-backend/core"
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
	"github.com/kigongo-vincent/my-broker-backend/fbs/gen/mybroker"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type GoogleProfile struct {
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// roomChatAggregates loads last message per room and unread counts for viewerUserID.
// Unread = messages the other person sent that viewer has not read yet (never count viewer's own messages).
func roomChatAggregates(db *gorm.DB, roomIDs []uint, viewerUserID uint) (lastByRoom map[uint]Message, unreadByRoom map[uint]int64, err error) {
	lastByRoom = make(map[uint]Message)
	unreadByRoom = make(map[uint]int64)
	if len(roomIDs) == 0 {
		return lastByRoom, unreadByRoom, nil
	}
	var lasts []Message
	err = db.Raw(`
		SELECT DISTINCT ON (room_id) messages.*
		FROM messages
		WHERE room_id IN ?
		ORDER BY room_id, id DESC
	`, roomIDs).Scan(&lasts).Error
	if err != nil {
		return nil, nil, err
	}
	for i := range lasts {
		lastByRoom[lasts[i].RoomID] = lasts[i]
	}
	type cntRow struct {
		RoomID uint
		Cnt    int64
	}
	var counts []cntRow
	err = db.Model(&Message{}).
		Select("room_id, COUNT(*) as cnt").
		Where("room_id IN ? AND is_read = ? AND user_id <> ?", roomIDs, false, viewerUserID).
		Group("room_id").
		Scan(&counts).Error
	if err != nil {
		return nil, nil, err
	}
	for _, r := range counts {
		unreadByRoom[r.RoomID] = r.Cnt
	}
	return lastByRoom, unreadByRoom, nil
}

func validatePIN(pin string) error {
	if len(pin) < 4 || len(pin) > 8 {
		return fmt.Errorf("PIN must be 4–8 digits")
	}
	for _, r := range pin {
		if r < '0' || r > '9' {
			return fmt.Errorf("PIN must be numeric")
		}
	}
	return nil
}

// CheckPhoneForPin tells the client whether to collect a new PIN (201) or sign in with PIN (202). No SMS / OTP.
func CheckPhoneForPin(c *fiber.Ctx, db *gorm.DB) error {
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	phoneRaw, err := ParseSignInPhone(req)
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	normalizedPhone, err := core.NormalizeUGPhoneNumber(phoneRaw)
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	var existing User
	lookup := db.Where("phone_number IN ?", core.UGPhoneCandidates(normalizedPhone)).Limit(1).Find(&existing)
	if lookup.Error != nil {
		return fbcodec.SendError(c, 500, lookup.Error.Error())
	}
	if lookup.RowsAffected == 0 {
		return fbcodec.SendOTP(c, 201, "Create a PIN for this number", "", 0)
	}
	if existing.PhoneNumber != normalizedPhone {
		if err := db.Model(&existing).Update("phone_number", normalizedPhone).Error; err != nil {
			return fbcodec.SendError(c, 400, err.Error())
		}
	}
	if existing.PinHash == "" {
		return fbcodec.SendOTP(c, 201, "Set a PIN for your account", "", 0)
	}
	return fbcodec.SendOTP(c, 202, "Enter your PIN", "", 0)
}

// CompletePhonePin registers (pin + matching confirm), sets first PIN on legacy users, or logs in (pin only).
func CompletePhonePin(c *fiber.Ctx, db *gorm.DB) error {
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	p, err := ParsePhonePin(req)
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	phoneRaw := string(p.PhoneNumber())
	pin := string(p.Pin())
	confirm := string(p.PinConfirm())
	if err := validatePIN(pin); err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	normalizedPhone, err := core.NormalizeUGPhoneNumber(phoneRaw)
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	cands := core.UGPhoneCandidates(normalizedPhone)

	if confirm != "" {
		if err := validatePIN(confirm); err != nil {
			return fbcodec.SendError(c, 400, err.Error())
		}
		if pin != confirm {
			return fbcodec.SendError(c, 400, "PINs do not match")
		}
		var u User
		err := db.Where("phone_number IN ?", cands).First(&u).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hash, herr := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
			if herr != nil {
				return fbcodec.SendError(c, 500, "could not set PIN")
			}
			u = User{
				PhoneNumber: normalizedPhone,
				PinHash:     string(hash),
				Verified:    "false",
			}
			if err := db.Create(&u).Error; err != nil {
				return fbcodec.SendError(c, 400, err.Error())
			}
			token, terr := core.IssueJWT(u.ID)
			if terr != nil {
				return fbcodec.SendError(c, 500, "failed to issue auth token")
			}
			return fbcodec.SendAuthOK(c, "welcome", token, UserToIn(u))
		}
		if err != nil {
			return fbcodec.SendError(c, 500, err.Error())
		}
		if u.PinHash != "" {
			return fbcodec.SendError(c, 400, "this number already has a PIN; sign in instead")
		}
		hash, herr := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
		if herr != nil {
			return fbcodec.SendError(c, 500, "could not set PIN")
		}
		if err := db.Model(&u).Update("pin_hash", string(hash)).Error; err != nil {
			return fbcodec.SendError(c, 500, err.Error())
		}
		u.PinHash = string(hash)
		token, terr := core.IssueJWT(u.ID)
		if terr != nil {
			return fbcodec.SendError(c, 500, "failed to issue auth token")
		}
		return fbcodec.SendAuthOK(c, "PIN set successfully", token, UserToIn(u))
	}

	var u User
	if err := db.Where("phone_number IN ?", cands).First(&u).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fbcodec.SendError(c, 400, "no account for this number")
		}
		return fbcodec.SendError(c, 500, err.Error())
	}
	if u.PinHash == "" {
		return fbcodec.SendError(c, 400, "set your PIN first")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PinHash), []byte(pin)); err != nil {
		return fbcodec.SendError(c, 400, "wrong PIN")
	}
	token, terr := core.IssueJWT(u.ID)
	if terr != nil {
		return fbcodec.SendError(c, 500, "failed to issue auth token")
	}
	return fbcodec.SendAuthOK(c, "signed in", token, UserToIn(u))
}

func GoogleSignin(c *fiber.Ctx, db *gorm.DB) error {
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, "invalid google payload")
	}
	g, err := ParseGoogleAuth(req)
	if err != nil {
		return fbcodec.SendError(c, 400, "invalid google payload")
	}
	idTok := string(g.IdToken())
	if idTok == "" {
		return fbcodec.SendError(c, 400, "id token is required")
	}
	clientIDs := configuredGoogleClientIDs()
	if len(clientIDs) == 0 {
		return fbcodec.SendError(c, 500, "google oauth is not configured")
	}

	var profile GoogleProfile

	verifiedProfile, err := verifyGoogleIDToken(idTok, clientIDs)
	if err != nil {
		return fbcodec.SendError(c, 401, "failed to verify google id token")
	}
	profile = verifiedProfile
	if profile.Email == "" {
		return fbcodec.SendError(c, 401, "google account email not available")
	}
	foundUser, err := upsertGoogleUser(db, profile)
	if err != nil {
		return fbcodec.SendError(c, 500, "failed to lookup or create google user")
	}
	token, err := core.IssueJWT(foundUser.ID)
	if err != nil {
		return fbcodec.SendError(c, 500, "failed to issue auth token")
	}
	return fbcodec.SendAuthOK(c, "google signin successful", token, UserToIn(foundUser))
}

// upsertGoogleUser finds or creates a user from a verified Google profile (email verified upstream).
func upsertGoogleUser(db *gorm.DB, profile GoogleProfile) (User, error) {
	var foundUser User
	email := strings.TrimSpace(strings.ToLower(profile.Email))
	emailPtr := &email
	createCandidate := User{
		Name:        profile.Name,
		Email:       emailPtr,
		Photo:       profile.Picture,
		PhoneNumber: fmt.Sprintf("g_%d", time.Now().UnixNano()),
		Status:      "user",
		Verified:    "false",
	}
	if err := db.Where("email = ?", email).Attrs(createCandidate).FirstOrCreate(&foundUser).Error; err != nil {
		return User{}, err
	}
	return foundUser, nil
}

func configuredGoogleClientIDs() []string {
	keys := []string{
		"GOOGLE_WEB_CLIENT_ID",
		"GOOGLE_IOS_CLIENT_ID",
		"GOOGLE_ANDROID_CLIENT_ID",
	}
	ids := make([]string, 0, len(keys))
	for _, k := range keys {
		v := strings.TrimSpace(os.Getenv(k))
		if v != "" {
			ids = append(ids, v)
		}
	}
	return ids
}

func verifyGoogleIDToken(idToken string, allowedAudiences []string) (GoogleProfile, error) {
	endpoint := "https://oauth2.googleapis.com/tokeninfo?id_token=" + url.QueryEscape(idToken)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return GoogleProfile{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return GoogleProfile{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return GoogleProfile{}, fmt.Errorf("tokeninfo returned %d", res.StatusCode)
	}
	type tokenInfo struct {
		Aud           string `json:"aud"`
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}
	var info tokenInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return GoogleProfile{}, err
	}
	matchesAudience := false
	for _, aud := range allowedAudiences {
		if info.Aud == aud {
			matchesAudience = true
			break
		}
	}
	if !matchesAudience {
		return GoogleProfile{}, fmt.Errorf("id token audience mismatch")
	}
	if info.Email == "" || info.EmailVerified != "true" {
		return GoogleProfile{}, fmt.Errorf("email not verified")
	}
	return GoogleProfile{Email: info.Email, Name: info.Name, Picture: info.Picture}, nil
}

func GetRooms(c *fiber.Ctx, db *gorm.DB, UserID uint) error {
	var uu User
	if err := db.Preload("Rooms").Preload("Rooms.Users").First(&uu, UserID).Error; err != nil {
		return fbcodec.SendError(c, 400, "failed to get rooms")
	}
	roomIDs := make([]uint, 0, len(uu.Rooms))
	for _, r := range uu.Rooms {
		roomIDs = append(roomIDs, r.ID)
	}
	lastBy, unreadBy, err := roomChatAggregates(db, roomIDs, UserID)
	if err != nil {
		return fbcodec.SendError(c, 400, "failed to get room messages count")
	}
	chats := make([]Chat, 0, len(uu.Rooms))
	for _, room := range uu.Rooms {
		peer := getParticipant(room.Users, UserID)
		lm := lastBy[room.ID]
		chats = append(chats, Chat{
			Id:          room.ID,
			User:        peer,
			LastMessage: core.Trim(lm.Text, 30),
			NewMessages: uint(unreadBy[room.ID]),
		})
	}
	chatRows := make([]fbcodec.ChatIn, len(chats))
	for i := range chats {
		chatRows[i] = ChatToIn(chats[i])
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadChatRowList, 65536, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildChatRowList(b, chatRows)
	})
}

func GetOrCreateRoomByParticipants(c *fiber.Ctx, db *gorm.DB, UserID uint, ParticipantID uint) error {
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
	_ = db.Model(&Message{}).
		Where("room_id = ? AND user_id <> ? AND is_read = ?", room.ID, UserID, false).
		Update("is_read", true).Error
	var messages []Message
	_ = db.Where("room_id = ?", room.ID).
		Preload("Post").
		Preload("Post.User").
		Order("id ASC").
		Find(&messages).Error
	rows := make([]fbcodec.ChatMsgRow, 0, len(messages))
	for _, m := range messages {
		rows = append(rows, fbcodec.ChatMsgRow{
			ID: m.ID, Text: m.Text, Image: m.Attachment, Post: PostToIn(m.Post),
			CreatedAtUnixMs: m.CreatedAt.UnixMilli(), SeenByRecipient: m.IsRead,
			// true = incoming for the authenticated user (peer sent this message)
			IsOwnedByRecipient: m.UserID != UserID,
		})
	}
	peer := getParticipant(room.Users, UserID)
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadChatDetail, 262144, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildChatDetail(b, room.ID, UserToIn(peer), rows)
	})
}

func getParticipant(users []User, UserID uint) User {
	for _, u := range users {
		if u.ID != UserID {
			return u
		}
	}
	return User{}
}

func UpdateProfile(c *fiber.Ctx, db *gorm.DB) error {
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return core.ThrowNewError(c, "failed to parse user data")
	}
	body, err := ParseUpdateProfile(req)
	if err != nil {
		return core.ThrowNewError(c, "failed to parse user data")
	}
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 || authUserID != uint(body.Id()) {
		return fbcodec.SendError(c, 403, "forbidden")
	}
	var uu User
	if err := db.First(&uu, body.Id()).Error; err != nil {
		return core.ThrowNewError(c, "user not found")
	}
	if un := string(body.Username()); un != "" {
		if err := db.Model(&uu).Update("name", un).Error; err != nil {
			return core.ThrowNewError(c, "failed to update username")
		}
	}
	if ph := string(body.Photo()); ph != "" {
		if err := db.Model(&uu).Update("photo", ph).Error; err != nil {
			return core.ThrowNewError(c, "failed to update photo")
		}
	}
	if body.ShowContactSet() {
		if err := db.Model(&uu).Update("show_contact", body.ShowContact()).Error; err != nil {
			return core.ThrowNewError(c, "failed to update contact visibility")
		}
	}
	if body.IsBrokerSet() {
		if err := db.Model(&uu).Update("is_broker", body.IsBroker()).Error; err != nil {
			return core.ThrowNewError(c, "failed to update broker status")
		}
	}
	if body.BrokerFeesSet() {
		if err := db.Model(&uu).Update("broker_fees", string(body.BrokerFees())).Error; err != nil {
			return core.ThrowNewError(c, "failed to update broker fees")
		}
	}
	if body.AcceptedPsSet() {
		if err := db.Model(&uu).Update("accepted_ps", body.AcceptedPs()).Error; err != nil {
			return core.ThrowNewError(c, "failed to update terms acceptance")
		}
	}
	return fbcodec.SendEmpty(c, 202, "")
}

func GetChatsAndFavourites(c *fiber.Ctx, db *gorm.DB, UserID uint) error {
	var uu User
	if err := db.Preload("Rooms").Preload("Rooms.Users").Preload("Liked").Preload("Liked.Likers").First(&uu, UserID).Error; err != nil {
		return fbcodec.SendError(c, 400, "failed to get rooms")
	}
	roomIDs := make([]uint, 0, len(uu.Rooms))
	for _, r := range uu.Rooms {
		roomIDs = append(roomIDs, r.ID)
	}
	lastBy, unreadBy, err := roomChatAggregates(db, roomIDs, UserID)
	if err != nil {
		return fbcodec.SendError(c, 400, "failed to get room messages count")
	}
	chats := make([]Chat, 0, len(uu.Rooms))
	for _, room := range uu.Rooms {
		peer := getParticipant(room.Users, UserID)
		lm := lastBy[room.ID]
		chats = append(chats, Chat{
			Id:          room.ID,
			User:        peer,
			LastMessage: core.Trim(lm.Text, 30),
			NewMessages: uint(unreadBy[room.ID]),
		})
	}
	unreadTotal := int32(0)
	for _, ch := range chats {
		unreadTotal += int32(ch.NewMessages)
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadChatsAndFavs, 4096, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildChatsAndFavs(b, int(unreadTotal), len(uu.Liked))
	})
}

func UpdateID(c *fiber.Ctx, db *gorm.DB) error {
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return core.ThrowNewError(c, "failed to parse body")
	}
	body, err := ParseUpdateID(req)
	if err != nil {
		return core.ThrowNewError(c, "failed to parse body")
	}
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 || authUserID != uint(body.Id()) {
		return fbcodec.SendError(c, 403, "forbidden")
	}
	var uu User
	if err := db.First(&uu, body.Id()).Error; err != nil {
		return core.ThrowNewError(c, "failed to get user")
	}
	val := string(body.Value())
	column := ""
	switch string(body.Type()) {
	case "phone":
		column = "phone_number"
		normalizedPhone, err := core.NormalizeUGPhoneNumber(val)
		if err != nil {
			return fbcodec.SendError(c, 400, err.Error())
		}
		val = normalizedPhone
	case "email":
		column = "email"
	default:
		return core.ThrowNewError(c, "failed to parse body")
	}
	if err := db.Model(&uu).Update(column, val).Error; err != nil {
		return core.ThrowNewError(c, "failed to update profile")
	}
	return fbcodec.SendEmpty(c, 200, "updated successfully")
}

func UpdateLastSeen(c *fiber.Ctx, db *gorm.DB) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	if err := db.Model(&User{}).Where("id = ?", userID).Update("last_seen", time.Now().UTC().Format(time.RFC3339)).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to update last seen")
	}
	return fbcodec.SendEmpty(c, 202, "")
}

// SyncSession returns the current user row for the JWT bearer so clients can refresh local profile after cold start.
func SyncSession(c *fiber.Ctx, db *gorm.DB) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	var u User
	if err := db.First(&u, userID).Error; err != nil {
		return fbcodec.SendError(c, 401, "user not found")
	}
	auth := strings.TrimSpace(c.Get("Authorization"))
	auth = strings.TrimPrefix(auth, "Bearer ")
	auth = strings.TrimSpace(auth)
	return fbcodec.SendAuthOK(c, "session synchronized", auth, UserToIn(u))
}

func GetProfileByID(c *fiber.Ctx, db *gorm.DB) error {
	userID := c.Query("user_id")
	if userID == "" {
		return fbcodec.SendError(c, 400, "user_id is required")
	}
	var profile User
	if err := db.First(&profile, userID).Error; err != nil {
		return fbcodec.SendError(c, 404, "user not found")
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadAuthOk, 8192, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildAuthOk(b, UserToIn(profile))
	})
}

func BlockUser(c *fiber.Ctx, db *gorm.DB) error {
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, "invalid block payload")
	}
	body, err := ParseBlock(req)
	if err != nil || body.BlockedUserId() == 0 || uint(body.BlockedUserId()) == authUserID {
		return fbcodec.SendError(c, 400, "invalid block payload")
	}
	if body.Blocked() {
		record := BlockedUser{UserID: authUserID, BlockedUserID: uint(body.BlockedUserId())}
		if err := db.Where("user_id = ? AND blocked_user_id = ?", authUserID, body.BlockedUserId()).FirstOrCreate(&record).Error; err != nil {
			return fbcodec.SendError(c, 500, "failed to block user")
		}
		return fbcodec.SendEmpty(c, 200, "user blocked")
	}
	if err := db.Where("user_id = ? AND blocked_user_id = ?", authUserID, body.BlockedUserId()).Delete(&BlockedUser{}).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to unblock user")
	}
	return fbcodec.SendEmpty(c, 200, "user unblocked")
}

func ReportUser(c *fiber.Ctx, db *gorm.DB) error {
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, "invalid report payload")
	}
	body, err := ParseReport(req)
	if err != nil || body.ReportedId() == 0 || uint(body.ReportedId()) == authUserID {
		return fbcodec.SendError(c, 400, "invalid report payload")
	}
	reason := strings.TrimSpace(string(body.Reason()))
	if reason == "" {
		reason = "no reason provided"
	}
	report := UserReport{ReporterID: authUserID, ReportedID: uint(body.ReportedId()), Reason: reason}
	if err := db.Create(&report).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to submit report")
	}
	return fbcodec.SendEmpty(c, 200, "report submitted")
}

func ClearChat(c *fiber.Ctx, db *gorm.DB) error {
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, "invalid room id")
	}
	body, err := ParseRoomID(req)
	if err != nil || body.RoomId() == 0 {
		return fbcodec.SendError(c, 400, "invalid room id")
	}
	var rel UserRoom
	if err := db.Where("room_id = ? AND user_id = ?", body.RoomId(), authUserID).First(&rel).Error; err != nil {
		return fbcodec.SendError(c, 403, "forbidden")
	}
	if err := db.Where("room_id = ?", body.RoomId()).Delete(&Message{}).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to clear chat")
	}
	return fbcodec.SendEmpty(c, 200, "chat cleared")
}

func DeleteChatRoom(c *fiber.Ctx, db *gorm.DB) error {
	authUserID, ok := c.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, "invalid room id")
	}
	body, err := ParseRoomID(req)
	if err != nil || body.RoomId() == 0 {
		return fbcodec.SendError(c, 400, "invalid room id")
	}
	roomID := uint(body.RoomId())
	var rel UserRoom
	if err := db.Where("room_id = ? AND user_id = ?", roomID, authUserID).First(&rel).Error; err != nil {
		return fbcodec.SendError(c, 403, "forbidden")
	}
	if err := db.Where("room_id = ? AND user_id = ?", roomID, authUserID).Delete(&UserRoom{}).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to leave room")
	}
	var remaining int64
	if err := db.Model(&UserRoom{}).Where("room_id = ?", roomID).Count(&remaining).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to inspect room members")
	}
	if remaining == 0 {
		if err := db.Where("room_id = ?", roomID).Delete(&Message{}).Error; err != nil {
			return fbcodec.SendError(c, 500, "failed to delete room messages")
		}
		if err := db.Delete(&Room{}, roomID).Error; err != nil {
			return fbcodec.SendError(c, 500, "failed to delete room")
		}
	}
	return fbcodec.SendEmpty(c, 200, "room deleted")
}

func ListUsersForAdmin(c *fiber.Ctx, db *gorm.DB) error {
	adminID := c.Locals("userID").(uint)
	var admin User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return fbcodec.SendError(c, 403, "admin access required")
	}
	page := 1
	limit := 30
	if q := c.Query("page"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil || n < 1 {
			return fbcodec.SendError(c, 400, "invalid page")
		}
		page = n
	}
	if q := c.Query("limit"); q != "" {
		n, err := strconv.Atoi(q)
		if err != nil || n < 1 {
			return fbcodec.SendError(c, 400, "invalid limit")
		}
		if n > 100 {
			n = 100
		}
		limit = n
	}
	var users []User
	if err := db.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to fetch users")
	}
	userPins := make([]fbcodec.UserIn, len(users))
	for i := range users {
		userPins[i] = UserToIn(users[i])
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadUsersList, 262144, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildUsersList(b, userPins)
	})
}

func ApproveUserID(c *fiber.Ctx, db *gorm.DB) error {
	adminID := c.Locals("userID").(uint)
	var admin User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return fbcodec.SendError(c, 403, "admin access required")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, "invalid user id")
	}
	body, err := ParseApproveUser(req)
	if err != nil || body.UserId() == 0 {
		return fbcodec.SendError(c, 400, "invalid user id")
	}
	if err := db.Model(&User{}).Where("id = ?", body.UserId()).Updates(map[string]any{
		"verified":               "true",
		"id_verification_status": "approved",
	}).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to approve user id")
	}
	return fbcodec.SendEmpty(c, 200, "user id approved")
}

func RejectUserID(c *fiber.Ctx, db *gorm.DB) error {
	adminID := c.Locals("userID").(uint)
	var admin User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return fbcodec.SendError(c, 403, "admin access required")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, "invalid user id")
	}
	body, err := ParseApproveUser(req)
	if err != nil || body.UserId() == 0 {
		return fbcodec.SendError(c, 400, "invalid user id")
	}
	if err := db.Model(&User{}).Where("id = ?", body.UserId()).Updates(map[string]any{
		"verified":               "false",
		"id_verification_status": "rejected",
	}).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to reject user id")
	}
	return fbcodec.SendEmpty(c, 200, "user id rejected")
}
