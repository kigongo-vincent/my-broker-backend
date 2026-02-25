package post

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	user "github.com/kigongo-vincent/my-broker-backend/User"
	"github.com/kigongo-vincent/my-broker-backend/core"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreatePost(ctx *fiber.Ctx, db *gorm.DB) error {

	var post user.Post
	post.UserID = 1

	if err := ctx.BodyParser(&post); err != nil {
		return ctx.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	if err := db.Create(&post).First(&post).Error; err != nil {
		return ctx.Status(400).JSON(fiber.Map{"msg": err.Error()})
	}

	return ctx.Status(201).JSON(fiber.Map{"msg": "post created successfully", "data": post})

}

// func GetPaginatedPosts(c *fiber.Ctx, db *gorm.DB) error {

// 	var posts []user.Post
// 	var page int = 1
// 	var limit int = 10
// 	var total int64

// 	// get the controls
// 	page, pageErr := strconv.Atoi(c.Query("page"))
// 	limit, limitErr := strconv.Atoi(c.Query("limit"))
// 	query := c.Query("query")

// 	if pageErr != nil || limitErr != nil {
// 		return c.Status(400).JSON(fiber.Map{"msg": "page or limit missing"})
// 	}

// 	if query != "" {
// 		query = "%" + query + "%"
// 		if err := db.Where("is_approved = ?", true).Where(db.Where("location_name LIKE ?", query).Or("CAST(bedrooms as TEXT) LIKE ?", query).Or("CAST(bathrooms as TEXT) LIKE ?", query).Or("CAST(toilets as TEXT) LIKE ?", query).Or("CAST(price_amount as TEXT) LIKE ?", query)).Limit(limit).Offset((page - 1) * limit).Preload("User").Find(&posts).Error; err != nil {
// 			return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
// 		}
// 	} else {

// 		if err := db.Where("is_approved = ?", true).Limit(limit).Offset((page - 1) * limit).Preload("User").Find(&posts).Error; err != nil {
// 			return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
// 		}
// 	}

// 	if len(posts) == 0 {
// 		return c.Status(202).JSON(fiber.Map{"msg": "no posts found"})
// 	}

// 	if err := db.Model(&user.Post{}).Where("is_approved = ?", true).Count(&total).Error; err != nil {
// 		return c.Status(400).JSON(fiber.Map{"msg": err.Error()})
// 	}

// 	return c.JSON(fiber.Map{"data": posts, "total": total, "page": page, "limit": limit})
// }

func GetPaginatedPosts(c *fiber.Ctx, db *gorm.DB) error {
	var posts []user.Post
	var total int64

	// Parse pagination params (required)
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	if pageStr == "" || limitStr == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "page and limit are required"})
	}

	page, pageErr := strconv.Atoi(pageStr)
	limit, limitErr := strconv.Atoi(limitStr)

	if pageErr != nil || limitErr != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid page or limit value"})
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Build base query
	dbQuery := db.Where("is_approved = ?", true)

	// Apply optional filters
	dbQuery = applyFilters(c, dbQuery, db)

	// Get total count with all filters applied
	countQuery := dbQuery.Model(&user.Post{})
	if err := countQuery.Count(&total).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to count posts", "error": err.Error()})
	}

	// Get paginated results
	if err := dbQuery.
		Limit(limit).
		Offset((page - 1) * limit).
		Preload("User").
		Preload("Likers").
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to fetch posts", "error": err.Error()})
	}

	// Return results (empty array if no posts found)
	return c.Status(200).JSON(fiber.Map{
		"data":  posts,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func applyFilters(c *fiber.Ctx, query *gorm.DB, db *gorm.DB) *gorm.DB {
	// General search query (searches across multiple fields)
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

	// Location filter (specific location search)
	if location := c.Query("location"); location != "" && location != "undefined" {
		locationPattern := "%" + location + "%"
		query = query.Where("location_name LIKE ?", locationPattern)
	}

	// Price range filters
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

	// Negotiable filter
	if isNegotiable := c.Query("isNegotiable"); isNegotiable != "" && isNegotiable != "undefined" {
		if negotiable, err := strconv.ParseBool(isNegotiable); err == nil {
			query = query.Where("is_negotiable = ?", negotiable)
		}
	}

	// Bedrooms filter
	if bedrooms := c.Query("bedrooms"); bedrooms != "" && bedrooms != "undefined" && bedrooms != "0" {
		query = query.Where("bedrooms = ?", bedrooms)
	}

	// Bathrooms filter
	if bathrooms := c.Query("bathrooms"); bathrooms != "" && bathrooms != "undefined" && bathrooms != "0" {
		query = query.Where("bathrooms = ?", bathrooms)
	}

	// Toilets filter
	if toilets := c.Query("toilets"); toilets != "" && toilets != "undefined" && toilets != "0" {
		query = query.Where("toilets = ?", toilets)
	}

	return query
}

func UpdateLikerStatus(c *fiber.Ctx, db *gorm.DB) error {

	PostID, pErr := strconv.ParseUint(c.Query("post_id"), 10, 32)
	UserID, uErr := strconv.ParseUint(c.Query("user_id"), 10, 32)
	Liked, _ := strconv.ParseBool(c.Query("liked"))

	if pErr != nil || uErr != nil {
		return core.ThrowNewError(c, "invalid post or user id")
	}

	if Liked {
		if err := db.Table("post_likes").Where("user_id = ? AND post_id = ?", UserID, PostID).Delete(nil).Error; err != nil {
			return core.ThrowNewError(c, "failed to unlike")
		}
	} else {
		if err := db.Table("post_likes").Clauses(clause.OnConflict{DoNothing: true}).Create(map[string]interface{}{
			"post_id": uint(PostID),
			"user_id": uint(UserID),
		}).Error; err != nil {
			return core.ThrowNewError(c, "failed to like post"+err.Error())
		}
	}

	return c.JSON(fiber.Map{"msg": "like status updated"})

}

func GetPostsByCreator(c *fiber.Ctx, db *gorm.DB) error {

	UserID, uErr := strconv.ParseUint(c.Query("user_id"), 10, 32)
	if uErr != nil {
		return core.ThrowNewError(c, "failed to get user id")
	}

	var posts []user.Post
	if err := db.Where("user_id = ?", uint(UserID)).Preload("User").Find(&posts).Error; err != nil {
		return core.ThrowNewError(c, "failed to get posts")
	}

	return c.JSON(fiber.Map{"data": posts})

}

func GetPostByID(c *fiber.Ctx, db *gorm.DB) error {

	var post user.Post
	PostID, pErr := strconv.ParseUint(c.Query("post_d"), 10, 32)

	if pErr != nil {
		return core.ThrowNewError(c, "missing post id")
	}

	if err := db.First(&post, PostID).Error; err != nil {
		return core.ThrowNewError(c, "failed to get post")
	}

	return c.JSON(fiber.Map{"data": post})

}

func GetPostLocations(c *fiber.Ctx, db *gorm.DB) error {

	var postLocations []PostLocationI
	var posts []user.Post
	if err := db.Where("is_approved = ?", true).Find(&posts).Error; err != nil {
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

	return c.JSON(fiber.Map{"msg": "", "data": postLocations})
}

func GetPostDetails(c *fiber.Ctx, db *gorm.DB) error {

	PostID, pErr := strconv.ParseUint(c.Query("post_id"), 10, 32)
	if pErr != nil || PostID == 0 {
		return core.ThrowNewError(c, "failed to get user id")
	}

	var post user.Post
	if err := db.Preload("User").Preload("Likers").First(&post, uint(PostID)).Error; err != nil {
		return core.ThrowNewError(c, "failed to get the post")
	}

	return c.JSON(fiber.Map{"data": TransformPost(post)})

}

// TransformPost converts a user.Post into a NestedPost
func TransformPost(p user.Post) NestedPost {
	// Unmarshal images JSON to []string
	var images []string
	if len(p.Images) > 0 {
		_ = json.Unmarshal(p.Images, &images)
	}

	// Transform likers
	likers := make([]NestedUser, len(p.Likers))
	for i, l := range p.Likers {
		likers[i] = NestedUser{
			ID:          l.ID,
			Name:        l.Name,
			PhoneNumber: l.PhoneNumber,
			Photo:       l.Photo,
			Email:       l.Email,
			LastSeen:    l.LastSeen,
			Status:      l.Status,
			Verified:    l.Verified,
			ShowContact: l.ShowContact,
		}
	}

	// Transform author/user (here we assume User is both Author & User)
	author := &NestedUser{
		ID:          p.User.ID,
		Name:        p.User.Name,
		PhoneNumber: p.User.PhoneNumber,
		Photo:       p.User.Photo,
		Email:       p.User.Email,
		LastSeen:    p.User.LastSeen,
		Status:      p.User.Status,
		Verified:    p.User.Verified,
		ShowContact: p.User.ShowContact,
	}

	return NestedPost{
		ID:           p.ID,
		Id:           p.ID,
		Type:         p.Type,
		Author:       author,
		User:         author,
		Price:        p.Price,
		Location:     p.Location,
		IsLiked:      false, // default false, you can compute based on context
		Bedrooms:     p.Bedrooms,
		Bathrooms:    p.Bathrooms,
		Toilets:      p.Toilets,
		Images:       images,
		Likers:       likers,
		IsNegotiable: p.IsNegotiable,
	}
}

func GetFavourites(c *fiber.Ctx, db *gorm.DB, UserID uint) error {

	var user user.User
	if err := db.Preload("Liked.Likers").Preload("Liked").First(&user, UserID).Error; err != nil {
		return core.ThrowNewError(c, "failed to get user")
	}
	return c.JSON(fiber.Map{"data": user.Liked})
}
