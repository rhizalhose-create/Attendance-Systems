package service

import (
    "encoding/json"
    "strings"
    "errors"
    "log"
    "time"
    "fmt"
    "AttendanceManagementSystem/config"
    "AttendanceManagementSystem/models"
    "AttendanceManagementSystem/utils"
)

// CreateEventWithQRUpdate creates event with course/year selection and updates QR codes
func CreateEventWithQRUpdate(req *models.CreateEventRequest, createdBy string) (*models.Event, error) {
    // Validate time
    if req.StartTime.After(req.EndTime) {
        return nil, errors.New("start time cannot be after end time")
    }

    if req.StartTime.Before(time.Now()) {
        return nil, errors.New("start time cannot be in the past")
    }

    // Validate required arrays
    if len(req.TargetCourses) == 0 {
        return nil, errors.New("at least one target course is required")
    }
    if len(req.TargetYearLevels) == 0 {
        return nil, errors.New("at least one target year level is required")
    }

    // Validate QR code type
    if req.QRCodeType == "" {
        return nil, errors.New("QR code type is required")
    }

    // PROPER JSON marshaling - ensure it's valid JSON arrays
    coursesJSON, err := json.Marshal(req.TargetCourses)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal courses: %v", err)
    }
    
    yearLevelsJSON, err := json.Marshal(req.TargetYearLevels)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal year levels: %v", err)
    }
    
    var sectionsJSON []byte
    if len(req.TargetSections) > 0 {
        sectionsJSON, err = json.Marshal(req.TargetSections)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal sections: %v", err)
        }
    } else {
        sectionsJSON, _ = json.Marshal([]string{}) // Empty array
    }

    log.Printf("🎯 CREATING EVENT WITH TARGETING:")
    log.Printf("  Courses: %v", req.TargetCourses)
    log.Printf("  Year Levels: %v", req.TargetYearLevels)
    log.Printf("  Sections: %v", req.TargetSections)
    log.Printf("  QR Code Type: %s", req.QRCodeType)

    event := &models.Event{
        EventName:       req.EventName,
        EventType:       req.EventType,
        Description:     req.Description,
        Location:        req.Location,
        TargetCourses:   string(coursesJSON),
        TargetYearLevels: string(yearLevelsJSON),
        TargetSections:  string(sectionsJSON),
        QRCodeType:      req.QRCodeType,
        Department:      req.Department,
        College:         req.College,
        CreatedBy:       createdBy,
        IsActive:        true,
        StartTime:       req.StartTime,
        EndTime:         req.EndTime,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    // Create event
    if err := config.DB.Create(event).Error; err != nil {
        return nil, err
    }

    log.Printf("✅ Event created with ID: %d", event.ID)

    // Update QR codes for SPECIFIC students only
    updatedCount, err := updateStudentQRCodesForEvent(event)
    if err != nil {
        log.Printf("⚠️ Event created but QR update had issues: %v", err)
        // Continue anyway - event was created successfully
    } else {
        log.Printf("✅ Updated QR codes for %d students", updatedCount)
    }

    return event, nil
}

