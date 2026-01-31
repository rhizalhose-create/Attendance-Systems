// handlers/web_dashboard.go
package handlers

import (
    "encoding/json"
    "log"
    "time"
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"

    "github.com/gofiber/fiber/v2"
)

// GetDashboardStats returns comprehensive dashboard statistics
func GetDashboardStats(c *fiber.Ctx) error {
    // Check if user is superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Superadmin access required"})
    }

    var stats struct {
        TotalUsers     int64
        ActiveUsers    int64
        TotalAdmins    int64
        TotalStudents  int64
        TotalEvents    int64
        ActiveEvents   int64
        TotalScans     int64
        TodayScans     int64
        PendingVerifications int64
    }

    today := time.Now().Format("2006-01-02")

    config.DB.Model(&models.User{}).Count(&stats.TotalUsers)
    config.DB.Model(&models.User{}).Where("is_verified = ?", true).Count(&stats.ActiveUsers)
    config.DB.Model(&models.User{}).Where("role IN ?", []string{"admin", "superadmin"}).Count(&stats.TotalAdmins)
    config.DB.Model(&models.User{}).Where("role = ?", "student").Count(&stats.TotalStudents)
    config.DB.Model(&models.Event{}).Count(&stats.TotalEvents)
    config.DB.Model(&models.Event{}).Where("is_active = ? AND end_time > ?", true, time.Now()).Count(&stats.ActiveEvents)
    config.DB.Model(&models.Attendance{}).Count(&stats.TotalScans)
    config.DB.Model(&models.Attendance{}).Where("DATE(scan_time) = ?", today).Count(&stats.TodayScans)
    config.DB.Model(&models.TempUser{}).Count(&stats.PendingVerifications)

    // Recent activity
    var recentScans []models.Attendance
    config.DB.Where("scan_time > ?", time.Now().Add(-24*time.Hour)).
        Order("scan_time DESC").
        Limit(10).
        Find(&recentScans)

    var recentUsers []models.User
    config.DB.Where("created_at > ?", time.Now().Add(-7*24*time.Hour)).
        Order("created_at DESC").
        Limit(10).
        Find(&recentUsers)

    return c.JSON(fiber.Map{
        "stats": stats,
        "recent_activity": fiber.Map{
            "scans": recentScans,
            "users": recentUsers,
        },
        "timestamp": time.Now(),
    })
}

// GetAllUsersForDashboard returns all users with detailed info for dashboard
func GetAllUsersForDashboard(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Superadmin access required"})
    }

    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 50)
    offset := (page - 1) * limit

    var users []models.User
    var total int64

    query := config.DB.Model(&models.User{})
    
    // Apply filters
    if role := c.Query("role"); role != "" {
        query = query.Where("role = ?", role)
    }
    if verified := c.Query("verified"); verified != "" {
        query = query.Where("is_verified = ?", verified == "true")
    }
    if search := c.Query("search"); search != "" {
        query = query.Where("student_id LIKE ? OR email LIKE ? OR first_name LIKE ? OR last_name LIKE ?",
            "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
    }

    query.Count(&total)
    query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&users)

    var userList []fiber.Map
    for _, user := range users {
        // Get attendance count for this user
        var attendanceCount int64
        config.DB.Model(&models.Attendance{}).Where("student_id = ?", user.StudentID).Count(&attendanceCount)

        // Get last attendance time (as proxy for last activity)
        var lastAttendance models.Attendance
        config.DB.Where("student_id = ?", user.StudentID).
            Order("scan_time DESC").
            First(&lastAttendance)

        userList = append(userList, fiber.Map{
            "id":            user.ID,
            "student_id":    user.StudentID,
            "email":         user.Email,
            "username":      user.Username,
            "first_name":    user.FirstName,
            "last_name":     user.LastName,
            "role":          user.Role,
            "is_verified":   user.IsVerified,
            "course":        user.Course,
            "year_level":    user.YearLevel,
            "qr_code_type":  user.QRCodeType,
            "attendance_count": attendanceCount,
            "created_at":    user.CreatedAt,
            "verified_at":   user.VerifiedAt,
            "last_activity": lastAttendance.ScanTime, // Use last attendance as activity indicator
        })
    }

    return c.JSON(fiber.Map{
        "users":      userList,
        "total":      total,
        "page":       page,
        "limit":      limit,
        "total_pages": (int(total) + limit - 1) / limit,
    })
}

