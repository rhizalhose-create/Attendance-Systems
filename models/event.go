package models

import "time"

type Event struct {
    ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    EventName   string    `json:"event_name" gorm:"not null;type:varchar(255)"`
    EventType   string    `json:"event_type" gorm:"not null;type:varchar(100)"`
    Description string    `json:"description" gorm:"type:text"`
    Location    string    `json:"location,omitempty" gorm:"type:varchar(255)"`
    
    // EVENT SCOPE FIELDS
    TargetCourses    string `json:"target_courses" gorm:"type:text"`   // JSON array of courses
    TargetYearLevels string `json:"target_year_levels" gorm:"type:text"` // JSON array of year levels
    TargetSections   string `json:"target_sections,omitempty" gorm:"type:text"` // JSON array of sections
    
    // QR CODE TYPE FOR THIS EVENT
    QRCodeType  string    `json:"qr_code_type" gorm:"type:varchar(100)"`
    
    Department  string    `json:"department,omitempty" gorm:"type:varchar(100)"`
    College     string    `json:"college,omitempty" gorm:"type:varchar(100)"`
    CreatedBy   string    `json:"created_by" gorm:"type:varchar(255)"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    StartTime   time.Time `json:"start_time" gorm:"type:timestamp"`
    EndTime     time.Time `json:"end_time" gorm:"type:timestamp"`
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CreateEventRequest struct {
    EventName   string    `json:"event_name" binding:"required"`
    EventType   string    `json:"event_type" binding:"required"`
    Description string    `json:"description" binding:"required"`
    Location    string    `json:"location,omitempty"`
    
    // Course and Year Level Selection
    TargetCourses   []string `json:"target_courses" binding:"required"`     // ["Information Systems", "Computer Science"]
    TargetYearLevels []string `json:"target_year_levels" binding:"required"` // ["1st Year", "2nd Year"]
    TargetSections  []string `json:"target_sections,omitempty"`             // ["A", "B"] - optional
    
    // QR Code Type for this event
    QRCodeType  string    `json:"qr_code_type" binding:"required"`
    
    Department  string    `json:"department,omitempty"`
    College     string    `json:"college,omitempty"`
    StartTime   time.Time `json:"start_time" binding:"required"`
    EndTime     time.Time `json:"end_time" binding:"required"`
}

type UpdateEventRequest struct {
    EventName   string    `json:"event_name,omitempty"`
    EventType   string    `json:"event_type,omitempty"`
    Description string    `json:"description,omitempty"`
    Location    string    `json:"location,omitempty"`
    
    TargetCourses   []string `json:"target_courses,omitempty"`
    TargetYearLevels []string `json:"target_year_levels,omitempty"`
    TargetSections  []string `json:"target_sections,omitempty"`
    
    QRCodeType  string    `json:"qr_code_type,omitempty"`
    
    Department  string    `json:"department,omitempty"`
    College     string    `json:"college,omitempty"`
    StartTime   time.Time `json:"start_time,omitempty"`
    EndTime     time.Time `json:"end_time,omitempty"`
    IsActive    *bool     `json:"is_active,omitempty"`
}