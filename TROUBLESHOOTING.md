# Troubleshooting Guide - If Nothing Happens

## Problem: "Nothing is happening" after verification

This guide helps you diagnose and fix issues with the email verification and account activation flow.

---

## Quick Diagnostics

### Step 1: Verify the Server is Running
```bash
curl http://localhost:9090/health
```

**Expected Response:**
```json
{
  "status": "OK",
  "service": "Attendance System API"
}
```

If this fails:
- [ ] Start the server: `go run main.go`
- [ ] Check if port 9090 is already in use
- [ ] Try a different port: `SERVER_PORT=9091 go run main.go`

---

## Issue 1: Registration Works But Verification Fails

### Symptoms:
- Registration creates temp_user ✅
- Verification endpoint returns error
- Account not moving to users table

### Cause & Fix:

**Check 1: Verification Code Mismatch**
```sql
SELECT email, verification_code, expires_at FROM temp_users LIMIT 5;
```

- Verify the code you're submitting matches the database
- Check if code has expired (should be valid for 24 hours)

**Check 2: Email Format Issue**
- Make sure email in request exactly matches what's in database
- Email is case-sensitive in database queries
- Check for extra spaces or special characters

**Check 3: Code Expiration**
```sql
SELECT email, expires_at, NOW() as current_time FROM temp_users 
WHERE email = 'your-email@example.com';
```

If `expires_at` < `current_time`, the code has expired. Register again.

---

## Issue 2: Verification Succeeds But Account Not Moving

### Symptoms:
- Verification endpoint returns success
- Student ID is returned
- But `temp_users` record still exists
- `users` table has no new record

### Cause & Fix:

**Check 1: Database Transaction Rollback**

Look in server logs for:
```
❌ Failed to create user in users table
❌ Failed to update student_id
❌ Failed to delete temp user
```

**Solutions:**
- [ ] Check database connection is working
- [ ] Verify users table exists and is writable
- [ ] Check disk space on database server
- [ ] Check database user has INSERT/UPDATE/DELETE permissions

**Check 2: Unique Constraint Violation**

If you see error like "Duplicate entry":
- Student ID might already exist in users table
- Email might already exist in users table

```sql
-- Check for duplicates
SELECT student_id, COUNT(*) as count FROM users GROUP BY student_id HAVING count > 1;
SELECT email, COUNT(*) as count FROM users GROUP BY email HAVING count > 1;
```

**Fix:** Delete duplicates from users table or use different email

---

## Issue 3: Account Moved But Email Not Received

### Symptoms:
- Verification succeeds ✅
- Account in users table ✅
- Temp user deleted ✅
- But no Student ID email received

### Cause & Fix:

**Check 1: SMTP Configuration**

Verify environment variables:
```bash
# In .env file
SMTP_EMAIL=your-email@gmail.com
SMTP_PASSWORD=your-app-password  (NOT your Gmail password)
SMTP_HOST=smtp.gmail.com
```

**Check 2: Gmail Setup**
- [ ] Enable 2-factor authentication on Gmail account
- [ ] Generate App Password (not regular password)
- [ ] Use App Password in SMTP_PASSWORD
- [ ] Allow Less Secure Apps (if using regular Gmail)

**Check 3: Check Server Logs**

Look for:
```
✅ Student ID email sent successfully to: user@example.com
```

If you see:
```
❌ Failed to send student ID email: SMTP connection failed
```

**Solutions:**
- [ ] Test SMTP manually: `telnet smtp.gmail.com 587`
- [ ] Verify credentials are correct
- [ ] Check firewall/network blocks outgoing SMTP
- [ ] Try different email provider (Mailgun, SendGrid)

**Check 4: Email in Spam Folder**
- Check email spam/junk folder
- Add sender email to contacts
- Check email provider's email logs

---

## Issue 4: Login Fails After Verification

### Symptoms:
- Verification succeeds ✅
- Account in users table ✅
- Email received ✅
- But login fails with "Invalid StudentID/Password"

### Cause & Fix:

**Check 1: Wrong StudentID Format**

```sql
SELECT student_id, email FROM users WHERE email = 'your-email@example.com';
```

Make sure you're using the exact StudentID from the database, not a typo.

**Check 2: Account Not Marked as Verified**

```sql
SELECT student_id, email, is_verified FROM users WHERE email = 'your-email@example.com';
```

Should show `is_verified = 1` (or true)

If not:
- Verification didn't complete
- Check server logs for transaction errors
- Try verification again

**Check 3: Password Incorrect**

- Password was hashed during registration
- You must use the same password you registered with
- Passwords are case-sensitive
- Common issue: accidental space at beginning/end

