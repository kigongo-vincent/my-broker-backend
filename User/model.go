package user

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Price struct {
	Currency string
	Amount   int
}

type Location struct {
	Lat  float64
	Lon  float64
	Name string
}

type Post struct {
	gorm.Model
	UserID       uint           `json:"user_id"`
	User         User           `json:"user" gorm:"foreignKey:UserID"`
	Price        Price          `json:"price" gorm:"embedded;embeddedPrefix:price_"`
	Location     Location       `json:"location" gorm:"embedded;embeddedPrefix:location_"`
	Likers       datatypes.JSON `json:"likers" gorm:"type:json"`
	Bedrooms     string         `json:"bedrooms"`
	Bathrooms    string         `json:"bathrooms"`
	Toilets      string         `json:"toilets"`
	Images       datatypes.JSON `json:"images" gorm:"type:json"`
	Units        string         `json:"units"`
	IsNegotiable bool           `json:"is_negotiable"`
	IsApproved   bool           `json:"is_approved" gorm:"default:false"`
}

type User struct {
	gorm.Model
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number" gorm:"unique" validate:"required"`
	OTP         int    `json:"otp"`
	Photo       string `json:"photo"`
	Email       string `json:"email" gorm:"omit-empty"`
	LastSeen    string `json:"last_seen"`
	Status      string `json:"status" gorm:"default:'user'"`
	Verified    string `json:"verified" gorm:"default:false"`
	Rooms       []Room `json:"rooms" gorm:"many2many:user_rooms"`
}

type Message struct {
	gorm.Model
	PostID     uint   `json:"post_id"`
	Post       Post   `json:"post,omitempty" gorm:"foreignKey:PostID"`
	Text       string `json:"text"`
	Attachment string `json:"attachment"`
	UserID     uint   `json:"user_id"`
	User       User   `json:"user" gorm:"foreignKey:UserID"`
	RoomID     uint   `json:"room_id"`
	Room       Room   `json:"room" gorm:"foreignKey:RoomID"`
	IsRead     bool   `json:"is_read"`
}

type Room struct {
	gorm.Model
	Users []User `json:"users" gorm:"many2many:user_rooms"`
}
