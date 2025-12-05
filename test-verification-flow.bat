@echo off
REM Comprehensive Test Script for Email Verification & Account Activation
REM This script tests the complete flow: Register -> Verify -> Login

setlocal enabledelayedexpansion

set API_URL=http://localhost:9090
set TEST_EMAIL=test_%RANDOM%@example.com
set TEST_PASSWORD=Test@12345
set TEST_USERNAME=testuser%RANDOM%

echo.
echo =========================================
echo Email Verification & Account Activation Test
echo =========================================
echo.

REM Step 1: Register
echo [1/4] REGISTERING NEW ACCOUNT...
echo Email: %TEST_EMAIL%
echo Password: %TEST_PASSWORD%
echo Username: %TEST_USERNAME%
echo.

curl -s -X POST "%API_URL%/register" ^
  -H "Content-Type: application/json" ^
  -d "{\"email\": \"%TEST_EMAIL%\", \"password\": \"%TEST_PASSWORD%\", \"username\": \"%TEST_USERNAME%\", \"first_name\": \"Test\", \"last_name\": \"User\", \"course\": \"BS Computer Science\", \"year_level\": \"1st Year\", \"section\": \"A\", \"department\": \"CICS\", \"college\": \"College of Science\"}" > register_response.json

echo.
echo Response saved to register_response.json
type register_response.json
echo.

echo ⚠️  Check the response above for verification code
echo 📧  Check your email inbox for verification code
echo.
set /p TEST_CODE=Enter the verification code from your email: 

echo.
REM Step 2: Verify Email
echo [2/4] VERIFYING EMAIL...
echo Email: %TEST_EMAIL%
echo Code: %TEST_CODE%
echo.

curl -s -X POST "%API_URL%/verify" ^
  -H "Content-Type: application/json" ^
  -d "{\"email\": \"%TEST_EMAIL%\", \"code\": \"%TEST_CODE%\"}" > verify_response.json

echo Response saved to verify_response.json
type verify_response.json
echo.

for /f "tokens=2 delims=:," %%a in ('findstr /C:"student_id" verify_response.json') do (
    set STUDENT_ID=%%a
    set STUDENT_ID=!STUDENT_ID:"=!
    set STUDENT_ID=!STUDENT_ID: =!
)

if "%STUDENT_ID%"=="" (
    echo ❌ ERROR: Could not extract student ID from response!
    pause
    exit /b 1
)

echo ✅ Verification successful!
echo Student ID: %STUDENT_ID%
echo.

REM Step 3: Login with Student ID
echo [3/4] LOGGING IN WITH STUDENT ID...
echo Student ID: %STUDENT_ID%
echo Password: %TEST_PASSWORD%
echo.

curl -s -X POST "%API_URL%/login" ^
  -H "Content-Type: application/json" ^
  -d "{\"student_id\": \"%STUDENT_ID%\", \"password\": \"%TEST_PASSWORD%\"}" > login_response.json

echo Response saved to login_response.json
type login_response.json
echo.

findstr /C:"Login successful" login_response.json >nul
if %ERRORLEVEL% EQU 0 (
    echo ✅ Login successful!
) else (
    echo ❌ Login failed!
    pause
    exit /b 1
)

echo.
echo =========================================
echo ✅ ALL TESTS PASSED!
echo =========================================
echo.
echo Summary:
echo   Email: %TEST_EMAIL%
echo   Username: %TEST_USERNAME%
echo   Student ID: %STUDENT_ID%
echo.
echo Next steps:
echo   1. Check email inbox for Student ID email
echo   2. Verify account was moved from temp_users to users table
echo   3. Verify temp_users record was deleted
echo.

pause
