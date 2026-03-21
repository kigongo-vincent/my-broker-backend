package db

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	user "github.com/kigongo-vincent/my-broker-backend/User"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SeedDatabase(db *gorm.DB) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	firstNames := []string{"Liam", "Noah", "Ethan", "Mason", "Aiden", "Lucas", "Elijah", "James", "Benjamin", "Daniel", "Grace", "Faith", "Ava", "Mia", "Emma", "Nora", "Stella", "Luna", "Zara", "Ivy"}
	lastNames := []string{"Kato", "Ssenfuma", "Nabwire", "Mugisha", "Byaruhanga", "Namutebi", "Okello", "Achieng", "Nsubuga", "Ssekandi", "Ssembatya", "Nakanwagi"}
	areas := []string{"Kololo", "Naguru", "Ntinda", "Bukoto", "Kisaasi", "Muyenga", "Munyonyo", "Naalya", "Kireka", "Namugongo", "Bweyogerere", "Kawempe", "Makindye", "Wandegeya"}
	propertyTypes := []string{"business", "office", "apartment", "single-room"}
	amenitiesPool := []string{"wifi", "parking", "water-tank", "security", "fence", "cctv", "balcony", "garden", "near-road", "power-backup"}
	messageSamples := []string{
		"Hi, is this unit still available?",
		"Can I schedule a viewing this weekend?",
		"Is the price slightly negotiable?",
		"How far is it from the main road?",
		"Do you allow small pets?",
		"Is water included in the rent?",
		"I can pay two months upfront.",
		"Please share more photos of the bedrooms.",
	}

	var userCount int64
	if err := db.Model(&user.User{}).Count(&userCount).Error; err != nil {
		return err
	}
	if userCount >= 30 {
		return nil
	}

	users := make([]user.User, 0, 35)
	for i := 0; i < 35; i++ {
		name := fmt.Sprintf("%s %s", firstNames[r.Intn(len(firstNames))], lastNames[r.Intn(len(lastNames))])
		phone := fmt.Sprintf("+2567%08d", 10000000+i)
		email := fmt.Sprintf("user%d@mybroker.ug", i+1)
		status := "user"
		if i < 2 {
			status = "admin"
		} else if i%5 == 0 {
			status = "broker"
		}
		users = append(users, user.User{
			Name:        name,
			PhoneNumber: phone,
			Email:       email,
			Photo:       "https://images.unsplash.com/photo-1544005313-94ddf0286df2",
			LastSeen:    time.Now().Add(-time.Duration(r.Intn(120)) * time.Minute).Format(time.RFC3339),
			Status:      status,
			Verified:    map[bool]string{true: "true", false: "false"}[r.Intn(3) == 0],
			IsBroker:    status == "broker",
			BrokerFees:  fmt.Sprintf("%d", 20000+r.Intn(120000)),
			AcceptedPS:  r.Intn(2) == 0,
			ShowContact: r.Intn(4) != 0,
		})
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(users, 20).Error; err != nil {
		return err
	}

	var dbUsers []user.User
	if err := db.Find(&dbUsers).Error; err != nil {
		return err
	}
	if len(dbUsers) < 30 {
		return fmt.Errorf("not enough users after seed")
	}

	posts := make([]user.Post, 0, 40)
	for i := 0; i < 40; i++ {
		u := dbUsers[r.Intn(len(dbUsers))]
		imgs, _ := json.Marshal([]string{
			"https://images.unsplash.com/photo-1560185127-6ed189bf02f4",
			"https://images.unsplash.com/photo-1493666438817-866a91353ca9",
		})
		amenities, _ := json.Marshal([]string{
			amenitiesPool[r.Intn(len(amenitiesPool))],
			amenitiesPool[r.Intn(len(amenitiesPool))],
			amenitiesPool[r.Intn(len(amenitiesPool))],
		})
		bedrooms := []string{"0", "1", "2", "3", "4", "5+"}[r.Intn(6)]
		bathrooms := []string{"1", "2", "3", "4", "5+"}[r.Intn(5)]
		toilets := []string{"1", "2", "3", "4", "5+"}[r.Intn(5)]
		ptype := propertyTypes[r.Intn(len(propertyTypes))]
		posts = append(posts, user.Post{
			UserID:                  u.ID,
			Bedrooms:                bedrooms,
			Bathrooms:               bathrooms,
			Toilets:                 toilets,
			Images:                  datatypes.JSON(imgs),
			Ammenities:              datatypes.JSON(amenities),
			PayWaterBills:           r.Intn(2) == 0,
			PayElectricityBills:     r.Intn(2) == 0,
			PayForTrash:             r.Intn(2) == 0,
			HasParking:              r.Intn(2) == 0,
			RequiredFirstMonthsPaid: 1 + r.Intn(3),
			Units:                   []string{"1", "2", "3", "4", "5+"}[r.Intn(5)],
			IsNegotiable:            r.Intn(2) == 0,
			IsApproved:              u.Status == "admin" || r.Intn(3) != 0,
			ReviewDisclaimerAgreed:  true,
			Type:                    ptype,
			Price: user.Price{
				Currency: "UGX",
				Amount:   250000 + r.Intn(2750000),
			},
			Location: user.Location{
				Lat:  0.25 + r.Float64()*0.2,
				Lon:  32.50 + r.Float64()*0.2,
				Name: fmt.Sprintf("%s, Kampala", areas[r.Intn(len(areas))]),
			},
		})
	}
	if err := db.CreateInBatches(posts, 20).Error; err != nil {
		return err
	}

	var dbPosts []user.Post
	if err := db.Find(&dbPosts).Error; err != nil {
		return err
	}

	rooms := make([]user.Room, 0, 30)
	for i := 0; i < 30; i++ {
		rooms = append(rooms, user.Room{})
	}
	if err := db.CreateInBatches(rooms, 30).Error; err != nil {
		return err
	}
	var dbRooms []user.Room
	if err := db.Find(&dbRooms).Error; err != nil {
		return err
	}

	for i := range dbRooms {
		u1 := dbUsers[r.Intn(len(dbUsers))]
		u2 := dbUsers[r.Intn(len(dbUsers))]
		for u2.ID == u1.ID {
			u2 = dbUsers[r.Intn(len(dbUsers))]
		}
		if err := db.Model(&dbRooms[i]).Association("Users").Append(&u1, &u2); err != nil {
			return err
		}
	}

	messages := make([]user.Message, 0, 90)
	for i := 0; i < 90; i++ {
		room := dbRooms[r.Intn(len(dbRooms))]
		var roomWithUsers user.Room
		if err := db.Preload("Users").First(&roomWithUsers, room.ID).Error; err != nil {
			return err
		}
		sender := roomWithUsers.Users[r.Intn(len(roomWithUsers.Users))]
		p := dbPosts[r.Intn(len(dbPosts))]
		messages = append(messages, user.Message{
			PostID:     p.ID,
			Text:       messageSamples[r.Intn(len(messageSamples))],
			Attachment: "",
			UserID:     sender.ID,
			RoomID:     room.ID,
			IsRead:     r.Intn(2) == 0,
		})
	}
	if err := db.CreateInBatches(messages, 30).Error; err != nil {
		return err
	}

	for i := 0; i < 60; i++ {
		u := dbUsers[r.Intn(len(dbUsers))]
		p := dbPosts[r.Intn(len(dbPosts))]
		_ = db.Table("post_likes").Clauses(clause.OnConflict{DoNothing: true}).Create(map[string]any{
			"user_id": u.ID,
			"post_id": p.ID,
		}).Error
	}

	return nil
}
