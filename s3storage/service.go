package s3storage

import (
	"context"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Service uploads objects to Amazon S3 (or S3-compatible endpoints).
type Service struct {
	client     *s3.Client
	bucket     string
	publicBase string // optional; trailing slash stripped
}

// NewFromEnv builds an S3 client using default AWS credential chain and region.
// Required: AWS_REGION, AWS_S3_BUCKET.
// Optional: AWS_S3_PUBLIC_BASE_URL (e.g. https://cdn.example.com or https://bucket.s3.region.amazonaws.com),
// AWS_S3_ENDPOINT (custom / LocalStack), AWS_S3_USE_PATH_STYLE=true
func NewFromEnv(ctx context.Context) (*Service, error) {
	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	bucket := strings.TrimSpace(os.Getenv("AWS_S3_BUCKET"))
	if region == "" || bucket == "" {
		return nil, fmt.Errorf("AWS_REGION and AWS_S3_BUCKET are required for S3 storage")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("aws config: %w", err)
	}

	opt := func(o *s3.Options) {
		if ep := strings.TrimSpace(os.Getenv("AWS_S3_ENDPOINT")); ep != "" {
			o.BaseEndpoint = aws.String(ep)
		}
		if strings.EqualFold(strings.TrimSpace(os.Getenv("AWS_S3_USE_PATH_STYLE")), "true") {
			o.UsePathStyle = true
		}
	}
	return &Service{
		client:     s3.NewFromConfig(cfg, opt),
		bucket:     bucket,
		publicBase: strings.TrimRight(strings.TrimSpace(os.Getenv("AWS_S3_PUBLIC_BASE_URL")), "/"),
	}, nil
}

// UploadFile uploads a local file and returns a public HTTPS URL.
// folder is prepended to the object key (sanitized). scalePercent is ignored (reserved for Cloudinary parity).
func (s *Service) UploadFile(ctx context.Context, filePath, folder string, scalePercent int) (string, error) {
	_ = scalePercent
	if filePath == "" {
		return "", fmt.Errorf("empty file path")
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return "", err
	}
	if !st.Mode().IsRegular() {
		return "", fmt.Errorf("not a regular file: %s", filePath)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	key := buildObjectKey(folder, ext)
	ct := mime.TypeByExtension(ext)
	if ct == "" {
		ct = "application/octet-stream"
	}

	upCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	_, err = s.client.PutObject(upCtx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        f,
		ContentType: aws.String(ct),
	})
	if err != nil {
		return "", fmt.Errorf("s3 put object: %w", err)
	}

	return s.objectURL(key), nil
}

func buildObjectKey(folder, ext string) string {
	folder = strings.Trim(strings.TrimSpace(folder), "/")
	base := fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	if folder != "" {
		return folder + "/" + base + ext
	}
	return base + ext
}

func (s *Service) objectURL(key string) string {
	if s.publicBase != "" {
		return s.publicBase + "/" + key
	}
	// Virtual-hosted–style URL (works for public-read or presigned GET elsewhere).
	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, region, key)
}
