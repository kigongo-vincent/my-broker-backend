package post

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	flatbuffers "github.com/google/flatbuffers/go"
	cld "github.com/kigongo-vincent/my-broker-backend/Cloudinary"
	usr "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/core"
	"github.com/kigongo-vincent/my-broker-backend/fbcodec"
	"github.com/kigongo-vincent/my-broker-backend/fbs/gen/mybroker"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func userEmailValue(email *string) string {
	if email == nil {
		return ""
	}
	return *email
}

func CreatePost(ctx *fiber.Ctx, db *gorm.DB) error {
	uid, ok := ctx.Locals("userID").(uint)
	if !ok || uid == 0 {
		return fbcodec.SendError(ctx, 401, "unauthorized")
	}
	req, err := fbcodec.OpenRequest(ctx.Body())
	if err != nil {
		return fbcodec.SendError(ctx, 400, err.Error())
	}
	body, err := ParseCreatePostBody(req)
	if err != nil {
		return fbcodec.SendError(ctx, 400, err.Error())
	}
	created, err := PostFromCreateBody(body)
	if err != nil {
		return fbcodec.SendError(ctx, 400, err.Error())
	}
	created.UserID = uid
	created.IsAvailable = true
	if !created.ReviewDisclaimerAgreed {
		return fbcodec.SendError(ctx, 400, "review disclaimer must be accepted")
	}
	var creator usr.User
	if err := db.First(&creator, uid).Error; err == nil && creator.Status == "admin" {
		created.IsApproved = true
	}
	if err := db.Create(&created).First(&created).Error; err != nil {
		return fbcodec.SendError(ctx, 400, err.Error())
	}
	pin := postModelToWire(created)
	return fbcodec.BuildAndSend(ctx, 201, "post created successfully", "", "", 0, mybroker.ApiPayloadPostList, 16384, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostList(b, []fbcodec.PostIn{pin})
	})
}

func GetPaginatedPosts(c *fiber.Ctx, db *gorm.DB) error {
	var posts []usr.Post
	var total int64

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	if pageStr == "" || limitStr == "" {
		return fbcodec.SendError(c, 400, "page and limit are required")
	}
	page, pageErr := strconv.Atoi(pageStr)
	limit, limitErr := strconv.Atoi(limitStr)
	if pageErr != nil || limitErr != nil {
		return fbcodec.SendError(c, 400, "invalid page or limit value")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	dbQuery := db.Where("is_approved = ? AND is_available = ?", true, true)
	dbQuery = applyFilters(c, dbQuery, db)
	countQuery := dbQuery.Model(&usr.Post{})
	if err := countQuery.Count(&total).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to count posts")
	}
	if err := dbQuery.
		Limit(limit).
		Offset((page - 1) * limit).
		Preload("User").
		Preload("Likers").
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to fetch posts")
	}
	pins := make([]fbcodec.PostIn, len(posts))
	for i := range posts {
		pins[i] = postModelToWire(posts[i])
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadPostPage, 65536, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostPage(b, pins, total, page, limit)
	})
}

func applyFilters(c *fiber.Ctx, query *gorm.DB, db *gorm.DB) *gorm.DB {
	if search := c.Query("query"); search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where(
			db.Where("location_name LIKE ?", searchPattern).
				Or("CAST(bedrooms AS TEXT) LIKE ?", searchPattern).
				Or("CAST(bathrooms AS TEXT) LIKE ?", searchPattern).
				Or("CAST(toilets AS TEXT) LIKE ?", searchPattern).
				Or("CAST(price_amount AS TEXT) LIKE ?", searchPattern),
		)
	}
	if location := c.Query("location"); location != "" && location != "undefined" {
		query = query.Where("location_name LIKE ?", "%"+location+"%")
	}
	if startPrice := c.Query("startPrice"); startPrice != "" && startPrice != "undefined" {
		if price, err := strconv.ParseFloat(startPrice, 64); err == nil {
			query = query.Where("price_amount >= ?", price)
		}
	}
	if endPrice := c.Query("endPrice"); endPrice != "" && endPrice != "undefined" {
		if price, err := strconv.ParseFloat(endPrice, 64); err == nil {
			query = query.Where("price_amount <= ?", price)
		}
	}
	if isNegotiable := c.Query("isNegotiable"); isNegotiable != "" && isNegotiable != "undefined" {
		if negotiable, err := strconv.ParseBool(isNegotiable); err == nil {
			query = query.Where("is_negotiable = ?", negotiable)
		}
	}
	if bedrooms := c.Query("bedrooms"); bedrooms != "" && bedrooms != "undefined" && bedrooms != "0" {
		if bedrooms == "5+" || bedrooms == "more than 5" {
			query = query.Where("bedrooms = ? OR bedrooms = ?", "5+", "more than 5")
		} else {
			query = query.Where("bedrooms = ?", bedrooms)
		}
	}
	if bathrooms := c.Query("bathrooms"); bathrooms != "" && bathrooms != "undefined" && bathrooms != "0" {
		if bathrooms == "5+" || bathrooms == "more than 5" {
			query = query.Where("bathrooms = ? OR bathrooms = ?", "5+", "more than 5")
		} else {
			query = query.Where("bathrooms = ?", bathrooms)
		}
	}
	if toilets := c.Query("toilets"); toilets != "" && toilets != "undefined" && toilets != "0" {
		if toilets == "5+" || toilets == "more than 5" {
			query = query.Where("toilets = ? OR toilets = ?", "5+", "more than 5")
		} else {
			query = query.Where("toilets = ?", toilets)
		}
	}
	if postType := c.Query("type"); postType != "" && postType != "undefined" {
		query = query.Where("type = ?", postType)
	}
	return query
}

