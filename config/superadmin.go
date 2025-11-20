package config

import (
    "log"
    "time"
    "AttendanceManagementSystem/models"
    "golang.org/x/crypto/bcrypt"
)

func createSuperAdmin() {
    var superAdmin models.User
    result := DB.Where("email = ?", "superadmin@system.com").First(&superAdmin)
    
    if result.Error != nil {
    
        hash, _ := bcrypt.GenerateFromPassword([]byte("superadmin123"), 14)
        
        superAdmin := models.User{
            UserID:        "U2025-0000",
            Email:         "superadmin@system.com",
            Password:      string(hash),
            Username:      "superadmin",
            Role:          "superadmin",
            IsVerified:    true,
            FirstName:     "Super",
            LastName:      "Admin",
            StudentNumber: "ADMIN-001",
            Course:        "System Administration",
            YearLevel:     "N/A",
            Department:    "System Administration",
            College:       "Administration",
            ContactNumber: "+639000000000",
            Address:       "Main Administration Building",
            CreatedAt:     time.Now(),
            VerifiedAt:    time.Now(),
        }
        
        if err := DB.Create(&superAdmin).Error; err != nil {
            log.Printf("Warning: Failed to create superadmin: %v", err)
        } else {
            log.Println(" Superadmin account created successfully!")
            log.Println(" Email: superadmin@system.com")
            log.Println(" Password: superadmin123")
            log.Println(" User ID: U2025-0000")
            log.Println(" Role: superadmin")
        }
    } else {
        log.Println(" Superadmin account already exists")
        log.Println(" Email: superadmin@system.com")
        log.Println(" Password: superadmin123")
    }
}