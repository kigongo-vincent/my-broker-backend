package db

import (
	"fmt"
	"log"
	"os"

	user "github.com/kigongo-vincent/my-broker-backend/User"
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

	DB.AutoMigrate(&user.User{}, &user.Post{}, &user.Room{}, &user.Message{})

	return DB
}
