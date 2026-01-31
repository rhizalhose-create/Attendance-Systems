//utils/conttants.gp

package utils

// Constants for error messages, headers and queries
const (
    // Error messages - Auth
    ErrCannotParseJSON              = "Cannot parse JSON"
    ErrInvalidStudentIDPass         = "Invalid student ID or password"
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
    
    // Error messages - Password Reset (UPDATED FOR STUDENT ID FLOW)
    ErrEmailRequired                = "Email is required"
    ErrStudentIDRequired            = "Student ID is required" // NEW
    ErrUserNotFoundForEmail         = "No user found with this email"
    ErrUserNotFoundForStudentID     = "No user found with this Student ID" // NEW
    ErrResetTokenInvalid            = "Invalid or expired reset token"
    ErrResetTokenExpired            = "Reset token has expired"
    ErrPasswordRequired             = "New password is required"
    ErrConfirmPasswordRequired      = "Confirm password is required" // NEW
    ErrPasswordsDoNotMatch          = "Passwords do not match" // NEW
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

    
    // Success messages (UPDATED FOR STUDENT ID FLOW)
    MsgResetEmailSent               = "Password reset email sent successfully"
    MsgResetLinkSent                = "Password reset link sent successfully" // NEW
    MsgPasswordResetSuccess         = "Password reset successfully"
    MsgResetTokenValid              = "Reset token is valid"
    MsgEventCreated                 = "Event created successfully"
    MsgEventUpdated                 = "Event updated successfully"
    MsgEventDeleted                 = "Event deleted successfully"
    
    // Log messages (UPDATED FOR STUDENT ID FLOW)
    LogEventCreated                 = "Event created successfully: %s by %s"
    LogEventUpdated                 = "Event updated successfully: %s"
    LogEventDeleted                 = "Event deleted successfully: %s"
    LogFailedToCreateEvent          = "Failed to create event: %v"
    LogFailedToUpdateEvent          = "Failed to update event: %v"
    LogFailedToDeleteEvent          = "Failed to delete event: %v"
    LogFailedToFetchEvents          = "Failed to fetch events: %v"
    LogResetRequestedStudentID      = "Password reset requested for Student ID: %s - Email: %s - Token: %s" // NEW
    LogResetSuccessfulStudentID     = "Password reset successful for Student ID: %s" // NEW
    LogInvalidResetTokenStudentID   = "Invalid or expired reset token for Student ID %s: %s" // NEW
    
    // Header names
    HeaderUserRole   = "X-User-Role"
    HeaderStudentID  = "X-Student-ID"
    
    // Query constants (UPDATED FOR STUDENT ID FLOW)
    QueryEmailWhere        = "email = ?"
    QueryStudentIDWhere    = "student_id = ?"
    QueryVerificationCode  = "verification_code = ?"
    QueryTypeNameWhere     = "type_name = ?"
    QueryIsActive          = "is_active = ?"
    QueryCourseWhere       = "course = ?"
    QueryYearLevelWhere    = "year_level = ?"
    QueryRoleWhere         = "role = ?"
    QueryResetTokenWhere   = "reset_token = ?"
    
    // Combined queries (UPDATED FOR STUDENT ID FLOW)
    QueryTypeNameAndActive = "type_name = ? AND is_active = ?"
    QueryCourseYearRole    = "course = ? AND year_level = ? AND role = ?"
    QueryActiveAndEndTime  = "is_active = ? AND end_time > ?"
    QueryResetTokenValid   = "reset_token = ? AND reset_token_expiry > ?"
    QueryEventIDWhere      = "id = ?"
    QueryStudentIDResetToken = "student_id = ? AND reset_token = ? AND reset_token_expiry > ?" // NEW
)