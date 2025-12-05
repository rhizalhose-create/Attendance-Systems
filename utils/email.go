package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	smtpPortStr := os.Getenv("SMTP_PORT")

	// ADDED DEBUG LOGS
	log.Printf("🔧 DEBUG EMAIL CONFIG:")
	log.Printf("   SMTP_EMAIL: '%s'", smtpEmail)
	log.Printf("   SMTP_PASSWORD length: %d", len(smtpPassword))
	log.Printf("   SMTP_HOST: '%s'", smtpHost)
	log.Printf("   SMTP_PORT: '%s'", smtpPortStr)

	if smtpHost == "" {
		smtpHost = "smtp.gmail.com"
		log.Printf("   Using default SMTP_HOST: %s", smtpHost)
	}

	smtpPort := 587
	if smtpPortStr != "" {
		if port, err := strconv.Atoi(smtpPortStr); err == nil {
			smtpPort = port
		}
	}
	log.Printf("   Final SMTP_PORT: %d", smtpPort)

	if smtpEmail == "" || smtpPassword == "" {
		log.Printf("❌ CRITICAL: Email credentials not configured!")
		log.Printf("   Check .env file for SMTP_EMAIL and SMTP_PASSWORD")
	}

	return &EmailConfig{
		SMTPEmail:    smtpEmail,
		SMTPPassword: smtpPassword,
		SMTPHost:     smtpHost,
		SMTPPort:     smtpPort,
	}
}

// SendEmail sends a generic email
func SendEmail(to, subject, htmlBody string) error {
	config := GetEmailConfig()

	// ADDED DEBUG LOGS
	log.Printf("📧 DEBUG SEND EMAIL:")
	log.Printf("   To: %s", to)
	log.Printf("   Subject: %s", subject)
	log.Printf("   From: %s", config.SMTPEmail)
	log.Printf("   SMTP Host: %s:%d", config.SMTPHost, config.SMTPPort)

	if config.SMTPEmail == "" || config.SMTPPassword == "" {
		log.Printf("❌ EMAIL CREDENTIALS NOT SET!")
		log.Printf("   SMTP_EMAIL: '%s'", config.SMTPEmail)
		log.Printf("   SMTP_PASSWORD length: %d", len(config.SMTPPassword))
		return fmt.Errorf("email service not configured. Check .env file")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTPEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.SMTPEmail, config.SMTPPassword)

	log.Printf("   Testing SMTP connection...")

	// Test the connection first
	var s gomail.SendCloser
	var err error
	if s, err = d.Dial(); err != nil {
		log.Printf("❌ SMTP Connection FAILED: %v", err)
		return fmt.Errorf("SMTP connection failed: %v", err)
	}
	defer s.Close()
	log.Printf("   SMTP Connection SUCCESSFUL")

	log.Printf("   Sending email...")
	if err := d.DialAndSend(m); err != nil {
		log.Printf("❌ Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("✅ Email sent successfully to %s", to)
	return nil
}

// SendStudentIDEmail sends the student ID to the user's email after verification
func SendStudentIDEmail(email, studentID, firstName string) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Your Student ID - Login Credentials</title>
  <style>
    body { font-family: 'Inter', Arial, sans-serif; background: linear-gradient(135deg, #667eea, #764ba2); padding: 40px 20px; margin: 0; }
    .container { background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(10px); padding: 48px 40px; border-radius: 20px; max-width: 520px; margin: 0 auto; box-shadow: 0 8px 32px rgba(102, 126, 234, 0.15); border: 1px solid rgba(102, 126, 234, 0.1); }
    .logo { text-align: center; font-size: 30px; font-weight: 800; color: #667eea; letter-spacing: -0.6px; text-transform: uppercase; }
    .header { color: #2d3748; text-align: center; margin-top: 28px; font-size: 24px; font-weight: 700; letter-spacing: -0.4px; }
    p { color: #4a5568; font-size: 15.5px; line-height: 1.7; margin: 16px 0; }
    .student-id-box { font-size: 18px; font-weight: 600; color: #667eea; text-align: center; background: linear-gradient(90deg, #f0f4ff, #f5f9ff); padding: 24px; border-radius: 14px; margin: 32px 0; box-shadow: inset 0 0 8px rgba(102, 126, 234, 0.1); border-left: 4px solid #667eea; }
    .note { color: #718096; font-size: 14px; margin-top: 16px; padding: 12px; background: #f7fafc; border-radius: 8px; border-left: 3px solid #cbd5e0; }
    .footer { margin-top: 48px; text-align: center; color: #a0aec0; font-size: 13.5px; border-top: 1px solid #e2e8f0; padding-top: 16px; }
    .highlight { color: #667eea; font-weight: 600; }
    .cta-button { display: inline-block; background: linear-gradient(135deg, #667eea, #764ba2); color: white; padding: 14px 36px; border-radius: 8px; text-decoration: none; font-weight: 600; margin-top: 16px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="logo">Attendance System</div>
    <h1 class="header">Account Verified!</h1>
    <p>Hi <strong>%s</strong>,</p>
    <p>Congratulations! Your email has been verified and your account is now active. Your <strong class="highlight">Student ID</strong> (login credential) is shown below:</p>
    <div class="student-id-box">
      <div style="font-size: 12px; color: #718096; margin-bottom: 8px;">YOUR STUDENT ID</div>
      <div style="font-size: 28px; font-weight: 800; letter-spacing: 2px; word-break: break-all;">%s</div>
    </div>
    <p><strong>📝 Important:</strong> Use your <strong class="highlight">Student ID</strong> and password to log in to the Attendance System app.</p>
    <div class="note">
      <strong>💡 Tip:</strong> Save your Student ID in a secure place. You'll need it every time you log in.
    </div>
    <p style="text-align: center; margin-top: 24px;">
      Ready to get started? Log in with your Student ID and password.
    </p>
    <div class="footer"><p>© 2025 Attendance System • Complete Management Solution</p></div>
  </div>
</body>
</html>
	`, firstName, studentID)

	log.Printf("📧 Sending Student ID email to: %s with Student ID: %s", email, studentID)
	return SendEmail(email, "Your Student ID - Account Verified", htmlBody)
}