// handlers/password-reset.go
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

const (
    ErrCannotParseJSON     = "Cannot parse JSON"
    ErrStudentIDRequired   = "Student ID is required"
)

func RequestPasswordReset(c *fiber.Ctx) error {
    type Request struct {
        StudentID string `json:"student_id"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": ErrCannotParseJSON})
    }

    if req.StudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": ErrStudentIDRequired})
    }

    var user models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
        return c.JSON(fiber.Map{
            "success": true,
            "message": "If the Student ID exists, a verification code will be sent to your registered email",
        })
    }

    if !user.IsVerified {
        return c.Status(400).JSON(fiber.Map{"error": "Please verify your email first before resetting password"})
    }

    var existingReset models.PasswordReset
    if err := config.DB.Where("student_id = ? AND used = ?", req.StudentID, false).First(&existingReset).Error; err == nil {
        if time.Until(existingReset.ExpiresAt) > 0 {
            return c.Status(400).JSON(fiber.Map{
                "error": "Reset request already pending. Please check your email or wait for it to expire.",
            })
        } else {
            config.DB.Delete(&existingReset)
        }
    }

    verificationCode := utils.GenerateSixDigitCode()
    resetToken := utils.GenerateResetToken()
    expiresAt := time.Now().Add(10 * time.Minute)

    if resetToken == "" {
        resetToken = "fallback_token_" + verificationCode
    }

    passwordReset := models.PasswordReset{
        StudentID:        req.StudentID,
        Email:            user.Email,
        VerificationCode: verificationCode,
        Token:            resetToken,
        ExpiresAt:        expiresAt,
        Used:             false,
        CreatedAt:        time.Now(),
    }

    if err := config.DB.Create(&passwordReset).Error; err != nil {
        log.Printf("Failed to create password reset entry: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to process reset request"})
    }

    // Verify the entry was saved correctly
    var savedReset models.PasswordReset
    if err := config.DB.Where("student_id = ? AND verification_code = ?", req.StudentID, verificationCode).First(&savedReset).Error; err != nil {
        log.Printf("CRITICAL: Could not find saved reset entry: %v", err)
    } else {
        log.Printf("Password reset entry saved - ID: %d, Code: %s, Token: %s", savedReset.ID, savedReset.VerificationCode, savedReset.Token)
    }

    if err := utils.SendVerificationCodeEmail(user.Email, verificationCode, user.StudentID); err != nil {
        log.Printf("Failed to send verification email: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to send verification code"})
    }

    return c.JSON(fiber.Map{
        "success":     true,
        "message":    "6-digit verification code sent to your email",
        "email":      user.Email,
        "student_id": user.StudentID,
        "expires_in": "10 minutes",
    })
}

func VerifyResetCode(c *fiber.Ctx) error {
    type Request struct {
        StudentID string `json:"student_id"`
        Code      string `json:"code"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        log.Printf("VerifyResetCode - JSON parse error: %v", err)
        return c.Status(400).JSON(fiber.Map{"error": ErrCannotParseJSON})
    }

    if req.StudentID == "" {
        log.Printf("VerifyResetCode - Student ID is empty")
        return c.Status(400).JSON(fiber.Map{"error": ErrStudentIDRequired})
    }

    if req.Code == "" {
        log.Printf("VerifyResetCode - Code is empty for Student ID: %s", req.StudentID)
        return c.Status(400).JSON(fiber.Map{"error": "Verification code is required"})
    }

    log.Printf("VerifyResetCode - Searching for Student ID: %s, Code: %s", req.StudentID, req.Code)

    // First, check what entries exist for this student
    var allResets []models.PasswordReset
    config.DB.Where("student_id = ?", req.StudentID).Find(&allResets)
    log.Printf("VerifyResetCode - Found %d reset entries for student", len(allResets))
    
    for i, reset := range allResets {
        log.Printf("VerifyResetCode - Entry %d: Code=%s, Token=%s, Used=%v, Expires=%v", 
            i+1, reset.VerificationCode, reset.Token, reset.Used, reset.ExpiresAt)
    }

    var passwordReset models.PasswordReset
    err := config.DB.Where("student_id = ? AND verification_code = ? AND used = ? AND expires_at > ?", 
        req.StudentID, req.Code, false, time.Now()).First(&passwordReset).Error

    if err != nil {
        log.Printf("VerifyResetCode - Query failed: %v", err)
        log.Printf("VerifyResetCode - Query: student_id=%s, code=%s, used=false, expires_at>%v", 
            req.StudentID, req.Code, time.Now())
        
        // Try simpler query without expiration check
        var simpleReset models.PasswordReset
        if simpleErr := config.DB.Where("student_id = ? AND verification_code = ?", req.StudentID, req.Code).First(&simpleReset).Error; simpleErr == nil {
            log.Printf("VerifyResetCode - Found with simple query but: Used=%v, Expired=%v", 
                simpleReset.Used, time.Now().After(simpleReset.ExpiresAt))
            
            if simpleReset.Used {
                return c.Status(400).JSON(fiber.Map{"error": "This verification code has already been used"})
            }
            if time.Now().After(simpleReset.ExpiresAt) {
                return c.Status(400).JSON(fiber.Map{"error": "This verification code has expired"})
            }
        }
        
        return c.Status(400).JSON(fiber.Map{"error": "Invalid verification code"})
    }

    log.Printf("VerifyResetCode - Found valid reset: Token=%s", passwordReset.Token)

    if passwordReset.Token == "" {
        log.Printf("VerifyResetCode - Token is empty")
        return c.Status(500).JSON(fiber.Map{"error": "Reset token not generated properly"})
    }

    log.Printf("VerifyResetCode - Verification successful")

    return c.JSON(fiber.Map{
        "success":     true,
        "message":     "Verification code verified successfully",
        "token":       passwordReset.Token,
        "email":       passwordReset.Email,
        "student_id":  passwordReset.StudentID,
        "expires_in":  time.Until(passwordReset.ExpiresAt).Round(time.Minute).String(),
    })
}

