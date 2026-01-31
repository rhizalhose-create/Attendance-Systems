// handlers/attendance.go
package handlers

import (
    "log"
    "time"
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"

    "github.com/gofiber/fiber/v2"
)

// RecordAttendance records when a student is scanned
func RecordAttendance(c *fiber.Ctx) error {
    var req models.AttendanceRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    scannerID := c.Get(utils.HeaderStudentID)
    
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Only admins can record attendance"})
    }

    // Get event details
    var event models.Event
    if err := config.DB.Where("id = ?", req.EventID).First(&event).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
    }

    // Get student details
    var student models.User
    if err := config.DB.Where("student_id = ?", req.StudentID).First(&student).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Student not found"})
    }

    // Check if attendance already recorded for this event today
    var existingAttendance models.Attendance
    today := time.Now().Format("2006-01-02")
    if err := config.DB.Where("student_id = ? AND event_id = ? AND DATE(scan_time) = ?", 
        req.StudentID, req.EventID, today).First(&existingAttendance).Error; err == nil {
        return c.Status(400).JSON(fiber.Map{"error": "Attendance already recorded for today"})
    }

    // Record attendance
    attendance := models.Attendance{
        StudentID:   req.StudentID,
        EventID:     req.EventID,
        EventName:   event.EventName,
        QRCodeType:  req.QRCodeType,
        ScannedBy:   scannerID,
        ScannerRole: userRole,
        ScanTime:    time.Now(),
        Status:      "present",
    }

    if err := config.DB.Create(&attendance).Error; err != nil {
        log.Printf("Failed to record attendance: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to record attendance"})
    }

    log.Printf("✅ Attendance recorded: %s for event %s by %s", 
        req.StudentID, event.EventName, scannerID)

    return c.JSON(fiber.Map{
        "success": true,
        "message": "Attendance recorded successfully",
        "attendance": fiber.Map{
            "id":          attendance.ID,
            "student_id":  attendance.StudentID,
            "student_name": student.FirstName + " " + student.LastName,
            "event_name":  attendance.EventName,
            "scan_time":   attendance.ScanTime,
            "scanned_by":  attendance.ScannedBy,
        },
    })
}

// GetEventAttendance gets all attendance for a specific event
func GetEventAttendance(c *fiber.Ctx) error {
    eventID, err := c.ParamsInt("event_id")
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid event ID"})
    }

    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Only admins can view attendance"})
    }

    var attendances []models.Attendance
    if err := config.DB.Where("event_id = ?", eventID).
        Order("scan_time DESC").
        Find(&attendances).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch attendance"})
    }

    // Enrich with student details
    var response []models.AttendanceResponse
    for _, att := range attendances {
        var student models.User
        config.DB.Where("student_id = ?", att.StudentID).First(&student)
        
        response = append(response, models.AttendanceResponse{
            ID:          att.ID,
            StudentID:   att.StudentID,
            EventName:   att.EventName,
            QRCodeType:  att.QRCodeType,
            ScannedBy:   att.ScannedBy,
            ScanTime:    att.ScanTime,
            Status:      att.Status,
            StudentName: student.FirstName + " " + student.LastName,
            Course:      student.Course,
            YearLevel:   student.YearLevel,
        })
    }

    return c.JSON(fiber.Map{
        "event_id":   eventID,
        "attendances": response,
        "count":      len(response),
    })
}

// GetStudentAttendance gets attendance history for a student
func GetStudentAttendance(c *fiber.Ctx) error {
    studentID := c.Params("student_id")

    // Check if user is admin or superadmin OR the student themselves
    userRole := c.Get(utils.HeaderUserRole)
    requestingStudentID := c.Get(utils.HeaderStudentID)
    
    if userRole == "student" && requestingStudentID != studentID {
        return c.Status(403).JSON(fiber.Map{"error": "Cannot view other students' attendance"})
    }

    var attendances []models.Attendance
    if err := config.DB.Where("student_id = ?", studentID).
        Order("scan_time DESC").
        Find(&attendances).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch attendance"})
    }

    return c.JSON(fiber.Map{
        "student_id": studentID,
        "attendances": attendances,
        "count":      len(attendances),
    })
}

// GetAttendanceStats gets attendance statistics
func GetAttendanceStats(c *fiber.Ctx) error {
    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Only admins can view statistics"})
    }

    var stats struct {
        TotalScans      int64
        TodayScans      int64
        UniqueStudents  int64
        EventsCount     int64
    }

    today := time.Now().Format("2006-01-02")
    
    config.DB.Model(&models.Attendance{}).Count(&stats.TotalScans)
    config.DB.Model(&models.Attendance{}).Where("DATE(scan_time) = ?", today).Count(&stats.TodayScans)
    config.DB.Model(&models.Attendance{}).Select("COUNT(DISTINCT student_id)").Scan(&stats.UniqueStudents)
    config.DB.Model(&models.Event{}).Where("is_active = ?", true).Count(&stats.EventsCount)

    return c.JSON(fiber.Map{
        "statistics": fiber.Map{
            "total_scans":      stats.TotalScans,
            "today_scans":      stats.TodayScans,
            "unique_students":  stats.UniqueStudents,
            "active_events":    stats.EventsCount,
        },
    })
}