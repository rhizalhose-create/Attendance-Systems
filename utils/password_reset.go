// utils/password_reset.go - ADD THESE FUNCTIONS
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// GenerateSixDigitCode generates a random 6-digit code
func GenerateSixDigitCode() string {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback to simple random
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// SendVerificationCodeEmail sends 6-digit code to email for password reset
func SendVerificationCodeEmail(email, code, studentID string) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Password Reset Verification</title>
  <style>
    body { font-family: 'Inter', Arial, sans-serif; background: linear-gradient(135deg, #667eea, #764ba2); padding: 40px 20px; margin: 0; }
    .container { background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(10px); padding: 48px 40px; border-radius: 20px; max-width: 520px; margin: 0 auto; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1); border: 1px solid rgba(255, 255, 255, 0.2); }
    .logo { text-align: center; font-size: 30px; font-weight: 800; color: #667eea; letter-spacing: -0.6px; text-transform: uppercase; }
    .header { color: #2d3748; text-align: center; margin-top: 28px; font-size: 24px; font-weight: 700; letter-spacing: -0.4px; }
    p { color: #4a5568; font-size: 15.5px; line-height: 1.7; margin: 16px 0; }
    .verification-code { font-size: 42px; font-weight: 800; text-align: center; letter-spacing: 12px; background: linear-gradient(135deg, #60B5FF, #4299e1); color: white; padding: 25px; margin: 30px 0; border-radius: 15px; box-shadow: 0 8px 25px rgba(96, 181, 255, 0.4); }
    .note { color: #718096; font-size: 14px; text-align: center; margin-top: 10px; }
    .footer { margin-top: 48px; text-align: center; color: #a0aec0; font-size: 13.5px; border-top: 1px solid #e2e8f0; padding-top: 16px; }
    .highlight { color: #667eea; font-weight: 600; }
    .instruction { background: #f7fafc; padding: 16px; border-radius: 8px; border-left: 4px solid #667eea; margin: 20px 0; }
    .student-id { background: #f0fff4; padding: 12px 20px; border-radius: 8px; border: 1px solid #c6f6d5; margin: 15px 0; text-align: center; }
    .warning { color: #e53e3e; font-weight: 600; text-align: center; }
  </style>
</head>
<body>
  <div class="container">
    <div class="logo">Attendify</div>
    <h1 class="header">Password Reset Verification</h1>
    <p>Hi there,</p>
    <p>We received a request to reset your password for your <strong class="highlight">Attendify</strong> account.</p>
    
    <div class="student-id">
      <p style="margin: 0; font-size: 16px;"><strong>Student ID:</strong> %s</p>
    </div>

    <div class="instruction">
      <p><strong>How to reset your password:</strong></p>
      <ol>
        <li>Enter the verification code below in the app</li>
        <li>You'll be taken to the reset password page</li>
        <li>Enter your new password and confirm it</li>
        <li>Click "Reset Password" to complete the process</li>
      </ol>
    </div>

    <div class="verification-code">%s</div>

    <p class="warning">This verification code will expire in <strong>10 minutes</strong>.</p>
    
    <p style="text-align: center; color: #4a5568; font-size: 14.5px; margin-top: 16px;">
      Enter this code in the app to verify your identity and reset your password.
    </p>

    <p style="text-align: center; color: #e53e3e; font-size: 14px; margin-top: 24px;">
      If you didn't request this reset, you can safely ignore this email.
    </p>

    <div class="footer">
      <p>© 2025 Attendify • Complete Attendance Management Solution</p>
    </div>
  </div>
</body>
</html>
    `, studentID, code)

	return SendEmail(email, "Password Reset Verification Code - Attendify", htmlBody)
}

// GenerateResetToken generates a secure random token for password reset
func GenerateResetToken() string {
    return GenerateRandomString(32) // 32-character random string
}

// GenerateRandomString generates a cryptographically secure random string
func GenerateRandomString(length int) string {
    bytes := make([]byte, length)
    _, err := rand.Read(bytes)
    if err != nil {
        // Fallback to simpler random if crypto/rand fails
        return fallbackRandomString(length)
    }
    return hex.EncodeToString(bytes)
}

// fallbackRandomString generates a random string using math/rand (less secure fallback)
func fallbackRandomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, length)
    for i := range result {
        num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
        result[i] = charset[num.Int64()]
    }
    return string(result)
}