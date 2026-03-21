package wschat

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	fiberws "github.com/gofiber/websocket/v2"
	"github.com/fasthttp/websocket"
	"gorm.io/gorm"

	usr "github.com/kigongo-vincent/my-broker-backend/User"
)

type hub struct {
	mu     sync.Mutex
	byUser map[uint][]*fiberws.Conn
}

var defaultHub = &hub{byUser: make(map[uint][]*fiberws.Conn)}

func (h *hub) register(uid uint, c *fiberws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.byUser[uid] = append(h.byUser[uid], c)
}

func (h *hub) unregister(uid uint, target *fiberws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	list := h.byUser[uid]
	for i, c := range list {
		if c == target {
			h.byUser[uid] = append(list[:i], list[i+1:]...)
			break
		}
	}
	if len(h.byUser[uid]) == 0 {
		delete(h.byUser, uid)
	}
}

func (h *hub) broadcastRoom(db *gorm.DB, roomID uint, payload []byte) {
	var userIDs []uint
	if err := db.Table("user_rooms").Where("room_id = ?", roomID).Pluck("user_id", &userIDs).Error; err != nil {
		log.Printf("wschat: room members: %v", err)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, uid := range userIDs {
		for _, c := range h.byUser[uid] {
			if c == nil {
				continue
			}
			if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Printf("wschat: write user %d: %v", uid, err)
			}
		}
	}
}

type inbound struct {
	UserID     uint   `json:"user_id"`
	RoomID     uint   `json:"room_id"`
	Content    string `json:"content"`
	Type       string `json:"type"`
	PostID     uint   `json:"post_id"`
	Attachment string `json:"attachment"`
}

type outbound struct {
	UserID      uint   `json:"user_id"`
	RoomID      uint   `json:"room_id"`
	Content     string `json:"content"`
	Type        string `json:"type"`
	CreatedAt   string `json:"created_at"`
	PostID      uint   `json:"post_id,omitempty"`
	Attachment  string `json:"attachment,omitempty"`
	ID          uint   `json:"id"`
}

// ServeChat handles one WebSocket: JWT user id must be in conn Locals as "wsUserID" (uint).
func ServeChat(conn *fiberws.Conn, db *gorm.DB) {
	defer conn.Close()

	uidVal := conn.Locals("wsUserID")
	uid, ok := uidVal.(uint)
	if !ok || uid == 0 {
		return
	}

	defaultHub.register(uid, conn)
	defer defaultHub.unregister(uid, conn)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if len(msg) == 0 {
			continue
		}

		var in inbound
		if err := json.Unmarshal(msg, &in); err != nil {
			continue
		}
		if in.Type != "" && in.Type != "message" {
			continue
		}
		if in.UserID != uid {
			continue
		}
		if in.RoomID == 0 {
			continue
		}
		if in.Content == "" && in.Attachment == "" {
			continue
		}

		var n int64
		if err := db.Model(&usr.UserRoom{}).Where("room_id = ? AND user_id = ?", in.RoomID, uid).Count(&n).Error; err != nil || n == 0 {
			continue
		}

		m := usr.Message{
			UserID:     uid,
			RoomID:     in.RoomID,
			Text:       in.Content,
			Attachment: in.Attachment,
			IsRead:     false,
		}
		if in.PostID != 0 {
			m.PostID = in.PostID
		}

		var createErr error
		if in.PostID == 0 {
			createErr = db.Omit("PostID", "Post").Create(&m).Error
		} else {
			createErr = db.Create(&m).Error
		}
		if createErr != nil {
			log.Printf("wschat: create message: %v", createErr)
			continue
		}

		out := outbound{
			UserID:     m.UserID,
			RoomID:     m.RoomID,
			Content:    m.Text,
			Type:       "message",
			CreatedAt:  m.CreatedAt.UTC().Format(time.RFC3339Nano),
			Attachment: m.Attachment,
			ID:         m.ID,
		}
		if m.PostID != 0 {
			out.PostID = m.PostID
		}

		payload, err := json.Marshal(out)
		if err != nil {
			continue
		}
		defaultHub.broadcastRoom(db, in.RoomID, payload)
	}
}
