// package config

// import (
//     "fmt"
//     "log"
//     "AttendanceManagementSystem/models"
//     "os"
//     "time"

//     "github.com/joho/godotenv"
//     "golang.org/x/crypto/bcrypt"
//     "gorm.io/driver/postgres"
//     "gorm.io/gorm"
// )

// var DB *gorm.DB

// // Define constants for commonly used query conditions
// const (
//     roleCondition = "role = ?"
// )

// // Define constants for common column types
// const (
//     varchar100     = "VARCHAR(100)"
//     varchar100NotNull = "VARCHAR(100) NOT NULL DEFAULT ''"
//     varchar50      = "VARCHAR(50)"
//     varchar20      = "VARCHAR(20)"
//     textType       = "TEXT"
//     timestampType  = "TIMESTAMP"
// )

// func ConnectDB() {
//     err := godotenv.Load()
//     if err != nil {
//         log.Println("Warning: .env file not found")
//     }

//     dsn := fmt.Sprintf(
//         "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Manila",
//         getEnv("DB_HOST", "localhost"),
//         getEnv("DB_USER", "postgres"),
//         getEnv("DB_PASSWORD", "123456"),
//         getEnv("DB_NAME", "Attendance"),
//         getEnv("DB_PORT", "5432"),
//     )

//     DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
//     if err != nil {
//         log.Fatal("Failed to connect to database:", err)
//     }

//     // Use manual migration instead of AutoMigrate to avoid constraint errors
//     manualMigrate()

//     // Auto-create superadmin account ONLY
//     createSuperAdmin()

//     log.Println("Database connected successfully")
// }

// func manualMigrate() {
//     log.Println("Starting manual migration...")
    
//     // Check if users table exists
//     if !DB.Migrator().HasTable(&models.User{}) {
//         log.Println("Creating users table...")
//         err := DB.AutoMigrate(&models.User{}, &models.TempUser{}, &models.QRCodeType{}, &models.QRCodeEvent{}, &models.QRCodeScan{})
//         if err != nil {
//             log.Fatal("Failed to create tables:", err)
//         }
//         log.Println("All tables created successfully")
        
//         // Create default QR code types
//         createDefaultQRCodeTypes()
//     } else {
//         log.Println("Users table already exists, adding missing columns...")
//         addMissingColumns()
        
//         // Create new tables if they don't exist
//         if !DB.Migrator().HasTable(&models.TempUser{}) {
//             DB.AutoMigrate(&models.TempUser{})
//             log.Println("TempUser table created")
//         }
//         if !DB.Migrator().HasTable(&models.QRCodeType{}) {
//             DB.AutoMigrate(&models.QRCodeType{}, &models.QRCodeEvent{}, &models.QRCodeScan{})
//             createDefaultQRCodeTypes()
//         }
//     }
    
//     // Clean up expired temporary users
//     cleanUpExpiredTempUsers()
// }

// func cleanUpExpiredTempUsers() {
//     result := DB.Where("expires_at < ?", time.Now()).Delete(&models.TempUser{})
//     if result.Error == nil && result.RowsAffected > 0 {
//         log.Printf("Cleaned up %d expired temporary users", result.RowsAffected)
//     }
// }

// func createDefaultQRCodeTypes() {
//     defaultTypes := []models.QRCodeType{
//         {
//             TypeName:    "student_id",
//             Description: "Student Identification QR Code",
//             CreatedBy:   "system",
//             IsActive:    true,
//             CreatedAt:   time.Now(),
//             UpdatedAt:   time.Now(),
//         },
//         {
//             TypeName:    "attendance",
//             Description: "Attendance Tracking QR Code",
//             CreatedBy:   "system",
//             IsActive:    true,
//             CreatedAt:   time.Now(),
//             UpdatedAt:   time.Now(),
//         },
//         {
//             TypeName:    "library",
//             Description: "Library Access QR Code",
//             CreatedBy:   "system",
//             IsActive:    true,
//             CreatedAt:   time.Now(),
//             UpdatedAt:   time.Now(),
//         },
//         {
//             TypeName:    "event",
//             Description: "Event Participation QR Code",
//             CreatedBy:   "system",
//             IsActive:    true,
//             CreatedAt:   time.Now(),
//             UpdatedAt:   time.Now(),
//         },
//         {
//             TypeName:    "business",
//             Description: "Business Purpose QR Code",
//             CreatedBy:   "system",
//             IsActive:    true,
//             CreatedAt:   time.Now(),
//             UpdatedAt:   time.Now(),
//         },
//         {
//             TypeName:    "activity",
//             Description: "Activity Participation QR Code",
//             CreatedBy:   "system",
//             IsActive:    true,
//             CreatedAt:   time.Now(),
//             UpdatedAt:   time.Now(),
//         },
//     }

