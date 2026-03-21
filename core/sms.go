package core

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func SendOTP(phoneNumber string, otp int) error {
	dryRun := strings.ToLower(strings.TrimSpace(os.Getenv("SMS_DRY_RUN")))
	// Default to dry-run for local/dev safety unless explicitly disabled.
	if dryRun == "" || dryRun == "true" || dryRun == "1" {
		return nil
	}

	provider := strings.ToLower(strings.TrimSpace(os.Getenv("SMS_PROVIDER")))
	if provider == "" || provider == "africastalking" {
		return sendWithAfricasTalking(phoneNumber, otp)
	}
	return fmt.Errorf("unsupported SMS provider: %s", provider)
}

func sendWithAfricasTalking(phoneNumber string, otp int) error {
	apiKey := os.Getenv("AT_API_KEY")
	username := os.Getenv("AT_USERNAME")
	sender := os.Getenv("AT_SENDER_ID")

	if apiKey == "" || username == "" {
		return fmt.Errorf("AT_API_KEY or AT_USERNAME missing")
	}

	msg := fmt.Sprintf("Your My Broker OTP is %d. It expires shortly.", otp)
	form := url.Values{}
	form.Set("username", username)
	form.Set("to", phoneNumber)
	form.Set("message", msg)
	if sender != "" {
		form.Set("from", sender)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.africastalking.com/version1/messaging", strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("SMS provider returned status %d", res.StatusCode)
	}
	return nil
}