// GetAllEventsForDashboard returns all events for dashboard
func GetAllEventsForDashboard(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" && userRole != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    var events []models.Event
    if err := config.DB.Order("created_at DESC").Find(&events).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch events"})
    }

    var eventList []fiber.Map
    for _, event := range events {
        // Get attendance count for this event
        var attendanceCount int64
        config.DB.Model(&models.Attendance{}).Where("event_id = ?", event.ID).Count(&attendanceCount)

        // Parse target arrays (handle empty strings)
        var targetCourses, targetYearLevels, targetSections []string
        
        if event.TargetCourses != "" {
            json.Unmarshal([]byte(event.TargetCourses), &targetCourses)
        }
        if event.TargetYearLevels != "" {
            json.Unmarshal([]byte(event.TargetYearLevels), &targetYearLevels)
        }
        if event.TargetSections != "" {
            json.Unmarshal([]byte(event.TargetSections), &targetSections)
        }

        eventList = append(eventList, fiber.Map{
            "id":                event.ID,
            "event_name":        event.EventName,
            "event_type":        event.EventType,
            "description":       event.Description,
            "location":          event.Location,
            "target_courses":    targetCourses,
            "target_year_levels": targetYearLevels,
            "target_sections":   targetSections,
            "qr_code_type":      event.QRCodeType,
            "created_by":        event.CreatedBy,
            "is_active":         event.IsActive,
            "start_time":        event.StartTime,
            "end_time":          event.EndTime,
            "attendance_count":  attendanceCount,
            "created_at":        event.CreatedAt,
            "updated_at":        event.UpdatedAt,
        })
    }

    return c.JSON(fiber.Map{
        "events": eventList,
        "count":  len(eventList),
    })
}

