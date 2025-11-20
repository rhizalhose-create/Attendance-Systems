package handlers

import (
    "log"
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"
    "time"

    "github.com/gofiber/fiber/v2"
    "golang.org/x/crypto/bcrypt"
)



var verificationCodes = make(map[string]string)

func Register(c *fiber.Ctx) error {
    var req models.RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Validate required fields
    if req.Email == "" || req.Password == "" || req.Username == "" || 
       req.FirstName == "" || req.LastName == "" || req.Course == "" || req.YearLevel == "" {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrAllFieldsRequired})
    }

    // Check if user already exists (verified user)
    var existingUser models.User
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&existingUser).Error; err == nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrUserExists})
    }

    // Check if email is already in temporary storage (pending verification)
    var existingTempUser models.TempUser
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&existingTempUser).Error; err == nil {
        // Check if verification is still valid
        if time.Until(existingTempUser.ExpiresAt) > 0 {
            return c.Status(400).JSON(fiber.Map{
                "error": utils.ErrEmailPendingVerification,
                "expires_in": time.Until(existingTempUser.ExpiresAt).Round(time.Minute).String(),
            })
        } else {
            // Delete expired temporary user
            config.DB.Delete(&existingTempUser)
        }
    }

    // Hash password
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrHashPassword})
    }

    // Generate verification code
    verificationCode := utils.GenerateVerificationCode()
    
    log.Printf("📧 Generated verification code for %s: %s", req.Email, verificationCode)

    // Create temporary user
    tempUser := models.TempUser{
        Email:           req.Email,
        Password:        string(hash),
        Username:        req.Username,
        UserID:          req.UserID,
        FirstName:       req.FirstName,
        LastName:        req.LastName,
        MiddleName:      req.MiddleName,
        Course:          req.Course,
        YearLevel:       req.YearLevel,
        Section:         req.Section,
        Department:      req.Department,
        College:         req.College,
        ContactNumber:   req.ContactNumber,
        Address:         req.Address,
        VerificationCode: verificationCode,
        ExpiresAt:       time.Now().Add(24 * time.Hour),
        CreatedAt:       time.Now(),
    }

    // Save to temporary table
    if err := config.DB.Create(&tempUser).Error; err != nil {
        log.Printf("❌ Failed to create temp user: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedProcessRegistration})
    }

    // Store verification code in memory
    verificationCodes[req.Email] = verificationCode
    log.Printf("💾 Saved verification code in memory for: %s", req.Email)

    // Send verification email
    if err := utils.SendVerificationEmail(req.Email, verificationCode); err != nil {
        log.Printf("⚠️ Failed to send email, but registration continues: %v", err)
    }

    return c.JSON(fiber.Map{
        "message":    "Registration successful. Please check your email for verification code.",
        "email":      req.Email,
        "expires_in": "24 hours",
        "note":       "Verification code: " + verificationCode,
    })
}

func VerifyEmail(c *fiber.Ctx) error {
    var req models.VerifyRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    log.Printf("🔍 Verifying email: %s with code: %s", req.Email, req.Code)

    // Step 1: Check in-memory storage first
    if storedCode, exists := verificationCodes[req.Email]; exists {
        log.Printf("📱 Found code in memory for %s", req.Email)
        if storedCode == req.Code {
            log.Printf("✅ Memory verification successful for %s", req.Email)
            return completeVerification(req.Email, c)
        }
    }

    // Step 2: Check database
    var tempUser models.TempUser
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&tempUser).Error; err != nil {
        log.Printf("❌ Temp user not found for email: %s, error: %v", req.Email, err)
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrInvalidVerificationCodeOrEmail})
    }

    log.Printf(" Found temp user: %s, Code: %s, Expires: %s", 
        tempUser.Email, tempUser.VerificationCode, tempUser.ExpiresAt.Format(time.RFC3339))

   
    if tempUser.VerificationCode != req.Code {
        log.Printf("❌ Code mismatch. Expected: %s, Got: %s", tempUser.VerificationCode, req.Code)
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrInvalidVerifyCode})
    }

    if time.Now().After(tempUser.ExpiresAt) {
        log.Printf(" Verification code expired for: %s", req.Email)
        // Delete expired temporary user
        config.DB.Where(utils.QueryEmailWhere, req.Email).Delete(&models.TempUser{})
        delete(verificationCodes, req.Email)
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrVerificationExpired})
    }

    log.Printf("✅ Database verification successful for %s", req.Email)
    return completeVerification(req.Email, c)
}

