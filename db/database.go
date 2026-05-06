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

func columnType(db *gorm.DB, tableName, columnName string) (string, error) {
	var udtName string
	err := db.Raw(`
SELECT c.udt_name
FROM information_schema.columns c
WHERE c.table_schema = 'public'
  AND c.table_name = ?
  AND c.column_name = ?
LIMIT 1`, tableName, columnName).Scan(&udtName).Error
	if err != nil {
		return "", err
	}
	return udtName, nil
}

func upgradeColumnToBigintIfInteger(db *gorm.DB, tableName, columnName string) error {
	stmt := fmt.Sprintf(`
DO $$
BEGIN
	IF EXISTS (
		SELECT 1
		FROM information_schema.columns
		WHERE table_schema = 'public'
		  AND table_name = '%s'
		  AND column_name = '%s'
		  AND data_type = 'integer'
	) THEN
		EXECUTE 'ALTER TABLE "%s" ALTER COLUMN "%s" TYPE bigint';
	END IF;
END $$;`, tableName, columnName, tableName, columnName)
	return db.Exec(stmt).Error
}

func relaxLegacyUsersColumnNotNull(db *gorm.DB, columnName string) error {
	var exists bool
	if err := db.Raw(`
SELECT EXISTS (
	SELECT 1
	FROM information_schema.columns
	WHERE table_schema = 'public'
	  AND table_name = 'users'
	  AND column_name = ?
)`, columnName).Scan(&exists).Error; err != nil {
		return err
	}
	if !exists {
		return nil
	}

	stmt := fmt.Sprintf(`ALTER TABLE "users" ALTER COLUMN "%s" DROP NOT NULL`, columnName)
	if err := db.Exec(stmt).Error; err != nil {
		return err
	}
	return nil
}

func rebuildUsersIDIfUUID(db *gorm.DB) error {
	typ, err := columnType(db, "users", "id")
	if err != nil {
		return err
	}
	if typ != "uuid" {
		return nil
	}

	var usersCount int64
	if err := db.Raw(`SELECT COUNT(*) FROM "users"`).Scan(&usersCount).Error; err != nil {
		return err
	}
	if usersCount > 0 {
		return fmt.Errorf("users.id is uuid with %d existing rows; manual migration required", usersCount)
	}

	var incomingFKCount int64
	if err := db.Raw(`
SELECT COUNT(*)
FROM pg_constraint
WHERE contype = 'f'
  AND confrelid = 'users'::regclass`).Scan(&incomingFKCount).Error; err != nil {
		return err
	}
	if incomingFKCount > 0 {
		return fmt.Errorf("users.id is uuid and has %d foreign key references; manual migration required", incomingFKCount)
	}

	if err := db.Exec(`ALTER TABLE "users" DROP CONSTRAINT IF EXISTS "users_pkey"`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE "users" DROP COLUMN IF EXISTS "id"`).Error; err != nil {
		return err
	}
	if err := db.Exec(`ALTER TABLE "users" ADD COLUMN "id" bigserial PRIMARY KEY`).Error; err != nil {
		return err
	}
	return nil
}

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
	if err := rebuildUsersIDIfUUID(DB); err != nil {
		log.Fatal("failed to normalize users.id type: " + err.Error())
	}
	// Legacy installs may still have SERIAL (int4) IDs. Align them with current uint-based models.
	if err := upgradeColumnToBigintIfInteger(DB, "users", "id"); err != nil {
		log.Fatal("failed to upgrade users.id type: " + err.Error())
	}
	if err := upgradeColumnToBigintIfInteger(DB, "posts", "user_id"); err != nil {
		log.Fatal("failed to upgrade posts.user_id type: " + err.Error())
	}
	if err := upgradeColumnToBigintIfInteger(DB, "user_rooms", "user_id"); err != nil {
		log.Fatal("failed to upgrade user_rooms.user_id type: " + err.Error())
	}
	if err := upgradeColumnToBigintIfInteger(DB, "post_likes", "user_id"); err != nil {
		log.Fatal("failed to upgrade post_likes.user_id type: " + err.Error())
	}
	if err := upgradeColumnToBigintIfInteger(DB, "messages", "user_id"); err != nil {
		log.Fatal("failed to upgrade messages.user_id type: " + err.Error())
	}
	if err := upgradeColumnToBigintIfInteger(DB, "blocked_users", "user_id"); err != nil {
		log.Fatal("failed to upgrade blocked_users.user_id type: " + err.Error())
	}
	if err := upgradeColumnToBigintIfInteger(DB, "blocked_users", "blocked_user_id"); err != nil {
		log.Fatal("failed to upgrade blocked_users.blocked_user_id type: " + err.Error())
	}
	if err := upgradeColumnToBigintIfInteger(DB, "user_reports", "reporter_id"); err != nil {
		log.Fatal("failed to upgrade user_reports.reporter_id type: " + err.Error())
	}
	if err := upgradeColumnToBigintIfInteger(DB, "user_reports", "reported_id"); err != nil {
		log.Fatal("failed to upgrade user_reports.reported_id type: " + err.Error())
	}
	// The app schema evolved away from some legacy users columns; old DBs may
	// still keep NOT NULL constraints that break inserts.
	if err := relaxLegacyUsersColumnNotNull(DB, "password"); err != nil {
		log.Fatal("failed to relax legacy users.password constraint: " + err.Error())
	}
	if err := relaxLegacyUsersColumnNotNull(DB, "role"); err != nil {
		log.Fatal("failed to relax legacy users.role constraint: " + err.Error())
	}

	// Ensure the many2many schema uses our explicit join model definition.
	if err := DB.SetupJoinTable(&user.User{}, "Rooms", &user.UserRoom{}); err != nil {
		log.Fatal("failed to setup user-rooms join table: " + err.Error())
	}
	if err := DB.SetupJoinTable(&user.Room{}, "Users", &user.UserRoom{}); err != nil {
		log.Fatal("failed to setup room-users join table: " + err.Error())
	}
	if err := DB.SetupJoinTable(&user.User{}, "Liked", &user.PostLike{}); err != nil {
		log.Fatal("failed to setup user-liked join table: " + err.Error())
	}
	if err := DB.SetupJoinTable(&user.Post{}, "Likers", &user.PostLike{}); err != nil {
		log.Fatal("failed to setup post-likers join table: " + err.Error())
	}

	// Keep migrations non-destructive for existing data.
	if err := DB.AutoMigrate(&user.User{}, &user.Post{}, &user.Room{}, &user.UserRoom{}, &user.PostLike{}, &user.Message{}, &user.BlockedUser{}, &user.UserReport{}); err != nil {
		log.Fatal("database migration failed: " + err.Error())
	}
	if err := EnsureDefaultAdminFromEnv(DB); err != nil {
		log.Printf("admin bootstrap failed: %v", err)
	}
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
