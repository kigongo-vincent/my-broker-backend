package core

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"os"
	"strings"
)

// EmailOTPPepper secrets OTP hashes; prefer EMAIL_OTP_PEPPER, else JWT_SECRET.
func EmailOTPPepper() string {
	p := strings.TrimSpace(os.Getenv("EMAIL_OTP_PEPPER"))
	if p != "" {
		return p
	}
	return strings.TrimSpace(os.Getenv("JWT_SECRET"))
}

// HashEmailOTP returns a hex-encoded SHA-256 of pepper + code (trimmed).
func HashEmailOTP(code string) string {
	sum := sha256.Sum256([]byte(EmailOTPPepper() + "|otp|" + strings.TrimSpace(code)))
	return hex.EncodeToString(sum[:])
}

// EmailOTPCodeMatches compares stored hex hash with a freshly hashed code (constant-time).
func EmailOTPCodeMatches(storedHex, code string) bool {
	if storedHex == "" || code == "" {
		return false
	}
	got := HashEmailOTP(code)
	return subtle.ConstantTimeCompare([]byte(storedHex), []byte(got)) == 1
}

// EmailDryRun skips outbound SES when true (default true if unset).
func EmailDryRun() bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("EMAIL_DRY_RUN")))
	if v == "" {
		return true
	}
	return v == "true" || v == "1"
}
