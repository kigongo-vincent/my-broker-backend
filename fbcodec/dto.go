package fbcodec

// Wire DTOs keep fbcodec free of imports on domain packages (no import cycles).

type PriceIn struct {
	Currency string
	Amount   int
}

type LocationIn struct {
	Lat, Lon float64
	Name     string
}

type UserIn struct {
	ID                                uint
	Name, PhoneNumber, Photo          string
	Email, LastSeen, Status           string
	Verified, BrokerFees              string
	ShowContact, IsBroker, AcceptedPS bool
}

type PostIn struct {
	ID, UserID                                                  uint
	Price                                                       PriceIn
	Location                                                    LocationIn
	Bedrooms, Bathrooms, Toilets                                string
	Images                                                      []string
	Amenities                                                   string
	PayWaterBills, PayElectricityBills, PayForTrash, HasParking bool
	RequiredFirstMonthsPaid                                     int
	Units, Type                                                 string
	IsNegotiable, IsApproved, ReviewDisclaimerAgreed            bool
	User                                                        UserIn
	Likers                                                      []UserIn
}

type ChatIn struct {
	Id          uint
	User        UserIn
	LastMessage string
	NewMessages uint
}

// NestedPostIn is the post-detail payload (detail screen).
type NestedPostIn struct {
	ID, Id                       uint
	Type                         string
	Author, User                 *UserIn
	Price                        PriceIn
	Location                     LocationIn
	IsLiked, IsNegotiable        bool
	Bedrooms, Bathrooms, Toilets string
	Images                       []string
	Likers                       []UserIn
	HideUserInfo, Selected       *bool
}
