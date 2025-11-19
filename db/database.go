package db

import (
	"fmt"
	"log"
	"os"

	post "github.com/kigongo-vincent/my-broker-backend/Post"
	user "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/chat"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var DBERROR error

func ConnectToDB() *gorm.DB {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))
	DB, DBERROR = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if DBERROR != nil {
		log.Fatal("failed to connect to database " + DBERROR.Error())
	}
	fmt.Println("Connected to DB successully!")

	DB.AutoMigrate(&user.User{}, &post.Post{}, &chat.Message{}, &chat.Room{})

	return DB
}
