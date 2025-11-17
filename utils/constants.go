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
    
    // Error messages - Password Reset
    ErrEmailRequired                = "Email is required"
    ErrUserNotFoundForEmail         = "No user found with this email"
    ErrResetTokenInvalid            = "Invalid or expired reset token"
    ErrResetTokenExpired            = "Reset token has expired"
    ErrPasswordRequired             = "New password is required"
    ErrPasswordTooShort             = "Password must be at least 6 characters"
    ErrResetAttemptsExceeded        = "Too many reset attempts. Please try again later."
    ErrFailedToProcessReset         = "Failed to process password reset request"
    ErrFailedToResetPassword        = "Failed to reset password"
    
    // Error messages - QR Code
    ErrQRCodeTypeExists             = "QR code type already exists"
    ErrInvalidQRCodeType            = "Invalid QR code type"
    ErrFailedToFetchStudents        = "Failed to fetch students"
    ErrFailedToFetchQRTypes         = "Failed to fetch QR code types"
    ErrFailedToFetchEvents          = "Failed to fetch events"
    ErrCourseYearLevelRequired      = "Course and year_level parameters are required"
    
    // Error messages - Events
    ErrEventNotFound                = "Event not found"
    ErrEventTimeInvalid             = "Event time is invalid"
    ErrEventFieldsRequired          = "Event name, type, and description are required"
    ErrEventStartTimePast           = "Start time cannot be in the past"
    ErrEventStartAfterEnd           = "Start time cannot be after end time"
    ErrFailedToCreateEvent          = "Failed to create event"
    ErrFailedToUpdateEvent          = "Failed to update event"
    ErrFailedToDeleteEvent          = "Failed to delete event"

    
    // Success messages
    MsgResetEmailSent               = "Password reset email sent successfully"
    MsgPasswordResetSuccess         = "Password reset successfully"
    MsgResetTokenValid              = "Reset token is valid"
    MsgEventCreated                 = "Event created successfully"
    MsgEventUpdated                 = "Event updated successfully"
    MsgEventDeleted                 = "Event deleted successfully"
    
    // Log messages
    LogEventCreated                 = "Event created successfully: %s by %s"
    LogEventUpdated                 = "Event updated successfully: %s"
    LogEventDeleted                 = "Event deleted successfully: %s"
    LogFailedToCreateEvent          = "Failed to create event: %v"
    LogFailedToUpdateEvent          = "Failed to update event: %v"
    LogFailedToDeleteEvent          = "Failed to delete event: %v"
    LogFailedToFetchEvents          = "Failed to fetch events: %v"
    
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
    QueryResetTokenWhere   = "reset_token = ?"
    
    // Combined queries
    QueryTypeNameAndActive = "type_name = ? AND is_active = ?"
    QueryCourseYearRole    = "course = ? AND year_level = ? AND role = ?"
    QueryActiveAndEndTime  = "is_active = ? AND end_time > ?"
    QueryResetTokenValid   = "reset_token = ? AND reset_token_expiry > ?"



    QueryEventIDWhere = "id = ?"
)