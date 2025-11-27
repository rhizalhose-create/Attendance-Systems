// utils/password_reset.go - FIXED MODERN EMAIL VERSION
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"time"
)

// GenerateSixDigitCode generates a random 6-digit code
func GenerateSixDigitCode() string {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// Fallback to simple random
		fallback := time.Now().UnixNano() % 1000000
		if fallback < 0 {
			fallback = -fallback
		}
		code := fmt.Sprintf("%06d", fallback)
		log.Printf("🔑 UTILS - Generated fallback code: %s", code)
		return code
	}
	code := fmt.Sprintf("%06d", n.Int64())
	log.Printf("🔑 UTILS - Generated 6-digit code: %s", code)
	return code
}

// GenerateResetToken generates a secure random token for password reset
func GenerateResetToken() string {
    log.Printf("🔑 UTILS - Starting token generation...")
    
    bytes := make([]byte, 32)
    n, err := rand.Read(bytes)
    
    log.Printf("🔑 UTILS - rand.Read bytes: %d, error: %v", n, err)
    
    if err != nil {
        log.Printf("❌ UTILS - Crypto rand failed: %v", err)
        token := fmt.Sprintf("reset_%d", time.Now().UnixNano())
        log.Printf("🔑 UTILS - Generated fallback token: %s (length: %d)", token, len(token))
        return token
    }
    
    token := hex.EncodeToString(bytes)
    log.Printf("✅ UTILS - Generated crypto token: %s (length: %d)", token, len(token))
    return token
}

