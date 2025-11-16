package models

import "time"

type User struct {
    ID              uint      `json:"-" gorm:"primaryKey;autoIncrement"`
    UserID          string    `json:"user_id" gorm:"uniqueIndex;type:varchar(255)"`
    Email           string    `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)"`
    Password        string    `json:"password" gorm:"not null;type:varchar(255)"`
    Username        string    `json:"username" gorm:"not null;type:varchar(255)"`
    Role            string    `json:"role" gorm:"not null;type:varchar(50);default:'student'"`
    IsVerified      bool      `json:"is_verified" gorm:"default:false"`
    CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
    VerifiedAt      time.Time `json:"verified_at"`
    

    StudentNumber   string    `json:"student_number,omitempty" gorm:"type:varchar(100)"`
    FirstName       string    `json:"first_name" gorm:"not null;type:varchar(100)"`
    LastName        string    `json:"last_name" gorm:"not null;type:varchar(100)"`
    MiddleName      string    `json:"middle_name,omitempty" gorm:"type:varchar(100)"`
    Course          string    `json:"course" gorm:"type:varchar(100)"`
    YearLevel       string    `json:"year_level" gorm:"type:varchar(50)"`
    Section         string    `json:"section,omitempty" gorm:"type:varchar(50)"`
    Department      string    `json:"department,omitempty" gorm:"type:varchar(100)"`
    College         string    `json:"college,omitempty" gorm:"type:varchar(100)"`
    ContactNumber   string    `json:"contact_number,omitempty" gorm:"type:varchar(20)"`
    Address         string    `json:"address,omitempty" gorm:"type:text"`

    ResetToken      string    `json:"-" gorm:"type:varchar(64)"`
    ResetTokenExpiry time.Time `json:"-" gorm:"type:timestamp"`
    ResetAttempts   int       `json:"-" gorm:"default:0"`
    LastResetRequest time.Time `json:"-" gorm:"type:timestamp"`
    

    QRCodeData      string    `json:"qr_code_data,omitempty" gorm:"type:text"`
    QRCodeType      string    `json:"qr_code_type,omitempty" gorm:"type:varchar(50);default:'student_id'"`
}


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

type RegisterRequest struct {
    Email         string `json:"email" binding:"required,email"`
    Password      string `json:"password" binding:"required,min=6"`
    Username      string `json:"username" binding:"required"`
    

    StudentNumber string `json:"student_number,omitempty"`
    FirstName     string `json:"first_name" binding:"required"`
    LastName      string `json:"last_name" binding:"required"`
    MiddleName    string `json:"middle_name,omitempty"`
    Course        string `json:"course" binding:"required"`
    YearLevel     string `json:"year_level" binding:"required"`
    Section       string `json:"section,omitempty"`
    Department    string `json:"department,omitempty"`
    College       string `json:"college,omitempty"`
    ContactNumber string `json:"contact_number,omitempty"`
    Address       string `json:"address,omitempty"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type VerifyRequest struct {
    Email string `json:"email" binding:"required,email"`
    Code  string `json:"code" binding:"required"`
}


type CreateQRCodeTypeRequest struct {
    TypeName    string `json:"type_name" binding:"required"`
    Description string `json:"description" binding:"required"`
}

type CreateEventRequest struct {
    EventName   string    `json:"event_name" binding:"required"`
    EventType   string    `json:"event_type" binding:"required"`
    Description string    `json:"description" binding:"required"`
    Course      string    `json:"course,omitempty"`
    Department  string    `json:"department,omitempty"`
    College     string    `json:"college,omitempty"`
    StartTime   time.Time `json:"start_time" binding:"required"`
    EndTime     time.Time `json:"end_time" binding:"required"`
}

type UpdateUserQRCodeRequest struct {
    UserID     string `json:"user_id" binding:"required"`
    QRCodeType string `json:"qr_code_type" binding:"required"`
}

type PromoteToAdminRequest struct {
    TargetUserID string `json:"target_user_id" binding:"required"`
    Department   string `json:"department,omitempty"`
    College      string `json:"college,omitempty"`
}

type DemoteToStudentRequest struct {
    TargetUserID string `json:"target_user_id" binding:"required"`
}

type AdminListResponse struct {
    UserID       string `json:"user_id"`
    Email        string `json:"email"`
    Username     string `json:"username"`
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Role         string `json:"role"`
    Department   string `json:"department"`
    College      string `json:"college"`
    IsVerified   bool   `json:"is_verified"`
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

type TempUser struct {
    ID                uint      `json:"-" gorm:"primaryKey;autoIncrement"`
    Email             string    `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)"`
    Password          string    `json:"password" gorm:"not null;type:varchar(255)"`
    Username          string    `json:"username" gorm:"not null;type:varchar(255)"`
    StudentNumber     string    `json:"student_number,omitempty" gorm:"type:varchar(100)"`
    FirstName         string    `json:"first_name" gorm:"not null;type:varchar(100)"`
    LastName          string    `json:"last_name" gorm:"not null;type:varchar(100)"`
    MiddleName        string    `json:"middle_name,omitempty" gorm:"type:varchar(100)"`
    Course            string    `json:"course" gorm:"type:varchar(100)"`
    YearLevel         string    `json:"year_level" gorm:"type:varchar(50)"`
    Section           string    `json:"section,omitempty" gorm:"type:varchar(50)"`
    Department        string    `json:"department,omitempty" gorm:"type:varchar(100)"`
    College           string    `json:"college,omitempty" gorm:"type:varchar(100)"`
    ContactNumber     string    `json:"contact_number,omitempty" gorm:"type:varchar(20)"`
    Address           string    `json:"address,omitempty" gorm:"type:text"`
    VerificationCode  string    `json:"-" gorm:"type:varchar(10)"`
    ExpiresAt         time.Time `json:"expires_at"`
    CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
}



func (TempUser) TableName() string {
    return "temp_users"
}