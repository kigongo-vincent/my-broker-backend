package post

import user "github.com/kigongo-vincent/my-broker-backend/User"

type PostLocationI struct {
	Id        uint    `json:"id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Title     string  `json:"title"`
	Price     string  `json:"price"`
	Address   string  `json:"address"`
}

// NestedPost is the fully nested struct we defined before
type NestedPost struct {
	ID           uint          `json:"ID"`
	Id           uint          `json:"id"`
	Type         string        `json:"type"`
	Author       *NestedUser   `json:"author,omitempty"`
	User         *NestedUser   `json:"user,omitempty"`
	Price        user.Price    `json:"price"`
	Location     user.Location `json:"location"`
	IsLiked      bool          `json:"is_liked"`
	IsNegotiable bool          `json:"is_negotiable"`
	Bedrooms     string        `json:"bedrooms"`
	Bathrooms    string        `json:"bathrooms"`
	Toilets      string        `json:"toilets"`
	Images       []string      `json:"images"`
	Likers       []NestedUser  `json:"likers"`
	HideUserInfo *bool         `json:"hideUserInfo,omitempty"`
	Selected     *bool         `json:"selected,omitempty"`
}

// NestedUser is a simplified version of user.User
type NestedUser struct {
	ID          uint   `json:"ID"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Photo       string `json:"photo,omitempty"`
	Email       string `json:"email,omitempty"`
	LastSeen    string `json:"last_seen,omitempty"`
	Status      string `json:"status,omitempty"`
	Verified    string `json:"verified,omitempty"`
	ShowContact bool   `json:"show_contact"`
}
