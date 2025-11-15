package handlers

import (
    "fmt"
    "log"
    "myapp/config"
    "myapp/models"
    "myapp/utils"
    "time"

    "github.com/gofiber/fiber/v2"
)

// GetQRCodeTypes - Get all available QR code types
func GetQRCodeTypes(c *fiber.Ctx) error {
    var qrTypes []models.QRCodeType
    
    if err := config.DB.Where(utils.QueryIsActive, true).Find(&qrTypes).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedToFetchQRTypes})
    }

    return c.JSON(fiber.Map{
        "qr_code_types": qrTypes,
        "count":         len(qrTypes),
    })
}

// GetEvents - Get all active events
func GetEvents(c *fiber.Ctx) error {
    var events []models.QRCodeEvent
    
    if err := config.DB.Where(utils.QueryActiveAndEndTime, true, time.Now()).Find(&events).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedToFetchEvents})
    }

    return c.JSON(fiber.Map{
        "events": events,
        "count":  len(events),
    })
}

// GetUserQRCode - Get user's current QR code
func GetUserQRCode(c *fiber.Ctx) error {
    userID := c.Params("user_id")

    var user models.User
    if err := config.DB.Where(utils.QueryUserIDWhere, userID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": utils.ErrUserNotFound})
    }

    return c.JSON(fiber.Map{
        "user_id":      user.UserID,
        "qr_code_type": user.QRCodeType,
        "qr_code_data": user.QRCodeData,
        "email":        user.Email,
        "name":         user.FirstName + " " + user.LastName,
    })
}

// UpdateCourseQRCodeType - Update QR code type for ALL students in specific course and year level
func UpdateCourseQRCodeType(c *fiber.Ctx) error {
    var req models.UpdateCourseQRCodeRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
    }

    // Verify QR code type exists
    var qrType models.QRCodeType
    if err := config.DB.Where(utils.QueryTypeNameAndActive, req.QRCodeType, true).First(&qrType).Error; err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrInvalidQRCodeType})
    }

    // Find all students in the specified course and year level
    var students []models.User
    if err := config.DB.Where(utils.QueryCourseYearRole, 
        req.Course, req.YearLevel, "student").Find(&students).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedToFetchStudents})
    }

    if len(students) == 0 {
        return c.Status(404).JSON(fiber.Map{
            "error":      "No students found for the specified course and year level",
            "course":     req.Course,
            "year_level": req.YearLevel,
        })
    }

    updatedCount := 0
    var updateErrors []string

    // Update QR code for each student
    for _, student := range students {
        // Generate new QR code based on type
        var qrCodeData string
        var err error

        switch req.QRCodeType {
        case "student_id":
            qrCodeData, err = utils.GenerateStudentQRCode(student.UserID, student.Email, student.FirstName, student.LastName, student.Course)
        case "attendance":
            customData := fmt.Sprintf("attendance|course:%s|year:%s", req.Course, req.YearLevel)
            qrCodeData, err = utils.GenerateCustomQRCode(student.UserID, "attendance", customData)
        case "event":
            customData := fmt.Sprintf("event|course:%s|year:%s", req.Course, req.YearLevel)
            qrCodeData, err = utils.GenerateCustomQRCode(student.UserID, "event", customData)
        case "business":
            customData := fmt.Sprintf("business|course:%s|year:%s", req.Course, req.YearLevel)
            qrCodeData, err = utils.GenerateCustomQRCode(student.UserID, "business", customData)
        case "activity":
            customData := fmt.Sprintf("activity|course:%s|year:%s", req.Course, req.YearLevel)
            qrCodeData, err = utils.GenerateCustomQRCode(student.UserID, "activity", customData)
        case "library":
            customData := fmt.Sprintf("library|course:%s|year:%s", req.Course, req.YearLevel)
            qrCodeData, err = utils.GenerateCustomQRCode(student.UserID, "library", customData)
        default:
            customData := fmt.Sprintf("%s|course:%s|year:%s", req.QRCodeType, req.Course, req.YearLevel)
            qrCodeData, err = utils.GenerateCustomQRCode(student.UserID, req.QRCodeType, customData)
        }

        if err != nil {
            updateErrors = append(updateErrors, fmt.Sprintf("Failed to generate QR for %s: %v", student.Email, err))
            continue
        }

        // Update student's QR code
        updates := map[string]interface{}{
            "qr_code_data": qrCodeData,
            "qr_code_type": req.QRCodeType,
        }

        if err := config.DB.Model(&student).Updates(updates).Error; err != nil {
            updateErrors = append(updateErrors, fmt.Sprintf("Failed to update %s: %v", student.Email, err))
        } else {
            updatedCount++
            log.Printf("Updated QR code for student: %s (%s) to type: %s", 
                student.Email, student.UserID, req.QRCodeType)
        }
    }

    response := models.BulkQRCodeUpdateResponse{
        Message:      "QR code update completed",
        UpdatedCount: updatedCount,
        Course:       req.Course,
        YearLevel:    req.YearLevel,
        QRCodeType:   req.QRCodeType,
    }

    if len(updateErrors) > 0 {
        response.Message = fmt.Sprintf("QR code update completed with %d errors", len(updateErrors))
        return c.Status(207).JSON(fiber.Map{
            "bulk_update": response,
            "errors":      updateErrors,
        })
    }

    return c.JSON(fiber.Map{
        "bulk_update": response,
    })
}

// GetStudentsByCourse - Get all students by course and year level
func GetStudentsByCourse(c *fiber.Ctx) error {
    course := c.Query("course")
    yearLevel := c.Query("year_level")

    if course == "" || yearLevel == "" {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCourseYearLevelRequired})
    }

    var students []models.User
    if err := config.DB.Where(utils.QueryCourseYearRole, 
        course, yearLevel, "student").Find(&students).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": utils.ErrFailedToFetchStudents})
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
            "section":       student.Section,
            "qr_code_type":  student.QRCodeType,
            "qr_code_data":  student.QRCodeData,
            "is_verified":   student.IsVerified,
        })
    }

    return c.JSON(fiber.Map{
        "students":   studentList,
        "count":      len(studentList),
        "course":     course,
        "year_level": yearLevel,
    })
}