// SendVerificationCodeEmail sends 6-digit code to email for password reset - FIXED FORMATTING
func SendVerificationCodeEmail(email, code, studentID string) error {
    currentTime := time.Now().Format("January 2, 2006 at 3:04 PM")
    
    htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Reset Your Password • Attendify</title>
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap');
    
    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    
    body {
      font-family: 'Inter', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
      padding: 40px 20px;
      margin: 0;
      min-height: 100vh;
    }
    
    .email-container {
      max-width: 520px;
      margin: 0 auto;
      background: #ffffff;
      border-radius: 24px;
      overflow: hidden;
      box-shadow: 
        0 20px 60px rgba(0, 0, 0, 0.15),
        0 4px 12px rgba(0, 0, 0, 0.08);
      border: 1px solid rgba(255, 255, 255, 0.2);
    }
    
    .header-gradient {
      background: linear-gradient(135deg, #60B5FF 0%%, #4299e1 100%%);
      padding: 40px 32px;
      text-align: center;
      position: relative;
      overflow: hidden;
    }
    
    .header-gradient::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: url("data:image/svg+xml,%%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%%3E%%3Cg fill='none' fill-rule='evenodd'%%3E%%3Cg fill='%%23ffffff' fill-opacity='0.1'%%3E%%3Ccircle cx='30' cy='30' r='4'/%%3E%%3C/g%%3E%%3C/g%%3E%%3C/svg%%3E");
    }
    
    .logo {
      font-size: 32px;
      font-weight: 800;
      color: white;
      letter-spacing: -0.5px;
      margin-bottom: 16px;
      position: relative;
      z-index: 2;
    }
    
    .logo-subtitle {
      color: rgba(255, 255, 255, 0.9);
      font-size: 14px;
      font-weight: 500;
      letter-spacing: 0.5px;
      text-transform: uppercase;
      position: relative;
      z-index: 2;
    }
    
    .content {
      padding: 48px 40px;
    }
    
    .title {
      color: #1a202c;
      font-size: 28px;
      font-weight: 700;
      line-height: 1.2;
      margin-bottom: 16px;
      text-align: center;
      letter-spacing: -0.3px;
    }
    
    .subtitle {
      color: #718096;
      font-size: 16px;
      line-height: 1.6;
      text-align: center;
      margin-bottom: 32px;
    }
    
    .info-card {
      background: #f8fafc;
      border-radius: 16px;
      padding: 24px;
      margin-bottom: 32px;
      border: 1px solid #e2e8f0;
    }
    
    .info-row {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;
    }
    
    .info-row:last-child {
      margin-bottom: 0;
    }
    
    .info-label {
      color: #4a5568;
      font-size: 14px;
      font-weight: 500;
    }
    
    .info-value {
      color: #2d3748;
      font-size: 15px;
      font-weight: 600;
    }
    
    .verification-section {
      text-align: center;
      margin-bottom: 40px;
    }
    
    .verification-label {
      color: #4a5568;
      font-size: 14px;
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 1px;
      margin-bottom: 16px;
    }
    
    .verification-code {
      font-size: 52px;
      font-weight: 800;
      color: #1a202c;
      letter-spacing: 8px;
      background: linear-gradient(135deg, #60B5FF, #4299e1);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
      background-clip: text;
      margin: 8px 0;
      padding: 0 8px;
    }
    
    .steps {
      display: grid;
      gap: 20px;
      margin-bottom: 40px;
    }
    
    .step {
      display: flex;
      align-items: flex-start;
      gap: 16px;
    }
    
    .step-number {
      background: #60B5FF;
      color: white;
      width: 32px;
      height: 32px;
      border-radius: 50%%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      font-weight: 700;
      flex-shrink: 0;
    }
    
    .step-content {
      flex: 1;
    }
    
    .step-title {
      color: #2d3748;
      font-size: 16px;
      font-weight: 600;
      margin-bottom: 4px;
    }
    
    .step-description {
      color: #718096;
      font-size: 14px;
      line-height: 1.5;
    }
    
    .warning-card {
      background: linear-gradient(135deg, #fed7d7, #feb2b2);
      border-radius: 16px;
      padding: 20px;
      text-align: center;
      margin-bottom: 32px;
      border: 1px solid #fc8181;
    }
    
    .warning-icon {
      font-size: 24px;
      margin-bottom: 8px;
    }
    
    .warning-text {
      color: #c53030;
      font-size: 14px;
      font-weight: 600;
      line-height: 1.4;
    }
    
    .expiry-notice {
      background: #fffaf0;
      border: 1px solid #faf089;
      border-radius: 12px;
      padding: 16px;
      text-align: center;
      margin-bottom: 32px;
    }
    
    .expiry-text {
      color: #d69e2e;
      font-size: 14px;
      font-weight: 600;
    }
    
    .footer {
      text-align: center;
      padding-top: 32px;
      border-top: 1px solid #e2e8f0;
    }
    
    .footer-text {
      color: #a0aec0;
      font-size: 13px;
      line-height: 1.5;
    }
    
    .security-notice {
      background: #f0fff4;
      border: 1px solid #9ae6b4;
      border-radius: 12px;
      padding: 16px;
      text-align: center;
      margin-top: 24px;
    }
    
    .security-text {
      color: #38a169;
      font-size: 13px;
      font-weight: 500;
    }
    
    @media (max-width: 600px) {
      body {
        padding: 20px 16px;
      }
      
      .content {
        padding: 32px 24px;
      }
      
      .header-gradient {
        padding: 32px 24px;
      }
      
      .title {
        font-size: 24px;
      }
      
      .verification-code {
        font-size: 42px;
        letter-spacing: 6px;
      }
    }
  </style>
</head>
<body>
  <div class="email-container">
    <!-- Header with Gradient -->
    <div class="header-gradient">
      <div class="logo">Attendify</div>
      <div class="logo-subtitle">Smart Attendance Management</div>
    </div>
    
    <!-- Main Content -->
    <div class="content">
      <h1 class="title">Reset Your Password</h1>
      <p class="subtitle">Enter the verification code below to secure your account</p>
      
      <!-- User Information -->
      <div class="info-card">
        <div class="info-row">
          <span class="info-label">Student ID:</span>
          <span class="info-value">%s</span>
        </div>
        <div class="info-row">
          <span class="info-label">Request Time:</span>
          <span class="info-value">%s</span>
        </div>
      </div>
      
      <!-- Verification Code -->
      <div class="verification-section">
        <div class="verification-label">Your Verification Code</div>
        <div class="verification-code">%s</div>
        <div class="verification-label">Valid for 10 minutes</div>
      </div>
      
      <!-- Steps -->
      <div class="steps">
        <div class="step">
          <div class="step-number">1</div>
          <div class="step-content">
            <div class="step-title">Enter the Code</div>
            <div class="step-description">Return to the Attendify app and enter this 6-digit verification code</div>
          </div>
        </div>
        
        <div class="step">
          <div class="step-number">2</div>
          <div class="step-content">
            <div class="step-title">Set New Password</div>
            <div class="step-description">Create a strong, unique password for your account</div>
          </div>
        </div>
        
        <div class="step">
          <div class="step-number">3</div>
          <div class="step-content">
            <div class="step-title">Secure Your Account</div>
            <div class="step-description">You'll be redirected to login with your new credentials</div>
          </div>
        </div>
      </div>
      
      <!-- Expiry Notice -->
      <div class="expiry-notice">
        <div class="expiry-text">⏰ This code expires in 10 minutes for security reasons</div>
      </div>
      
      <!-- Warning -->
      <div class="warning-card">
        <div class="warning-icon">⚠️</div>
        <div class="warning-text">
          If you didn't request this password reset, please ignore this email and ensure your account is secure.
        </div>
      </div>
      
      <!-- Footer -->
      <div class="footer">
        <p class="footer-text">
          © 2025 Attendify • Complete Attendance Management Solution<br>
          This is an automated message, please do not reply to this email.
        </p>
        
        <div class="security-notice">
          <div class="security-text">🔒 Your security is our priority. This code ensures only you can reset your password.</div>
        </div>
      </div>
    </div>
  </div>
</body>
</html>
`, studentID, currentTime, code)

    log.Printf("📧 UTILS - Sending password reset email to: %s", email)
    log.Printf("📧 UTILS - Student ID: %s", studentID)
    log.Printf("📧 UTILS - Verification Code: %s", code)
    log.Printf("📧 UTILS - Request Time: %s", currentTime)

    return SendEmail(email, "Reset Your Password • Attendify", htmlBody)
}