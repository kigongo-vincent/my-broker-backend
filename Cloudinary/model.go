package cloudinary

import (
	"context"

	"github.com/cloudinary/cloudinary-go/v2"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
	ctx context.Context
}
