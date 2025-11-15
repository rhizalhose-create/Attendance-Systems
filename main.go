package main

import (
    "log"
    "myapp/config"
    "myapp/handlers"
    "os"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("⚠️  .env file not found, using system environment variables")
    }

    // Connect to database
    config.ConnectDB()
    
    // Display database info
    config.DisplayTableStructure()
    config.DisplayUserStats()

    // Create Fiber app
    app := fiber.New(fiber.Config{
        AppName: "Attendance System API",
    })

    // Basic middleware
    app.Use(func(c *fiber.Ctx) error {
        // Simple CORS
        c.Set("Access-Control-Allow-Origin", "*")
        c.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-Role, X-User-ID")
        
        if c.Method() == "OPTIONS" {
            return c.SendStatus(fiber.StatusOK)
        }
        
        return c.Next()
    })

    // Public routes
    app.Get("/health", healthCheck)
    app.Post("/register", handlers.Register)
    app.Post("/verify", handlers.VerifyEmail)
    app.Post("/login", handlers.Login)
    app.Get("/user/:user_id", handlers.GetUserProfile)
    app.Post("/resend-verification", handlers.ResendVerificationCode)

    // Admin management routes (Superadmin only)
    adminRoutes := app.Group("/admin")
    adminRoutes.Post("/promote", handlers.PromoteToAdmin)
    adminRoutes.Post("/demote", handlers.DemoteToStudent)
    adminRoutes.Get("/admins", handlers.GetAllAdmins)
    adminRoutes.Get("/students", handlers.GetAllStudents)
    adminRoutes.Delete("/cleanup-expired", handlers.CleanupExpiredRegistrations)

    // QR Code Management routes
    qrRoutes := app.Group("/qrcode")
    qrRoutes.Post("/types", handlers.CreateQRCodeType)          
    qrRoutes.Get("/types", handlers.GetQRCodeTypes)             
    qrRoutes.Post("/events", handlers.CreateEvent)              
    qrRoutes.Get("/events", handlers.GetEvents)                 
    qrRoutes.Put("/user", handlers.UpdateUserQRCodeType)        
    qrRoutes.Get("/user/:user_id", handlers.GetUserQRCode)      
    qrRoutes.Put("/course", handlers.UpdateCourseQRCodeType)        
    qrRoutes.Get("/course/students", handlers.GetStudentsByCourse)

    // 404 Handler
    app.Use(notFoundHandler)

    // Start server
    port := getPort()
    log.Printf(" Server starting on :%s", port)
    log.Printf(" Superadmin Login:")
    log.Printf("    Email: superadmin@system.com")
    log.Printf("    Password: superadmin123")
    log.Printf("    User ID: U2025-0000")
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