func UpdateLikerStatus(c *fiber.Ctx, db *gorm.DB) error {
	PostID, pErr := strconv.ParseUint(c.Query("post_id"), 10, 32)
	userID, ok := c.Locals("userID").(uint)
	Liked, _ := strconv.ParseBool(c.Query("liked"))
	if pErr != nil || !ok || userID == 0 {
		return core.ThrowNewError(c, "invalid post or user id")
	}
	if Liked {
		if err := db.Table("post_likes").Where("user_id = ? AND post_id = ?", userID, PostID).Delete(nil).Error; err != nil {
			return core.ThrowNewError(c, "failed to unlike")
		}
	} else {
		if err := db.Table("post_likes").Clauses(clause.OnConflict{DoNothing: true}).Create(map[string]interface{}{
			"post_id": uint(PostID),
			"user_id": userID,
		}).Error; err != nil {
			return core.ThrowNewError(c, "failed to like post"+err.Error())
		}
	}
	return fbcodec.SendEmpty(c, 200, "like status updated")
}

func GetPostsByCreator(c *fiber.Ctx, db *gorm.DB, requesterID uint) error {
	creatorID := requesterID
	if q := c.Query("user_id"); q != "" {
		u, err := strconv.ParseUint(q, 10, 32)
		if err != nil || u == 0 {
			return fbcodec.SendError(c, 400, "invalid user_id")
		}
		creatorID = uint(u)
	}
	q := db.Where("user_id = ?", creatorID)
	if creatorID != requesterID {
		q = q.Where("is_approved = ?", true)
	}
	var posts []usr.Post
	if err := q.Preload("User").Find(&posts).Error; err != nil {
		return core.ThrowNewError(c, "failed to get posts")
	}
	pins := make([]fbcodec.PostIn, len(posts))
	for i := range posts {
		pins[i] = postModelToWire(posts[i])
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadPostList, 65536, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostList(b, pins)
	})
}

func GetPostByID(c *fiber.Ctx, db *gorm.DB) error {
	PostID, pErr := strconv.ParseUint(c.Query("post_id"), 10, 32)
	if pErr != nil {
		return core.ThrowNewError(c, "missing post id")
	}
	var post usr.Post
	if err := db.First(&post, PostID).Error; err != nil {
		return core.ThrowNewError(c, "failed to get post")
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadPostList, 32768, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostList(b, []fbcodec.PostIn{postModelToWire(post)})
	})
}

func GetPostLocations(c *fiber.Ctx, db *gorm.DB) error {
	var postLocations []PostLocationI
	var posts []usr.Post
	if err := db.Where("is_approved = ? AND is_available = ?", true, true).Find(&posts).Error; err != nil {
		return core.ThrowNewError(c, "failed to get posts")
	}
	for _, p := range posts {
		postLocations = append(postLocations, PostLocationI{
			Id:        p.ID,
			Latitude:  p.Location.Lat,
			Longitude: p.Location.Lon,
			Price:     fmt.Sprintf("%s %s", p.Price.Currency, strconv.Itoa(p.Price.Amount)),
		})
	}
	locs := make([]fbcodec.PostLocationIn, len(postLocations))
	for i := range postLocations {
		locs[i] = fbcodec.PostLocationIn{
			Id: postLocations[i].Id, Latitude: postLocations[i].Latitude, Longitude: postLocations[i].Longitude,
			Title: postLocations[i].Title, Price: postLocations[i].Price, Address: postLocations[i].Address,
		}
	}
	return fbcodec.BuildAndSend(c, 200, "", "", "", 0, mybroker.ApiPayloadPostLocationList, 65536, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostLocationList(b, locs)
	})
}