// Helper function to complete verification process
func completeVerification(email string, c *fiber.Ctx) error {
    // Find the temporary user
    var tempUser models.TempUser
    if err := config.DB.Where(utils.QueryEmailWhere, email).First(&tempUser).Error; err != nil {
        log.Printf("❌ Temp user not found during completion: %s, error: %v", email, err)
        return c.Status(404).JSON(fiber.Map{"error": utils.ErrTempUserNotFound})
    }

    // Create actual user in main users table
    user := models.User{
        Email:         tempUser.Email,
        Password:      tempUser.Password,
        Username:      tempUser.Username,
        Role:          "student",
        IsVerified:    true,
        CreatedAt:     time.Now(),
        VerifiedAt:    time.Now(),
        UserID:        tempUser.UserID,
        FirstName:     tempUser.FirstName,
        LastName:      tempUser.LastName,
        MiddleName:    tempUser.MiddleName,
        Course:        tempUser.Course,
        YearLevel:     tempUser.YearLevel,
        Section:       tempUser.Section,
        Department:    tempUser.Department,
        College:       tempUser.College,
        ContactNumber: tempUser.ContactNumber,
        Address:       tempUser.Address,
    }

    // Save to main users table
    result := config.DB.Create(&user)
    if result.Error != nil {
        log.Printf("❌ Failed to create user: %v", result.Error)
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedCreateUserAccount})
    }

    // Generate custom user ID
    customUserID := utils.GenerateCustomUserID(user.ID)
    if err := config.DB.Model(&user).Update("user_id", customUserID).Error; err != nil {
        log.Printf("❌ Failed to update user_id: %v", err)
    }

    // Generate QR code after user ID is set
    qrCodeData, qrErr := utils.GenerateStudentQRCode(customUserID, user.Email, user.FirstName, user.LastName, user.Course)
    if qrErr != nil {
        log.Printf("⚠️ Failed to generate QR code for user %s: %v", customUserID, qrErr)
    } else {
        // Update user with QR code data
        if err := config.DB.Model(&user).Updates(map[string]interface{}{
            "qr_code_data": qrCodeData,
            "qr_code_type": "student_id",
        }).Error; err != nil {
            log.Printf("⚠️ Failed to update QR code data: %v", err)
        }
        log.Printf("✅ QR code generated successfully for user: %s", customUserID)
    }

    // Clean up: Delete temporary user and verification code
    if err := config.DB.Where(utils.QueryEmailWhere, email).Delete(&models.TempUser{}).Error; err != nil {
        log.Printf("⚠️ Failed to delete temp user: %v", err)
    }
    delete(verificationCodes, email)

    log.Printf("🎉 User verification completed successfully: %s", customUserID)

    return c.JSON(fiber.Map{
        "message":          "Email verified successfully! Your account has been created.",
        "user_id":          customUserID,
        "qr_code_generated": qrErr == nil,
        "student_info": fiber.Map{
            "name":       user.FirstName + " " + user.LastName,
            "course":     user.Course,
            "year_level": user.YearLevel,
            "section":    user.Section,
        },
    })
}

