package cloudinary

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func NewCloudinaryService() (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromURL(os.Getenv("CLOUDINARY_URL"))
	if err != nil {
		return nil, err
	}
	return &CloudinaryService{
		cld: cld,
		ctx: context.Background(),
	}, nil
}

// UploadFile uploads a file and optionally scales it down.
// scalePercent: 100 = original size, 50 = 50%, etc.
func (s *CloudinaryService) UploadFile(filePath, folder string, scalePercent int) (string, error) {
	if scalePercent <= 0 || scalePercent > 100 {
		scalePercent = 100
	}

	resp, err := s.cld.Upload.Upload(s.ctx, filePath, uploader.UploadParams{
		Folder:         folder,
		Transformation: fmt.Sprintf("c_scale,w_%d", scalePercent), // scales width proportionally
	})
	if err != nil {
		return "", err
	}
	return resp.SecureURL, nil
}
