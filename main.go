package main

import (
	"log"
	"os"
	"time"

	"AttendanceManagementSystem/config"
	"AttendanceManagementSystem/handlers"
	"AttendanceManagementSystem/models"
	"AttendanceManagementSystem/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using system environment variables")
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

	// REMOVED: Setup web dashboard routes (if routes package doesn't exist)
	// routes.SetupWebDashboard(app)

	// Add new endpoints
	app.Post("/attendance/record", handlers.RecordAttendance)
	app.Get("/attendance/event/:event_id", handlers.GetEventAttendance)
	app.Get("/attendance/student/:student_id", handlers.GetStudentAttendance)
	app.Get("/attendance/stats", handlers.GetAttendanceStats)
	
	// Student-specific events - REMOVED or replaced
	// app.Get("/events/for-student", handlers.GetEventsForStudent)

	// ADDED: Debug and test endpoints
	app.Get("/test-email", testEmailHandler)
	app.Get("/debug/temp-users", debugTempUsersHandler)
	app.Get("/debug/users", debugUsersHandler)
	app.Get("/debug/email-config", debugEmailConfigHandler)

	// Health check
	app.Get("/health", healthCheck)
	app.Get("/debug/qr-types", handlers.DebugQRCodeTypes)
	app.Post("/events/reset-qr-simple", handlers.QuickResetAllQRCodes)
	app.Post("/events/reset-all-qr", handlers.ResetAllQRCodesToDefault)

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
	app.Get("/user/:student_id", handlers.GetUserProfile)

	// Event Management Routes
	eventRoutes := app.Group("/events")
	eventRoutes.Post("/", handlers.CreateEventHandler)
	eventRoutes.Get("/", handlers.GetEventsHandler)
	eventRoutes.Get("/:id", handlers.GetEventByIDHandler)
	eventRoutes.Put("/:id", handlers.UpdateEventHandler)
	eventRoutes.Delete("/:id", handlers.DeleteEventHandler)
	eventRoutes.Get("/options/available", handlers.GetAvailableOptions)
	eventRoutes.Get("/:id/students", handlers.GetEventStudents)
	eventRoutes.Post("/:id/refresh-qr", handlers.RefreshEventQRCodes)

	app.Get("/my-events", handlers.GetMyEventsHandler)

	// 404 Handler
	app.Use(notFoundHandler)

	port := getPort()
	log.Printf("🚀 Server starting on :%s", port)
	log.Printf("")
	log.Printf("🔑 Superadmin Login:")
	log.Printf("    Email: superadmin@system.com")
	log.Printf("    Password: superadmin123")
	log.Printf("    Student ID: U2025-0000")
	log.Printf("")
	log.Printf("🔧 Debug Endpoints:")
	log.Printf("    GET  /test-email?email=test@example.com")
	log.Printf("    GET  /debug/temp-users")
	log.Printf("    GET  /debug/users")
	log.Printf("    GET  /debug/email-config")
	log.Printf("")
	log.Printf("📋 Essential Endpoints:")
	log.Printf("    POST /register - User registration")
	log.Printf("    POST /verify - Email verification")
	log.Printf("    POST /login - User login")
	log.Printf("    GET  /health - Health check")
	log.Printf("")
	log.Printf("🔍 Note: Email sending is disabled for debugging.")
	log.Printf("    Verification codes are returned in the response.")

	log.Fatal(app.Listen(":" + port))
}

