package fbcodec

import (
	"github.com/gofiber/fiber/v2"
	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/kigongo-vincent/my-broker-backend/fbs/gen/mybroker"
)

// PostLocationIn is one map pin row.
type PostLocationIn struct {
	Id                    uint
	Latitude, Longitude   float64
	Title, Price, Address string
}

func str(b *flatbuffers.Builder, s string) flatbuffers.UOffsetT {
	return b.CreateString(s)
}

func BuildPrice(b *flatbuffers.Builder, p PriceIn) flatbuffers.UOffsetT {
	c := str(b, p.Currency)
	mybroker.PriceStart(b)
	mybroker.PriceAddAmount(b, int32(p.Amount))
	mybroker.PriceAddCurrency(b, c)
	return mybroker.PriceEnd(b)
}

func BuildLocation(b *flatbuffers.Builder, loc LocationIn) flatbuffers.UOffsetT {
	n := str(b, loc.Name)
	mybroker.LocationStart(b)
	mybroker.LocationAddName(b, n)
	mybroker.LocationAddLon(b, loc.Lon)
	mybroker.LocationAddLat(b, loc.Lat)
	return mybroker.LocationEnd(b)
}

func BuildUser(b *flatbuffers.Builder, user UserIn) flatbuffers.UOffsetT {
	n := str(b, user.Name)
	ph := str(b, user.PhoneNumber)
	photo := str(b, user.Photo)
	em := str(b, user.Email)
	ls := str(b, user.LastSeen)
	st := str(b, user.Status)
	vf := str(b, user.Verified)
	bf := str(b, user.BrokerFees)
	mybroker.UserStart(b)
	mybroker.UserAddAcceptedPs(b, user.AcceptedPS)
	mybroker.UserAddBrokerFees(b, bf)
	mybroker.UserAddIsBroker(b, user.IsBroker)
	mybroker.UserAddShowContact(b, user.ShowContact)
	mybroker.UserAddVerified(b, vf)
	mybroker.UserAddStatus(b, st)
	mybroker.UserAddLastSeen(b, ls)
	mybroker.UserAddEmail(b, em)
	mybroker.UserAddPhoto(b, photo)
	mybroker.UserAddPhoneNumber(b, ph)
	mybroker.UserAddName(b, n)
	mybroker.UserAddId(b, uint32(user.ID))
	return mybroker.UserEnd(b)
}

func BuildNestedUser(b *flatbuffers.Builder, user UserIn) flatbuffers.UOffsetT {
	n := str(b, user.Name)
	ph := str(b, user.PhoneNumber)
	photo := str(b, user.Photo)
	em := str(b, user.Email)
	ls := str(b, user.LastSeen)
	st := str(b, user.Status)
	vf := str(b, user.Verified)
	mybroker.NestedUserStart(b)
	mybroker.NestedUserAddShowContact(b, user.ShowContact)
	mybroker.NestedUserAddVerified(b, vf)
	mybroker.NestedUserAddStatus(b, st)
	mybroker.NestedUserAddLastSeen(b, ls)
	mybroker.NestedUserAddEmail(b, em)
	mybroker.NestedUserAddPhoto(b, photo)
	mybroker.NestedUserAddPhoneNumber(b, ph)
	mybroker.NestedUserAddName(b, n)
	mybroker.NestedUserAddId(b, uint32(user.ID))
	return mybroker.NestedUserEnd(b)
}

