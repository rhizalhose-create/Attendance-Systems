package handlers

import (
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"

    "github.com/gofiber/fiber/v2"
)

// Middleware to check if user is superadmin
func SuperAdminOnly(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{
            "error": "Access denied. Superadmin privileges required.",
        })
    }
    return c.Next()
}

// Middleware to check if user is admin or superadmin
func AdminOrSuperAdminOnly(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{
            "error": "Access denied. Admin privileges required.",
        })
    }
    return c.Next()
}

// Promote user to admin (Superadmin only)
func PromoteToAdmin(c *fiber.Ctx) error {
    var req models.PromoteToAdminRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Check if user is superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Superadmin access required"})
    }

    // Validate required fields
    if req.TargetStudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Target student ID is required"})
    }

    // Find target user
    var targetUser models.User
    if err := config.DB.Where("student_id = ?", req.TargetStudentID).First(&targetUser).Error; err != nil {
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
            "student_id": targetUser.StudentID,
            "email":      targetUser.Email,
            "username":   targetUser.Username,
            "first_name": targetUser.FirstName,
            "last_name":  targetUser.LastName,
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
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Check if user is superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Superadmin access required"})
    }

    // Validate required fields
    if req.TargetStudentID == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Target student ID is required"})
    }

    // Find target user
    var targetUser models.User
    if err := config.DB.Where("student_id = ?", req.TargetStudentID).First(&targetUser).Error; err != nil {
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

    // Update user role to student and clear admin-specific fields
    updates := map[string]interface{}{
        "role":       "student",
        "department": "", // Clear department
        "college":    "", // Clear college
    }

    if err := config.DB.Model(&targetUser).Updates(updates).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to demote user"})
    }

    return c.JSON(fiber.Map{
        "message": "User demoted to student successfully",
        "user": fiber.Map{
            "student_id": targetUser.StudentID,
            "email":      targetUser.Email,
            "username":   targetUser.Username,
            "first_name": targetUser.FirstName,
            "last_name":  targetUser.LastName,
            "role":       "student",
        },
    })
}

// Get all admins (Superadmin only)
func GetAllAdmins(c *fiber.Ctx) error {
    // Check if user is superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Superadmin access required"})
    }

    var admins []models.User
    
    if err := config.DB.Where("role IN ?", []string{"admin", "superadmin"}).Find(&admins).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch admins"})
    }

    var adminList []fiber.Map
    for _, admin := range admins {
        adminList = append(adminList, fiber.Map{
            "student_id":   admin.StudentID,
            "email":        admin.Email,
            "username":     admin.Username,
            "first_name":   admin.FirstName,
            "last_name":    admin.LastName,
            "role":         admin.Role,
            "department":   admin.Department,
            "college":      admin.College,
            "is_verified":  admin.IsVerified,
            "created_at":   admin.CreatedAt,
            "verified_at":  admin.VerifiedAt,
        })
    }

    return c.JSON(fiber.Map{
        "admins": adminList,
        "count":  len(adminList),
    })
}

// Get all students (Admin/Superadmin only)
func GetAllStudents(c *fiber.Ctx) error {
    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    var students []models.User
    
    if err := config.DB.Where("role = ?", "student").Find(&students).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
    }

    var studentList []fiber.Map
    for _, student := range students {
        studentList = append(studentList, fiber.Map{
            "student_id":    student.StudentID,
            "email":         student.Email,
            "username":      student.Username,
            "first_name":    student.FirstName,
            "last_name":     student.LastName,
            "course":        student.Course,
            "year_level":    student.YearLevel,
            "section":       student.Section,
            "department":    student.Department,
            "college":       student.College,
            "is_verified":   student.IsVerified,
            "created_at":    student.CreatedAt,
            "verified_at":   student.VerifiedAt,
            "qr_code_type":  student.QRCodeType,
        })
    }

    return c.JSON(fiber.Map{
        "students": studentList,
        "count":    len(studentList),
    })
}

// Get user statistics (Admin/Superadmin only)
func GetUserStats(c *fiber.Ctx) error {
    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    var stats struct {
        TotalUsers    int64
        SuperAdmins   int64
        Admins        int64
        Students      int64
        VerifiedUsers int64
        UsersWithQR   int64
    }

    config.DB.Model(&models.User{}).Count(&stats.TotalUsers)
    config.DB.Model(&models.User{}).Where("role = ?", "superadmin").Count(&stats.SuperAdmins)
    config.DB.Model(&models.User{}).Where("role = ?", "admin").Count(&stats.Admins)
    config.DB.Model(&models.User{}).Where("role = ?", "student").Count(&stats.Students)
    config.DB.Model(&models.User{}).Where("is_verified = ?", true).Count(&stats.VerifiedUsers)
    config.DB.Model(&models.User{}).Where("qr_code_data IS NOT NULL AND qr_code_data != ''").Count(&stats.UsersWithQR)

    return c.JSON(fiber.Map{
        "user_statistics": fiber.Map{
            "total_users":     stats.TotalUsers,
            "super_admins":    stats.SuperAdmins,
            "admins":          stats.Admins,
            "students":        stats.Students,
            "verified_users":  stats.VerifiedUsers,
            "users_with_qr":   stats.UsersWithQR,
        },
    })
}

// Get user by student ID (Admin/Superadmin only)
func GetUserByStudentID(c *fiber.Ctx) error {
    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    studentID := c.Params("student_id")

    var user models.User
    if err := config.DB.Where("student_id = ?", studentID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    return c.JSON(fiber.Map{
        "user": fiber.Map{
            "student_id":    user.StudentID,
            "email":         user.Email,
            "username":      user.Username,
            "first_name":    user.FirstName,
            "last_name":     user.LastName,
            "middle_name":   user.MiddleName,
            "course":        user.Course,
            "year_level":    user.YearLevel,
            "section":       user.Section,
            "department":    user.Department,
            "college":       user.College,
            "contact_number": user.ContactNumber,
            "address":       user.Address,
            "role":          user.Role,
            "is_verified":   user.IsVerified,
            "created_at":    user.CreatedAt,
            "verified_at":   user.VerifiedAt,
            "qr_code_type":  user.QRCodeType,
            "qr_code_data":  user.QRCodeData,
        },
    })
}