// ADDED: Test email endpoint
func testEmailHandler(c *fiber.Ctx) error {
	email := c.Query("email")
	if email == "" {
		email = "test@example.com"
	}

	log.Printf("🧪 TEST EMAIL REQUESTED for: %s", email)

	code := utils.GenerateVerificationCode()

	// First test basic email
	err := utils.SendEmail(email, "Test Email - Attendance System",
		"<h1>Test Email</h1><p>If you received this, email is working!</p>")

	if err != nil {
		log.Printf("❌ TEST BASIC EMAIL FAILED: %v", err)
		return c.JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Basic email test failed",
		})
	}

	log.Printf("✅ BASIC TEST EMAIL SENT to: %s", email)

	// Then test verification email
	err = utils.SendVerificationEmail(email, code)
	if err != nil {
		log.Printf("❌ TEST VERIFICATION EMAIL FAILED: %v", err)
		return c.JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
			"message": "Verification email test failed",
			"code":    code,
		})
	}

	log.Printf("✅ TEST VERIFICATION EMAIL SENT: %s -> %s", email, code)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Both test emails sent successfully",
		"code":    code,
		"email":   email,
	})
}

// ADDED: Debug temp users endpoint
func debugTempUsersHandler(c *fiber.Ctx) error {
	email := c.Query("email")

	var tempUsers []models.TempUser
	query := config.DB
	if email != "" {
		query = query.Where("email = ?", email)
	}
	query.Find(&tempUsers)

	var result []map[string]interface{}
	for _, tu := range tempUsers {
		result = append(result, map[string]interface{}{
			"id":         tu.ID,
			"email":      tu.Email,
			"student_id": tu.StudentID,
			"code":       tu.VerificationCode,
			"expires_at": tu.ExpiresAt,
			"created_at": tu.CreatedAt,
			"is_expired": time.Now().After(tu.ExpiresAt),
			"expires_in": time.Until(tu.ExpiresAt).Round(time.Second).String(),
			"first_name": tu.FirstName,
			"last_name":  tu.LastName,
		})
	}

	log.Printf("📊 DEBUG: Found %d temp users", len(result))

	return c.JSON(fiber.Map{
		"count":        len(result),
		"temp_users":   result,
		"current_time": time.Now(),
		"query_email":  email,
	})
}

// ADDED: Debug users endpoint
func debugUsersHandler(c *fiber.Ctx) error {
	email := c.Query("email")
	studentID := c.Query("student_id")

	var users []models.User
	query := config.DB
	if email != "" {
		query = query.Where("email = ?", email)
	}
	if studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}
	query.Limit(50).Find(&users)

	var result []map[string]interface{}
	for _, u := range users {
		result = append(result, map[string]interface{}{
			"id":           u.ID,
			"student_id":   u.StudentID,
			"email":        u.Email,
			"username":     u.Username,
			"role":         u.Role,
			"is_verified":  u.IsVerified,
			"qr_code_type": u.QRCodeType,
			"created_at":   u.CreatedAt,
			"verified_at":  u.VerifiedAt,
			"first_name":   u.FirstName,
			"last_name":    u.LastName,
			"course":       u.Course,
			"year_level":   u.YearLevel,
		})
	}

	log.Printf("📊 DEBUG: Found %d users", len(result))

	return c.JSON(fiber.Map{
		"count":            len(result),
		"users":            result,
		"current_time":     time.Now(),
		"query_email":      email,
		"query_student_id": studentID,
	})
}

// ADDED: Debug email config endpoint
func debugEmailConfigHandler(c *fiber.Ctx) error {
	config := utils.GetEmailConfig()

	return c.JSON(fiber.Map{
		"smtp_email":               config.SMTPEmail,
		"smtp_password":            "***" + config.SMTPPassword[len(config.SMTPPassword)-4:], // Hide most of password
		"smtp_host":                config.SMTPHost,
		"smtp_port":                config.SMTPPort,
		"env_smtp_email":           os.Getenv("SMTP_EMAIL"),
		"env_smtp_password_length": len(os.Getenv("SMTP_PASSWORD")),
		"env_smtp_host":            os.Getenv("SMTP_HOST"),
		"env_smtp_port":            os.Getenv("SMTP_PORT"),
	})
}

// Health check endpoint
func healthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "OK",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "Attendance System API",
		"version":   "1.0.0",
		"database":  "Connected",
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
	return "9090"
}