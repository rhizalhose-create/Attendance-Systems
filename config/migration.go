package config

import (
	"fmt"
	"log"
	"time"
	"AttendanceManagementSystem/models"
)

func manualMigrate() error {
	log.Println("Starting manual migration...")

	if !DB.Migrator().HasTable(&models.User{}) {
		log.Println("Creating users table...")
		if err := DB.AutoMigrate(&models.User{}, &models.TempUser{}, &models.QRCodeType{}, &models.QRCodeEvent{}, &models.QRCodeScan{}); err != nil {
			return fmt.Errorf("failed to create tables: %v", err)
		}
		createDefaultQRCodeTypes()
	} else {
		addMissingColumns()
		if !DB.Migrator().HasTable(&models.TempUser{}) {
			if err := DB.AutoMigrate(&models.TempUser{}); err != nil {
				return err
			}
			log.Println("TempUser table created")
		}
		if !DB.Migrator().HasTable(&models.QRCodeType{}) {
			if err := DB.AutoMigrate(&models.QRCodeType{}, &models.QRCodeEvent{}, &models.QRCodeScan{}); err != nil {
				return err
			}
			createDefaultQRCodeTypes()
		}
	}

	cleanUpExpiredTempUsers()
	log.Println("Manual migration completed successfully")
	return nil
}

func cleanUpExpiredTempUsers() {
	result := DB.Where("expires_at < ?", time.Now()).Delete(&models.TempUser{})
	if result.Error == nil && result.RowsAffected > 0 {
		log.Printf("Cleaned up %d expired temporary users", result.RowsAffected)
	}
}

func createDefaultQRCodeTypes() {
	defaultTypes := []models.QRCodeType{
		{
			TypeName:    "student_id",
			Description: "Student Identification QR Code",
			CreatedBy:   "system",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			TypeName:    "attendance",
			Description: "Attendance Tracking QR Code",
			CreatedBy:   "system",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			TypeName:    "library",
			Description: "Library Access QR Code",
			CreatedBy:   "system",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			TypeName:    "event",
			Description: "Event Participation QR Code",
			CreatedBy:   "system",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			TypeName:    "business",
			Description: "Business Purpose QR Code",
			CreatedBy:   "system",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			TypeName:    "activity",
			Description: "Activity Participation QR Code",
			CreatedBy:   "system",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, qrType := range defaultTypes {
		var existing models.QRCodeType
		if err := DB.Where("type_name = ?", qrType.TypeName).First(&existing).Error; err != nil {
			DB.Create(&qrType)
			log.Printf("Created default QR code type: %s", qrType.TypeName)
		}
	}
}

func addMissingColumns() {
	columnsToAdd := map[string]string{
		"role":               "VARCHAR(50) DEFAULT 'student'",
		"student_number":     varchar100,
		"first_name":         varchar100NotNull,
		"last_name":          varchar100NotNull,
		"middle_name":        varchar100,
		"course":             varchar100,
		"year_level":         varchar50,
		"section":            varchar50,
		"department":         varchar100,
		"college":            varchar100,
		"contact_number":     varchar20,
		"address":            textType,
		"qr_code_data":       textType,
		"qr_code_type":       "VARCHAR(50) DEFAULT 'student_id'",
		"verified_at":        timestampType,
		"reset_token":        "VARCHAR(64)",
		"reset_token_expiry": "TIMESTAMP",
		"reset_attempts":     "INTEGER DEFAULT 0",
		"last_reset_request": "TIMESTAMP",
	}

	for columnName, columnType := range columnsToAdd {
		if !DB.Migrator().HasColumn(&models.User{}, columnName) {
			log.Printf("Adding column: %s", columnName)
			sql := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", columnName, columnType)
			if err := DB.Exec(sql).Error; err != nil {
				log.Printf("Warning: Failed to add column %s: %v", columnName, err)
			} else {
				log.Printf("Column %s added successfully", columnName)
			}
		}
	}

	// Set default values for existing users
	DB.Exec("UPDATE users SET first_name = username WHERE first_name = '' OR first_name IS NULL")
	DB.Exec("UPDATE users SET last_name = 'User' WHERE last_name = '' OR last_name IS NULL")
	DB.Exec("UPDATE users SET role = 'student' WHERE role IS NULL OR role = ''")
	DB.Exec("UPDATE users SET qr_code_type = 'student_id' WHERE qr_code_type IS NULL OR qr_code_type = ''")
	DB.Exec("UPDATE users SET reset_attempts = 0 WHERE reset_attempts IS NULL")
}
