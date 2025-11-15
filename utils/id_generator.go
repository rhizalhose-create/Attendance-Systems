package utils

import (
    "crypto/rand"  
    "fmt"
    "strings"
    "time"
)

func GenerateCustomUserID(dbID uint) string {
    year := time.Now().Format("2006")
    return fmt.Sprintf("U%s-%04d", year, dbID)
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