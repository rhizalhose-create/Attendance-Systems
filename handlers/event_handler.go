package handlers

import (
    "encoding/json"
    "log"
    "strconv"
    "time"
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/service"
    "AttendanceManagementSystem/utils"

    "github.com/gofiber/fiber/v2"
)

func CreateEventHandler(c *fiber.Ctx) error {
    var req models.CreateEventRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
    }

    // Validate required fields
    if req.EventName == "" || req.EventType == "" || req.Description == "" || req.QRCodeType == "" {
        return c.Status(400).JSON(fiber.Map{"error": "Event name, type, description, and QR code type are required"})
    }

    // Validate course and year level selection
    if len(req.TargetCourses) == 0 {
        return c.Status(400).JSON(fiber.Map{"error": "At least one target course is required"})
    }
    if len(req.TargetYearLevels) == 0 {
        return c.Status(400).JSON(fiber.Map{"error": "At least one target year level is required"})
    }

    // Get user ID from context
    userID := c.Get(utils.HeaderStudentID)

    log.Printf("🎯 CREATE EVENT REQUEST:")
    log.Printf("   Event: %s", req.EventName)
    log.Printf("   Type: %s", req.EventType)
    log.Printf("   QR Code Type: %s", req.QRCodeType)
    log.Printf("   Target Courses: %v", req.TargetCourses)
    log.Printf("   Target Year Levels: %v", req.TargetYearLevels)
    log.Printf("   Target Sections: %v", req.TargetSections)
    log.Printf("   Created By: %s", userID)

    // Use service to create event and update QR codes
    event, err := service.CreateEventWithQRUpdate(&req, userID)
    if err != nil {
        log.Printf("❌ Failed to create event: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }

    // Parse back the JSON arrays for response
    var targetCourses []string
    var targetYearLevels []string
    var targetSections []string
    
    json.Unmarshal([]byte(event.TargetCourses), &targetCourses)
    json.Unmarshal([]byte(event.TargetYearLevels), &targetYearLevels)
    json.Unmarshal([]byte(event.TargetSections), &targetSections)

    log.Printf("✅ Event created successfully: %s by %s", event.EventName, userID)
    log.Printf("📊 Affected: Courses=%v, YearLevels=%v, Sections=%v", targetCourses, targetYearLevels, targetSections)
    log.Printf("🏷️ QR Code Type: %s", event.QRCodeType)

    // DEBUG: Check if students were actually updated
    checkQRCodeUpdates(targetCourses, targetYearLevels, targetSections, event.QRCodeType)

    return c.JSON(fiber.Map{
        "message": "Event created successfully and student QR codes updated",
        "event": fiber.Map{
            "id":                event.ID,
            "event_name":        event.EventName,
            "event_type":        event.EventType,
            "description":       event.Description,
            "location":          event.Location,
            "target_courses":    targetCourses,
            "target_year_levels": targetYearLevels,
            "target_sections":   targetSections,
            "qr_code_type":      event.QRCodeType,
            "department":        event.Department,
            "college":           event.College,
            "created_by":        event.CreatedBy,
            "start_time":        event.StartTime,
            "end_time":          event.EndTime,
            "is_active":         event.IsActive,
        },
    })
}

// Add this debug function to check QR code updates
func checkQRCodeUpdates(targetCourses, targetYearLevels, targetSections []string, qrCodeType string) {
    var students []models.User
    query := config.DB.Where("role = ?", "student")
    
    if len(targetCourses) > 0 {
        query = query.Where("course IN ?", targetCourses)
    }
    if len(targetYearLevels) > 0 {
        query = query.Where("year_level IN ?", targetYearLevels)
    }
    if len(targetSections) > 0 {
        query = query.Where("section IN ?", targetSections)
    }

    if err := query.Find(&students).Error; err != nil {
        log.Printf("❌ DEBUG: Failed to fetch students for verification: %v", err)
        return
    }

    log.Printf("🔍 DEBUG: Found %d students matching event criteria", len(students))
    
    updatedCount := 0
    for _, student := range students {
        if student.QRCodeType == qrCodeType {
            updatedCount++
        } else {
            log.Printf("   ❌ Student %s (%s %s) still has QR type: %s, expected: %s", 
                student.StudentID, student.FirstName, student.LastName, 
                student.QRCodeType, qrCodeType)
        }
    }
    
    log.Printf("📊 DEBUG: %d/%d students have correct QR code type", updatedCount, len(students))
}
// GetEventsHandler - Get all active events
func GetEventsHandler(c *fiber.Ctx) error {
    var events []models.Event
    
    // Get query parameters for filtering
    eventType := c.Query("type")
    department := c.Query("department")
    college := c.Query("college")

    query := config.DB.Where("is_active = ? AND end_time > ?", true, time.Now())

    // Apply filters if provided
    if eventType != "" {
        query = query.Where("event_type = ?", eventType)
    }
    if department != "" {
        query = query.Where("department = ?", department)
    }
    if college != "" {
        query = query.Where("college = ?", college)
    }

    if err := query.Order("start_time ASC").Find(&events).Error; err != nil {
        log.Printf("❌ Failed to fetch events: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch events"})
    }

    // Parse target arrays for each event
    var eventList []fiber.Map
    for _, event := range events {
        var targetCourses []string
        var targetYearLevels []string
        var targetSections []string
        
        json.Unmarshal([]byte(event.TargetCourses), &targetCourses)
        json.Unmarshal([]byte(event.TargetYearLevels), &targetYearLevels)
        json.Unmarshal([]byte(event.TargetSections), &targetSections)

        eventList = append(eventList, fiber.Map{
            "id":                event.ID,
            "event_name":        event.EventName,
            "event_type":        event.EventType,
            "description":       event.Description,
            "location":          event.Location,
            "target_courses":    targetCourses,
            "target_year_levels": targetYearLevels,
            "target_sections":   targetSections,
            "department":        event.Department,
            "college":           event.College,
            "qr_code_type":      event.QRCodeType,
            "created_by":        event.CreatedBy,
            "start_time":        event.StartTime,
            "end_time":          event.EndTime,
            "is_active":         event.IsActive,
            "created_at":        event.CreatedAt,
            "updated_at":        event.UpdatedAt,
        })
    }

    return c.JSON(fiber.Map{
        "events": eventList,
        "count":  len(eventList),
    })
}

