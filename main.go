package main

import (
	"log"
	"os"
	"time"

	"AttendanceManagementSystem/config"
	"AttendanceManagementSystem/handlers"

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
	app.Get("/debug/qr-types", handlers.DebugQRCodeTypes)
	app.Get("", handlers.ResetAllQRCodesToDefault)
	app.Get("/events/reset-qr-simple", handlers.QuickResetAllQRCodes)

	// Authentication routes
	app.Post("/register", handlers.Register)
	app.Post("/verify", handlers.VerifyEmail)
	app.Post("/login", handlers.Login)
	app.Post("/resend-verification", handlers.ResendVerificationCode)
	

	// Password reset routes
	app.Post("/forgot-password", handlers.RequestPasswordReset)
	app.Post("/verify-reset-code", handlers.VerifyResetCode)
	app.Post("/reset-password", handlers.ResetPassword)

	// User profile
	app.Get("/user/:user_id", handlers.GetUserProfile)

	// QR Code routes
	qrRoutes := app.Group("/qrcode")
	qrRoutes.Post("/types", handlers.CreateQRCodeType)
	qrRoutes.Get("/types", handlers.GetQRCodeTypes)
	// Remove duplicate event routes from here
	qrRoutes.Put("/user", handlers.UpdateUserQRCodeType)
	qrRoutes.Get("/user/:user_id", handlers.GetUserQRCode)
	

	// Event Management Routes - Use the new handler names
	eventRoutes := app.Group("/events")
	eventRoutes.Post("/", handlers.CreateEventHandler)
	eventRoutes.Get("/", handlers.GetEventsHandler)
	eventRoutes.Get("/:id", handlers.GetEventByIDHandler)   // Changed
	eventRoutes.Put("/:id", handlers.UpdateEventHandler)    // Changed
	eventRoutes.Delete("/:id", handlers.DeleteEventHandler) // Changed
	eventRoutes.Get("/options/available", handlers.GetAvailableOptions) // Get dropdown options
eventRoutes.Get("/:id/students", handlers.GetEventStudents)        // Get affected students
eventRoutes.Post("/:id/refresh-qr", handlers.RefreshEventQRCodes)  // Refresh QR codes



	app.Get("/my-events", handlers.GetMyEventsHandler) // Changed

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
	
	log.Printf(" Event Management Endpoints:")
	log.Printf("    POST /events - Create new event")
	log.Printf("    GET /events - Get all events")
	log.Printf("    GET /events/:id - Get event by ID")
	log.Printf("    PUT /events/:id - Update event")
	log.Printf("    DELETE /events/:id - Delete event")
	log.Printf("    GET /my-events - Get user-specific events")
	
	
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

func getPort() string {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		return port
	}
	return "9090"
}
