package post

import (
	user "github.com/kigongo-vincent/my-broker-backend/User"
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
	Id           int            `json:"id" gorm:"primaryKey"`
	UserID       int            `json:"user_id"`
	User         user.User      `json:"user" gorm:"foreignKey:UserID"`
	Price        Price          `json:"price" gorm:"embedded;embeddedPrefix:price_"`
	Location     Location       `json:"location" gorm:"embedded;embeddedPrefix:location_"`
	Likers       datatypes.JSON `json:"likers" gorm:"type:json"`
	Bedrooms     string         `json:"bedrooms"`
	Bathrooms    string         `json:"bathrooms"`
	Toilets      string         `json:"toilets"`
	Images       datatypes.JSON `json:"images" gorm:"type:json"`
	Units        string         `json:"units"`
	IsNegotiable bool           `json:"is_negotiable"`
}
