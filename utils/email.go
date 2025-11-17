package utils

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/gomail.v2"
)

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPEmail    string
	SMTPPassword string
	SMTPHost     string
	SMTPPort     int
}

// GetEmailConfig retrieves email configuration from environment variables
func GetEmailConfig() *EmailConfig {
	smtpEmail := os.Getenv("SMTP_EMAIL")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	
	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
	}

	return &EmailConfig{
		SMTPEmail:    smtpEmail,
		SMTPPassword: smtpPassword,
		SMTPHost:     smtpHost,
		SMTPPort:     587,
	}
}

// SendEmail sends a generic email
func SendEmail(to, subject, htmlBody string) error {
	config := GetEmailConfig()

	if config.SMTPEmail == "" || config.SMTPPassword == "" {
		log.Printf(" Email credentials not set. Would send to %s: %s", to, subject)
		return fmt.Errorf("email service not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTPEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPEmail, config.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		log.Printf(" Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf(" Email sent to %s", to)
	return nil
}