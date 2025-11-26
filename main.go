package main

import (
    "log"
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/handlers"
    "os"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/joho/godotenv"
)

func main() {
    if err := godotenv.Load(); err != nil {
        log.Println("  .env file not found, using system environment variables")
    }

    config.ConnectDB()
    
    config.DisplayTableStructure()
    config.DisplayUserStats()
    
    app := fiber.New(fiber.Config{
        AppName: "Attendance System API",
    })

    // CORS middleware
    app.Use(func(c *fiber.Ctx) error {
        c.Set("Access-Control-Allow-Origin", "*")
        c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-Role, X-User-ID")
        
        if c.Method() == "OPTIONS" {
            return c.SendStatus(fiber.StatusOK)
        }
        
        return c.Next()
    })

    // Health check
    app.Get("/health", healthCheck)

    // Authentication routes
    app.Post("/register", handlers.Register)
    app.Post("/verify", handlers.VerifyEmail)
    app.Post("/login", handlers.Login)
    app.Post("/resend-verification", handlers.ResendVerificationCode)

    // Password reset routes
    app.Post("/forgot-password", handlers.RequestPasswordReset)
    app.Post("/reset-password", handlers.ResetPassword)


    // User profile
    app.Get("/user/:user_id", handlers.GetUserProfile)

    // QR Code routes
    qrRoutes := app.Group("/qrcode")
    qrRoutes.Post("/types", handlers.CreateQRCodeType)
    qrRoutes.Get("/types", handlers.GetQRCodeTypes)
    qrRoutes.Post("/events", handlers.CreateEvent)
    qrRoutes.Get("/events", handlers.GetEvents)
    qrRoutes.Put("/user", handlers.UpdateUserQRCodeType)
    qrRoutes.Get("/user/:user_id", handlers.GetUserQRCode)

    // 404 Handler
    app.Use(notFoundHandler)

    port := getPort()
    log.Printf(" Server starting on :%s", port)
    log.Printf(" Superadmin Login:")
    log.Printf("    Email: superadmin@system.com")
    log.Printf("    Password: superadmin123")
    log.Printf("    User ID: U2025-0000")
    
    log.Printf(" Essential Endpoints:")
    log.Printf("    POST /register - User registration")
    log.Printf("    POST /login - User login")
    log.Printf("    POST /forgot-password - Request password reset")
    log.Printf("    POST /reset-password - Reset password")
    log.Printf("    GET /health - Health check")
    
    log.Fatal(app.Listen(":" + port))
}

// Health check endpoint
func healthCheck(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{
        "status":    "OK",
        "timestamp": time.Now().Format(time.RFC3339),
        "service":   "Attendance System API",
        "version":   "1.0.0",
    })
}

// 404 Handler
func notFoundHandler(c *fiber.Ctx) error {
    return c.Status(404).JSON(fiber.Map{
        "error":   "Endpoint not found",
        "path":    c.Path(),
        "method":  c.Method(),
        "message": "Check the API documentation for available endpoints",
    })
}

// Get port from environment or default
func getPort() string {
    if port := os.Getenv("SERVER_PORT"); port != "" {
        return port
    }
    return "8080"
}