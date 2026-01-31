// models/attendance.go
package models

import "time"

type Attendance struct {
    ID           uint      `json:"id" gorm:"primaryKey;autoIncrement"`
    StudentID    string    `json:"student_id" gorm:"type:varchar(255);index"`
    EventID      uint      `json:"event_id" gorm:"index"`
    EventName    string    `json:"event_name" gorm:"type:varchar(255)"`
    QRCodeType   string    `json:"qr_code_type" gorm:"type:varchar(100)"`
    ScannedBy    string    `json:"scanned_by" gorm:"type:varchar(255)"` // Admin who scanned
    ScannerRole  string    `json:"scanner_role" gorm:"type:varchar(50)"` // admin or superadmin
    ScanTime     time.Time `json:"scan_time" gorm:"autoCreateTime"`
    Location     string    `json:"location,omitempty" gorm:"type:varchar(255)"`
    Status       string    `json:"status" gorm:"type:varchar(20);default:'present'"`
    Notes        string    `json:"notes,omitempty" gorm:"type:text"`
}

type AttendanceRequest struct {
    StudentID  string `json:"student_id" binding:"required"`
    EventID    uint   `json:"event_id" binding:"required"`
    QRCodeType string `json:"qr_code_type" binding:"required"`
}

type AttendanceResponse struct {
    ID           uint      `json:"id"`
    StudentID    string    `json:"student_id"`
    EventName    string    `json:"event_name"`
    QRCodeType   string    `json:"qr_code_type"`
    ScannedBy    string    `json:"scanned_by"`
    ScanTime     time.Time `json:"scan_time"`
    Status       string    `json:"status"`
    StudentName  string    `json:"student_name"`
    Course       string    `json:"course"`
    YearLevel    string    `json:"year_level"`
}