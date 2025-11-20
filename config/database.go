package config

import (
    "fmt"
    "log"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/joho/godotenv"
)

var DB *gorm.DB 

func ConnectDB() {
    err := godotenv.Load()
    if err != nil {
        log.Println("Warning: .env file not found")
    }

    dsn := fmt.Sprintf(
        "host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Manila",
        getEnv("DB_HOST", "localhost"),
        getEnv("DB_USER", "postgres"),
        getEnv("DB_PASSWORD", "123456"),
        getEnv("DB_NAME", "Attendance"),
        getEnv("DB_PORT", "5432"),
    )

    DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

  
    manualMigrate()

    // Auto-create superadmin account ONLY
    createSuperAdmin()

    log.Println("Database connected successfully")
}