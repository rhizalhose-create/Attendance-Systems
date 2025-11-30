package handlers

import (
    "log"
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"

    "github.com/gofiber/fiber/v2"
)

// QuickResetAllQRCodes resets all student QR codes to default
func QuickResetAllQRCodes(c *fiber.Ctx) error {
    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
    }

    // Simple bulk update - reset ALL students to default
    result := config.DB.Model(&models.User{}).
        Where("role = ?", "student").
        Updates(map[string]interface{}{
            "qr_code_type": "student_id",
        })

    if result.Error != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to reset QR codes"})
    }

    log.Printf("✅ Quick reset: %d students updated to default QR", result.RowsAffected)

    return c.JSON(fiber.Map{
        "message": "All student QR codes reset to default",
        "updated_count": result.RowsAffected,
    })
}

// ResetAllQRCodesToDefault resets all student QR codes with proper QR generation
func ResetAllQRCodesToDefault(c *fiber.Ctx) error {
    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
    }

    log.Printf("🔄 Resetting ALL student QR codes to default...")

    // Get all students
    var students []models.User
    if err := config.DB.Where("role = ?", "student").Find(&students).Error; err != nil {
        log.Printf("❌ Failed to fetch students: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
    }

    updatedCount := 0
    var updateErrors []string

    // Reset each student's QR code to default
    for _, student := range students {
        // Generate default student ID QR code
        qrCodeData, err := utils.GenerateStudentQRCode(
            student.StudentID, 
            student.Email, 
            student.FirstName, 
            student.LastName, 
            student.Course,
        )
        
        if err != nil {
            updateErrors = append(updateErrors, "Failed to generate QR for "+student.StudentID)
            continue
        }

        // Update to default student_id QR code
        updates := map[string]interface{}{
            "qr_code_data": qrCodeData,
            "qr_code_type": "student_id",
        }

        if err := config.DB.Model(&models.User{}).
            Where("student_id = ?", student.StudentID).
            Updates(updates).Error; err != nil {
            updateErrors = append(updateErrors, "Failed to update "+student.StudentID)
        } else {
            updatedCount++
        }
    }

    log.Printf("✅ Reset complete: %d/%d students updated", updatedCount, len(students))

    response := fiber.Map{
        "message": "All student QR codes reset to default",
        "updated_count": updatedCount,
        "total_students": len(students),
    }

    if len(updateErrors) > 0 {
        response["errors"] = updateErrors
        response["error_count"] = len(updateErrors)
    }

    return c.JSON(response)
}