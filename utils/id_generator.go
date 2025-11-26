// utils/id_generator.go

package utils

import (
    "crypto/rand"  
    "fmt"
    "strings"
    "time"
)

func GenerateCustomStudentID(dbID uint) string {
    year := time.Now().Format("2006")
    return fmt.Sprintf("S%s-%04d", year, dbID)
}

func GenerateCustomStudentIDWithCourse(dbID uint, course string) string {
    year := time.Now().Format("2006")
    
    // Create course prefix (first 2-3 letters of course)
    coursePrefix := "S" // default
    if len(course) >= 2 {
        // Extract first letters from course name
        words := strings.Fields(course)
        if len(words) > 0 {
            firstWord := words[0]
            if strings.HasPrefix(strings.ToUpper(firstWord), "BS") && len(words) > 1 {
                // For "BS Computer Science", use "CSC"
                coursePrefix = strings.ToUpper(firstWord[2:4] + words[1][:1])
            } else {
                coursePrefix = strings.ToUpper(firstWord[:2])
            }
        }
    }
    
    return fmt.Sprintf("%s%s-%04d", coursePrefix, year, dbID)
}

func generateRandomString(length int) string {
    const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    bytes := make([]byte, length)
    
    if _, err := rand.Read(bytes); err != nil {
        return fmt.Sprintf("%d", time.Now().UnixNano())[:length]
    }
    
    for i, b := range bytes {
        bytes[i] = chars[b%byte(len(chars))]
    }
    
    return strings.ToUpper(string(bytes))
}
