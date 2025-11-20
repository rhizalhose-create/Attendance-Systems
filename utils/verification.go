package utils

import (
	"fmt"
	"math/rand"
	"time"
)


func GenerateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}


func GenerateResetCode() string {
	return GenerateVerificationCode()
}


func SendVerificationEmail(email, verificationCode string) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Email Verification</title>
  <style>
    /* Your existing CSS styles here */
    body { font-family: 'Inter', Arial, sans-serif; background: linear-gradient(135deg, #ffecd2, #fcb69f); padding: 40px 20px; margin: 0; }
    .container { background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(10px); padding: 48px 40px; border-radius: 20px; max-width: 520px; margin: 0 auto; box-shadow: 0 8px 32px rgba(255, 107, 107, 0.15); border: 1px solid rgba(255, 107, 107, 0.1); }
    .logo { text-align: center; font-size: 30px; font-weight: 800; color: #ff6b6b; letter-spacing: -0.6px; text-transform: uppercase; }
    .header { color: #2d3748; text-align: center; margin-top: 28px; font-size: 24px; font-weight: 700; letter-spacing: -0.4px; }
    p { color: #4a5568; font-size: 15.5px; line-height: 1.7; margin: 16px 0; }
    .code { font-size: 40px; font-weight: 800; color: #ff6b6b; text-align: center; letter-spacing: 6px; background: linear-gradient(90deg, #ffe4e6, #fff1f2); padding: 18px 0; border-radius: 14px; margin: 36px 0; box-shadow: inset 0 0 8px rgba(255, 107, 107, 0.15); }
    .note { color: #718096; font-size: 14px; text-align: center; margin-top: 10px; }
    .footer { margin-top: 48px; text-align: center; color: #a0aec0; font-size: 13.5px; border-top: 1px solid #e2e8f0; padding-top: 16px; }
    .highlight { color: #ff6b6b; font-weight: 600; }
  </style>
</head>
<body>
  <div class="container">
    <div class="logo">Attendance System</div>
    <h1 class="header">Verify Your Email</h1>
    <p>Hi there,</p>
    <p>Welcome to <strong class="highlight">Attendance System</strong> — your complete attendance management solution. To activate your account, please use the verification code below:</p>
    <div class="code">%s</div>
    <p class="note">This code will expire in <strong>24 hours</strong>.</p>
    <p style="text-align: center; color: #4a5568; font-size: 14.5px; margin-top: 16px;">If you didn't request this, you can safely ignore this email.</p>
    <div class="footer"><p>© 2025 Attendance System • Complete Management Solution</p></div>
  </div>
</body>
</html>
	`, verificationCode)

	return SendEmail(email, "Verify Your Email - Attendance System", htmlBody)
}