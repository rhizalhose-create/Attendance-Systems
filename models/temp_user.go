package models

import "time"

type TempUser struct {
    ID                uint      `json:"-" gorm:"primaryKey;autoIncrement"`
    Email             string    `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)"`
    Password          string    `json:"password" gorm:"not null;type:varchar(255)"`
    Username          string    `json:"username" gorm:"not null;type:varchar(255)"`
    UserID            string    `json:"user_id" gorm:"not null;uniqueIndex"`
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