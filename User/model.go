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
	UserID                  uint           `json:"user_id" gorm:"not null;index"`
	User                    User           `json:"user" gorm:"foreignKey:UserID"`
	Price                   Price          `json:"price" gorm:"embedded;embeddedPrefix:price_"`
	Location                Location       `json:"location" gorm:"embedded;embeddedPrefix:location_"`
	Likers                  []User         `json:"likers" gorm:"many2many:post_likes"`
	Bedrooms                string         `json:"bedrooms" gorm:"not null;default:''"`
	Bathrooms               string         `json:"bathrooms" gorm:"not null;default:''"`
	Toilets                 string         `json:"toilets" gorm:"not null;default:''"`
	Images                  datatypes.JSON `json:"images" gorm:"type:json"`
	Ammenities              datatypes.JSON `json:"ammenities" gorm:"type:json"`
	PayWaterBills           bool           `json:"pay_water_bills"`
	PayElectricityBills     bool           `json:"pay_electricity_bills"`
	PayForTrash             bool           `json:"pay_for_trash"`
	HasParking              bool           `json:"has_parking"`
	RequiredFirstMonthsPaid int            `json:"required_first_months_paid"`
	Units                   string         `json:"units" gorm:"not null;default:'1'"`
	IsNegotiable            bool           `json:"is_negotiable"`
	IsApproved              bool           `json:"is_approved" gorm:"default:false"`
	IsAvailable             bool           `json:"is_available" gorm:"not null;default:true"`
	ReviewDisclaimerAgreed  bool           `json:"review_disclaimer_agreed" gorm:"not null;default:false"`
	Type                    string         `json:"type" gorm:"not null;default:''"`
}

type User struct {
	gorm.Model
	Name        string  `json:"name" gorm:"not null;default:''"`
	PhoneNumber string  `json:"phone_number" gorm:"not null;uniqueIndex" validate:"required"`
	OTP         int     `json:"otp"`
	// PinHash is bcrypt hash of numeric PIN; empty means user must set PIN (legacy OTP users or new flow).
	PinHash string `json:"-" gorm:"not null;default:''"`
	Photo       string  `json:"photo" gorm:"not null;default:''"`
	Email       *string `json:"email,omitempty"`
	LastSeen    string  `json:"last_seen" gorm:"not null;default:''"`
	Status      string  `json:"status" gorm:"default:'user'"`
	Verified    string  `json:"verified" gorm:"not null;default:'false'"` // admin ID verification; only "true" shows badge
	IsBroker    bool    `json:"is_broker" gorm:"default:false"`
	BrokerFees  string  `json:"broker_fees" gorm:"default:''"`
	AcceptedPS  bool    `json:"accepted_ps" gorm:"default:false"`
	Rooms       []Room  `json:"rooms" gorm:"many2many:user_rooms"`
	ShowContact bool    `json:"show_contact" gorm:"default:true"`
	Liked       []Post  `json:"liked" gorm:"many2many:post_likes"`

	// ID verification (admin approves via existing approve-id; URLs stored for review)
	IDVerificationSelfie   string `json:"id_verification_selfie" gorm:"not null;default:''"`
	IDVerificationDocument string `json:"id_verification_document" gorm:"not null;default:''"`
	IDVerificationStatus   string `json:"id_verification_status" gorm:"not null;default:''"` // submitted | approved | rejected
}

type UserRoom struct {
	RoomID uint
	UserID uint
}

type PostLike struct {
	PostID uint
	UserID uint
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

type BlockedUser struct {
	gorm.Model
	UserID        uint `json:"user_id" gorm:"not null;index:idx_user_blocked,unique"`
	BlockedUserID uint `json:"blocked_user_id" gorm:"not null;index:idx_user_blocked,unique"`
}

type UserReport struct {
	gorm.Model
	ReporterID uint   `json:"reporter_id" gorm:"not null;index"`
	ReportedID uint   `json:"reported_id" gorm:"not null;index"`
	Reason     string `json:"reason" gorm:"not null;default:''"`
}

type Chat struct {
	Id          uint   `json:"id"`
	User        User   `json:"user"`
	LastMessage string `json:"lastMessage"`
	NewMessages uint   `json:"newMessages"`
}
