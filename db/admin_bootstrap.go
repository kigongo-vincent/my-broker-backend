package db

import (
	"fmt"
	"log"
	"os"
	"strings"

	user "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/core"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func EnsureDefaultAdminFromEnv(db *gorm.DB) error {
	phoneRaw := strings.TrimSpace(os.Getenv("ADMIN_PHONE"))
	if phoneRaw == "" {
		return nil
	}

	normalizedPhone, err := core.NormalizeUGPhoneNumber(phoneRaw)
	if err != nil {
		return err
	}

	adminPin := strings.TrimSpace(os.Getenv("ADMIN_PIN"))
	var pinHash string
	if adminPin != "" {
		h, herr := bcrypt.GenerateFromPassword([]byte(adminPin), bcrypt.DefaultCost)
		if herr != nil {
			return fmt.Errorf("ADMIN_PIN bcrypt: %w", herr)
		}
		pinHash = string(h)
	}

	name := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	if name == "" {
		name = "System Admin"
	}
	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))

	var existing user.User
	lookup := db.Where("phone_number IN ?", core.UGPhoneCandidates(normalizedPhone)).Limit(1).Find(&existing)
	if lookup.Error != nil {
		return lookup.Error
	}

	if existing.ID == 0 {
		var emailPtr *string
		if email != "" {
			emailPtr = &email
		}
		admin := user.User{
			Name:        name,
			PhoneNumber: normalizedPhone,
			Email:       emailPtr,
			Status:      "admin",
			Verified:    "true",
			IsBroker:    false,
			ShowContact: true,
			AcceptedPS:  true,
			PinHash:     pinHash,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		log.Printf("created bootstrap admin user with phone %s", normalizedPhone)
		if adminPin != "" {
			log.Printf("bootstrap admin PIN set from ADMIN_PIN env")
		}
		return nil
	}

	updates := map[string]any{
		"status":   "admin",
		"verified": "true",
	}
	if name != "" {
		updates["name"] = name
	}
	if email != "" {
		updates["email"] = email
	}
	if pinHash != "" {
		updates["pin_hash"] = pinHash
	}
	if err := db.Model(&existing).Updates(updates).Error; err != nil {
		return err
	}
	log.Printf("ensured bootstrap admin privileges for user %d", existing.ID)
	if adminPin != "" {
		log.Printf("bootstrap admin PIN updated from ADMIN_PIN env")
	}
	return nil
}
