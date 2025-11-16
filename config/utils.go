package config

import (

    "log"
    "os"
    "AttendanceManagementSystem/models"
)

// Define constants for commonly used query conditions
const (
    roleCondition = "role = ?"
)

// Define constants for common column types
const (
    varchar100     = "VARCHAR(100)"
    varchar100NotNull = "VARCHAR(100) NOT NULL DEFAULT ''"
    varchar50      = "VARCHAR(50)"
    varchar20      = "VARCHAR(20)"
    textType       = "TEXT"
    timestampType  = "TIMESTAMP"
)

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// Function to display current table structure
func DisplayTableStructure() {
    type ColumnInfo struct {
        ColumnName string `gorm:"column:column_name"`
        DataType   string `gorm:"column:data_type"`
        IsNullable string `gorm:"column:is_nullable"`
    }
    
    var columns []ColumnInfo
    result := DB.Raw(`
        SELECT column_name, data_type, is_nullable 
        FROM information_schema.columns 
        WHERE table_name = 'users' 
        ORDER BY ordinal_position
    `).Scan(&columns)
    
    if result.Error != nil {
        log.Printf("Error getting table structure: %v", result.Error)
        return
    }
    
    log.Println(" Current users table structure:")
    log.Println("Column Name | Data Type | Nullable")
    log.Println("----------------------------------")
    for _, col := range columns {
        log.Printf("%s | %s | %s", col.ColumnName, col.DataType, col.IsNullable)
    }
}

// Function to display user statistics
func DisplayUserStats() {
    var stats struct {
        TotalUsers    int64
        SuperAdmins   int64
        Admins        int64
        Students      int64
        VerifiedUsers int64
        UsersWithQR   int64
    }

    DB.Model(&models.User{}).Count(&stats.TotalUsers)
    DB.Model(&models.User{}).Where(roleCondition, "superadmin").Count(&stats.SuperAdmins)
    DB.Model(&models.User{}).Where(roleCondition, "admin").Count(&stats.Admins)
    DB.Model(&models.User{}).Where(roleCondition, "student").Count(&stats.Students)
    DB.Model(&models.User{}).Where("is_verified = ?", true).Count(&stats.VerifiedUsers)
    DB.Model(&models.User{}).Where("qr_code_data IS NOT NULL AND qr_code_data != ''").Count(&stats.UsersWithQR)

    log.Println(" User Statistics:")
    log.Printf("   Total Users: %d", stats.TotalUsers)
    log.Printf("   Super Admins: %d", stats.SuperAdmins)
    log.Printf("   Admins: %d", stats.Admins)
    log.Printf("   Students: %d", stats.Students)
    log.Printf("   Verified Users: %d", stats.VerifiedUsers)
    log.Printf("   Users with QR Codes: %d", stats.UsersWithQR)
}