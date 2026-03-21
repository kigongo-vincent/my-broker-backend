package db

import (
	"fmt"
	"log"
	"os"
	"strings"

	user "github.com/kigongo-vincent/my-broker-backend/User"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var DBERROR error

func ConnectToDB() *gorm.DB {

	sslmode := strings.TrimSpace(os.Getenv("DB_SSLMODE"))
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		sslmode,
	)
	if cert := strings.TrimSpace(os.Getenv("DB_SSLROOTCERT")); cert != "" {
		dsn += " sslrootcert=" + cert
	}
	DB, DBERROR = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if DBERROR != nil {
		log.Fatal("failed to connect to database " + DBERROR.Error())
	}

	DB.AutoMigrate(&user.User{}, &user.Post{}, &user.Room{}, &user.Message{}, &user.BlockedUser{}, &user.UserReport{})
	if err := DB.Exec(`CREATE INDEX IF NOT EXISTS idx_posts_approved_created_at ON posts (is_approved, created_at DESC)`).Error; err != nil {
		log.Printf("posts index: %v", err)
	}
	if os.Getenv("DB_SEED") == "true" {
		if err := SeedDatabase(DB); err != nil {
			log.Printf("database seed failed: %v", err)
		}
	}

	return DB
}
