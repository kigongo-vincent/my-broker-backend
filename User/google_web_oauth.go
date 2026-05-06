package user

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kigongo-vincent/my-broker-backend/core"
	"gorm.io/gorm"
)

const (
	googleAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	googleTokenURL = "https://oauth2.googleapis.com/token"
)

func googleWebRedirectURI() string {
	if u := strings.TrimSpace(os.Getenv("GOOGLE_OAUTH_REDIRECT_URI")); u != "" {
		return u
	}
	base := strings.TrimSpace(os.Getenv("PUBLIC_API_URL"))
	if base == "" {
		base = "http://localhost:3001"
	}
	return strings.TrimRight(base, "/") + "/auth/google/web/callback"
}

func webAppOrigin() string {
	o := strings.TrimSpace(os.Getenv("WEB_APP_ORIGIN"))
	if o == "" {
		o = "http://localhost:8081"
	}
	return strings.TrimRight(o, "/")
}

func cookieSecure() bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(os.Getenv("PUBLIC_API_URL"))), "https://") ||
		strings.HasPrefix(strings.ToLower(webAppOrigin()), "https://")
}

func randomOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sanitizeReturnPath(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "/(tabs)/home"
	}
	if !strings.HasPrefix(s, "/") || strings.HasPrefix(s, "//") {
		return "/(tabs)/home"
	}
	for _, c := range s {
		if c == '?' || c == '#' || c == '\\' {
			return "/(tabs)/home"
		}
	}
	return s
}

// GoogleWebOAuthStart redirects the browser to Google OAuth (web client only).
func GoogleWebOAuthStart(c *fiber.Ctx) error {
	clientID := strings.TrimSpace(os.Getenv("GOOGLE_WEB_CLIENT_ID"))
	if clientID == "" {
		return c.Status(fiber.StatusServiceUnavailable).SendString("GOOGLE_WEB_CLIENT_ID is not configured")
	}
	state, err := randomOAuthState()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("failed to generate oauth state")
	}
	ret := sanitizeReturnPath(c.Query("return_to", "/(tabs)/home"))

	secure := cookieSecure()
	c.Cookie(&fiber.Cookie{
		Name:     "g_oauth_state",
		Value:    state,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   secure,
		MaxAge:   600,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "g_oauth_return",
		Value:    ret,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		Secure:   secure,
		MaxAge:   600,
	})

	q := url.Values{}
	q.Set("client_id", clientID)
	q.Set("redirect_uri", googleWebRedirectURI())
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	q.Set("access_type", "online")
	q.Set("prompt", "select_account")

	return c.Redirect(googleAuthURL+"?"+q.Encode(), fiber.StatusFound)
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	IDToken     string `json:"id_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Error       string `json:"error"`
	Description string `json:"error_description"`
}

// GoogleWebOAuthCallback completes the OAuth code exchange, verifies the ID token, upserts the user, and redirects to the web app with a JWT in the URL fragment.
func GoogleWebOAuthCallback(c *fiber.Ctx, db *gorm.DB) error {
	origin := webAppOrigin()
	signIn := origin + "/auth/signin"

	if errParam := strings.TrimSpace(c.Query("error")); errParam != "" {
		desc := strings.TrimSpace(c.Query("error_description"))
		q := url.Values{}
		q.Set("error", "google_oauth")
		if desc != "" {
			q.Set("detail", desc)
		} else if errParam != "" {
			q.Set("detail", errParam)
		}
		return c.Redirect(signIn+"?"+q.Encode(), fiber.StatusFound)
	}

	code := strings.TrimSpace(c.Query("code"))
	state := strings.TrimSpace(c.Query("state"))
	cookieState := strings.TrimSpace(c.Cookies("g_oauth_state"))
	returnPath := sanitizeReturnPath(c.Cookies("g_oauth_return"))

	clearOAuthCookies := func() {
		secure := cookieSecure()
		exp := &fiber.Cookie{
			Name:     "g_oauth_state",
			Value:    "",
			Path:     "/",
			HTTPOnly: true,
			MaxAge:   -1,
			Secure:   secure,
		}
		c.Cookie(exp)
		exp.Name = "g_oauth_return"
		c.Cookie(exp)
	}
	defer clearOAuthCookies()

	if code == "" || state == "" || cookieState == "" {
		return c.Redirect(signIn+"?error=google_oauth&detail=missing_code_or_state", fiber.StatusFound)
	}
	if len(state) != len(cookieState) || subtle.ConstantTimeCompare([]byte(state), []byte(cookieState)) != 1 {
		return c.Redirect(signIn+"?error=google_oauth&detail=invalid_state", fiber.StatusFound)
	}

	clientID := strings.TrimSpace(os.Getenv("GOOGLE_WEB_CLIENT_ID"))
	clientSecret := strings.TrimSpace(os.Getenv("GOOGLE_WEB_CLIENT_SECRET"))
	if clientID == "" || clientSecret == "" {
		return c.Redirect(signIn+"?error=google_oauth&detail=server_not_configured", fiber.StatusFound)
	}

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", clientID)
	form.Set("client_secret", clientSecret)
	form.Set("redirect_uri", googleWebRedirectURI())
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequest(http.MethodPost, googleTokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return c.Redirect(signIn+"?error=google_oauth&detail=token_request", fiber.StatusFound)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return c.Redirect(signIn+"?error=google_oauth&detail=token_exchange", fiber.StatusFound)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return c.Redirect(signIn+"?error=google_oauth&detail=token_http", fiber.StatusFound)
	}

	var tok googleTokenResponse
	if err := json.Unmarshal(body, &tok); err != nil || tok.IDToken == "" {
		return c.Redirect(signIn+"?error=google_oauth&detail=no_id_token", fiber.StatusFound)
	}
	if tok.Error != "" {
		return c.Redirect(signIn+"?error=google_oauth&detail="+url.QueryEscape(tok.Error), fiber.StatusFound)
	}

	webIDs := []string{clientID}
	profile, err := verifyGoogleIDToken(tok.IDToken, webIDs)
	if err != nil {
		return c.Redirect(signIn+"?error=google_oauth&detail=invalid_id_token", fiber.StatusFound)
	}
	if profile.Email == "" {
		return c.Redirect(signIn+"?error=google_oauth&detail=no_email", fiber.StatusFound)
	}

	u, err := upsertGoogleUser(db, profile)
	if err != nil {
		return c.Redirect(signIn+"?error=google_oauth&detail=user_upsert", fiber.StatusFound)
	}
	jwt, err := core.IssueJWT(u.ID)
	if err != nil {
		return c.Redirect(signIn+"?error=google_oauth&detail=jwt", fiber.StatusFound)
	}

	frag := url.Values{}
	frag.Set("jwt", jwt)
	loc := origin + returnPath + "#" + frag.Encode()
	return c.Redirect(loc, fiber.StatusFound)
}