// GetUserDetails returns detailed user information
func GetUserDetails(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" && userRole != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    studentID := c.Params("student_id")

    var user models.User
    if err := config.DB.Where("student_id = ?", studentID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    // Get user's attendance history
    var attendance []models.Attendance
    config.DB.Where("student_id = ?", studentID).
        Order("scan_time DESC").
        Limit(20).
        Find(&attendance)

    // Get events user is eligible for (based on course/year)
    var eligibleEvents []models.Event
    config.DB.Where("is_active = ? AND end_time > ?", true, time.Now()).
        Find(&eligibleEvents)

    var eventsForUser []fiber.Map
    for _, event := range eligibleEvents {
        var targetCourses []string
        if event.TargetCourses != "" {
            json.Unmarshal([]byte(event.TargetCourses), &targetCourses)
        }
        
        // Check if user's course is in target courses
        for _, course := range targetCourses {
            if course == user.Course {
                eventsForUser = append(eventsForUser, fiber.Map{
                    "id":          event.ID,
                    "event_name":  event.EventName,
                    "event_type":  event.EventType,
                    "start_time":  event.StartTime,
                    "end_time":    event.EndTime,
                    "qr_code_type": event.QRCodeType,
                })
                break
            }
        }
    }

    return c.JSON(fiber.Map{
        "user": fiber.Map{
            "student_id":    user.StudentID,
            "email":         user.Email,
            "username":      user.Username,
            "first_name":    user.FirstName,
            "last_name":     user.LastName,
            "middle_name":   user.MiddleName,
            "role":          user.Role,
            "is_verified":   user.IsVerified,
            "course":        user.Course,
            "year_level":    user.YearLevel,
            "section":       user.Section,
            "department":    user.Department,
            "college":       user.College,
            "contact_number": user.ContactNumber,
            "address":       user.Address,
            "qr_code_type":  user.QRCodeType,
            "created_at":    user.CreatedAt,
            "verified_at":   user.VerifiedAt,
        },
        "attendance_history": attendance,
        "eligible_events":    eventsForUser,
        "attendance_count":   len(attendance),
    })
}

// UpdateUserDetails allows superadmin to update user information
func UpdateUserDetails(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Superadmin access required"})
    }

    studentID := c.Params("student_id")
    
    type UpdateRequest struct {
        Role        string `json:"role,omitempty"`
        Course      string `json:"course,omitempty"`
        YearLevel   string `json:"year_level,omitempty"`
        Department  string `json:"department,omitempty"`
        College     string `json:"college,omitempty"`
        IsVerified  *bool  `json:"is_verified,omitempty"`
    }

    var req UpdateRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    var user models.User
    if err := config.DB.Where("student_id = ?", studentID).First(&user).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "User not found"})
    }

    updates := make(map[string]interface{})
    
    if req.Role != "" {
        if req.Role == "superadmin" {
            return c.Status(400).JSON(fiber.Map{"error": "Cannot assign superadmin role via API"})
        }
        updates["role"] = req.Role
    }
    if req.Course != "" {
        updates["course"] = req.Course
    }
    if req.YearLevel != "" {
        updates["year_level"] = req.YearLevel
    }
    if req.Department != "" {
        updates["department"] = req.Department
    }
    if req.College != "" {
        updates["college"] = req.College
    }
    if req.IsVerified != nil {
        updates["is_verified"] = *req.IsVerified
        if *req.IsVerified && user.VerifiedAt.IsZero() {
            updates["verified_at"] = time.Now()
        }
    }

    if len(updates) > 0 {
        updates["updated_at"] = time.Now()
        if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
            return c.Status(500).JSON(fiber.Map{"error": "Failed to update user"})
        }
    }

    log.Printf("✅ User updated: %s by superadmin", studentID)
    
    return c.JSON(fiber.Map{
        "success": true,
        "message": "User updated successfully",
        "user": fiber.Map{
            "student_id": user.StudentID,
            "email":      user.Email,
            "role":       user.Role,
            "is_verified": user.IsVerified,
        },
    })
}

// GenerateAttendanceReport generates comprehensive attendance report
func GenerateAttendanceReport(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" && userRole != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    startDate := c.Query("start_date")
    endDate := c.Query("end_date")
    eventID := c.QueryInt("event_id", 0)

    query := config.DB.Model(&models.Attendance{})
    
    if startDate != "" {
        query = query.Where("DATE(scan_time) >= ?", startDate)
    }
    if endDate != "" {
        query = query.Where("DATE(scan_time) <= ?", endDate)
    }
    if eventID > 0 {
        query = query.Where("event_id = ?", eventID)
    }

    var attendances []models.Attendance
    if err := query.Order("scan_time DESC").Find(&attendances).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to generate report"})
    }

    // Enrich with user details
    var report []fiber.Map
    for _, att := range attendances {
        var student models.User
        config.DB.Where("student_id = ?", att.StudentID).First(&student)
        
        var event models.Event
        config.DB.Where("id = ?", att.EventID).First(&event)

        report = append(report, fiber.Map{
            "scan_id":      att.ID,
            "student_id":   att.StudentID,
            "student_name": student.FirstName + " " + student.LastName,
            "course":       student.Course,
            "year_level":   student.YearLevel,
            "event_name":   att.EventName,
            "event_id":     att.EventID,
            "qr_code_type": att.QRCodeType,
            "scanned_by":   att.ScannedBy,
            "scanner_role": att.ScannerRole,
            "scan_time":    att.ScanTime,
            "status":       att.Status,
            "location":     att.Location,
        })
    }

    // Generate summary statistics
    var summary struct {
        TotalScans     int64
        UniqueStudents int64
        TotalEvents    int64
    }

    if startDate != "" && endDate != "" {
        config.DB.Model(&models.Attendance{}).
            Where("DATE(scan_time) BETWEEN ? AND ?", startDate, endDate).
            Count(&summary.TotalScans)
        
        config.DB.Model(&models.Attendance{}).
            Select("COUNT(DISTINCT student_id)").
            Where("DATE(scan_time) BETWEEN ? AND ?", startDate, endDate).
            Scan(&summary.UniqueStudents)
    } else {
        config.DB.Model(&models.Attendance{}).Count(&summary.TotalScans)
        config.DB.Model(&models.Attendance{}).Select("COUNT(DISTINCT student_id)").Scan(&summary.UniqueStudents)
    }
    
    config.DB.Model(&models.Event{}).Count(&summary.TotalEvents)

    return c.JSON(fiber.Map{
        "report": report,
        "summary": fiber.Map{
            "total_scans":     summary.TotalScans,
            "unique_students": summary.UniqueStudents,
            "total_events":    summary.TotalEvents,
            "date_range":      startDate + " to " + endDate,
        },
        "generated_at": time.Now(),
    })
}