---

## Debug Steps - Run These in Order

### Step 1: Check if Temp User Was Created
```sql
SELECT id, email, student_id, verification_code, expires_at 
FROM temp_users 
WHERE email = 'test@example.com';
```

**Expected:**
- Row exists
- `verification_code` is not empty (6 digits)
- `expires_at` is in the future

### Step 2: Check Server Logs During Verification

Start server with:
```bash
go run main.go 2>&1 | tee server.log
```

Then run verification and look for:
```
✅ Verification Success! Temp User ID: 123
📝 Transaction started for email verification
✅ User created in users table with ID: 456
✅ Student ID successfully set to: STU2025-001
🗑️ Temp user deleted from database
✅ Transaction committed successfully
✅ Student ID email sent successfully
```

If you see errors like `❌ Failed to`, that's your problem.

### Step 3: Verify Database State

After verification attempt:

```sql
-- Should be empty (or no matching row)
SELECT COUNT(*) as temp_count FROM temp_users WHERE email = 'test@example.com';

-- Should have 1 row
SELECT COUNT(*) as user_count FROM users WHERE email = 'test@example.com';

-- Check the user record
SELECT student_id, email, is_verified, verified_at FROM users WHERE email = 'test@example.com';
```

### Step 4: Test Email Sending Separately

Create a test script:

```bash
curl -X POST http://localhost:9090/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "code": "123456"
  }' 2>&1 | tee verify_output.txt
```

Check output for:
- `"message": "Email verified successfully"` - Success
- `"error": "Verification failed"` - Code/Email mismatch
- `"error": "Failed to activate account"` - Database error

---

## Common Error Messages & Solutions

### "Verification failed"
- Wrong verification code
- Code has expired (24 hours)
- Email doesn't match what's in database

**Solution:** Register again and use correct code from email

### "Failed to activate account"
- Database connection error
- Users table doesn't exist or not writable
- Disk space issue on database server

**Solution:** Check database permissions and connection

### "Failed to generate student ID"
- Database query error
- StudentID column not updatable
- Duplicate StudentID

**Solution:** Check database schema and logs

### "email service not configured"
- SMTP_EMAIL or SMTP_PASSWORD not set in .env
- Credentials missing

**Solution:** 
```bash
# In .env
SMTP_EMAIL=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

---

## Test Data for Manual Testing

### Minimal Registration
```json
{
  "email": "test@example.com",
  "password": "Test@12345",
  "username": "testuser",
  "first_name": "Test",
  "last_name": "User",
  "course": "BS CS",
  "year_level": "1st Year"
}
```

### Verification
```json
{
  "email": "test@example.com",
  "code": "123456"
}
```

### Login
```json
{
  "student_id": "STU2025-001",
  "password": "Test@12345"
}
```

---

## Enable Debug Logging

Add to your `.env`:
```
LOG_LEVEL=debug
DEBUG=true
```

Or modify `main.go` to increase verbosity:
```go
log.SetFlags(log.LstdFlags | log.Lshortfile)
```

---

## Request Support

When asking for help, include:

1. **Server Logs**
   ```bash
   go run main.go 2>&1 | grep -A 5 "Verification" > logs.txt
   ```

2. **Database State**
   ```sql
   SELECT * FROM temp_users WHERE email = 'test@example.com'\G
   SELECT * FROM users WHERE email = 'test@example.com'\G
   ```

3. **Request/Response**
   ```json
   {
     "request": { "email": "...", "code": "..." },
     "response": { "error": "..." }
   }
   ```

4. **Environment**
   - Operating System
   - Go version: `go version`
   - Database type (MySQL/PostgreSQL/SQLite)
   - Server port being used

---

## Quick Checklist

Before assuming something is broken:

- [ ] Verify endpoint is being called (check logs)
- [ ] Verify verification code is correct
- [ ] Verify code hasn't expired
- [ ] Verify database connection works
- [ ] Verify SMTP credentials are correct
- [ ] Check server logs for `❌` errors
- [ ] Check database for account creation
- [ ] Check email spam folder
- [ ] Check firewall/network connectivity
- [ ] Verify temp_users table exists
- [ ] Verify users table exists
- [ ] Verify database user has permissions

---

## Still Not Working?

1. Check `server.log` for detailed error messages
2. Run `go build` to ensure code compiles
3. Restart the server: `go run main.go`
4. Try with fresh registration (don't reuse same email)
5. Check all environment variables are set
6. Verify database is running and accessible
7. Check database logs for errors
8. Try on a different database to isolate issue
9. Review the changes made to `/handlers/auth.go`
10. Contact support with logs and details

