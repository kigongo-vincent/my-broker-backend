package user

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type adminUpdateUserJSON struct {
	UserID   uint    `json:"user_id"`
	Status   *string `json:"status"`
	IsBroker *bool   `json:"is_broker"`
	Verified *bool   `json:"verified"`
}

type adminDeleteUserJSON struct {
	UserID uint `json:"user_id"`
}

func requireAdmin(c *fiber.Ctx, db *gorm.DB) (uint, bool, error) {
	adminID, ok := c.Locals("userID").(uint)
	if !ok || adminID == 0 {
		return 0, false, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"msg": "unauthorized"})
	}
	var admin User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return 0, false, c.Status(fiber.StatusForbidden).JSON(fiber.Map{"msg": "admin access required"})
	}
	return adminID, true, nil
}

func AdminUpdateUserJSON(c *fiber.Ctx, db *gorm.DB) error {
	adminID, ok, err := requireAdmin(c, db)
	if err != nil || !ok {
		return err
	}

	var body adminUpdateUserJSON
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "invalid json body"})
	}
	if body.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "user_id is required"})
	}

	var target User
	if err := db.First(&target, body.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "user not found"})
	}

	updates := map[string]any{}
	if body.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*body.Status))
		if status != "user" && status != "broker" && status != "admin" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "invalid status"})
		}
		if target.ID == adminID && status != "admin" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "you cannot remove your own admin role"})
		}
		updates["status"] = status
		if status == "broker" {
			updates["is_broker"] = true
		}
		if status == "user" {
			updates["is_broker"] = false
		}
	}
	if body.IsBroker != nil {
		updates["is_broker"] = *body.IsBroker
		if target.Status != "admin" && body.Status == nil {
			if *body.IsBroker {
				updates["status"] = "broker"
			} else {
				updates["status"] = "user"
			}
		}
	}
	if body.Verified != nil {
		if *body.Verified {
			updates["verified"] = "true"
			updates["id_verification_status"] = "approved"
		} else {
			updates["verified"] = "false"
			updates["id_verification_status"] = "rejected"
		}
	}
	if len(updates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "no updates provided"})
	}
	if err := db.Model(&target).Updates(updates).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "failed to update user"})
	}
	return c.JSON(fiber.Map{"msg": "user updated"})
}

func AdminDeleteUserJSON(c *fiber.Ctx, db *gorm.DB) error {
	adminID, ok, err := requireAdmin(c, db)
	if err != nil || !ok {
		return err
	}

	var body adminDeleteUserJSON
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "invalid json body"})
	}
	if body.UserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "user_id is required"})
	}
	if body.UserID == adminID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "you cannot delete your own account"})
	}

	var target User
	if err := db.First(&target, body.UserID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"msg": "user not found"})
	}
	if target.Status == "admin" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"msg": "demote admin users before deleting them"})
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`DELETE FROM post_likes WHERE user_id = ? OR post_id IN (SELECT id FROM posts WHERE user_id = ?)`, body.UserID, body.UserID).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ? OR blocked_user_id = ?", body.UserID, body.UserID).Delete(&BlockedUser{}).Error; err != nil {
			return err
		}
		if err := tx.Where("reporter_id = ? OR reported_id = ?", body.UserID, body.UserID).Delete(&UserReport{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", body.UserID).Delete(&Message{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", body.UserID).Delete(&Post{}).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", body.UserID).Delete(&UserRoom{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&User{}, body.UserID).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"msg": "failed to delete user"})
	}
	return c.JSON(fiber.Map{"msg": "user deleted"})
}
