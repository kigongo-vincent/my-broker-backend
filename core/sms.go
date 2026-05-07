package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

func SendOTP(phoneNumber string, otp int) error {
	dryRun := strings.ToLower(strings.TrimSpace(os.Getenv("SMS_DRY_RUN")))
	// Default to dry-run for local/dev safety unless explicitly disabled.
	if dryRun == "" || dryRun == "true" || dryRun == "1" {
		return nil
	}

	provider := strings.ToLower(strings.TrimSpace(os.Getenv("SMS_PROVIDER")))
	switch provider {
	case "", "africastalking":
		return sendWithAfricasTalking(phoneNumber, otp)
	case "sns":
		return sendWithSNS(phoneNumber, otp)
	default:
		return fmt.Errorf("unsupported SMS provider: %s", provider)
	}
}

func sendWithSNS(phoneNumber string, otp int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	if region == "" {
		return fmt.Errorf("AWS_REGION is required for SMS_PROVIDER=sns")
	}

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	client := sns.NewFromConfig(cfg)
	e164 := SMSDestinationE164(phoneNumber)
	if !strings.HasPrefix(e164, "+") {
		return fmt.Errorf("sns: phone must be E.164 (got %q)", e164)
	}

	msg := fmt.Sprintf("Your My Broker OTP is %d. It expires shortly.", otp)
	_, err = client.Publish(ctx, &sns.PublishInput{
		PhoneNumber: aws.String(e164),
		Message:     aws.String(msg),
	})
	if err != nil {
		return fmt.Errorf("sns publish: %w", err)
	}
	return nil
}

func africasTalkingAPIKey() string {
	k := strings.TrimSpace(os.Getenv("SMS_API_KEY"))
	if k != "" {
		return k
	}
	return strings.TrimSpace(os.Getenv("AT_API_KEY"))
}

func sendWithAfricasTalking(phoneNumber string, otp int) error {
	apiKey := africasTalkingAPIKey()
	username := strings.TrimSpace(os.Getenv("AT_USERNAME"))
	sender := strings.TrimSpace(os.Getenv("AT_SENDER_ID"))

	if apiKey == "" || username == "" {
		return fmt.Errorf("SMS_API_KEY (or AT_API_KEY) and AT_USERNAME are required for Africa's Talking")
	}

	to := SMSDestinationE164(phoneNumber)
	if to == "" {
		return fmt.Errorf("africastalking: empty SMS destination")
	}

	msg := fmt.Sprintf("Your My Broker OTP is %d. It expires shortly.", otp)
	form := url.Values{}
	form.Set("username", username)
	form.Set("to", to)
	form.Set("message", msg)
	if sender != "" {
		form.Set("from", sender)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.africastalking.com/version1/messaging", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("africastalking request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("africastalking send: %w", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(res.Body, 4096))

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		msg := strings.TrimSpace(string(body))
		if msg == "" {
			return fmt.Errorf("africastalking returned status %d", res.StatusCode)
		}
		return fmt.Errorf("africastalking returned status %d: %s", res.StatusCode, msg)
	}

	if err := validateAfricasTalkingResponse(body); err != nil {
		return fmt.Errorf("africastalking response not accepted: %w", err)
	}
	return nil
}

type atSMSResponse struct {
	SMSMessageData struct {
		Recipients []struct {
			Number    string `json:"number"`
			Status    string `json:"status"`
			StatusCode int   `json:"statusCode,string"`
			MessageID string `json:"messageId"`
		} `json:"recipients"`
	} `json:"SMSMessageData"`
}

func validateAfricasTalkingResponse(body []byte) error {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return fmt.Errorf("empty response body")
	}
	var resp atSMSResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("invalid JSON response: %w", err)
	}
	if len(resp.SMSMessageData.Recipients) == 0 {
		return fmt.Errorf("no recipients in provider response")
	}
	for _, r := range resp.SMSMessageData.Recipients {
		status := strings.ToLower(strings.TrimSpace(r.Status))
		accepted := strings.Contains(status, "success") || strings.Contains(status, "sent") || r.StatusCode == 101 || r.StatusCode == 100
		if accepted {
			return nil
		}
	}
	r := resp.SMSMessageData.Recipients[0]
	return fmt.Errorf("recipient status=%q code=%d number=%q", strings.TrimSpace(r.Status), r.StatusCode, strings.TrimSpace(r.Number))
}
