package utils

import "fmt"

// SendPasswordResetEmail sends password reset email with 6-digit code
func SendPasswordResetEmail(email, resetCode string) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Password Reset</title>
  <style>
    /* Your existing CSS styles for reset email */
    body { font-family: 'Inter', Arial, sans-serif; background: linear-gradient(135deg, #667eea, #764ba2); padding: 40px 20px; margin: 0; }
    .container { background: rgba(255, 255, 255, 0.95); backdrop-filter: blur(10px); padding: 48px 40px; border-radius: 20px; max-width: 520px; margin: 0 auto; box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1); border: 1px solid rgba(255, 255, 255, 0.2); }
    .logo { text-align: center; font-size: 30px; font-weight: 800; color: #667eea; letter-spacing: -0.6px; text-transform: uppercase; }
    .header { color: #2d3748; text-align: center; margin-top: 28px; font-size: 24px; font-weight: 700; letter-spacing: -0.4px; }
    p { color: #4a5568; font-size: 15.5px; line-height: 1.7; margin: 16px 0; }
    .reset-code { font-size: 40px; font-weight: 800; color: #667eea; text-align: center; letter-spacing: 6px; background: linear-gradient(90deg, #f0f4ff, #f8faff); padding: 18px 0; border-radius: 14px; margin: 36px 0; box-shadow: inset 0 0 8px rgba(102, 126, 234, 0.15); }
    .note { color: #718096; font-size: 14px; text-align: center; margin-top: 10px; }
    .footer { margin-top: 48px; text-align: center; color: #a0aec0; font-size: 13.5px; border-top: 1px solid #e2e8f0; padding-top: 16px; }
    .highlight { color: #667eea; font-weight: 600; }
    .instruction { background: #f7fafc; padding: 16px; border-radius: 8px; border-left: 4px solid #667eea; margin: 20px 0; }
  </style>
</head>
<body>
  <div class="container">
    <div class="logo">Attendance System</div>
    <h1 class="header">Reset Your Password</h1>
    <p>Hi there,</p>
    <p>We received a request to reset your password for your <strong class="highlight">Attendance System</strong> account. Use the 6-digit reset code below:</p>
    <div class="reset-code">%s</div>
    <div class="instruction">
      <p><strong>How to use this code:</strong></p>
      <ol>
        <li>Go to the password reset page</li>
        <li>Enter your email and this 6-digit code</li>
        <li>Create your new password</li>
      </ol>
    </div>
    <p class="note">This reset code will expire in <strong>1 hour</strong>.</p>
    <p style="text-align: center; color: #4a5568; font-size: 14.5px; margin-top: 16px;">If you didn't request this reset, you can safely ignore this email.</p>
    <div class="footer"><p>© 2025 Attendance System • Complete Management Solution</p></div>
  </div>
</body>
</html>
    `, resetCode)

	return SendEmail(email, "Reset Your Password - Attendance System", htmlBody)
}