//     for _, qrType := range defaultTypes {
//         var existingType models.QRCodeType
//         if err := DB.Where("type_name = ?", qrType.TypeName).First(&existingType).Error; err != nil {
//             DB.Create(&qrType)
//             log.Printf("Created default QR code type: %s", qrType.TypeName)
//         }
//     }
// }

// func addMissingColumns() {
//     // List of columns to check and add
//     columnsToAdd := map[string]string{
//         "role":           "VARCHAR(50) DEFAULT 'student'",
//         "student_number": varchar100,
//         "first_name":     varchar100NotNull,
//         "last_name":      varchar100NotNull,
//         "middle_name":    varchar100,
//         "course":         varchar100,
//         "year_level":     varchar50,
//         "section":        varchar50,
//         "department":     varchar100,
//         "college":        varchar100,
//         "contact_number": varchar20,
//         "address":        textType,
//         "qr_code_data":   textType,
//         "qr_code_type":   "VARCHAR(50) DEFAULT 'student_id'",
//         "verified_at":    timestampType,
        
//         // ADD THESE PASSWORD RESET COLUMNS
//         "reset_token":         "VARCHAR(64)",
//         "reset_token_expiry":  "TIMESTAMP",
//         "reset_attempts":      "INTEGER DEFAULT 0",
//         "last_reset_request":  "TIMESTAMP",
//     }

//     for columnName, columnType := range columnsToAdd {
//         if !DB.Migrator().HasColumn(&models.User{}, columnName) {
//             log.Printf("Adding column: %s", columnName)
//             sql := fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", columnName, columnType)
//             if err := DB.Exec(sql).Error; err != nil {
//                 log.Printf("Warning: Failed to add column %s: %v", columnName, err)
//             } else {
//                 log.Printf("Column %s added successfully", columnName)
//             }
//         } else {
//             log.Printf("Column %s already exists", columnName)
//         }
//     }

//     // Update existing records with default values
//     updateResults := DB.Exec("UPDATE users SET first_name = username WHERE first_name = '' OR first_name IS NULL")
//     if updateResults.Error == nil {
//         log.Printf("Updated %d records for first_name", updateResults.RowsAffected)
//     }

//     updateResults = DB.Exec("UPDATE users SET last_name = 'User' WHERE last_name = '' OR last_name IS NULL")
//     if updateResults.Error == nil {
//         log.Printf("Updated %d records for last_name", updateResults.RowsAffected)
//     }

//     updateResults = DB.Exec("UPDATE users SET role = 'student' WHERE role IS NULL OR role = ''")
//     if updateResults.Error == nil {
//         log.Printf("Updated %d records for role", updateResults.RowsAffected)
//     }

//     // Set default QR code type for existing users
//     updateResults = DB.Exec("UPDATE users SET qr_code_type = 'student_id' WHERE qr_code_type IS NULL OR qr_code_type = ''")
//     if updateResults.Error == nil {
//         log.Printf("Updated %d records for qr_code_type", updateResults.RowsAffected)
//     }

//     // Initialize reset_attempts for existing users
//     updateResults = DB.Exec("UPDATE users SET reset_attempts = 0 WHERE reset_attempts IS NULL")
//     if updateResults.Error == nil && updateResults.RowsAffected > 0 {
//         log.Printf("Initialized reset_attempts for %d users", updateResults.RowsAffected)
//     }

//     log.Println("Manual migration completed successfully")
// }

// func createSuperAdmin() {
//     var superAdmin models.User
//     result := DB.Where("email = ?", "superadmin@system.com").First(&superAdmin)
    
