package controllers

import (
    "myapp/config"
    "myapp/models"
    "myapp/utils"
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

    user := models.User{
        Email:      req.Email,
        Password:   string(hash),
        Username:   req.Username,
        IsVerified: false,
        CreatedAt:  time.Now(), 
    }

 
    result := config.DB.Create(&user)
    if result.Error != nil {
        return c.Status(500).JSON(fiber.Map{"error": result.Error.Error()})
    }


    customUserID := utils.GenerateCustomUserID(user.ID)
    
 
    config.DB.Model(&user).Update("user_id", customUserID)

    return c.JSON(fiber.Map{
        "message": "User registered successfully",
        "user_id": customUserID, 
        "db_id":   user.ID,      
    })
}

func Login(c *fiber.Ctx) error {
    type LoginRequest struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }


    var user models.User
    if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
    }

    // Check password
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
        return c.Status(401).JSON(fiber.Map{"error": "Invalid email or password"})
    }

    return c.JSON(fiber.Map{
        "message":  "Login successful",
        "user_id":  user.UserID, 
        "email":    user.Email,
        "username": user.Username,
    })
}