package utils

// Constants for error messages, headers and queries
const (
    // Error messages - Auth
    ErrCannotParseJSON              = "Cannot parse JSON"
    ErrInvalidEmailPass             = "Invalid email or password"
    ErrUserNotFound                 = "User not found"
    ErrUserExists                   = "User already exists"
    ErrEmailPassRequired            = "Email, password, and username are required"
    ErrHashPassword                 = "Failed to hash password"
    ErrInvalidVerifyCode            = "Invalid verification code"
    ErrVerifyEmailFirst             = "Please verify your email first"
    ErrTempUserNotFound             = "Registration not found or expired"
    ErrVerificationExpired          = "Verification code has expired"
    ErrAllFieldsRequired            = "All required fields must be filled"
    ErrEmailPendingVerification     = "Email already registered and pending verification. Please check your email or wait for the verification to expire."
    ErrInvalidVerificationCodeOrEmail = "Invalid verification code or email"
    ErrFailedProcessRegistration    = "Failed to process registration"
    ErrFailedCreateUserAccount      = "Failed to create user account"
    ErrRegistrationExpired          = "Registration has expired. Please register again."
    ErrNoPendingRegistration        = "No pending registration found for this email"
    ErrFailedResendVerification     = "Failed to resend verification code"
    ErrAdminAccessRequired          = "Admin access required"
    ErrFailedCleanupRegistrations   = "Failed to clean up expired registrations"
    
    // Error messages - QR Code
    ErrQRCodeTypeExists             = "QR code type already exists"
    ErrInvalidQRCodeType            = "Invalid QR code type"
    ErrFailedToFetchStudents        = "Failed to fetch students"
    ErrFailedToFetchQRTypes         = "Failed to fetch QR code types"
    ErrFailedToFetchEvents          = "Failed to fetch events"
    ErrCourseYearLevelRequired      = "Course and year_level parameters are required"
    
    // Header names
    HeaderUserRole = "X-User-Role"
    HeaderUserID   = "X-User-ID"
    
    // Query constants
    QueryEmailWhere        = "email = ?"
    QueryUserIDWhere       = "user_id = ?"
    QueryVerificationCode  = "verification_code = ?"
    QueryTypeNameWhere     = "type_name = ?"
    QueryIsActive          = "is_active = ?"
    QueryCourseWhere       = "course = ?"
    QueryYearLevelWhere    = "year_level = ?"
    QueryRoleWhere         = "role = ?"
    
    // Combined queries
    QueryTypeNameAndActive = "type_name = ? AND is_active = ?"
    QueryCourseYearRole    = "course = ? AND year_level = ? AND role = ?"
    QueryActiveAndEndTime  = "is_active = ? AND end_time > ?"
)