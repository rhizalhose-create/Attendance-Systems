// handlers/password-reset.go - COMPLETE UPDATED VERSION
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

// RequestPasswordReset - Step 1: Request password reset with Student ID (sends 6-digit code)
func RequestPasswordReset(c *fiber.Ctx) error {
    type Request struct {
        StudentID string `json:"student_id"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    if req.StudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
    }

    // Find user by Student ID
    var user models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
        // Return success even if user not found for security
        log.Printf(" User not found for password reset - Student ID: %s", req.StudentID)
        return c.JSON(fiber.Map{
            "success": true,
            "message": "If the Student ID exists, a verification code will be sent to your registered email",
        })
    }

    // Check if user is verified
    if !user.IsVerified {
        return c.Status(400).JSON(fiber.Map{"error": "Please verify your email first before resetting password"})
    }

    // Check reset attempts (prevent abuse)
    if user.ResetAttempts >= 5 && time.Since(user.LastResetRequest) < time.Hour {
        return c.Status(429).JSON(fiber.Map{"error": "Too many reset attempts. Please try again later."})
    }

    // Generate 6-digit verification code
    verificationCode := utils.GenerateSixDigitCode()
    resetToken := utils.GenerateResetToken()
    resetTokenExpiry := time.Now().Add(10 * time.Minute) // Code valid for 10 minutes

    // Update user with reset token and verification code
    updates := map[string]interface{}{
        "reset_token":         resetToken,
        "reset_token_expiry":  resetTokenExpiry,
        "reset_attempts":      user.ResetAttempts + 1,
        "last_reset_request":  time.Now(),
        "verification_code":   verificationCode, // Store the 6-digit code
    }

    if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
        log.Printf(" Failed to set reset token for Student ID %s: %v", req.StudentID, err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to process reset request"})
    }

    // Send email with 6-digit verification code
    if err := utils.SendVerificationCodeEmail(user.Email, verificationCode, user.StudentID); err != nil {
        log.Printf("⚠️ Failed to send verification code email to %s (Student ID: %s): %v", user.Email, req.StudentID, err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to send verification code"})
    }

    log.Printf(" Password reset requested for Student ID: %s - Email: %s - Code: %s", req.StudentID, user.Email, verificationCode)

    return c.JSON(fiber.Map{
        "success":     true,
        "message":    "6-digit verification code sent to your email",
        "note":       "Check your registered email for the verification code",
        "email":      user.Email,
        "student_id": user.StudentID,
        "expires_in": "10 minutes",
    })
}

// VerifyResetCode - Step 2: Verify the 6-digit code
func VerifyResetCode(c *fiber.Ctx) error {
    type Request struct {
        StudentID string `json:"student_id"`
        Code      string `json:"code"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    if req.StudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
    }

    if req.Code == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Verification code is required"})
    }

    // Find user by Student ID and check verification code
    var user models.User
    if err := config.DB.Where("student_id = ? AND verification_code = ? AND reset_token_expiry > ?", 
        req.StudentID, req.Code, time.Now()).First(&user).Error; err != nil {
        log.Printf(" Invalid verification code for Student ID %s: %s", req.StudentID, req.Code)
        return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired verification code"})
    }

    // Return the reset token for the next step
    return c.JSON(fiber.Map{
        "success":     true,
        "message":     "Verification code verified successfully",
        "token":       user.ResetToken,
        "email":       user.Email,
        "student_id":  user.StudentID,
        "expires_in":  time.Until(user.ResetTokenExpiry).Round(time.Minute).String(),
    })
}

// ResetPassword - Step 3: Reset password with token after code verification
func ResetPassword(c *fiber.Ctx) error {
    type Request struct {
        StudentID      string `json:"student_id"`
        Token          string `json:"token"`
        NewPassword    string `json:"new_password"`
        ConfirmPassword string `json:"confirm_password"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    if req.StudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
    }

    if req.Token == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Reset token is required"})
    }

    if req.NewPassword == "" {
        return c.Status(400).JSON(fiber.Map{"error": "New password is required"})
    }

    if req.ConfirmPassword == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Confirm password is required"})
    }

    if req.NewPassword != req.ConfirmPassword {
        return c.Status(400).JSON(fiber.Map{"error": "Passwords do not match"})
    }

    if len(req.NewPassword) < 6 {
        return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 6 characters"})
    }

    // Find user by Student ID and valid reset token
    var user models.User
    if err := config.DB.Where("student_id = ? AND reset_token = ? AND reset_token_expiry > ?", 
        req.StudentID, req.Token, time.Now()).First(&user).Error; err != nil {
        log.Printf(" Invalid or expired reset token for Student ID %s: %s", req.StudentID, req.Token)
        return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired reset token"})
    }

    // Hash new password
    hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
    }

    // Update user password and clear reset data
    updates := map[string]interface{}{
        "password":           string(hash),
        "reset_token":        nil,
        "reset_token_expiry": nil,
        "reset_attempts":     0,
        "verification_code":  nil, // Clear the verification code
    }

    if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
        log.Printf(" Failed to reset password for Student ID %s: %v", req.StudentID, err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to reset password"})
    }

    log.Printf(" Password reset successful for Student ID: %s", req.StudentID)

    return c.JSON(fiber.Map{
        "success":    true,
        "message":    "Password reset successfully",
        "student_id": user.StudentID,
        "email":      user.Email,
    })
}

// ResendVerificationCode - Resend 6-digit code
func ResendVerificationCode(c *fiber.Ctx) error {
    type Request struct {
        StudentID string `json:"student_id"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    if req.StudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
    }

    // Find user by Student ID
    var user models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student ID not found"})
    }

    // Generate new 6-digit code
    newCode := utils.GenerateSixDigitCode()
    newExpiry := time.Now().Add(10 * time.Minute)

    // Update verification code
    updates := map[string]interface{}{
        "verification_code":  newCode,
        "reset_token_expiry": newExpiry,
        "last_reset_request": time.Now(),
    }

    if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to generate new code"})
    }

    // Send new code via email
    if err := utils.SendVerificationCodeEmail(user.Email, newCode, user.StudentID); err != nil {
        log.Printf("⚠️ Failed to resend verification code to %s: %v", user.Email, err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to send verification code"})
    }

    return c.JSON(fiber.Map{
        "success":    true,
        "message":    "New verification code sent to your email",
        "email":      user.Email,
        "student_id": user.StudentID,
        "expires_in": "10 minutes",
    })
}