// GetEventByIDHandler - Get specific event by ID
func GetEventByIDHandler(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid event ID"})
    }

    var event models.Event
    if err := config.DB.Where("id = ?", id).First(&event).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
    }

    // Parse target arrays
    var targetCourses []string
    var targetYearLevels []string
    var targetSections []string
    
    json.Unmarshal([]byte(event.TargetCourses), &targetCourses)
    json.Unmarshal([]byte(event.TargetYearLevels), &targetYearLevels)
    json.Unmarshal([]byte(event.TargetSections), &targetSections)

    return c.JSON(fiber.Map{
        "event": fiber.Map{
            "id":                event.ID,
            "event_name":        event.EventName,
            "event_type":        event.EventType,
            "description":       event.Description,
            "location":          event.Location,
            "target_courses":    targetCourses,
            "target_year_levels": targetYearLevels,
            "target_sections":   targetSections,
            "department":        event.Department,
            "college":           event.College,
            "qr_code_type":      event.QRCodeType,
            "created_by":        event.CreatedBy,
            "start_time":        event.StartTime,
            "end_time":          event.EndTime,
            "is_active":         event.IsActive,
            "created_at":        event.CreatedAt,
            "updated_at":        event.UpdatedAt,
        },
    })
}

// UpdateEventHandler - Update event (Admin/SuperAdmin only)
func UpdateEventHandler(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid event ID"})
    }

    var req models.UpdateEventRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": utils.ErrCannotParseJSON})
    }

    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
    }

    // Find event
    var event models.Event
    if err := config.DB.Where("id = ?", id).First(&event).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
    }

    // Update fields if provided
    if req.EventName != "" {
        event.EventName = req.EventName
    }
    if req.EventType != "" {
        event.EventType = req.EventType
    }
    if req.Description != "" {
        event.Description = req.Description
    }
    if req.Location != "" {
        event.Location = req.Location
    }
    if req.Department != "" {
        event.Department = req.Department
    }
    if req.College != "" {
        event.College = req.College
    }
    if req.QRCodeType != "" {
        event.QRCodeType = req.QRCodeType
    }
    if !req.StartTime.IsZero() {
        event.StartTime = req.StartTime
    }
    if !req.EndTime.IsZero() {
        event.EndTime = req.EndTime
    }
    if req.IsActive != nil {
        event.IsActive = *req.IsActive
    }

    // Update target arrays if provided
    if len(req.TargetCourses) > 0 {
        coursesJSON, _ := json.Marshal(req.TargetCourses)
        event.TargetCourses = string(coursesJSON)
    }
    if len(req.TargetYearLevels) > 0 {
        yearLevelsJSON, _ := json.Marshal(req.TargetYearLevels)
        event.TargetYearLevels = string(yearLevelsJSON)
    }
    if len(req.TargetSections) > 0 {
        sectionsJSON, _ := json.Marshal(req.TargetSections)
        event.TargetSections = string(sectionsJSON)
    }

    event.UpdatedAt = time.Now()

    if err := config.DB.Save(&event).Error; err != nil {
        log.Printf("❌ Failed to update event: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to update event"})
    }

    log.Printf("✅ Event updated successfully: %s", event.EventName)
    return c.JSON(fiber.Map{
        "message": "Event updated successfully",
        "event":   event,
    })
}

// DeleteEventHandler - Soft delete event (Admin/SuperAdmin only)
func DeleteEventHandler(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid event ID"})
    }

    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
    }

    // Find event
    var event models.Event
    if err := config.DB.Where("id = ?", id).First(&event).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
    }

    // Soft delete by setting is_active to false
    event.IsActive = false
    event.UpdatedAt = time.Now()

    if err := config.DB.Save(&event).Error; err != nil {
        log.Printf("❌ Failed to delete event: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to delete event"})
    }

    log.Printf("✅ Event deleted successfully: %s", event.EventName)

    return c.JSON(fiber.Map{
        "message": "Event deleted successfully",
    })
}

