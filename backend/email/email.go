package email

import (
	"context"
	"fmt"
	"go-splitwise/model"
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

// SendMonthlyBalanceReminder sends an email with the user's current balances
func (s *EmailService) SendMonthlyBalanceReminder(recipient, userName string, balances []model.Balance) error {
	htmlBody := fmt.Sprintf(`
		<html>
		<head>
			<style>
				body { font-family: Arial, sans-serif; color: #333; line-height: 1.6; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #5bc5a7; color: white; padding: 10px; text-align: center; }
				.group { margin: 20px 0; border-bottom: 1px solid #eee; padding-bottom: 15px; }
				.group-name { font-weight: bold; background-color: #f5f5f5; padding: 5px; }
				.balance { margin: 10px 0; padding: 5px 0; }
				.owed { color: #5bc5a7; font-weight: bold; }
				.owing { color: #ff652f; font-weight: bold; }
				.settled { color: #999; }
				.footer { margin-top: 20px; font-size: 12px; color: #777; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Monthly Balance Summary</h1>
				</div>
				
				<p>Hello %s,</p>
				<p>Here's a summary of your current balances for this month:</p>
				
				%s
				
				<p>Please log in to settle your balances:</p>
				<p><a href="https://yoursplitwise.com/dashboard">Go to Dashboard</a></p>
				
				<div class="footer">
					<p>This is an automated monthly reminder. Please do not reply to this email.</p>
				</div>
			</div>
		</body>
		</html>
	`, userName, formatBalancesByGroup(balances))

	textBody := fmt.Sprintf(
		"Monthly Balance Summary\n\n"+
			"Hello %s,\n\n"+
			"Here's a summary of your current balances for this month:\n\n"+
			"%s\n\n"+
			"Please log in to settle your balances: https://yoursplitwise.com/dashboard\n\n"+
			"This is an automated monthly reminder. Please do not reply to this email.",
		userName, formatBalancesByGroupText(balances))

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
				Data:    aws.String("Your Monthly Splitwise Balance"),
			},
		},
		Source: aws.String(s.sender),
	}

	_, err := s.sesClient.SendEmail(context.Background(), input)
	if err != nil {
		return fmt.Errorf("failed to send balance reminder email: %w", err)
	}

	return nil
}

// Helper function to format balances by group for HTML email
func formatBalancesByGroup(balances []model.Balance) string {
	// Group balances by group name
	groupedBalances := make(map[string][]model.Balance)

	for _, balance := range balances {
		groupedBalances[balance.GroupName] = append(groupedBalances[balance.GroupName], balance)
	}

	var htmlContent string

	// Generate HTML for each group
	for groupName, groupBalances := range groupedBalances {
		htmlContent += fmt.Sprintf(`
			<div class="group">
				<div class="group-name">%s</div>
		`, groupName)

		for _, balance := range groupBalances {
			var amountClass, amountText string
			if balance.Amount > 0 {
				amountClass = "owed"
				amountText = fmt.Sprintf("You are owed %d", balance.Amount)
			} else if balance.Amount < 0 {
				amountClass = "owing"
				amountText = fmt.Sprintf("You owe %d", -balance.Amount)
			} else {
				amountClass = "settled"
				amountText = "You're settled up"
			}

			htmlContent += fmt.Sprintf(`
				<div class="balance">
					<strong>With %s:</strong> 
					<span class="%s">%s</span>
				</div>
			`, balance.OtherUserName, amountClass, amountText)
		}

		htmlContent += `</div>`
	}

	return htmlContent
}

// Helper function to format balances by group for plain text email
func formatBalancesByGroupText(balances []model.Balance) string {
	// Group balances by group name
	groupedBalances := make(map[string][]model.Balance)

	for _, balance := range balances {
		groupedBalances[balance.GroupName] = append(groupedBalances[balance.GroupName], balance)
	}

	var textContent string

	// Generate text for each group
	for groupName, groupBalances := range groupedBalances {
		textContent += fmt.Sprintf("== %s ==\n", groupName)

		for _, balance := range groupBalances {
			if balance.Amount > 0 {
				textContent += fmt.Sprintf("With %s: You are owed %d\n",
					balance.OtherUserName, balance.Amount)
			} else if balance.Amount < 0 {
				textContent += fmt.Sprintf("With %s: You owe %d\n",
					balance.OtherUserName, -balance.Amount)
			} else {
				textContent += fmt.Sprintf("With %s: You're settled up\n",
					balance.OtherUserName)
			}
		}

		textContent += "\n"
	}

	return textContent
}
