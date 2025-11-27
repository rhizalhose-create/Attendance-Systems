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

    // NEW: Log after DB save para confirm storage
    log.Printf("✅ TempUser saved to DB for %s: ID=%d, Code=%s, Expires=%v", 
               tempUser.Email, tempUser.ID, tempUser.VerificationCode, tempUser.ExpiresAt)

    // Cast to string kung int ang GenerateVerificationCode() (uncomment kung need)
    // verificationCodeStr := fmt.Sprintf("%d", verificationCode)  // e.g., int to string
    // verificationCodes[req.Email] = verificationCodeStr
    verificationCodes[req.Email] = verificationCode  // Assume string na

    // NEW: Confirm map save
    log.Printf("Saved verification code in memory for: %s -> %s", req.Email, verificationCodes[req.Email])

    // NEW: Better email send log
    if err := utils.SendVerificationEmail(req.Email, verificationCode); err != nil {
        log.Printf("❌ Failed to send email to %s with code %s: %v (but registration continues)", req.Email, verificationCode, err)
    } else {
        log.Printf("✅ Email sent successfully to %s with code %s", req.Email, verificationCode)
    }

    return c.JSON(fiber.Map{
        "message":    "Registration successful. Please check your email for verification code.",
        "email":      req.Email,
        "expires_in": "24 hours",
        "note":       "Verification code: " + verificationCode,  // Tanggalin sa prod para security
    })
}

func VerifyEmail(c *fiber.Ctx) error {
    type VerifyInput struct {
        Email string `json:"email"`
        Code  string `json:"code"`  // As string para safe
    }
    var input VerifyInput
    if err := c.BodyParser(&input); err != nil {
        log.Println("Parse Error:", err)  // Log input fail
        return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
    }

    log.Printf("VERIFY SCREEN - Verifying code: %s for email: %s", input.Code, input.Email)  // Match your log

    db := config.GetDB()
    var tempUser models.TempUser  // Assume model sa models/temp_user.go
    err := db.Where("email = ? AND verification_code = ?", input.Email, input.Code).First(&tempUser).Error
    if err != nil {
        log.Printf("DB Query Error: %v | Stored codes for %s: [check DB manually]", err, input.Email)  // Key log!
        return c.Status(400).JSON(fiber.Map{"error": "Verification failed"})  // Your generic error
    }

    // If success: Create real user from temp, delete temp, etc.
    log.Println("✅ Verification Success! User ID:", tempUser.ID)
    // ... (move to User table, send welcome email via utils/email.go)

    return c.Status(200).JSON(fiber.Map{"message": "Verified! Proceed to login."})
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