func GetPostDetails(c *fiber.Ctx, db *gorm.DB) error {
	PostID, pErr := strconv.ParseUint(c.Query("post_id"), 10, 32)
	if pErr != nil || PostID == 0 {
		return core.ThrowNewError(c, "failed to get user id")
	}
	var post usr.Post
	if err := db.Preload("User").Preload("Likers").First(&post, uint(PostID)).Error; err != nil {
		return core.ThrowNewError(c, "failed to get the post")
	}
	viewerID, _ := c.Locals("userID").(uint)
	if !post.IsAvailable && post.UserID != viewerID {
		var viewer usr.User
		if viewerID == 0 || db.First(&viewer, viewerID).Error != nil || viewer.Status != "admin" {
			return core.ThrowNewError(c, "post not found")
		}
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadPostList, 65536, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostList(b, []fbcodec.PostIn{postModelToWire(post)})
	})
}

func TransformPost(p usr.Post) NestedPost {
	var images []string
	if len(p.Images) > 0 {
		_ = json.Unmarshal(p.Images, &images)
	}
	likers := make([]NestedUser, len(p.Likers))
	for i, l := range p.Likers {
		likers[i] = NestedUser{
			ID: l.ID, Name: l.Name, PhoneNumber: l.PhoneNumber, Photo: l.Photo, Email: userEmailValue(l.Email),
			LastSeen: l.LastSeen, Status: l.Status, Verified: l.Verified, ShowContact: l.ShowContact,
		}
	}
	author := &NestedUser{
		ID: p.User.ID, Name: p.User.Name, PhoneNumber: p.User.PhoneNumber, Photo: p.User.Photo, Email: userEmailValue(p.User.Email),
		LastSeen: p.User.LastSeen, Status: p.User.Status, Verified: p.User.Verified, ShowContact: p.User.ShowContact,
	}
	return NestedPost{
		ID: p.ID, Id: p.ID, Type: p.Type, Author: author, User: author, Price: p.Price, Location: p.Location,
		IsLiked: false, Bedrooms: p.Bedrooms, Bathrooms: p.Bathrooms, Toilets: p.Toilets, Images: images,
		Likers: likers, IsNegotiable: p.IsNegotiable, IsAvailable: p.IsAvailable,
	}
}

func GetFavourites(c *fiber.Ctx, db *gorm.DB, UserID uint) error {
	var uu usr.User
	if err := db.Preload("Liked.User").Preload("Liked.Likers").Preload("Liked").First(&uu, UserID).Error; err != nil {
		return core.ThrowNewError(c, "failed to get user")
	}
	liked := make([]usr.Post, 0, len(uu.Liked))
	for _, p := range uu.Liked {
		if p.IsApproved && p.IsAvailable {
			liked = append(liked, p)
		}
	}
	lPins := make([]fbcodec.PostIn, len(liked))
	for i := range liked {
		lPins[i] = postModelToWire(liked[i])
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadPostList, 131072, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostList(b, lPins)
	})
}

func GetPendingPostsForAdmin(c *fiber.Ctx, db *gorm.DB) error {
	adminID := c.Locals("userID").(uint)
	var admin usr.User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return fbcodec.SendError(c, 403, "admin access required")
	}

	pageStr := c.Query("page")
	limitStr := c.Query("limit")
	if pageStr == "" || limitStr == "" {
		return fbcodec.SendError(c, 400, "page and limit are required")
	}
	page, pageErr := strconv.Atoi(pageStr)
	limit, limitErr := strconv.Atoi(limitStr)
	if pageErr != nil || limitErr != nil {
		return fbcodec.SendError(c, 400, "invalid page or limit value")
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var total int64
	countQ := db.Model(&usr.Post{}).Where("is_approved = ?", false)
	if err := countQ.Count(&total).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to count pending posts")
	}

	var posts []usr.Post
	if err := db.Where("is_approved = ?", false).
		Preload("User").
		Preload("Likers").
		Order("created_at DESC").
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&posts).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to fetch pending posts")
	}
	pendPins := make([]fbcodec.PostIn, len(posts))
	for i := range posts {
		pendPins[i] = postModelToWire(posts[i])
	}
	return fbcodec.BuildAndSend(c, 200, "ok", "", "", 0, mybroker.ApiPayloadPostPage, 131072, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostPage(b, pendPins, total, page, limit)
	})
}

