package user

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type verificationStatusJSON struct {
	Verified             string `json:"verified"`
	IDVerificationStatus string `json:"id_verification_status"`
}

type idVerificationSubmitJSON struct {
	SelfieURL   string `json:"selfie_url"`
	DocumentURL string `json:"document_url"`
}

func GetVerificationStatusJSON(c *fiber.Ctx, db *gorm.DB) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "unauthorized"})
	}
	var u User
	if err := db.First(&u, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "user not found"})
	}
	return c.JSON(verificationStatusJSON{
		Verified:             u.Verified,
		IDVerificationStatus: u.IDVerificationStatus,
	})
}

func SubmitIDVerificationJSON(c *fiber.Ctx, db *gorm.DB) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "unauthorized"})
	}
	var body idVerificationSubmitJSON
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "invalid json body"})
	}
	selfie := strings.TrimSpace(body.SelfieURL)
	doc := strings.TrimSpace(body.DocumentURL)
	if selfie == "" || doc == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "selfie_url and document_url are required"})
	}
	if !strings.HasPrefix(selfie, "https://") || !strings.HasPrefix(doc, "https://") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "urls must be https"})
	}
	var u User
	if err := db.First(&u, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "user not found"})
	}
	if u.Verified == "true" && u.IDVerificationStatus != "rejected" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "already verified"})
	}
	updates := map[string]any{
		"id_verification_selfie":   selfie,
		"id_verification_document": doc,
		"id_verification_status":   "submitted",
		"verified":                 "false",
	}
	if err := db.Model(&u).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "failed to save verification"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"msg": "verification submitted"})
}

type pendingIDVerificationRow struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	LastSeen    string `json:"last_seen"`
	Photo       string `json:"photo"`
	SelfieURL   string `json:"selfie_url"`
	DocumentURL string `json:"document_url"`
}

// ListPendingIDVerificationsForAdmin returns users awaiting ID review (admin only).
func ListPendingIDVerificationsForAdmin(c *fiber.Ctx, db *gorm.DB) error {
	adminID, ok := c.Locals("userID").(uint)
	if !ok || adminID == 0 {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "unauthorized"})
	}
	var admin User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"msg": "admin access required"})
	}
	var rows []User
	if err := db.Where("id_verification_status = ?", "submitted").Order("updated_at DESC").Find(&rows).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "failed to fetch verification queue"})
	}
	out := make([]pendingIDVerificationRow, 0, len(rows))
	for _, u := range rows {
		email := ""
		if u.Email != nil {
			email = strings.TrimSpace(*u.Email)
		}
		out = append(out, pendingIDVerificationRow{
			ID:          u.ID,
			Name:        u.Name,
			Email:       email,
			PhoneNumber: u.PhoneNumber,
			LastSeen:    strings.TrimSpace(u.LastSeen),
			Photo:       u.Photo,
			SelfieURL:   u.IDVerificationSelfie,
			DocumentURL: u.IDVerificationDocument,
		})
	}
	return c.JSON(out)
}