// updateStudentQRCodesForEvent updates QR codes for students matching event criteria
func updateStudentQRCodesForEvent(event *models.Event) (int, error) {
    // Parse the JSON arrays from event
    var targetCourses []string
    var targetYearLevels []string 
    var targetSections []string
    
    if err := json.Unmarshal([]byte(event.TargetCourses), &targetCourses); err != nil {
        log.Printf("❌ Failed to parse target courses: %v", err)
        return 0, err
    }
    if err := json.Unmarshal([]byte(event.TargetYearLevels), &targetYearLevels); err != nil {
        log.Printf("❌ Failed to parse target year levels: %v", err)
        return 0, err
    }
    if err := json.Unmarshal([]byte(event.TargetSections), &targetSections); err != nil {
        log.Printf("⚠️ Failed to parse target sections (might be empty): %v", err)
        targetSections = []string{} // Empty array
    }

    log.Printf("🎯 UPDATING STUDENT QR CODES FOR EVENT:")
    log.Printf("  Event: %s (ID: %d)", event.EventName, event.ID)
    log.Printf("  Target Courses: %v", targetCourses)
    log.Printf("  Target Year Levels: %v", targetYearLevels)
    log.Printf("  Target Sections: %v", targetSections)
    log.Printf("  New QR Code Type: %s", event.QRCodeType)

    // Build query to find SPECIFIC students only
    query := config.DB.Where("role = ?", "student")
    
    // Add course filter
    if len(targetCourses) > 0 {
        query = query.Where("course IN (?)", targetCourses)
    }
    
    // Add year level filter  
    if len(targetYearLevels) > 0 {
        query = query.Where("year_level IN (?)", targetYearLevels)
    }
    
    // Add section filter (if specified)
    if len(targetSections) > 0 {
        // Clean and normalize sections
        cleanSections := make([]string, len(targetSections))
        for i, section := range targetSections {
            cleanSections[i] = strings.ToUpper(strings.TrimSpace(section))
        }
        query = query.Where("UPPER(TRIM(section)) IN (?)", cleanSections)
    }

    // Get matching students
    var students []models.User
    if err := query.Find(&students).Error; err != nil {
        log.Printf("❌ Failed to find students: %v", err)
        return 0, err
    }

    log.Printf("📊 FOUND %d STUDENTS MATCHING CRITERIA:", len(students))
    for i, student := range students {
        log.Printf("  %d. %s: %s %s | %s | %s | Section: '%s'", 
            i+1, student.StudentID, student.FirstName, student.LastName,
            student.Course, student.YearLevel, student.Section)
    }

    if len(students) == 0 {
        log.Printf("❌ No students found matching the criteria")
        return 0, nil // Return 0 updated, but no error
    }

    // Update QR codes for EACH matched student
    updatedCount := 0
    for _, student := range students {
        // Generate NEW QR code data for this event
        qrCodeData, err := generateEventQRCode(student, event)
        if err != nil {
            log.Printf("❌ Failed to generate QR for %s: %v", student.StudentID, err)
            continue
        }

        // UPDATE THE USER'S qr_code_type and qr_code_data
        updates := map[string]interface{}{
            "qr_code_data": qrCodeData,
            "qr_code_type": event.QRCodeType, // This updates the qr_code_type in users table!
        }

        if err := config.DB.Model(&models.User{}).
            Where("student_id = ?", student.StudentID).
            Updates(updates).Error; err != nil {
            log.Printf("❌ Failed to update QR for %s: %v", student.StudentID, err)
        } else {
            updatedCount++
            log.Printf(" UPDATED: %s - QR type changed to: %s", 
                student.StudentID, event.QRCodeType)
        }
    }

    log.Printf(" UPDATE SUMMARY: %d/%d students updated successfully", 
        updatedCount, len(students))
    
    return updatedCount, nil
}

// generateEventQRCode creates event-specific QR code
func generateEventQRCode(student models.User, event *models.Event) (string, error) {
    // Create event-specific QR data
    qrData := fmt.Sprintf("EVENT|%s|%s|%d|%s|%d", 
        student.StudentID, 
        event.EventName, 
        event.ID, 
        event.EventType, 
        time.Now().Unix())

    log.Printf("🔗 Generating QR for %s: %s", student.StudentID, qrData)

    // Generate the actual QR code image
    qrCodeData, err := utils.GenerateCustomQRCode(
        student.StudentID, 
        event.QRCodeType, 
        qrData,
    )
    
    if err != nil {
        return "", fmt.Errorf("failed to generate QR code: %v", err)
    }
    
    return qrCodeData, nil
}

// FormatEventQRData creates formatted data for event QR codes
func FormatEventQRData(studentID, eventName string, eventID uint, eventType string) string {
    timestamp := time.Now().Unix()
    
    // Format: EVENT|StudentID|EventName|EventID|EventType|Timestamp
    return fmt.Sprintf("EVENT|%s|%s|%d|%s|%d", 
        studentID, eventName, eventID, eventType, timestamp)
}

// UpdateStudentQRCodesForEvent - Export the function for handlers
func UpdateStudentQRCodesForEvent(event *models.Event) error {
    _, err := updateStudentQRCodesForEvent(event)
    return err
}

// GetAvailableCourses returns list of all unique courses from users
func GetAvailableCourses() ([]string, error) {
    var courses []string
    if err := config.DB.Model(&models.User{}).Where("role = ?", "student").Distinct("course").Pluck("course", &courses).Error; err != nil {
        return nil, err
    }
    return courses, nil
}

// GetAvailableYearLevels returns list of all unique year levels from users
func GetAvailableYearLevels() ([]string, error) {
    var yearLevels []string
    if err := config.DB.Model(&models.User{}).Where("role = ?", "student").Distinct("year_level").Pluck("year_level", &yearLevels).Error; err != nil {
        return nil, err
    }
    return yearLevels, nil
}

// GetAvailableSections returns list of all unique sections from users
func GetAvailableSections() ([]string, error) {
    var sections []string
    if err := config.DB.Model(&models.User{}).Where("role = ? AND section IS NOT NULL AND section != ''", "student").Distinct("section").Pluck("section", &sections).Error; err != nil {
        return nil, err
    }
    return sections, nil
}