func BuildPostWire(b *flatbuffers.Builder, p PostIn) flatbuffers.UOffsetT {
	imgOffs := make([]flatbuffers.UOffsetT, len(p.Images))
	for i := range p.Images {
		imgOffs[i] = str(b, p.Images[i])
	}
	mybroker.PostWireStartImagesVector(b, len(imgOffs))
	for i := len(imgOffs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(imgOffs[i])
	}
	imgVec := b.EndVector(len(imgOffs))

	priceOff := BuildPrice(b, p.Price)
	locOff := BuildLocation(b, p.Location)
	userOff := BuildUser(b, p.User)

	likerOffs := make([]flatbuffers.UOffsetT, len(p.Likers))
	for i := range p.Likers {
		likerOffs[i] = BuildUser(b, p.Likers[i])
	}
	mybroker.PostWireStartLikersVector(b, len(likerOffs))
	for i := len(likerOffs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(likerOffs[i])
	}
	likVec := b.EndVector(len(likerOffs))

	bd := str(b, p.Bedrooms)
	bt := str(b, p.Bathrooms)
	to := str(b, p.Toilets)
	am := str(b, p.Amenities)
	un := str(b, p.Units)
	ty := str(b, p.Type)

	mybroker.PostWireStart(b)
	mybroker.PostWireAddIsAvailable(b, p.IsAvailable)
	mybroker.PostWireAddLikers(b, likVec)
	mybroker.PostWireAddUser(b, userOff)
	mybroker.PostWireAddType(b, ty)
	mybroker.PostWireAddReviewDisclaimerAgreed(b, p.ReviewDisclaimerAgreed)
	mybroker.PostWireAddIsApproved(b, p.IsApproved)
	mybroker.PostWireAddIsNegotiable(b, p.IsNegotiable)
	mybroker.PostWireAddUnits(b, un)
	mybroker.PostWireAddRequiredFirstMonthsPaid(b, int32(p.RequiredFirstMonthsPaid))
	mybroker.PostWireAddHasParking(b, p.HasParking)
	mybroker.PostWireAddPayForTrash(b, p.PayForTrash)
	mybroker.PostWireAddPayElectricityBills(b, p.PayElectricityBills)
	mybroker.PostWireAddPayWaterBills(b, p.PayWaterBills)
	mybroker.PostWireAddAmenities(b, am)
	mybroker.PostWireAddImages(b, imgVec)
	mybroker.PostWireAddToilets(b, to)
	mybroker.PostWireAddBathrooms(b, bt)
	mybroker.PostWireAddBedrooms(b, bd)
	mybroker.PostWireAddLocation(b, locOff)
	mybroker.PostWireAddPrice(b, priceOff)
	mybroker.PostWireAddUserId(b, uint32(p.UserID))
	mybroker.PostWireAddId(b, uint32(p.ID))
	return mybroker.PostWireEnd(b)
}

func BuildNestedPostT(b *flatbuffers.Builder, np NestedPostIn) flatbuffers.UOffsetT {
	imgOffs := make([]flatbuffers.UOffsetT, len(np.Images))
	for i := range np.Images {
		imgOffs[i] = str(b, np.Images[i])
	}
	mybroker.NestedPostTStartImagesVector(b, len(imgOffs))
	for i := len(imgOffs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(imgOffs[i])
	}
	imgVec := b.EndVector(len(imgOffs))

	priceOff := BuildPrice(b, np.Price)
	locOff := BuildLocation(b, np.Location)

	var authOff, usrOff flatbuffers.UOffsetT
	if np.Author != nil {
		authOff = BuildNestedUser(b, *np.Author)
	}
	if np.User != nil {
		usrOff = BuildNestedUser(b, *np.User)
	}

	likOffs := make([]flatbuffers.UOffsetT, len(np.Likers))
	for i := range np.Likers {
		likOffs[i] = BuildNestedUser(b, np.Likers[i])
	}
	mybroker.NestedPostTStartLikersVector(b, len(likOffs))
	for i := len(likOffs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(likOffs[i])
	}
	likVec := b.EndVector(len(likOffs))

	ty := str(b, np.Type)
	bd := str(b, np.Bedrooms)
	bt := str(b, np.Bathrooms)
	to := str(b, np.Toilets)

	hide := false
	if np.HideUserInfo != nil {
		hide = *np.HideUserInfo
	}
	sel := false
	if np.Selected != nil {
		sel = *np.Selected
	}

	mybroker.NestedPostTStart(b)
	mybroker.NestedPostTAddIsAvailable(b, np.IsAvailable)
	mybroker.NestedPostTAddSelected(b, sel)
	mybroker.NestedPostTAddHideUserInfo(b, hide)
	mybroker.NestedPostTAddLikers(b, likVec)
	mybroker.NestedPostTAddUser(b, usrOff)
	mybroker.NestedPostTAddAuthor(b, authOff)
	mybroker.NestedPostTAddImages(b, imgVec)
	mybroker.NestedPostTAddToilets(b, to)
	mybroker.NestedPostTAddBathrooms(b, bt)
	mybroker.NestedPostTAddBedrooms(b, bd)
	mybroker.NestedPostTAddIsNegotiable(b, np.IsNegotiable)
	mybroker.NestedPostTAddIsLiked(b, np.IsLiked)
	mybroker.NestedPostTAddLocation(b, locOff)
	mybroker.NestedPostTAddPrice(b, priceOff)
	mybroker.NestedPostTAddType(b, ty)
	mybroker.NestedPostTAddIdDup(b, uint32(np.Id))
	mybroker.NestedPostTAddId(b, uint32(np.ID))
	return mybroker.NestedPostTEnd(b)
}