func ResetPassword(c *fiber.Ctx) error {
    type Request struct {
        StudentID      string `json:"student_id"`
        Token          string `json:"token"`
        NewPassword    string `json:"new_password"`
        ConfirmPassword string `json:"confirm_password"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": ErrCannotParseJSON})
    }

    if req.StudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": ErrStudentIDRequired})
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

    var passwordReset models.PasswordReset
    if err := config.DB.Where("student_id = ? AND token = ? AND used = ? AND expires_at > ?", 
        req.StudentID, req.Token, false, time.Now()).First(&passwordReset).Error; err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid or expired reset token"})
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 14)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
    }

    var user models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    if err := config.DB.Model(&user).Update("password", string(hash)).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to reset password"})
    }

    config.DB.Model(&passwordReset).Update("used", true)

    return c.JSON(fiber.Map{
        "success":    true,
        "message":    "Password reset successfully",
        "student_id": user.StudentID,
        "email":      user.Email,
    })
}

func ResendVerificationCode(c *fiber.Ctx) error {
    type Request struct {
        StudentID string `json:"student_id"`
    }

    var req Request
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": ErrCannotParseJSON})
    }

    if req.StudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": ErrStudentIDRequired})
    }

    var user models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student ID not found"})
    }

    newCode := utils.GenerateSixDigitCode()
    newToken := utils.GenerateResetToken()
    newExpiry := time.Now().Add(10 * time.Minute)

    if newToken == "" {
        newToken = "resend_token_" + newCode
    }

    var passwordReset models.PasswordReset
    if err := config.DB.Where("student_id = ? AND used = ?", req.StudentID, false).First(&passwordReset).Error; err != nil {
        passwordReset = models.PasswordReset{
            StudentID:        req.StudentID,
            Email:            user.Email,
            VerificationCode: newCode,
            Token:            newToken,
            ExpiresAt:        newExpiry,
            Used:             false,
            CreatedAt:        time.Now(),
        }
        if err := config.DB.Create(&passwordReset).Error; err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to create reset request"})
        }
    } else {
        updates := map[string]interface{}{
            "verification_code": newCode,
            "token":             newToken,
            "expires_at":        newExpiry,
        }
        if err := config.DB.Model(&passwordReset).Updates(updates).Error; err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to generate new code"})
        }
    }

    if err := utils.SendVerificationCodeEmail(user.Email, newCode, user.StudentID); err != nil {
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

func DebugPasswordReset(c *fiber.Ctx) error {
    studentId := c.Query("student_id")
    
    if studentId == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
    }

    var user models.User
    if err := config.DB.Where("student_id = ?", studentId).First(&user).Error; err != nil {
        return c.JSON(fiber.Map{"user_found": false})
    }

    var passwordResets []models.PasswordReset
    config.DB.Where("student_id = ?", studentId).Order("created_at DESC").Find(&passwordResets)

    resetData := []map[string]interface{}{}
    for _, reset := range passwordResets {
        resetData = append(resetData, map[string]interface{}{
            "verification_code": reset.VerificationCode,
            "token":             reset.Token,
            "used":              reset.Used,
            "expires_at":        reset.ExpiresAt,
            "created_at":        reset.CreatedAt,
            "is_expired":        time.Now().After(reset.ExpiresAt),
        })
    }

    return c.JSON(fiber.Map{
        "user_found": true,
        "user": map[string]interface{}{
            "student_id":  user.StudentID,
            "email":       user.Email,
            "is_verified": user.IsVerified,
        },
        "password_resets": resetData,
        "current_time":    time.Now(),
    })
}