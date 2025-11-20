package models

import "time"

type QRCodeType struct {
    ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    TypeName    string    `json:"type_name" gorm:"uniqueIndex;type:varchar(100)"`
    Description string    `json:"description" gorm:"type:text"`
    CreatedBy   string    `json:"created_by" gorm:"type:varchar(255)"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type QRCodeEvent struct {
    ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    EventName   string    `json:"event_name" gorm:"type:varchar(255)"`
    EventType   string    `json:"event_type" gorm:"type:varchar(100)"`
    Description string    `json:"description" gorm:"type:text"`
    Course      string    `json:"course,omitempty" gorm:"type:varchar(100)"`
    Department  string    `json:"department,omitempty" gorm:"type:varchar(100)"`
    College     string    `json:"college,omitempty" gorm:"type:varchar(100)"`
    CreatedBy   string    `json:"created_by" gorm:"type:varchar(255)"`
    IsActive    bool      `json:"is_active" gorm:"default:true"`
    StartTime   time.Time `json:"start_time" gorm:"type:timestamp"`
    EndTime     time.Time `json:"end_time" gorm:"type:timestamp"`
    CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
    UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type QRCodeScan struct {
    ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    UserID      string    `json:"user_id" gorm:"type:varchar(255)"`
    EventID     uint      `json:"event_id" gorm:"type:integer"`
    QRCodeType  string    `json:"qr_code_type" gorm:"type:varchar(100)"`
    ScannedAt   time.Time `json:"scanned_at" gorm:"autoCreateTime"`
    ScannerID   string    `json:"scanner_id,omitempty" gorm:"type:varchar(255)"`
    Location    string    `json:"location,omitempty" gorm:"type:varchar(255)"`
}

type CreateQRCodeTypeRequest struct {
    TypeName    string `json:"type_name" binding:"required"`
    Description string `json:"description" binding:"required"`
}

type UpdateUserQRCodeRequest struct {
    UserID     string `json:"user_id" binding:"required"`
    QRCodeType string `json:"qr_code_type" binding:"required"`
}

type UpdateCourseQRCodeRequest struct {
    Course     string `json:"course" binding:"required"`
    YearLevel  string `json:"year_level" binding:"required"`
    QRCodeType string `json:"qr_code_type" binding:"required"`
}

type BulkQRCodeUpdateResponse struct {
    Message      string `json:"message"`
    UpdatedCount int    `json:"updated_count"`
    Course       string `json:"course"`
    YearLevel    string `json:"year_level"`
    QRCodeType   string `json:"qr_code_type"`
}