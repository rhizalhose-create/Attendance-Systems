package models

import "time"

type User struct {
    ID              uint      `json:"-" gorm:"primaryKey;autoIncrement"`
    UserID          string    `json:"user_id" gorm:"uniqueIndex;type:varchar(255);not null"`
    Email           string    `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)"`
    Password        string    `json:"password" gorm:"not null;type:varchar(255)"`
    Username        string    `json:"username" gorm:"not null;type:varchar(255)"`
    Role            string    `json:"role" gorm:"not null;type:varchar(50);default:'student'"`
    IsVerified      bool      `json:"is_verified" gorm:"default:false"`
    CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
    VerifiedAt      time.Time `json:"verified_at"`

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

    ResetToken       string    `json:"-" gorm:"type:varchar(64)"`
    ResetTokenExpiry time.Time `json:"-" gorm:"type:timestamp"`
    ResetAttempts    int       `json:"-" gorm:"default:0"`
    LastResetRequest time.Time `json:"-" gorm:"type:timestamp"`

    QRCodeData       string    `json:"qr_code_data,omitempty" gorm:"type:text"`
    QRCodeType       string    `json:"qr_code_type,omitempty" gorm:"type:varchar(50);default:'student_id'"`
}

type RegisterRequest struct {
    UserID        string `json:"user_id" binding:"required"`
    Email         string `json:"email" binding:"required,email"`
    Password      string `json:"password" binding:"required,min=6"`
    Username      string `json:"username" binding:"required"`
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
