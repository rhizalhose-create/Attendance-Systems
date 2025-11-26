// handlers/auth.go

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

    if req.Email == "" || req.Password == "" || req.Username == "" || 
       req.FirstName == "" || req.LastName == "" || req.Course == "" || req.YearLevel == "" {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrAllFieldsRequired})
    }

    var existingUser models.User
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&existingUser).Error; err == nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrUserExists})
    }

    var existingTempUser models.TempUser
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&existingTempUser).Error; err == nil {
        if time.Until(existingTempUser.ExpiresAt) > 0 {
            return c.Status(400).JSON(fiber.Map{
                "error": utils.ErrEmailPendingVerification,
                "expires_in": time.Until(existingTempUser.ExpiresAt).Round(time.Minute).String(),
            })
        } else {
            config.DB.Delete(&existingTempUser)
        }
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrHashPassword})
    }

    verificationCode := utils.GenerateVerificationCode()
    
    log.Printf("Generated verification code for %s: %s", req.Email, verificationCode)

    tempUser := models.TempUser{
        Email:           req.Email,
        Password:        string(hash),
        Username:        req.Username,
        StudentID:       req.StudentID,
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

    if err := config.DB.Create(&tempUser).Error; err != nil {
        log.Printf("Failed to create temp user: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedProcessRegistration})
    }

    verificationCodes[req.Email] = verificationCode
    log.Printf("Saved verification code in memory for: %s", req.Email)

    if err := utils.SendVerificationEmail(req.Email, verificationCode); err != nil {
        log.Printf("Failed to send email, but registration continues: %v", err)
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

    log.Printf("Verifying email: %s with code: %s", req.Email, req.Code)

    if storedCode, exists := verificationCodes[req.Email]; exists {
        log.Printf("Found code in memory for %s", req.Email)
        if storedCode == req.Code {
            log.Printf("Memory verification successful for %s", req.Email)
            return completeVerification(req.Email, c)
        }
    }

    var tempUser models.TempUser
    if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&tempUser).Error; err != nil {
        log.Printf("Temp user not found for email: %s, error: %v", req.Email, err)
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrInvalidVerificationCodeOrEmail})
    }

    log.Printf("Found temp user: %s, Code: %s, Expires: %s", 
        tempUser.Email, tempUser.VerificationCode, tempUser.ExpiresAt.Format(time.RFC3339))

    if tempUser.VerificationCode != req.Code {
        log.Printf("Code mismatch. Expected: %s, Got: %s", tempUser.VerificationCode, req.Code)
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrInvalidVerifyCode})
    }

    if time.Now().After(tempUser.ExpiresAt) {
        log.Printf("Verification code expired for: %s", req.Email)
        config.DB.Where(utils.QueryEmailWhere, req.Email).Delete(&models.TempUser{})
        delete(verificationCodes, req.Email)
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrVerificationExpired})
    }

    log.Printf("Database verification successful for %s", req.Email)
    return completeVerification(req.Email, c)
}

func completeVerification(email string, c *fiber.Ctx) error {
    var tempUser models.TempUser
    if err := config.DB.Where(utils.QueryEmailWhere, email).First(&tempUser).Error; err != nil {
        log.Printf("Temp user not found during completion: %s, error: %v", email, err)
        return c.Status(404).JSON(fiber.Map{"error": utils.ErrTempUserNotFound})
    }

    user := models.User{
        Email:         tempUser.Email,
        Password:      tempUser.Password,
        Username:      tempUser.Username,
        Role:          "student",
        IsVerified:    true,
        CreatedAt:     time.Now(),
        VerifiedAt:    time.Now(),
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

    result := config.DB.Create(&user)
    if result.Error != nil {
        log.Printf("Failed to create user: %v", result.Error)
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedCreateUserAccount})
    }

    customStudentID := utils.GenerateCustomStudentID(user.ID)
    if err := config.DB.Model(&user).Update("student_id", customStudentID).Error; err != nil {
        log.Printf("Failed to update student_id: %v", err)
    }

    qrCodeData, qrErr := utils.GenerateStudentQRCode(customStudentID, user.Email, user.FirstName, user.LastName, user.Course)
    if qrErr != nil {
        log.Printf("Failed to generate QR code for user %s: %v", customStudentID, qrErr)
    } else {
        if err := config.DB.Model(&user).Updates(map[string]interface{}{
            "qr_code_data": qrCodeData,
            "qr_code_type": "student_id",
        }).Error; err != nil {
            log.Printf("Failed to update QR code data: %v", err)
        }
        log.Printf("QR code generated successfully for user: %s", customStudentID)
    }

    if err := config.DB.Where(utils.QueryEmailWhere, email).Delete(&models.TempUser{}).Error; err != nil {
        log.Printf("Failed to delete temp user: %v", err)
    }
    delete(verificationCodes, email)

    log.Printf("User verification completed successfully: %s", customStudentID)

    return c.JSON(fiber.Map{
        "message":          "Email verified succexssfully! Your account has been created.",
        "student_id":       customStudentID,
        "qr_code_generated": qrErr == nil,
        "student_info": fiber.Map{
            "name":       user.FirstName + " " + user.LastName,
            "course":     user.Course,
            "year_level": user.YearLevel,
            "section":    user.Section,
        },
    })
}
func Login(c *fiber.Ctx) error {
    var req models.LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Validation
    if req.StudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Student ID is required"})
    }

    if req.Password == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Password is required"})
    }

    var user models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
        log.Printf("Student not found with ID: %s, error: %v", req.StudentID, err)
        return c.Status(401).JSON(fiber.Map{"error": utils.ErrInvalidStudentIDPass})
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        return c.Status(401).JSON(fiber.Map{"error": utils.ErrInvalidStudentIDPass})
    }

    if !user.IsVerified && user.Role != "superadmin" {
        return c.Status(401).JSON(fiber.Map{"error": utils.ErrVerifyEmailFirst})
    }

    return c.JSON(fiber.Map{
        "message":     "Login successful",
        "student_id":  user.StudentID,
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
func GetUserProfile(c *fiber.Ctx) error {
    studentID := c.Params("student_id")

    var user models.User
    if err := config.DB.Where("student_id = ?", studentID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": utils.ErrUserNotFound})
    }

    return c.JSON(fiber.Map{
        "student_id":   user.StudentID,
        "email":        user.Email,
        "username":     user.Username,
        "role":         user.Role,
        "is_verified":  user.IsVerified,
        "created_at":   user.CreatedAt,
        "qr_code_data": user.QRCodeData,
        "qr_code_type": user.QRCodeType,
        "student_info": fiber.Map{
            "StudentID":      user.StudentID,
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