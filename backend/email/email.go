package email

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// EmailService provides email functionality
type EmailService struct {
	sesClient *ses.Client
	sender    string
}

// NewEmailService creates a new email service
func NewEmailService(ctx context.Context, region, sender string) (*EmailService, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := ses.NewFromConfig(cfg)

	return &EmailService{
		sesClient: client,
		sender:    sender,
	}, nil
}

// GenerateVerificationCode generates a 6-digit code
func GenerateVerificationCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", r.Intn(1000000))
}

// SendPasswordResetCode sends a password reset code via email
func (s *EmailService) SendPasswordResetCode(recipient, code string) error {
	htmlBody := fmt.Sprintf(`
		<h1>Password Reset Code</h1>
		<p>You requested to reset your password. Use the following code to complete the process:</p>
		<h2 style="background-color: #f4f4f4; padding: 10px; font-size: 24px; text-align: center;">%s</h2>
		<p>If you did not request this, please ignore this email.</p>
		<p>This code will expire in 15 minutes.</p>
	`, code)

	textBody := fmt.Sprintf(
		"Password Reset Code\n\n"+
			"You requested to reset your password. Use the following code to complete the process:\n\n"+
			"%s\n\n"+
			"If you did not request this, please ignore this email.\n"+
			"This code will expire in 15 minutes.",
		code)

	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{recipient},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(htmlBody),
				},
				Text: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(textBody),
				},
			},
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String("Your Password Reset Code"),
			},
		},
		Source: aws.String(s.sender),
	}

	_, err := s.sesClient.SendEmail(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
