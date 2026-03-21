package post

import (
	"encoding/json"

	usr "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
)

// postModelToWire maps a GORM usr.Post into fbcodec.PostIn in this package so assignments
// to []fbcodec.PostIn use the same fbcodec instance as the rest of post (avoids cross-package
// "invalid type" issues with some toolchains when mixing user.PostToIn with post-local slices).
func postModelToWire(p usr.Post) fbcodec.PostIn {
	var imgs []string
	if len(p.Images) > 0 {
		_ = json.Unmarshal(p.Images, &imgs)
	}
	likers := make([]fbcodec.UserIn, len(p.Likers))
	for i := range p.Likers {
		likers[i] = userModelToWire(p.Likers[i])
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
		User: userModelToWire(p.User), Likers: likers,
	}
}

func userModelToWire(u usr.User) fbcodec.UserIn {
	return fbcodec.UserIn{
		ID: u.ID, Name: u.Name, PhoneNumber: u.PhoneNumber, Photo: u.Photo, Email: u.Email,
		LastSeen: u.LastSeen, Status: u.Status, Verified: u.Verified, BrokerFees: u.BrokerFees,
		ShowContact: u.ShowContact, IsBroker: u.IsBroker, AcceptedPS: u.AcceptedPS,
	}
}