func BuildPostPage(b *flatbuffers.Builder, posts []PostIn, total int64, page, limit int) flatbuffers.UOffsetT {
	offs := make([]flatbuffers.UOffsetT, len(posts))
	for i := range posts {
		offs[i] = BuildPostWire(b, posts[i])
	}
	mybroker.PostPageStartPostsVector(b, len(offs))
	for i := len(offs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offs[i])
	}
	vec := b.EndVector(len(offs))
	mybroker.PostPageStart(b)
	mybroker.PostPageAddLimit(b, int32(limit))
	mybroker.PostPageAddPage(b, int32(page))
	mybroker.PostPageAddTotal(b, total)
	mybroker.PostPageAddPosts(b, vec)
	return mybroker.PostPageEnd(b)
}

func BuildPostList(b *flatbuffers.Builder, posts []PostIn) flatbuffers.UOffsetT {
	offs := make([]flatbuffers.UOffsetT, len(posts))
	for i := range posts {
		offs[i] = BuildPostWire(b, posts[i])
	}
	mybroker.PostListStartPostsVector(b, len(offs))
	for i := len(offs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offs[i])
	}
	vec := b.EndVector(len(offs))
	mybroker.PostListStart(b)
	mybroker.PostListAddPosts(b, vec)
	return mybroker.PostListEnd(b)
}

func BuildPostLocationList(b *flatbuffers.Builder, locs []PostLocationIn) flatbuffers.UOffsetT {
	offs := make([]flatbuffers.UOffsetT, len(locs))
	for i := range locs {
		ti := str(b, locs[i].Title)
		pr := str(b, locs[i].Price)
		ad := str(b, locs[i].Address)
		mybroker.PostLocationRowStart(b)
		mybroker.PostLocationRowAddAddress(b, ad)
		mybroker.PostLocationRowAddPrice(b, pr)
		mybroker.PostLocationRowAddTitle(b, ti)
		mybroker.PostLocationRowAddLongitude(b, locs[i].Longitude)
		mybroker.PostLocationRowAddLatitude(b, locs[i].Latitude)
		mybroker.PostLocationRowAddId(b, uint32(locs[i].Id))
		offs[i] = mybroker.PostLocationRowEnd(b)
	}
	mybroker.PostLocationListStartItemsVector(b, len(offs))
	for i := len(offs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offs[i])
	}
	vec := b.EndVector(len(offs))
	mybroker.PostLocationListStart(b)
	mybroker.PostLocationListAddItems(b, vec)
	return mybroker.PostLocationListEnd(b)
}

func BuildChatRowList(b *flatbuffers.Builder, chats []ChatIn) flatbuffers.UOffsetT {
	offs := make([]flatbuffers.UOffsetT, len(chats))
	for i := range chats {
		uo := BuildUser(b, chats[i].User)
		lm := str(b, chats[i].LastMessage)
		mybroker.ChatRowStart(b)
		mybroker.ChatRowAddNewMessages(b, uint32(chats[i].NewMessages))
		mybroker.ChatRowAddLastMessage(b, lm)
		mybroker.ChatRowAddUser(b, uo)
		mybroker.ChatRowAddId(b, uint32(chats[i].Id))
		offs[i] = mybroker.ChatRowEnd(b)
	}
	mybroker.ChatRowListStartChatsVector(b, len(offs))
	for i := len(offs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offs[i])
	}
	vec := b.EndVector(len(offs))
	mybroker.ChatRowListStart(b)
	mybroker.ChatRowListAddChats(b, vec)
	return mybroker.ChatRowListEnd(b)
}

func BuildPostWireMinimal(b *flatbuffers.Builder, p PostIn) flatbuffers.UOffsetT {
	if p.ID == 0 {
		// All strings and child tables before PostWireStart — CreateString is illegal inside an open table.
		emptyStr := str(b, "")
		mybroker.PostWireStartImagesVector(b, 0)
		emptyImg := b.EndVector(0)
		pr := BuildPrice(b, PriceIn{})
		loc := BuildLocation(b, LocationIn{})
		u0 := BuildUser(b, UserIn{})
		mybroker.PostWireStartLikersVector(b, 0)
		emptyLik := b.EndVector(0)
		mybroker.PostWireStart(b)
		mybroker.PostWireAddIsAvailable(b, true)
		mybroker.PostWireAddLikers(b, emptyLik)
		mybroker.PostWireAddUser(b, u0)
		mybroker.PostWireAddType(b, emptyStr)
		mybroker.PostWireAddReviewDisclaimerAgreed(b, false)
		mybroker.PostWireAddIsApproved(b, false)
		mybroker.PostWireAddIsNegotiable(b, false)
		mybroker.PostWireAddUnits(b, emptyStr)
		mybroker.PostWireAddRequiredFirstMonthsPaid(b, 0)
		mybroker.PostWireAddHasParking(b, false)
		mybroker.PostWireAddPayForTrash(b, false)
		mybroker.PostWireAddPayElectricityBills(b, false)
		mybroker.PostWireAddPayWaterBills(b, false)
		mybroker.PostWireAddAmenities(b, emptyStr)
		mybroker.PostWireAddImages(b, emptyImg)
		mybroker.PostWireAddToilets(b, emptyStr)
		mybroker.PostWireAddBathrooms(b, emptyStr)
		mybroker.PostWireAddBedrooms(b, emptyStr)
		mybroker.PostWireAddLocation(b, loc)
		mybroker.PostWireAddPrice(b, pr)
		mybroker.PostWireAddUserId(b, 0)
		mybroker.PostWireAddId(b, 0)
		return mybroker.PostWireEnd(b)
	}
	return BuildPostWire(b, p)
}

