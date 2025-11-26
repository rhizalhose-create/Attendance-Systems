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

	log.Printf("🔧 Email Config - Host: %s, Port: %d, Email: %s", smtpHost, 587, smtpEmail)
	
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

	log.Printf(" Attempting to send email:")
	log.Printf("   To: %s", to)
	log.Printf("   Subject: %s", subject)
	log.Printf("   From: %s", config.SMTPEmail)
	log.Printf("   SMTP Host: %s:%d", config.SMTPHost, config.SMTPPort)

	if config.SMTPEmail == "" || config.SMTPPassword == "" {
		log.Printf(" Email credentials not set!")
		log.Printf("   SMTP_EMAIL: '%s'", config.SMTPEmail)
		log.Printf("   SMTP_PASSWORD length: %d", len(config.SMTPPassword))
		return fmt.Errorf("email service not configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTPEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPEmail, config.SMTPPassword)

	log.Printf(" Testing SMTP connection...")
	
	// Test the connection first
	var s gomail.SendCloser
	var err error
	if s, err = d.Dial(); err != nil {
		log.Printf(" SMTP Connection FAILED: %v", err)
		return fmt.Errorf("SMTP connection failed: %v", err)
	}
	defer s.Close()
	log.Printf(" SMTP Connection SUCCESSFUL")

	log.Printf(" Sending email...")
	if err := d.DialAndSend(m); err != nil {
		log.Printf(" Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf(" Email sent successfully to %s", to)
	return nil
}