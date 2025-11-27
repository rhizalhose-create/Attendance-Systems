package models

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type TempUser struct {
	ID               uint      `json:"-" gorm:"primaryKey;autoIncrement"`
	Email            string    `json:"email" gorm:"uniqueIndex;not null;type:varchar(255)"`
	Password         string    `json:"password" gorm:"not null;type:varchar(255)"`
	Username         string    `json:"username" gorm:"not null;type:varchar(255)"`
	StudentID        string    `json:"student_id" gorm:"not null;uniqueIndex"`
	FirstName        string    `json:"first_name" gorm:"not null;type:varchar(100)"`
	LastName         string    `json:"last_name" gorm:"not null;type:varchar(100)"`
	MiddleName       string    `json:"middle_name,omitempty" gorm:"type:varchar(100)"`
	Course           string    `json:"course" gorm:"type:varchar(100)"`
	YearLevel        string    `json:"year_level" gorm:"type:varchar(50)"`
	Section          string    `json:"section,omitempty" gorm:"type:varchar(50)"`
	Department       string    `json:"department,omitempty" gorm:"type:varchar(100)"`
	College          string    `json:"college,omitempty" gorm:"type:varchar(100)"`
	ContactNumber    string    `json:"contact_number,omitempty" gorm:"type:varchar(20)"`
	Address          string    `json:"address,omitempty" gorm:"type:text"`
	VerificationCode string    `json:"-" gorm:"type:varchar(10)"`
	ExpiresAt        time.Time `json:"expires_at"`
	CreatedAt        time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (TempUser) TableName() string {
	return "temp_users"
}

func (t *TempUser) Verify(db *gorm.DB, code string) error {

	log.Println("🔍 VERIFY START")
	log.Println("TempUser ID:", t.ID)
	log.Println("Stored Code:", t.VerificationCode)
	log.Println("Input Code:", code)
	log.Println("Expires At:", t.ExpiresAt)
	log.Println("Current Time:", time.Now())

	// 1. Check verification code
	if t.VerificationCode != code {
		log.Println("❌ CODE MISMATCH")
		return fmt.Errorf("Invalid verification code")
	}

	// 2. Check if expired
	if time.Now().After(t.ExpiresAt) {
		log.Println("❌ CODE EXPIRED")
		return fmt.Errorf("Verification code expired")
	}

	log.Println("✅ CODE VERIFIED - Creating User...")

	// 3. Convert TempUser → User
	user := User{
		StudentID:     t.StudentID,
		Email:         t.Email,
		Username:      t.Username,
		Password:      t.Password,
		FirstName:     t.FirstName,
		LastName:      t.LastName,
		MiddleName:    t.MiddleName,
		Course:        t.Course,
		YearLevel:     t.YearLevel,
		Section:       t.Section,
		Department:    t.Department,
		College:       t.College,
		ContactNumber: t.ContactNumber,
		Address:       t.Address,
		IsVerified:    true,
		CreatedAt:     time.Now(),
	}

	// 4. Save user
	if err := db.Create(&user).Error; err != nil {
		log.Println("❌ FAILED TO CREATE USER:", err)
		return fmt.Errorf("Failed to create user: %v", err)
	}

	log.Println("🗑️ DELETING TEMP USER:", t.ID)
	// 5. Delete temp user
	if err := db.Delete(&TempUser{}, t.ID).Error; err != nil {
		log.Println("❌ FAILED DELETE:", err)
		return fmt.Errorf("Failed to delete temp user: %v", err)
	}

	log.Println("🎉 TEMP USER DELETED, VERIFIED SUCCESSFULLY")
	return nil
}