// GenerateUsersReport generates users report
func GenerateUsersReport(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" && userRole != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    var users []models.User
    if err := config.DB.Order("created_at DESC").Find(&users).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch users"})
    }

    var report []fiber.Map
    for _, user := range users {
        var attendanceCount int64
        config.DB.Model(&models.Attendance{}).Where("student_id = ?", user.StudentID).Count(&attendanceCount)

        report = append(report, fiber.Map{
            "student_id":    user.StudentID,
            "email":         user.Email,
            "full_name":     user.FirstName + " " + user.LastName,
            "role":          user.Role,
            "course":        user.Course,
            "year_level":    user.YearLevel,
            "is_verified":   user.IsVerified,
            "attendance_count": attendanceCount,
            "created_at":    user.CreatedAt,
            "verified_at":   user.VerifiedAt,
        })
    }

    // Summary
    var verifiedCount, adminCount, studentCount int64
    config.DB.Model(&models.User{}).Where("is_verified = ?", true).Count(&verifiedCount)
    config.DB.Model(&models.User{}).Where("role IN ?", []string{"admin", "superadmin"}).Count(&adminCount)
    config.DB.Model(&models.User{}).Where("role = ?", "student").Count(&studentCount)

    return c.JSON(fiber.Map{
        "report": report,
        "summary": fiber.Map{
            "total_users":       len(users),
            "verified_users":    verifiedCount,
            "admin_users":       adminCount,
            "student_users":     studentCount,
        },
        "generated_at": time.Now(),
    })
}

// GenerateEventsReport generates events report
func GenerateEventsReport(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" && userRole != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    startDate := c.Query("start_date")
    endDate := c.Query("end_date")

    query := config.DB.Model(&models.Event{})
    
    if startDate != "" {
        query = query.Where("DATE(created_at) >= ?", startDate)
    }
    if endDate != "" {
        query = query.Where("DATE(created_at) <= ?", endDate)
    }

    var events []models.Event
    if err := query.Order("created_at DESC").Find(&events).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch events"})
    }

    var report []fiber.Map
    for _, event := range events {
        var attendanceCount int64
        config.DB.Model(&models.Attendance{}).Where("event_id = ?", event.ID).Count(&attendanceCount)

        var targetCourses, targetYearLevels, targetSections []string
        
        if event.TargetCourses != "" {
            json.Unmarshal([]byte(event.TargetCourses), &targetCourses)
        }
        if event.TargetYearLevels != "" {
            json.Unmarshal([]byte(event.TargetYearLevels), &targetYearLevels)
        }
        if event.TargetSections != "" {
            json.Unmarshal([]byte(event.TargetSections), &targetSections)
        }

        report = append(report, fiber.Map{
            "event_id":     event.ID,
            "event_name":   event.EventName,
            "event_type":   event.EventType,
            "description":  event.Description,
            "location":     event.Location,
            "target_courses": targetCourses,
            "target_year_levels": targetYearLevels,
            "target_sections": targetSections,
            "is_active":    event.IsActive,
            "start_time":   event.StartTime,
            "end_time":     event.EndTime,
            "attendance_count": attendanceCount,
            "created_by":   event.CreatedBy,
            "created_at":   event.CreatedAt,
        })
    }

    // Summary
    var activeEvents, completedEvents int64
    config.DB.Model(&models.Event{}).Where("is_active = ? AND end_time > ?", true, time.Now()).Count(&activeEvents)
    config.DB.Model(&models.Event{}).Where("end_time <= ?", time.Now()).Count(&completedEvents)

    return c.JSON(fiber.Map{
        "report": report,
        "summary": fiber.Map{
            "total_events":     len(events),
            "active_events":    activeEvents,
            "completed_events": completedEvents,
            "date_range":       startDate + " to " + endDate,
        },
        "generated_at": time.Now(),
    })
}

