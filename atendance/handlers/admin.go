package handlers

import (
    "myapp/config"
    "myapp/models"
    "github.com/gofiber/fiber/v2"
)

// Middleware to check if user is superadmin
func SuperAdminOnly(c *fiber.Ctx) error {
    userRole := c.Get("X-User-Role")
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{
            "error": "Access denied. Superadmin privileges required.",
        })
    }
    return c.Next()
}

// Promote user to admin (Superadmin only)
func PromoteToAdmin(c *fiber.Ctx) error {
    var req models.PromoteToAdminRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    // Find target user
    var targetUser models.User
    if err := config.DB.Where("user_id = ?", req.TargetUserID).First(&targetUser).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    // Check if user is already admin
    if targetUser.Role == "admin" || targetUser.Role == "superadmin" {
        return c.Status(400).JSON(fiber.Map{"error": "User is already an admin"})
    }

    // Update user role to admin
    updates := map[string]interface{}{
        "role": "admin",
    }

    // Update department and college if provided
    if req.Department != "" {
        updates["department"] = req.Department
    }
    if req.College != "" {
        updates["college"] = req.College
    }

    if err := config.DB.Model(&targetUser).Updates(updates).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to promote user"})
    }

    return c.JSON(fiber.Map{
        "message": "User promoted to admin successfully",
        "user": fiber.Map{
            "user_id":    targetUser.UserID,
            "email":      targetUser.Email,
            "username":   targetUser.Username,
            "role":       "admin",
            "department": req.Department,
            "college":    req.College,
        },
    })
}

// Demote admin to student (Superadmin only)
func DemoteToStudent(c *fiber.Ctx) error {
    var req models.DemoteToStudentRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot parse JSON"})
    }

    // Find target user
    var targetUser models.User
    if err := config.DB.Where("user_id = ?", req.TargetUserID).First(&targetUser).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    // Cannot demote superadmin
    if targetUser.Role == "superadmin" {
        return c.Status(400).JSON(fiber.Map{"error": "Cannot demote superadmin"})
    }

    // Check if user is already student
    if targetUser.Role == "student" {
        return c.Status(400).JSON(fiber.Map{"error": "User is already a student"})
    }

    // Update user role to student
    if err := config.DB.Model(&targetUser).Update("role", "student").Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to demote user"})
    }

    return c.JSON(fiber.Map{
        "message": "User demoted to student successfully",
        "user": fiber.Map{
            "user_id":  targetUser.UserID,
            "email":    targetUser.Email,
            "username": targetUser.Username,
            "role":     "student",
        },
    })
}

// Get all admins (Superadmin only)
func GetAllAdmins(c *fiber.Ctx) error {
    var admins []models.User
    
    if err := config.DB.Where("role IN ?", []string{"admin", "superadmin"}).Find(&admins).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch admins"})
    }

    var adminList []models.AdminListResponse
    for _, admin := range admins {
        adminList = append(adminList, models.AdminListResponse{
            UserID:     admin.UserID,
            Email:      admin.Email,
            Username:   admin.Username,
            FirstName:  admin.FirstName,
            LastName:   admin.LastName,
            Role:       admin.Role,
            Department: admin.Department,
            College:    admin.College,
            IsVerified: admin.IsVerified,
        })
    }

    return c.JSON(fiber.Map{
        "admins": adminList,
        "count":  len(adminList),
    })
}

// Get all students (for admin management)
func GetAllStudents(c *fiber.Ctx) error {
    var students []models.User
    
    if err := config.DB.Where("role = ?", "student").Find(&students).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
    }

    var studentList []fiber.Map
    for _, student := range students {
        studentList = append(studentList, fiber.Map{
            "user_id":       student.UserID,
            "email":         student.Email,
            "username":      student.Username,
            "first_name":    student.FirstName,
            "last_name":     student.LastName,
            "course":        student.Course,
            "year_level":    student.YearLevel,
            "department":    student.Department,
            "college":       student.College,
            "is_verified":   student.IsVerified,
            "created_at":    student.CreatedAt,
        })
    }

    return c.JSON(fiber.Map{
        "students": studentList,
        "count":    len(studentList),
    })
}