package controllers

import (
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"
    "time" 
    "log"

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
        return c.Status(400).JSON(fiber.Map{"error": "Email, password, and username are required"})
    }

    var existingUser models.User
    if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
        return c.Status(400).JSON(fiber.Map{"error": "User already exists"})
    }

    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 14)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to hash password"})
    }

    // FIX 1: GENERATE STUDENT ID FIRST
    var lastUser models.User
    config.DB.Order("id DESC").First(&lastUser)
    
    nextID := uint(1)
    if lastUser.ID > 0 {
        nextID = lastUser.ID + 1
    }
    
    customStudentID := utils.GenerateCustomStudentID(nextID)
    log.Printf("🎯 Generating Student ID: %s for new user", customStudentID)

    // FIX 2: CREATE USER WITH STUDENT ID ALREADY SET
    user := models.User{
        StudentID:   customStudentID, // SET STUDENT ID HERE
        Email:       req.Email,
        Password:    string(hash),
        Username:    req.Username,
        IsVerified:  false,
        CreatedAt:   time.Now(), 
    }

    // FIX 3: CREATE USER WITH STUDENT ID
    result := config.DB.Create(&user)
    if result.Error != nil {
        log.Printf("❌ Failed to create user: %v", result.Error)
        return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
    }

    log.Printf("✅ User registered successfully: %s (Student ID: %s)", user.Email, user.StudentID)

    return c.JSON(fiber.Map{
        "message": "User registered successfully",
        "student_id": customStudentID,
        "db_id":   user.ID,      
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

    // ADD DEBUG LOGGING
    log.Printf("🔍 LOGIN ATTEMPT - Student ID: '%s'", req.StudentID)

    var user models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&user).Error; err != nil {
        log.Printf("❌ STUDENT NOT FOUND - ID: '%s', Error: %v", req.StudentID, err)
        
        // Debug: List all student IDs
        var allUsers []models.User
        config.DB.Select("student_id, email").Find(&allUsers)
        log.Printf("📋 ALL STUDENT IDs IN DATABASE:")
        for _, u := range allUsers {
            log.Printf("   - %s (Email: %s)", u.StudentID, u.Email)
        }
        
        return c.Status(401).JSON(fiber.Map{"error": "Invalid student ID or password"})
    }

    log.Printf("✅ STUDENT FOUND - ID: '%s', Email: %s", user.StudentID, user.Email)

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