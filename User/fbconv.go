package user

import (
	"encoding/json"

	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
)

func UserToIn(u User) fbcodec.UserIn {
	return fbcodec.UserIn{
		ID: u.ID, Name: u.Name, PhoneNumber: u.PhoneNumber, Photo: u.Photo, Email: u.Email,
		LastSeen: u.LastSeen, Status: u.Status, Verified: u.Verified, BrokerFees: u.BrokerFees,
		ShowContact: u.ShowContact, IsBroker: u.IsBroker, AcceptedPS: u.AcceptedPS,
	}
}

func PostToIn(p Post) fbcodec.PostIn {
	var imgs []string
	if len(p.Images) > 0 {
		_ = json.Unmarshal(p.Images, &imgs)
	}
	likers := make([]fbcodec.UserIn, len(p.Likers))
	for i := range p.Likers {
		likers[i] = UserToIn(p.Likers[i])
	}
	return fbcodec.PostIn{
		ID: p.ID, UserID: p.UserID,
		Price:    fbcodec.PriceIn{Currency: p.Price.Currency, Amount: p.Price.Amount},
		Location: fbcodec.LocationIn{Lat: p.Location.Lat, Lon: p.Location.Lon, Name: p.Location.Name},
		Bedrooms: p.Bedrooms, Bathrooms: p.Bathrooms, Toilets: p.Toilets,
		Images: imgs, Amenities: string(p.Ammenities),
		PayWaterBills: p.PayWaterBills, PayElectricityBills: p.PayElectricityBills,
		PayForTrash: p.PayForTrash, HasParking: p.HasParking,
		RequiredFirstMonthsPaid: p.RequiredFirstMonthsPaid, Units: p.Units, Type: p.Type,
		IsNegotiable: p.IsNegotiable, IsApproved: p.IsApproved, ReviewDisclaimerAgreed: p.ReviewDisclaimerAgreed,
		User: UserToIn(p.User), Likers: likers,
	}
}

func ChatToIn(c Chat) fbcodec.ChatIn {
	return fbcodec.ChatIn{
		Id: c.Id, User: UserToIn(c.User), LastMessage: c.LastMessage, NewMessages: c.NewMessages,
	}
}
