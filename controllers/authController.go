package controllers

import (
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"
    "log"
    "time"

    "github.com/gofiber/fiber/v2"
    "golang.org/x/crypto/bcrypt"
)

func Register(c *fiber.Ctx) error {
    type RegisterRequest struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Username string `json:"username"`
    }

    var req RegisterRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    if req.Email == "" || req.Password == "" || req.Username == "" {
        return c.Status(400).JSON(fiber.Map{
            "error": "Email, password, and username are required",
        })
    }

    // Check if email already exists in USERS table
    var existingUser models.User
    if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
        return c.Status(400).JSON(fiber.Map{"error": "User already exists"})
    }

    // Check if email already exists in TEMP_USERS table
    var existingTemp models.TempUser
    if err := config.DB.Where("email = ?", req.Email).First(&existingTemp).Error; err == nil {
        return c.Status(400).JSON(fiber.Map{"error": "Email already registered but not verified"})
    }

    // Hash password
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
    }

    // Generate verification code
    verificationCode := utils.GenerateVerificationCode() // auto 6-digit or whatever your utils provides

    // Generate expiration (5 mins)
    expiresAt := time.Now().Add(5 * time.Minute)

    tempUser := models.TempUser{
        Email:            req.Email,
        Password:         string(hash),
        Username:         req.Username,
        VerificationCode: verificationCode,
        ExpiresAt:        expiresAt,
        CreatedAt:        time.Now(),
    }

    if err := config.DB.Create(&tempUser).Error; err != nil {
        log.Printf("❌ Failed to create temp user: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // Send the verification code via email or logs
    log.Printf("📧 Sending verification code '%s' to %s", verificationCode, tempUser.Email)

    return c.JSON(fiber.Map{
        "message":      "Verification code sent to your email",
        "temp_user_id": tempUser.ID,
        "expires_at":   expiresAt,
    })
}

func Login(c *fiber.Ctx) error {
    type LoginRequest struct {
        StudentID string `json:"student_id"`
        Password  string `json:"password"`
    }

    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    log.Printf("🔍 LOGIN ATTEMPT - Student ID: '%s'", req.StudentID)

    var user models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
        log.Printf("❌ STUDENT NOT FOUND - ID: '%s'", req.StudentID)
        return c.Status(401).JSON(fiber.Map{"error": "Invalid student ID or password"})
    }

    // Must be verified first
    if !user.IsVerified {
        return c.Status(403).JSON(fiber.Map{"error": "Account not verified"})
    }

    // Check password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid student ID or password"})
    }

    return c.JSON(fiber.Map{
        "message":    "Login successful",
        "student_id": user.StudentID,
        "email":      user.Email,
        "username":   user.Username,
    })
}