// CleanupSystem performs system cleanup tasks
func CleanupSystem(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Superadmin access required"})
    }

    // Cleanup expired temp users
    result := config.DB.Where("expires_at < ?", time.Now()).Delete(&models.TempUser{})
    expiredTempUsers := result.RowsAffected

    // Cleanup attendance records older than 90 days
    result = config.DB.Where("scan_time < ?", time.Now().AddDate(0, 0, -90)).Delete(&models.Attendance{})
    oldAttendance := result.RowsAffected

    return c.JSON(fiber.Map{
        "success": true,
        "message": "System cleanup completed",
        "results": fiber.Map{
            "expired_temp_users_deleted": expiredTempUsers,
            "old_attendance_deleted":     oldAttendance,
            "cleanup_time":               time.Now(),
        },
    })
}

// ResetAllQRCodes resets all QR codes to default type
func ResetAllQRCodes(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": "Superadmin access required"})
    }

    // Reset user QR codes
    result := config.DB.Model(&models.User{}).Updates(map[string]interface{}{
        "qr_code_type": "default",
        "updated_at":   time.Now(),
    })

    // Reset event QR codes
    eventResult := config.DB.Model(&models.Event{}).Updates(map[string]interface{}{
        "qr_code_type": "default",
        "updated_at":   time.Now(),
    })

    return c.JSON(fiber.Map{
        "success": true,
        "message": "All QR codes reset to default",
        "results": fiber.Map{
            "users_updated":  result.RowsAffected,
            "events_updated": eventResult.RowsAffected,
            "reset_time":     time.Now(),
        },
    })
}

// GetAllAttendance returns all attendance records for dashboard
func GetAllAttendance(c *fiber.Ctx) error {
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "superadmin" && userRole != "admin" {
        return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
    }

    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 50)
    offset := (page - 1) * limit

    var attendances []models.Attendance
    var total int64

    query := config.DB.Model(&models.Attendance{})
    
    // Apply filters
    if eventID := c.Query("event_id"); eventID != "" {
        query = query.Where("event_id = ?", eventID)
    }
    if studentID := c.Query("student_id"); studentID != "" {
        query = query.Where("student_id = ?", studentID)
    }
    if date := c.Query("date"); date != "" {
        query = query.Where("DATE(scan_time) = ?", date)
    }

    query.Count(&total)
    query.Order("scan_time DESC").Offset(offset).Limit(limit).Find(&attendances)

    // Enrich with details
    var attendanceList []fiber.Map
    for _, att := range attendances {
        var student models.User
        config.DB.Where("student_id = ?", att.StudentID).First(&student)
        
        var event models.Event
        config.DB.Where("id = ?", att.EventID).First(&event)

        attendanceList = append(attendanceList, fiber.Map{
            "id":           att.ID,
            "student_id":   att.StudentID,
            "student_name": student.FirstName + " " + student.LastName,
            "event_id":     att.EventID,
            "event_name":   att.EventName,
            "qr_code_type": att.QRCodeType,
            "scanned_by":   att.ScannedBy,
            "scanner_role": att.ScannerRole,
            "scan_time":    att.ScanTime,
            "status":       att.Status,
            "location":     att.Location,
        })
    }

    return c.JSON(fiber.Map{
        "attendance":   attendanceList,
        "total":        total,
        "page":         page,
        "limit":        limit,
        "total_pages":  (int(total) + limit - 1) / limit,
    })
}