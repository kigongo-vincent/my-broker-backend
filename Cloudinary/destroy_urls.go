package cloudinary

import (
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// PublicIDFromDeliveryURL extracts Cloudinary public_id from a typical HTTPS delivery URL.
// Skips transformation and version segments when possible.
func PublicIDFromDeliveryURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || !strings.Contains(strings.ToLower(u.Host), "cloudinary.com") {
		return ""
	}
	p := u.Path
	const marker = "/upload/"
	idx := strings.Index(p, marker)
	if idx < 0 {
		return ""
	}
	tail := p[idx+len(marker):]
	segs := strings.Split(tail, "/")
	var kept []string
	for _, s := range segs {
		if s == "" {
			continue
		}
		if strings.Contains(s, ",") {
			continue
		}
		if len(s) > 1 && s[0] == 'v' && strings.Trim(s[1:], "0123456789") == "" {
			continue
		}
		kept = append(kept, s)
	}
	if len(kept) == 0 {
		return ""
	}
	full := strings.Join(kept, "/")
	ext := path.Ext(full)
	if ext != "" {
		full = strings.TrimSuffix(full, ext)
	}
	return full
}

// DestroyDeliveryURLs deletes Cloudinary assets referenced by delivery URLs (best-effort; logs failures).
func DestroyDeliveryURLs(urls []string) {
	if len(urls) == 0 || strings.TrimSpace(os.Getenv("CLOUDINARY_URL")) == "" {
		return
	}
	svc, err := NewCloudinaryService()
	if err != nil {
		log.Printf("cloudinary: destroy skip (init): %v", err)
		return
	}
	for _, raw := range urls {
		pid := PublicIDFromDeliveryURL(raw)
		if pid == "" {
			continue
		}
		if _, err := svc.cld.Upload.Destroy(svc.ctx, uploader.DestroyParams{PublicID: pid}); err != nil {
			log.Printf("cloudinary: destroy %q: %v", pid, err)
		}
	}
}