func nestedUserToFB(n NestedUser) fbcodec.UserIn {
	return fbcodec.UserIn{
		ID: n.ID, Name: n.Name, PhoneNumber: n.PhoneNumber, Photo: n.Photo, Email: n.Email,
		LastSeen: n.LastSeen, Status: n.Status, Verified: n.Verified, ShowContact: n.ShowContact,
	}
}

func nestedPostToIn(np NestedPost) fbcodec.NestedPostIn {
	out := fbcodec.NestedPostIn{
		ID: np.ID, Id: np.Id, Type: np.Type,
		Price:    fbcodec.PriceIn{Currency: np.Price.Currency, Amount: np.Price.Amount},
		Location: fbcodec.LocationIn{Lat: np.Location.Lat, Lon: np.Location.Lon, Name: np.Location.Name},
		IsLiked:  np.IsLiked, IsNegotiable: np.IsNegotiable, IsAvailable: np.IsAvailable, Bedrooms: np.Bedrooms,
		Bathrooms: np.Bathrooms, Toilets: np.Toilets, Images: np.Images,
		HideUserInfo: np.HideUserInfo, Selected: np.Selected,
	}
	if np.Author != nil {
		v := nestedUserToFB(*np.Author)
		out.Author = &v
	}
	if np.User != nil {
		v := nestedUserToFB(*np.User)
		out.User = &v
	}
	for _, l := range np.Likers {
		out.Likers = append(out.Likers, nestedUserToFB(l))
	}
	return out
}

func ApprovePostByAdmin(c *fiber.Ctx, db *gorm.DB) error {
	adminID := c.Locals("userID").(uint)
	var admin usr.User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return fbcodec.SendError(c, 403, "admin access required")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	body, err := ParseApprovePost(req)
	if err != nil || body.PostId() == 0 {
		return fbcodec.SendError(c, 400, "invalid post id")
	}
	if err := db.Model(&usr.Post{}).Where("id = ?", body.PostId()).Update("is_approved", true).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to approve post")
	}
	return fbcodec.SendEmpty(c, 200, "post approved")
}

func RejectPostByAdmin(c *fiber.Ctx, db *gorm.DB) error {
	adminID := c.Locals("userID").(uint)
	var admin usr.User
	if err := db.First(&admin, adminID).Error; err != nil || admin.Status != "admin" {
		return fbcodec.SendError(c, 403, "admin access required")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	body, err := ParseApprovePost(req)
	if err != nil || body.PostId() == 0 {
		return fbcodec.SendError(c, 400, "invalid post id")
	}
	pid := uint(body.PostId())
	var p usr.Post
	if err := db.First(&p, pid).Error; err != nil {
		return fbcodec.SendError(c, 404, "post not found")
	}
	if p.IsApproved {
		return fbcodec.SendError(c, 400, "post is not pending")
	}
	var imageURLs []string
	if len(p.Images) > 0 {
		_ = json.Unmarshal(p.Images, &imageURLs)
	}
	cld.DestroyDeliveryURLs(imageURLs)
	if err := db.Exec(`DELETE FROM post_likes WHERE post_id = ?`, pid).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to delete likes")
	}
	if err := db.Delete(&usr.Post{}, pid).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to reject post")
	}
	return fbcodec.SendEmpty(c, 200, "post rejected")
}

