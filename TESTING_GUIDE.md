# Testing Guide - Email Verification & Account Activation

## Quick Start Testing Flow

### Step 1: Register a New Account
```bash
curl -X POST http://localhost:9090/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "Test@12345",
    "username": "testuser",
    "first_name": "Test",
    "last_name": "User",
    "course": "BS Computer Science",
    "year_level": "1st Year",
    "section": "A",
    "department": "CICS",
    "college": "College of Science"
  }'
```

**Expected Response:**
```json
{
  "message": "Registration successful. Please check your email for verification code.",
  "email": "testuser@example.com",
  "expires_in": "24 hours",
  "note": "Verification code: 123456"
}
```

### Step 2: Verify Email with Code
Extract the verification code from the response/email, then:

```bash
curl -X POST http://localhost:9090/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "code": "123456"
  }'
```

**Expected Response:**
```json
{
  "message": "Email verified successfully! Your account has been activated.",
  "student_id": "STU2025-001",
  "email": "testuser@example.com",
  "note": "Student ID has been sent to your email. Use it to login."
}
```

### Step 3: Check Database Tables

#### Verify temp_user is DELETED
```sql
SELECT * FROM temp_users WHERE email = 'testuser@example.com';
-- Result: (no rows)
```

#### Verify user is CREATED in users table
```sql
SELECT student_id, email, is_verified, verified_at, qr_code_data 
FROM users 
WHERE email = 'testuser@example.com';
```

**Expected Result:**
| student_id  | email                 | is_verified | verified_at         | qr_code_data |
|-------------|----------------------|-------------|---------------------|--------------|
| STU2025-001 | testuser@example.com | 1/true      | 2025-12-03 12:00:00 | [QR data]    |

### Step 4: Check Email Inbox
- Look for email from `SMTP_EMAIL` (from your .env)
- **Subject**: "Your Student ID - Account Verified"
- **Content**: Should display student ID prominently
- **Action**: Student can now use this ID to login

### Step 5: Login with Student ID
```bash
curl -X POST http://localhost:9090/login \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "STU2025-001",
    "password": "Test@12345"
  }'
```

**Expected Response:**
```json
{
  "message": "Login successful",
  "student_id": "STU2025-001",
  "email": "testuser@example.com",
  "username": "testuser",
  "role": "student",
  "is_verified": true,
  "student_info": {
    "first_name": "Test",
    "last_name": "User",
    "course": "BS Computer Science",
    "year_level": "1st Year",
    "section": "A"
  },
  "qr_code_data": "[QR code]",
  "qr_code_type": "student_id"
}
```

## Error Testing

### Test 1: Invalid Verification Code
```bash
curl -X POST http://localhost:9090/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "code": "999999"
  }'
```

**Expected Response (HTTP 400):**
```json
{
  "error": "Verification failed"
}
```

### Test 2: Expired Verification Code
Register, wait 24+ hours, then:
```bash
curl -X POST http://localhost:9090/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "code": "123456"
  }'
```

**Expected Response (HTTP 400):**
```json
{
  "error": "Verification failed"
}
```

### Test 3: Login Before Verification
Register but don't verify, then try to login:
```bash
curl -X POST http://localhost:9090/login \
  -H "Content-Type: application/json" \
  -d '{
    "student_id": "STU2025-001",
    "password": "Test@12345"
  }'
```

**Expected Response (HTTP 401):**
```json
{
  "error": "Please verify your email first"
}
```

### Test 4: Resend Verification Code
```bash
curl -X POST http://localhost:9090/resend-verification \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com"
  }'
```

**Expected Response:**
```json
{
  "message": "Verification code resent to your email"
}
```

## Database Verification Commands

### Check temp_users (before verification)
```sql
SELECT id, email, student_id, verification_code, expires_at 
FROM temp_users 
WHERE email = 'testuser@example.com';
```

### Check users (after verification)
```sql
SELECT id, student_id, email, is_verified, verified_at, qr_code_data 
FROM users 
WHERE email = 'testuser@example.com';
```

### Count records
```sql
-- Should show account moved, not duplicated
SELECT COUNT(*) as temp_count FROM temp_users WHERE email = 'testuser@example.com';
SELECT COUNT(*) as user_count FROM users WHERE email = 'testuser@example.com';
```

## Logs to Monitor

### On Registration (main.go logs)
```
✅ TempUser saved to DB for testuser@example.com: ID=123, Code=123456, Expires=...
✅ Email sent successfully to testuser@example.com with code 123456
```

### On Verification (handlers/auth.go logs)
```
✅ Verification Success! Temp User ID: 123
✅ User created in users table with ID: 456, StudentID: STU2025-001
✅ Student ID generated: STU2025-001
✅ QR code generated successfully for user: STU2025-001
🗑️ Temp user deleted from database
✅ Transaction committed successfully
✅ Student ID email sent successfully to: testuser@example.com
```

### On Email Send (utils/email.go logs)
```
📧 Sending Student ID email to: testuser@example.com with Student ID: STU2025-001
✅ Email sent successfully to testuser@example.com
```

## Environment Setup for Testing

Ensure these are in your `.env` file:

```env
# SMTP Configuration (Gmail example)
SMTP_EMAIL=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_HOST=smtp.gmail.com

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your-password
DB_NAME=attendance_system

# Server
SERVER_PORT=9090
```

### Gmail Setup
1. Enable 2-factor authentication on your Gmail account
2. Generate an "App Password" for Gmail
3. Use the App Password in `SMTP_PASSWORD`

## Things to Verify

✅ **Functionality**
- [ ] Account moves from temp_users to users
- [ ] temp_user record is deleted (not just marked)
- [ ] Student ID is generated/preserved
- [ ] QR code is created
- [ ] Email is sent with student ID
- [ ] User can login after verification
- [ ] User cannot login before verification

✅ **Database Integrity**
- [ ] No duplicate accounts
- [ ] No orphaned temp_user records
- [ ] All user fields preserved from temp_user
- [ ] Timestamps are correct (created_at, verified_at)

✅ **Error Handling**
- [ ] Invalid code returns error
- [ ] Expired code returns error
- [ ] SMTP failure doesn't block account activation
- [ ] Transaction rollback on DB error

✅ **Logging**
- [ ] Verification logs show progress
- [ ] Email sending is logged
- [ ] Errors are logged with details
- [ ] Transaction commits are logged

## Known Behavior

1. **Email is sent asynchronously** - Response returns before email is sent
2. **Account is activated even if email fails** - Check logs if email didn't arrive
3. **Verification code is valid for 24 hours** - After that, resend required
4. **Student ID is auto-generated** - Format depends on `GenerateCustomStudentID()`
5. **Transaction ensures atomicity** - If any step fails, entire operation rolls back