// GetMyEventsHandler - Get events created by current user
func GetMyEventsHandler(c *fiber.Ctx) error {
    userID := c.Get(utils.HeaderStudentID)
    userRole := c.Get(utils.HeaderUserRole)

    var events []models.Event
    query := config.DB.Where("is_active = ?", true)

    // If user is admin/superadmin, show events they created
    if userRole == "admin" || userRole == "superadmin" {
        query = query.Where("created_by = ?", userID)
    }
    // If user is student, show all active events (no filter by created_by)

    if err := query.Order("start_time ASC").Find(&events).Error; err != nil {
        log.Printf("❌ Failed to fetch events: %v", err)
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch events"})
    }

    // Parse target arrays for each event
    var eventList []fiber.Map
    for _, event := range events {
        var targetCourses []string
        var targetYearLevels []string
        var targetSections []string
        
        json.Unmarshal([]byte(event.TargetCourses), &targetCourses)
        json.Unmarshal([]byte(event.TargetYearLevels), &targetYearLevels)
        json.Unmarshal([]byte(event.TargetSections), &targetSections)

        eventList = append(eventList, fiber.Map{
            "id":                event.ID,
            "event_name":        event.EventName,
            "event_type":        event.EventType,
            "description":       event.Description,
            "location":          event.Location,
            "target_courses":    targetCourses,
            "target_year_levels": targetYearLevels,
            "target_sections":   targetSections,
            "department":        event.Department,
            "college":           event.College,
            "qr_code_type":      event.QRCodeType,
            "created_by":        event.CreatedBy,
            "start_time":        event.StartTime,
            "end_time":          event.EndTime,
            "is_active":         event.IsActive,
        })
    }

    return c.JSON(fiber.Map{
        "events": eventList,
        "count":  len(eventList),
    })
}

// GetAvailableOptions - Get available courses, year levels, and sections for dropdowns
func GetAvailableOptions(c *fiber.Ctx) error {
    courses, err := service.GetAvailableCourses()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch courses"})
    }

    yearLevels, err := service.GetAvailableYearLevels()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch year levels"})
    }

    sections, err := service.GetAvailableSections()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch sections"})
    }

    return c.JSON(fiber.Map{
        "courses":    courses,
        "year_levels": yearLevels,
        "sections":   sections,
    })
}

// GetEventStudents - Get all students affected by an event
func GetEventStudents(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid event ID"})
    }

    // Find event
    var event models.Event
    if err := config.DB.Where("id = ?", id).First(&event).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
    }

    // Parse target arrays
    var targetCourses []string
    var targetYearLevels []string
    var targetSections []string
    
    json.Unmarshal([]byte(event.TargetCourses), &targetCourses)
    json.Unmarshal([]byte(event.TargetYearLevels), &targetYearLevels)
    json.Unmarshal([]byte(event.TargetSections), &targetSections)

    // Get affected students
    var students []models.User
    query := config.DB.Where("role = ?", "student")
    
    if len(targetCourses) > 0 {
        query = query.Where("course IN ?", targetCourses)
    }
    if len(targetYearLevels) > 0 {
        query = query.Where("year_level IN ?", targetYearLevels)
    }
    if len(targetSections) > 0 {
        query = query.Where("section IN ?", targetSections)
    }

    if err := query.Find(&students).Error; err != nil {
        return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch students"})
    }

    var studentList []fiber.Map
    for _, student := range students {
        studentList = append(studentList, fiber.Map{
            "student_id":   student.StudentID,
            "email":        student.Email,
            "first_name":   student.FirstName,
            "last_name":    student.LastName,
            "course":       student.Course,
            "year_level":   student.YearLevel,
            "section":      student.Section,
            "qr_code_type": student.QRCodeType,
            "qr_code_data": student.QRCodeData,
        })
    }

    return c.JSON(fiber.Map{
        "event": fiber.Map{
            "id":           event.ID,
            "event_name":   event.EventName,
            "target_courses": targetCourses,
            "target_year_levels": targetYearLevels,
            "target_sections": targetSections,
        },
        "students": studentList,
        "count":    len(studentList),
    })
}

// RefreshEventQRCodes - Refresh QR codes for existing event
func RefreshEventQRCodes(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 32)
    if err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid event ID"})
    }

    // Check if user is admin or superadmin
    userRole := c.Get(utils.HeaderUserRole)
    if userRole != "admin" && userRole != "superadmin" {
        return c.Status(403).JSON(fiber.Map{"error": utils.ErrAdminAccessRequired})
    }

    // Find event
    var event models.Event
    if err := config.DB.Where("id = ?", id).First(&event).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Event not found"})
    }



    return c.JSON(fiber.Map{
        "message": "QR codes refreshed successfully for event",
        "event_id": event.ID,
        "event_name": event.EventName,
    })
}