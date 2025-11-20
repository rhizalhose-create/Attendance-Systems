package handlers

import (
    "log"
    "time"
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"

    "github.com/gofiber/fiber/v2"
    "golang.org/x/crypto/bcrypt"
)

// RequestPasswordReset - Step 1: Request password reset with 6-digit code
func RequestPasswordReset(c *fiber.Ctx) error {
    type Request struct {
        Email string `json:"email" binding:"required,email"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    if req.Email == "" {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrEmailRequired})
    }

    // Find user by email
    var user models.User
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&user).Error; err != nil {
        // Return success even if user not found for security
        log.Printf(" User not found for password reset: %s", req.Email)
        return c.JSON(fiber.Map{
            "message": "If the email exists, a reset code will be sent",
            "note":    "Check your email for 6-digit reset code",
        })
    }

    // Check if user is verified
    if !user.IsVerified {
        return c.Status(400).JSON(fiber.Map{"error": "Please verify your email first before resetting password"})
    }

    // Check reset attempts (prevent abuse)
    if user.ResetAttempts >= 5 && time.Since(user.LastResetRequest) < time.Hour {
        return c.Status(429).JSON(fiber.Map{"error": utils.ErrResetAttemptsExceeded})
    }

    // Generate 6-digit reset code
    resetCode := utils.GenerateResetCode()
    resetCodeExpiry := time.Now().Add(1 * time.Hour)

    // Update user with reset code (store in reset_token field)
    updates := map[string]interface{}{
        "reset_token":         resetCode,
        "reset_token_expiry":  resetCodeExpiry,
        "reset_attempts":      user.ResetAttempts + 1,
        "last_reset_request":  time.Now(),
    }

    if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
        log.Printf(" Failed to set reset code for %s: %v", req.Email, err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to process reset request"})
    }

    // Send reset email with 6-digit code
    if err := utils.SendPasswordResetEmail(user.Email, resetCode); err != nil {
        log.Printf("⚠️ Failed to send reset email to %s: %v", user.Email, err)
   
    }

    log.Printf(" Password reset requested for: %s - Code: %s", user.Email, resetCode)

    return c.JSON(fiber.Map{
        "message": "Password reset code sent successfully",
        "note":    "Check your email for 6-digit reset code",
        "code":    resetCode, // Only for development - remove in production
    })
}

// ResetPassword - Step 2: Reset password with 6-digit code
func ResetPassword(c *fiber.Ctx) error {
    type Request struct {
        Email       string `json:"email" binding:"required,email"`
        Code        string `json:"code" binding:"required"`
        NewPassword string `json:"new_password" binding:"required"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    if req.Email == "" {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrEmailRequired})
    }

    if req.Code == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Reset code is required"})
    }

    if req.NewPassword == "" {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrPasswordRequired})
    }

    if len(req.NewPassword) < 6 {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrPasswordTooShort})
    }

    // Find user by email and valid reset code
    var user models.User
    if err := config.DB.Where("email = ? AND reset_token = ? AND reset_token_expiry > ?", 
        req.Email, req.Code, time.Now()).First(&user).Error; err != nil {
        log.Printf(" Invalid or expired reset code for %s: %s", req.Email, req.Code)
        return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired reset code"})
    }

    // Hash new password
    hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrHashPassword})
    }

    // Update user password and clear reset code
    updates := map[string]interface{}{
        "password":           string(hash),
        "reset_token":        nil,
        "reset_token_expiry": nil,
        "reset_attempts":     0,
    }

    if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
        log.Printf("❌ Failed to reset password for %s: %v", user.Email, err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to reset password"})
    }

    log.Printf("✅ Password reset successful for: %s", user.Email)

    return c.JSON(fiber.Map{
        "message": "Password reset successfully",
        "user_id": user.UserID,
        "email":   user.Email,
    })
}

// VerifyResetCode - Step 1.5: Verify if reset code is valid (optional)
func VerifyResetCode(c *fiber.Ctx) error {
    email := c.Query("email")
    code := c.Query("code")
    
    if email == "" || code == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Email and reset code are required"})
    }

    var user models.User
    if err := config.DB.Where("email = ? AND reset_token = ? AND reset_token_expiry > ?", 
        email, code, time.Now()).First(&user).Error; err != nil {
        return c.Status(400).JSON(fiber.Map{
            "valid":   false,
            "error":   "Invalid or expired reset code",
            "message": "Invalid or expired reset code",
        })
    }

    return c.JSON(fiber.Map{
        "valid":    true,
        "message":  "Reset code is valid",
        "email":    user.Email,
        "user_id":  user.UserID,
        "expires_in": time.Until(user.ResetTokenExpiry).Round(time.Minute).String(),
    })
}