//     if result.Error != nil {
//         // Hash password
//         hash, _ := bcrypt.GenerateFromPassword([]byte("superadmin123"), 14)
        
//         superAdmin := models.User{
//             UserID:        "U2025-0000",
//             Email:         "superadmin@system.com",
//             Password:      string(hash),
//             Username:      "superadmin",
//             Role:          "superadmin",
//             IsVerified:    true,
//             FirstName:     "Super",
//             LastName:      "Admin",
//             StudentNumber: "ADMIN-001",
//             Course:        "System Administration",
//             YearLevel:     "N/A",
//             Department:    "System Administration",
//             College:       "Administration",
//             ContactNumber: "+639000000000",
//             Address:       "Main Administration Building",
//             CreatedAt:     time.Now(),
//             VerifiedAt:    time.Now(),
//         }
        
//         if err := DB.Create(&superAdmin).Error; err != nil {
//             log.Printf("Warning: Failed to create superadmin: %v", err)
//         } else {
//             log.Println(" Superadmin account created successfully!")
//             log.Println(" Email: superadmin@system.com")
//             log.Println(" Password: superadmin123")
//             log.Println(" User ID: U2025-0000")
//             log.Println(" Role: superadmin")
//         }
//     } else {
//         log.Println(" Superadmin account already exists")
//         log.Println(" Email: superadmin@system.com")
//         log.Println(" Password: superadmin123")
//     }
// }


// func getEnv(key, defaultValue string) string {
//     if value := os.Getenv(key); value != "" {
//         return value
//     }
//     return defaultValue
// }

// // Function to display current table structure
// func DisplayTableStructure() {
//     type ColumnInfo struct {
//         ColumnName string `gorm:"column:column_name"`
//         DataType   string `gorm:"column:data_type"`
//         IsNullable string `gorm:"column:is_nullable"`
//     }
    
//     var columns []ColumnInfo
//     result := DB.Raw(`
//         SELECT column_name, data_type, is_nullable 
//         FROM information_schema.columns 
//         WHERE table_name = 'users' 
//         ORDER BY ordinal_position
//     `).Scan(&columns)
    
//     if result.Error != nil {
//         log.Printf("Error getting table structure: %v", result.Error)
//         return
//     }
    
//     log.Println(" Current users table structure:")
//     log.Println("Column Name | Data Type | Nullable")
//     log.Println("----------------------------------")
//     for _, col := range columns {
//         log.Printf("%s | %s | %s", col.ColumnName, col.DataType, col.IsNullable)
//     }
// }

// // Function to display user statistics
// func DisplayUserStats() {
//     var stats struct {
//         TotalUsers    int64
//         SuperAdmins   int64
//         Admins        int64
//         Students      int64
//         VerifiedUsers int64
//         UsersWithQR   int64
//     }

//     DB.Model(&models.User{}).Count(&stats.TotalUsers)
//     DB.Model(&models.User{}).Where(roleCondition, "superadmin").Count(&stats.SuperAdmins)
//     DB.Model(&models.User{}).Where(roleCondition, "admin").Count(&stats.Admins)
//     DB.Model(&models.User{}).Where(roleCondition, "student").Count(&stats.Students)
//     DB.Model(&models.User{}).Where("is_verified = ?", true).Count(&stats.VerifiedUsers)
//     DB.Model(&models.User{}).Where("qr_code_data IS NOT NULL AND qr_code_data != ''").Count(&stats.UsersWithQR)

//     log.Println(" User Statistics:")
//     log.Printf("   Total Users: %d", stats.TotalUsers)
//     log.Printf("   Super Admins: %d", stats.SuperAdmins)
//     log.Printf("   Admins: %d", stats.Admins)
//     log.Printf("   Students: %d", stats.Students)
//     log.Printf("   Verified Users: %d", stats.VerifiedUsers)
//     log.Printf("   Users with QR Codes: %d", stats.UsersWithQR)
// }




package config

import "gorm.io/gorm"


func GetDB() *gorm.DB {
    return DB
}


func InitDatabase() {
    ConnectDB()
}

func ShowTableStructure() {
    DisplayTableStructure()
}

func ShowUserStats() {
    DisplayUserStats()
}