// ResendVerificationCode - Resend verification code for pending registration
func ResendVerificationCode(c *fiber.Ctx) error {
    type ResendRequest struct {
        Email string `json:"email" binding:"required,email"`
    }

    var req ResendRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Check if temporary user exists
    var tempUser models.TempUser
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&tempUser).Error; err != nil {
        log.Printf("❌ Temp user not found for resend: %s, error: %v", req.Email, err)
        return c.Status(404).JSON(fiber.Map{"error": utils.ErrNoPendingRegistration})
    }

    // Check if verification has expired
    if time.Until(tempUser.ExpiresAt) <= 0 {
        config.DB.Where(utils.QueryEmailWhere, req.Email).Delete(&models.TempUser{})
        delete(verificationCodes, req.Email)
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrRegistrationExpired})
    }

    // Generate new verification code
    newVerificationCode := utils.GenerateVerificationCode()

    // Update temporary user with new code
    if err := config.DB.Model(&tempUser).Updates(map[string]interface{}{
        "verification_code": newVerificationCode,
        "expires_at":        time.Now().Add(24 * time.Hour),
    }).Error; err != nil {
        log.Printf("❌ Failed to update temp user: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedResendVerification})
    }

    // Update memory storage
    verificationCodes[req.Email] = newVerificationCode

    // Send new verification email
    if err := utils.SendVerificationEmail(req.Email, newVerificationCode); err != nil {
        log.Printf("⚠️ Failed to send email: %v", err)
    }

    log.Printf("📨 Resent verification code to %s: %s", req.Email, newVerificationCode)

    return c.JSON(fiber.Map{
        "message":    "Verification code resent successfully. Please check your email.",
        "expires_in": "24 hours",
        "note":       "New verification code: " + newVerificationCode,
    })
}

// CleanupExpiredRegistrations - Admin function to clean up expired temporary users
func CleanupExpiredRegistrations(c *fiber.Ctx) error {
    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
    }

    result := config.DB.Where("expires_at < ?", time.Now()).Delete(&models.TempUser{})
    
    deletedCount := result.RowsAffected
    if result.Error != nil {
        log.Printf("❌ Failed to cleanup expired registrations: %v", result.Error)
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedCleanupRegistrations})
    }

    log.Printf("🧹 Cleaned up %d expired temporary users", deletedCount)

    return c.JSON(fiber.Map{
        "message":       "Cleanup completed successfully",
        "deleted_count": deletedCount,
    })
}

// Login function
func Login(c *fiber.Ctx) error {
    var req models.LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    var user models.User
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&user).Error; err != nil {
        return c.Status(401).JSON(fiber.Map{"error": utils.ErrInvalidEmailPass})
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        return c.Status(401).JSON(fiber.Map{"error": utils.ErrInvalidEmailPass})
    }

    if !user.IsVerified && user.Role != "superadmin" {
        return c.Status(401).JSON(fiber.Map{"error": utils.ErrVerifyEmailFirst})
    }

    return c.JSON(fiber.Map{
        "message":     "Login successful",
        "user_id":     user.UserID,
        "email":       user.Email,
        "username":    user.Username,
        "role":        user.Role,
        "is_verified": user.IsVerified,
        "student_info": fiber.Map{
            "first_name": user.FirstName,
            "last_name":  user.LastName,
            "course":     user.Course,
            "year_level": user.YearLevel,
            "section":    user.Section,
        },
        "qr_code_data": user.QRCodeData,
        "qr_code_type": user.QRCodeType,
    })
}

// GetUserProfile function
func GetUserProfile(c *fiber.Ctx) error {
    userID := c.Params("user_id")

    var user models.User
    if err := config.DB.Where(utils.QueryUserIDWhere, userID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": utils.ErrUserNotFound})
    }

    return c.JSON(fiber.Map{
        "user_id":      user.UserID,
        "email":        user.Email,
        "username":     user.Username,
        "role":         user.Role,
        "is_verified":  user.IsVerified,
        "created_at":   user.CreatedAt,
        "qr_code_data": user.QRCodeData,
        "qr_code_type": user.QRCodeType,
        "student_info": fiber.Map{
            "UserId":         user.UserID,
            "first_name":     user.FirstName,
            "last_name":      user.LastName,
            "middle_name":    user.MiddleName,
            "course":         user.Course,
            "year_level":     user.YearLevel,
            "section":        user.Section,
            "department":     user.Department,
            "college":        user.College,
            "contact_number": user.ContactNumber,
            "address":        user.Address,
        },
    })
}