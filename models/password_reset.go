// models/password_reset.go
package models

import "time"

type PasswordReset struct {
    ID                uint      `json:"-" gorm:"primaryKey;autoIncrement"`
    StudentID         string    `json:"student_id" gorm:"not null;type:varchar(255);index"`
    Email             string    `json:"email" gorm:"not null;type:varchar(255)"`
    VerificationCode  string    `json:"-" gorm:"type:varchar(6);not null"`
    Token             string    `json:"token" gorm:"type:varchar(64)"`
    ExpiresAt         time.Time `json:"expires_at" gorm:"not null"`
    Used              bool      `json:"used" gorm:"default:false"`
    CreatedAt         time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (PasswordReset) TableName() string {
    return "password_resets"
}

type ForgotPasswordRequest struct {
    StudentID string `json:"student_id" binding:"required"`
}

type VerifyCodeRequest struct {
    StudentID string `json:"student_id" binding:"required"`
    Code      string `json:"code" binding:"required,len=6"`
}

type ResetPasswordRequest struct {
    StudentID      string `json:"student_id" binding:"required"`
    Token          string `json:"token" binding:"required"`
    NewPassword    string `json:"new_password" binding:"required,min=6"`
    ConfirmPassword string `json:"confirm_password" binding:"required"`
}