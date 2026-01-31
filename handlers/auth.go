package handlers

import (
	"AttendanceManagementSystem/config"
	"AttendanceManagementSystem/models"
	"AttendanceManagementSystem/utils"
	"log"
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

	// ADDED DEBUG LOG
	log.Printf("🔍 DEBUG REGISTER:")
	log.Printf("   Email: %s", req.Email)
	log.Printf("   Username: %s", req.Username)
	log.Printf("   Name: %s %s", req.FirstName, req.LastName)

	var existingUser models.User
	if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&existingUser).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{"error": utils.ErrUserExists})
	}

	var existingTempUser models.TempUser
	if err := config.DB.Where(utils.QueryEmailWhere, req.Email).First(&existingTempUser).Error; err == nil {
		if time.Until(existingTempUser.ExpiresAt) > 0 {
			return c.Status(400).JSON(fiber.Map{
				"error":      utils.ErrEmailPendingVerification,
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

	// ADDED DEBUG LOGS
	log.Printf("🔑 DEBUG VERIFICATION CODE:")
	log.Printf("   Generated code: %s", verificationCode)
	log.Printf("   For email: %s", req.Email)
	log.Printf("   Will expire at: %v", time.Now().Add(24*time.Hour))

	// Generate StudentID if not provided during registration
	studentID := req.StudentID
	if studentID == "" {
		studentID = "TEMP_" + utils.GenerateVerificationCode()
		log.Printf("   Generated temporary StudentID: %s", studentID)
	}

	tempUser := models.TempUser{
		Email:            req.Email,
		Password:         string(hash),
		Username:         req.Username,
		StudentID:        studentID,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		MiddleName:       req.MiddleName,
		Course:           req.Course,
		YearLevel:        req.YearLevel,
		Section:          req.Section,
		Department:       req.Department,
		College:          req.College,
		ContactNumber:    req.ContactNumber,
		Address:          req.Address,
		VerificationCode: verificationCode,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now(),
	}

	// Save temp user and verify
	if err := config.DB.Create(&tempUser).Error; err != nil {
		log.Printf("❌ FAILED TO SAVE TEMP USER: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedProcessRegistration})
	}

	// ADDED: Verify it was saved
	var savedTemp models.TempUser
	if err := config.DB.Where("email = ?", req.Email).First(&savedTemp).Error; err != nil {
		log.Printf("❌ COULD NOT RETRIEVE SAVED TEMP USER: %v", err)
	} else {
		log.Printf("✅ Temp user saved: ID=%d, Code='%s', Expires=%v",
			savedTemp.ID, savedTemp.VerificationCode, savedTemp.ExpiresAt)
	}

	verificationCodes[req.Email] = verificationCode

	// NEW: TEMPORARY DISABLE EMAIL FOR DEBUGGING - UNCOMMENT LATER
	log.Printf("📧 DEBUG: Email sending temporarily disabled for testing")
	log.Printf("📧 VERIFICATION CODE for %s: %s", req.Email, verificationCode)
	log.Printf("📧 Use this code in the /verify endpoint")

	// COMMENTED OUT FOR DEBUGGING - UNCOMMENT WHEN READY

	if err := utils.SendVerificationEmail(req.Email, verificationCode); err != nil {
		log.Printf("❌ Failed to send email to %s with code %s: %v (but registration continues)", req.Email, verificationCode, err)
	} else {
		log.Printf("✅ Email sent successfully to %s with code %s", req.Email, verificationCode)
	}

	return c.JSON(fiber.Map{
		"message":            "Registration successful. Please use the verification code below.",
		"email":              req.Email,
		"expires_in":         "24 hours",
		"verification_code":  verificationCode, // For debugging - remove in production
		"note":               "Email sending disabled for debugging. Use the code above to verify.",
		"debug_temp_user_id": tempUser.ID,
	})
}

func VerifyEmail(c *fiber.Ctx) error {
	type VerifyInput struct {
		Email string `json:"email"`
		Code  string `json:"code"` // As string para safe
	}
	var input VerifyInput
	if err := c.BodyParser(&input); err != nil {
		log.Println("❌ PARSE ERROR:", err)
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// ADDED DEBUG LOGS
	log.Printf("🔍 VERIFY EMAIL DEBUG:")
	log.Printf("   Email: %s", input.Email)
	log.Printf("   Code: %s", input.Code)
	log.Printf("   Current time: %v", time.Now())

	db := config.GetDB()
	var tempUser models.TempUser

	// ADD THIS: Check what's in database
	var allTempUsers []models.TempUser
	db.Where("email = ?", input.Email).Find(&allTempUsers)
	log.Printf("📊 Found %d temp users for email %s:", len(allTempUsers), input.Email)
	for i, tu := range allTempUsers {
		log.Printf("   %d. ID: %d, Code: '%s', Expires: %v, IsExpired: %v",
			i+1, tu.ID, tu.VerificationCode, tu.ExpiresAt, time.Now().After(tu.ExpiresAt))
	}

	// FIXED QUERY: Added expiration check
	err := db.Where("email = ? AND verification_code = ? AND expires_at > ?",
		input.Email, input.Code, time.Now()).First(&tempUser).Error

	if err != nil {
		log.Printf("❌ VERIFICATION FAILED - DB Query Error: %v", err)
		log.Printf("   Query: email=%s, code=%s, expires_at>%v",
			input.Email, input.Code, time.Now())

		// Check if it exists but expired
		var expiredTemp models.TempUser
		if err2 := db.Where("email = ? AND verification_code = ?",
			input.Email, input.Code).First(&expiredTemp).Error; err2 == nil {
			log.Printf("⚠️ Found but expired: Expired at %v (now: %v)",
				expiredTemp.ExpiresAt, time.Now())
			return c.Status(400).JSON(fiber.Map{
				"error":        "Verification code expired. Please request a new one.",
				"expired_at":   expiredTemp.ExpiresAt,
				"current_time": time.Now(),
			})
		}

		// Check in-memory map as fallback (for debugging)
		if storedCode, ok := verificationCodes[input.Email]; ok {
			log.Printf("⚠️ Found in memory map: %s (input: %s)", storedCode, input.Code)
			if storedCode == input.Code {
				log.Printf("⚠️ CODE MATCHES IN MEMORY BUT NOT IN DATABASE")
			}
		}

		return c.Status(400).JSON(fiber.Map{"error": "Invalid verification code"})
	}

	// ✅ VERIFICATION SUCCESS! Now move to User table
	log.Println("✅ VERIFICATION SUCCESS! Temp User ID:", tempUser.ID)
	log.Printf("   Email: %s", tempUser.Email)
	log.Printf("   Code: %s", tempUser.VerificationCode)
	log.Printf("   Expires at: %v", tempUser.ExpiresAt)

	// START TRANSACTION for atomic move from temp_users to users
	tx := db.Begin()
	log.Printf("📝 Transaction started for email verification of: %s", tempUser.Email)

	// 1. Create user record from temp_user data (without StudentID - will generate below)
	user := models.User{
		Email:         tempUser.Email,
		Password:      tempUser.Password,
		Username:      tempUser.Username,
		StudentID:     tempUser.StudentID, // Will be generated after creation
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
		IsVerified:    true,
		Role:          "student",
		CreatedAt:     time.Now(),
		VerifiedAt:    time.Now(),
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Failed to create user in users table: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to activate account"})
	}
	log.Printf("✅ User created in users table with ID: %d", user.ID)

	// 2. Generate and set the final StudentID using the auto-increment ID
	customStudentID := utils.GenerateCustomStudentID(user.ID)
	log.Printf("📝 Generated StudentID: %s for user ID: %d", customStudentID, user.ID)

	// Update the user with the generated student ID
	if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("student_id", customStudentID).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Failed to update student_id: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate student ID"})
	}

	// Update the local user object with the new student ID
	user.StudentID = customStudentID
	log.Printf("✅ Student ID successfully set to: %s", customStudentID)

	// 3. Generate QR code for the user
	qrCodeData, qrErr := utils.GenerateStudentQRCode(user.StudentID, user.Email, user.FirstName, user.LastName, user.Course)
	if qrErr != nil {
		log.Printf("⚠️ Failed to generate QR code for user %s: %v (continuing anyway)", user.StudentID, qrErr)
	} else {
		if err := tx.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
			"qr_code_data": qrCodeData,
			"qr_code_type": "student_id",
		}).Error; err != nil {
			log.Printf("⚠️ Failed to update QR code data: %v (continuing anyway)", err)
		}
		log.Printf("✅ QR code generated successfully for user: %s", user.StudentID)
	}

	// 4. Delete temp user (atomic within transaction)
	if err := tx.Where("email = ?", input.Email).Delete(&models.TempUser{}).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Failed to delete temp user: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to complete verification"})
	}
	log.Printf("🗑️ Temp user deleted from database")

	// 5. Remove from in-memory verification codes map
	delete(verificationCodes, input.Email)

	// COMMIT TRANSACTION
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Failed to commit transaction: %v", err)
		return c.Status(500).JSON(fiber.Map{"error": "Failed to complete verification"})
	}
	log.Printf("✅ Transaction committed successfully")

	// Reload the user to get the updated student ID
	var finalUser models.User
	if err := db.Where("id = ?", user.ID).First(&finalUser).Error; err != nil {
		log.Printf("⚠️ Could not reload user: %v", err)
		finalUser = user
	} else {
		log.Printf("✅ User reloaded from DB with StudentID: %s", finalUser.StudentID)
	}

	// 6. Send student ID email (outside transaction)
	go func() {
		log.Printf("📧 Attempting to send Student ID email to: %s", finalUser.Email)
		if err := utils.SendStudentIDEmail(finalUser.Email, finalUser.StudentID, finalUser.FirstName); err != nil {
			log.Printf("❌ Failed to send student ID email to %s: %v (but account is verified)", finalUser.Email, err)
		} else {
			log.Printf("✅ Student ID email sent successfully to: %s", finalUser.Email)
		}
	}()

	return c.Status(200).JSON(fiber.Map{
		"message":    "Email verified successfully! Your account has been activated.",
		"student_id": finalUser.StudentID,
		"email":      finalUser.Email,
		"user_info": fiber.Map{
			"first_name": finalUser.FirstName,
			"last_name":  finalUser.LastName,
			"course":     finalUser.Course,
			"year_level": finalUser.YearLevel,
		},
		"note": "Student ID has been sent to your email. Use it to login.",
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
		log.Printf("❌ Student not found with ID: %s, error: %v", req.StudentID, err)
		return c.Status(401).JSON(fiber.Map{"error": utils.ErrInvalidStudentIDPass})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		log.Printf("❌ Invalid password for student: %s", req.StudentID)
		return c.Status(401).JSON(fiber.Map{"error": utils.ErrInvalidStudentIDPass})
	}

	if !user.IsVerified && user.Role != "superadmin" {
		log.Printf("❌ Account not verified for student: %s", req.StudentID)
		return c.Status(401).JSON(fiber.Map{"error": utils.ErrVerifyEmailFirst})
	}

	log.Printf("✅ Login successful for student: %s", req.StudentID)

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
		log.Printf("❌ User not found: %s", studentID)
		return c.Status(404).JSON(fiber.Map{"error": utils.ErrUserNotFound})
	}

	log.Printf("✅ User profile retrieved: %s", studentID)

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
