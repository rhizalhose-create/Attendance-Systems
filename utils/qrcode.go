package utils

import (
    "encoding/base64"
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/skip2/go-qrcode"
)

// GenerateStudentQRCode generates QR code for student registration
func GenerateStudentQRCode(userID, email, firstName, lastName, course string) (string, error) {
    qrData := fmt.Sprintf("STUDENT_ID|%s|%s|%s|%s|%s|%d", 
        userID, email, firstName, lastName, course, time.Now().Unix())

    return generateQRCodeBase64(qrData)
}

// GenerateEventQRCode generates QR code for specific event
func GenerateEventQRCode(userID, eventType, eventName string, eventID uint) (string, error) {
    qrData := fmt.Sprintf("EVENT|%s|%s|%s|%d|%d", 
        userID, eventType, eventName, eventID, time.Now().Unix())

    return generateQRCodeBase64(qrData)
}

// GenerateCustomQRCode generates QR code with custom type
func GenerateCustomQRCode(userID, qrCodeType, customData string) (string, error) {
    qrData := fmt.Sprintf("%s|%s|%s|%d", 
        qrCodeType, userID, customData, time.Now().Unix())

    return generateQRCodeBase64(qrData)
}

// GenerateQRCodeFile generates QR code and saves as file
func GenerateQRCodeFile(userID, qrCodeType, data string) (string, error) {
    qrData := fmt.Sprintf("%s|%s|%s|%d", qrCodeType, userID, data, time.Now().Unix())

    // Create directory if it doesn't exist
    qrDir := "qrcodes"
    if err := os.MkdirAll(qrDir, 0755); err != nil {
        return "", fmt.Errorf("failed to create QR code directory: %v", err)
    }

    // Generate file path
    fileName := fmt.Sprintf("qr_%s_%s.png", userID, qrCodeType)
    filePath := filepath.Join(qrDir, fileName)

    // Generate and save QR code
    err := qrcode.WriteFile(qrData, qrcode.Medium, 256, filePath)
    if err != nil {
        return "", fmt.Errorf("failed to generate QR code: %v", err)
    }

    return filePath, nil
}

// Helper function to generate base64 QR code
func generateQRCodeBase64(data string) (string, error) {
    qr, err := qrcode.New(data, qrcode.Medium)
    if err != nil {
        return "", fmt.Errorf("failed to generate QR code: %v", err)
    }

    // Convert to PNG bytes
    pngData, err := qr.PNG(256)
    if err != nil {
        return "", fmt.Errorf("failed to generate PNG data: %v", err)
    }

    // Convert to base64
    base64Str := base64.StdEncoding.EncodeToString(pngData)
    return "data:image/png;base64," + base64Str, nil
}


func ParseQRCodeData(qrData string) (map[string]string, error) {
  

    result := make(map[string]string)
    parts := splitQRData(qrData)

    if len(parts) < 3 {
        return nil, fmt.Errorf("invalid QR code data format")
    }

    result["type"] = parts[0]
    result["user_id"] = parts[1]
    result["timestamp"] = parts[len(parts)-1] 

   
    switch result["type"] {

 
    case "STUDENT_ID":
        if len(parts) >= 6 {
            result["email"] = parts[2]
            result["first_name"] = parts[3]
            result["last_name"] = parts[4]
            result["course"] = parts[5]
        }

   
    case "EVENT":
        if len(parts) >= 5 {
            result["event_type"] = parts[2]
            result["event_name"] = parts[3]
            result["event_id"] = parts[4]
        }


    case "ACTIVITY":
        if len(parts) >= 5 {
            result["activity_id"] = parts[2]
            result["activity_name"] = parts[3]
            result["activity_description"] = parts[4]
        }


    case "STUDENT_EVENT":
        if len(parts) >= 5 {
            result["student_id"] = parts[2]
            result["event_id"] = parts[3]
            result["action"] = parts[4] 
        }


    default:
        if len(parts) > 3 {
            result["custom_data"] = parts[2]
            for i := 3; i < len(parts)-1; i++ {
                result["custom_data"] += "|" + parts[i]
            }
        }
    }

    return result, nil
}

// Helper function to split QR data safely
func splitQRData(data string) []string {
    var parts []string
    start := 0
    for i, char := range data {
        if char == '|' {
            parts = append(parts, data[start:i])
            start = i + 1
        }
    }
    // Add the last part
    if start < len(data) {
        parts = append(parts, data[start:])
    }
    return parts
}