package config

import (
    "fmt"
    "log"
    "time"
    "AttendanceManagementSystem/models"
)

func manualMigrate() {
    log.Println("Starting manual migration...")
    
    // Check if users table exists
    if !DB.Migrator().HasTable(&models.User{}) {
        log.Println("Creating users table...")
        err := DB.AutoMigrate(&models.User{}, &models.TempUser{}, &models.QRCodeType{}, &models.QRCodeEvent{}, &models.QRCodeScan{})
        if err != nil {
            log.Fatal("Failed to create tables:", err)
        }
        log.Println("All tables created successfully")
        
        // Create default QR code types
        createDefaultQRCodeTypes()
    } else {
        log.Println("Users table already exists, adding missing columns...")
        addMissingColumns()
        
        // Create new tables if they don't exist
        if !DB.Migrator().HasTable(&models.TempUser{}) {
            DB.AutoMigrate(&models.TempUser{})
            log.Println("TempUser table created")
        }
        if !DB.Migrator().HasTable(&models.QRCodeType{}) {
            DB.AutoMigrate(&models.QRCodeType{}, &models.QRCodeEvent{}, &models.QRCodeScan{})
            createDefaultQRCodeTypes()
        }
    }
    
    // Clean up expired temporary users
    cleanUpExpiredTempUsers()
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
        var existingType models.QRCodeType
        if err := DB.Where("type_name = ?", qrType.TypeName).First(&existingType).Error; err != nil {
            DB.Create(&qrType)
            log.Printf("Created default QR code type: %s", qrType.TypeName)
        }
    }
}

func addMissingColumns() {
    // List of columns to check and add
    columnsToAdd := map[string]string{
        "role":           "VARCHAR(50) DEFAULT 'student'",
        "student_number": varchar100,
        "first_name":     varchar100NotNull,
        "last_name":      varchar100NotNull,
        "middle_name":    varchar100,
        "course":         varchar100,
        "year_level":     varchar50,
        "section":        varchar50,
        "department":     varchar100,
        "college":        varchar100,
        "contact_number": varchar20,
        "address":        textType,
        "qr_code_data":   textType,
        "qr_code_type":   "VARCHAR(50) DEFAULT 'student_id'",
        "verified_at":    timestampType,
        
        // ADD THESE PASSWORD RESET COLUMNS
        "reset_token":         "VARCHAR(64)",
        "reset_token_expiry":  "TIMESTAMP",
        "reset_attempts":      "INTEGER DEFAULT 0",
        "last_reset_request":  "TIMESTAMP",
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
        } else {
            log.Printf("Column %s already exists", columnName)
        }
    }

    // Update existing records with default values
    updateResults := DB.Exec("UPDATE users SET first_name = username WHERE first_name = '' OR first_name IS NULL")
    if updateResults.Error == nil {
        log.Printf("Updated %d records for first_name", updateResults.RowsAffected)
    }

    updateResults = DB.Exec("UPDATE users SET last_name = 'User' WHERE last_name = '' OR last_name IS NULL")
    if updateResults.Error == nil {
        log.Printf("Updated %d records for last_name", updateResults.RowsAffected)
    }

    updateResults = DB.Exec("UPDATE users SET role = 'student' WHERE role IS NULL OR role = ''")
    if updateResults.Error == nil {
        log.Printf("Updated %d records for role", updateResults.RowsAffected)
    }

    // Set default QR code type for existing users
    updateResults = DB.Exec("UPDATE users SET qr_code_type = 'student_id' WHERE qr_code_type IS NULL OR qr_code_type = ''")
    if updateResults.Error == nil {
        log.Printf("Updated %d records for qr_code_type", updateResults.RowsAffected)
    }

    // Initialize reset_attempts for existing users
    updateResults = DB.Exec("UPDATE users SET reset_attempts = 0 WHERE reset_attempts IS NULL")
    if updateResults.Error == nil && updateResults.RowsAffected > 0 {
        log.Printf("Initialized reset_attempts for %d users", updateResults.RowsAffected)
    }

    log.Println("Manual migration completed successfully")
}