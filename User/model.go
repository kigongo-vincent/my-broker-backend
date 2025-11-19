package user

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Id          int    `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number" gorm:"unique" validate:"required"`
	OTP         int    `json:"otp"`
	Photo       string `json:"photo"`
	Email       string
}
