package db

import (
	"log"
	"os"
	"strings"

	user "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/core"
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
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		log.Printf("created bootstrap admin user with phone %s", normalizedPhone)
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
	if err := db.Model(&existing).Updates(updates).Error; err != nil {
		return err
	}
	log.Printf("ensured bootstrap admin privileges for user %d", existing.ID)
	return nil
}
