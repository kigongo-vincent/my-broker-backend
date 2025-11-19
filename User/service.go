package user

import (
	"math/rand"

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
