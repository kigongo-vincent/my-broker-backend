package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// SendTransactionalEmail sends a simple HTML + text email via AWS SES v2.
func SendTransactionalEmail(toEmail, subject, textBody, htmlBody string) error {
	toEmail = strings.TrimSpace(strings.ToLower(toEmail))
	if toEmail == "" {
		return fmt.Errorf("empty recipient email")
	}
	if EmailDryRun() {
		log.Printf("[email dry-run] to=%s subject=%s body=%q", toEmail, subject, textBody)
		return nil
	}

	from := strings.TrimSpace(os.Getenv("SES_FROM_EMAIL"))
	if from == "" {
		return fmt.Errorf("SES_FROM_EMAIL is required")
	}
	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	if region == "" {
		return fmt.Errorf("AWS_REGION is required for SES")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("aws config: %w", err)
	}

	client := sesv2.NewFromConfig(cfg)
	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(from),
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data:    aws.String(subject),
					Charset: aws.String("UTF-8"),
				},
				Body: &types.Body{
					Text: &types.Content{
						Data:    aws.String(textBody),
						Charset: aws.String("UTF-8"),
					},
					Html: &types.Content{
						Data:    aws.String(htmlBody),
						Charset: aws.String("UTF-8"),
					},
				},
			},
		},
	}
	if reply := sesReplyToAddresses(); len(reply) > 0 {
		input.ReplyToAddresses = reply
	}

	_, err = client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("ses send: %w", err)
	}
	return nil
}

func sesReplyToAddresses() []string {
	raw := strings.TrimSpace(os.Getenv("SES_REPLY_TO"))
	if raw == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(raw, ",") {
		s := strings.TrimSpace(p)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