// DeleteMyPost removes a listing owned by the authenticated user.
func DeleteMyPost(c *fiber.Ctx, db *gorm.DB) error {
	uid, ok := c.Locals("userID").(uint)
	if !ok || uid == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	body, err := usr.ParsePostID(req)
	if err != nil || body.PostId() == 0 {
		return fbcodec.SendError(c, 400, "invalid post id")
	}
	pid := uint(body.PostId())
	var p usr.Post
	if err := db.First(&p, pid).Error; err != nil {
		return fbcodec.SendError(c, 404, "post not found")
	}
	var actor usr.User
	if err := db.First(&actor, uid).Error; err != nil {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	if p.UserID != uid && actor.Status != "admin" {
		return fbcodec.SendError(c, 403, "forbidden")
	}
	if err := db.Exec(`DELETE FROM post_likes WHERE post_id = ?`, pid).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to delete likes")
	}
	if err := db.Delete(&usr.Post{}, pid).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to delete post")
	}
	return fbcodec.SendEmpty(c, 200, "post deleted")
}

// UpdateMyPost updates mutable listing fields for the authenticated owner/admin.
func UpdateMyPost(c *fiber.Ctx, db *gorm.DB) error {
	uid, ok := c.Locals("userID").(uint)
	if !ok || uid == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	postID, err := strconv.ParseUint(c.Query("post_id"), 10, 32)
	if err != nil || postID == 0 {
		return fbcodec.SendError(c, 400, "invalid post_id")
	}
	req, err := fbcodec.OpenRequest(c.Body())
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	body, err := ParseCreatePostBody(req)
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	next, err := PostFromCreateBody(body)
	if err != nil {
		return fbcodec.SendError(c, 400, err.Error())
	}
	if !next.ReviewDisclaimerAgreed {
		return fbcodec.SendError(c, 400, "review disclaimer must be accepted")
	}

	var existing usr.Post
	if err := db.First(&existing, uint(postID)).Error; err != nil {
		return fbcodec.SendError(c, 404, "post not found")
	}
	var actor usr.User
	if err := db.First(&actor, uid).Error; err != nil {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	if existing.UserID != uid && actor.Status != "admin" {
		return fbcodec.SendError(c, 403, "forbidden")
	}

	existing.Price = next.Price
	existing.Location = next.Location
	existing.Bedrooms = next.Bedrooms
	existing.Bathrooms = next.Bathrooms
	existing.Toilets = next.Toilets
	existing.Images = next.Images
	existing.Ammenities = next.Ammenities
	existing.PayWaterBills = next.PayWaterBills
	existing.PayElectricityBills = next.PayElectricityBills
	existing.PayForTrash = next.PayForTrash
	existing.HasParking = next.HasParking
	existing.RequiredFirstMonthsPaid = next.RequiredFirstMonthsPaid
	existing.Units = next.Units
	existing.IsNegotiable = next.IsNegotiable
	existing.ReviewDisclaimerAgreed = next.ReviewDisclaimerAgreed
	existing.Type = next.Type

	// Non-admin listing edits require re-approval.
	if actor.Status != "admin" {
		existing.IsApproved = false
	}

	if err := db.Save(&existing).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to update post")
	}
	if err := db.Preload("User").Preload("Likers").First(&existing, existing.ID).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to load updated post")
	}
	return fbcodec.BuildAndSend(c, 200, "post updated successfully", "", "", 0, mybroker.ApiPayloadPostList, 16384, func(b *flatbuffers.Builder) flatbuffers.UOffsetT {
		return fbcodec.BuildPostList(b, []fbcodec.PostIn{postModelToWire(existing)})
	})
}

// SetMyPostAvailability sets is_available for the author's listing (hides from public feed when false).
func SetMyPostAvailability(c *fiber.Ctx, db *gorm.DB) error {
	uid, ok := c.Locals("userID").(uint)
	if !ok || uid == 0 {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	postID, err := strconv.ParseUint(c.Query("post_id"), 10, 32)
	if err != nil || postID == 0 {
		return fbcodec.SendError(c, 400, "invalid post_id")
	}
	available, err := strconv.ParseBool(c.Query("available"))
	if err != nil {
		return fbcodec.SendError(c, 400, "available must be true or false")
	}
	var p usr.Post
	if err := db.First(&p, uint(postID)).Error; err != nil {
		return fbcodec.SendError(c, 404, "post not found")
	}
	var actor usr.User
	if err := db.First(&actor, uid).Error; err != nil {
		return fbcodec.SendError(c, 401, "unauthorized")
	}
	if p.UserID != uid && actor.Status != "admin" {
		return fbcodec.SendError(c, 403, "forbidden")
	}
	if err := db.Model(&p).Update("is_available", available).Error; err != nil {
		return fbcodec.SendError(c, 500, "failed to update availability")
	}
	return fbcodec.SendEmpty(c, 200, "availability updated")
}