// ChatMsgRow is one message line for BuildChatDetail.
type ChatMsgRow struct {
	ID                 uint
	Text, Image        string
	Post               PostIn
	CreatedAtUnixMs    int64
	SeenByRecipient    bool
	IsOwnedByRecipient bool
}

func BuildChatDetail(b *flatbuffers.Builder, roomID uint, peer UserIn, messages []ChatMsgRow) flatbuffers.UOffsetT {
	msgOffs := make([]flatbuffers.UOffsetT, len(messages))
	for i := range messages {
		pw := BuildPostWireMinimal(b, messages[i].Post)
		tx := str(b, messages[i].Text)
		im := str(b, messages[i].Image)
		mybroker.ChatMsgStart(b)
		mybroker.ChatMsgAddIsOwnedByRecipient(b, messages[i].IsOwnedByRecipient)
		mybroker.ChatMsgAddSeenByRecipient(b, messages[i].SeenByRecipient)
		mybroker.ChatMsgAddCreatedAtUnixMs(b, messages[i].CreatedAtUnixMs)
		mybroker.ChatMsgAddPost(b, pw)
		mybroker.ChatMsgAddImage(b, im)
		mybroker.ChatMsgAddText(b, tx)
		mybroker.ChatMsgAddId(b, uint32(messages[i].ID))
		msgOffs[i] = mybroker.ChatMsgEnd(b)
	}
	mybroker.ChatDetailStartMessagesVector(b, len(msgOffs))
	for i := len(msgOffs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(msgOffs[i])
	}
	mvec := b.EndVector(len(msgOffs))
	uo := BuildUser(b, peer)
	mybroker.ChatDetailStart(b)
	mybroker.ChatDetailAddMessages(b, mvec)
	mybroker.ChatDetailAddUser(b, uo)
	mybroker.ChatDetailAddId(b, uint32(roomID))
	return mybroker.ChatDetailEnd(b)
}

func BuildChatsAndFavs(b *flatbuffers.Builder, unread, favs int) flatbuffers.UOffsetT {
	mybroker.ChatsAndFavsStart(b)
	mybroker.ChatsAndFavsAddFavourites(b, int32(favs))
	mybroker.ChatsAndFavsAddUnreadTotal(b, int32(unread))
	return mybroker.ChatsAndFavsEnd(b)
}

func BuildAuthOk(b *flatbuffers.Builder, user UserIn) flatbuffers.UOffsetT {
	uo := BuildUser(b, user)
	mybroker.AuthOkStart(b)
	mybroker.AuthOkAddUser(b, uo)
	return mybroker.AuthOkEnd(b)
}

func BuildUsersList(b *flatbuffers.Builder, users []UserIn) flatbuffers.UOffsetT {
	offs := make([]flatbuffers.UOffsetT, len(users))
	for i := range users {
		offs[i] = BuildUser(b, users[i])
	}
	mybroker.UsersListStartUsersVector(b, len(offs))
	for i := len(offs) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offs[i])
	}
	vec := b.EndVector(len(offs))
	mybroker.UsersListStart(b)
	mybroker.UsersListAddUsers(b, vec)
	return mybroker.UsersListEnd(b)
}

func SendOTP(c *fiber.Ctx, httpStatus int, msg, warning string, otp int32) error {
	return BuildAndSend(c, httpStatus, msg, "", warning, otp, mybroker.ApiPayloadEmpty, 384, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		mybroker.EmptyStart(b)
		return mybroker.EmptyEnd(b)
	})
}

func SendAuthOK(c *fiber.Ctx, msg, token string, usr UserIn) error {
	return BuildAndSend(c, 200, msg, token, "", 0, mybroker.ApiPayloadAuthOk, 8192, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return BuildAuthOk(b, usr)
	})
}
