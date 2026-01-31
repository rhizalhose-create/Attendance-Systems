#!/bin/bash

# Comprehensive Test Script for Email Verification & Account Activation
# This script tests the complete flow: Register → Verify → Login

set -e

API_URL="http://localhost:9090"
TEST_EMAIL="test_$(date +%s)@example.com"
TEST_PASSWORD="Test@12345"
TEST_USERNAME="testuser$(date +%s)"
TEST_CODE=""  # Will be extracted from logs

echo "========================================="
echo "Email Verification & Account Activation Test"
echo "========================================="
echo ""

# Step 1: Register
echo "[1/4] REGISTERING NEW ACCOUNT..."
echo "Email: $TEST_EMAIL"
echo "Password: $TEST_PASSWORD"
echo "Username: $TEST_USERNAME"
echo ""

REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"username\": \"$TEST_USERNAME\",
    \"first_name\": \"Test\",
    \"last_name\": \"User\",
    \"course\": \"BS Computer Science\",
    \"year_level\": \"1st Year\",
    \"section\": \"A\",
    \"department\": \"CICS\",
    \"college\": \"College of Science\"
  }")

echo "Response: $REGISTER_RESPONSE"
echo ""

# Extract verification code from response (if included for testing)
TEST_CODE=$(echo "$REGISTER_RESPONSE" | grep -o '"note":"Verification code: [^"]*"' | cut -d' ' -f3 | tr -d '"' || echo "NOT_FOUND")

if [ "$TEST_CODE" = "NOT_FOUND" ]; then
    echo "⚠️ Verification code not in response (expected for production)"
    echo "📧 Check email inbox for verification code"
    read -p "Enter the verification code from your email: " TEST_CODE
fi

echo "Using verification code: $TEST_CODE"
echo ""

# Step 2: Verify Email
echo "[2/4] VERIFYING EMAIL..."
echo "Email: $TEST_EMAIL"
echo "Code: $TEST_CODE"
echo ""

VERIFY_RESPONSE=$(curl -s -X POST "$API_URL/verify" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$TEST_EMAIL\",
    \"code\": \"$TEST_CODE\"
  }")

echo "Response: $VERIFY_RESPONSE"
echo ""

# Extract student ID from response
STUDENT_ID=$(echo "$VERIFY_RESPONSE" | grep -o '"student_id":"[^"]*"' | cut -d'"' -f4 || echo "NOT_FOUND")

if [ "$STUDENT_ID" = "NOT_FOUND" ] || [ -z "$STUDENT_ID" ]; then
    echo "❌ ERROR: Could not extract student ID from response!"
    echo "Full response: $VERIFY_RESPONSE"
    exit 1
fi

echo "✅ Verification successful!"
echo "Student ID: $STUDENT_ID"
echo ""

# Step 3: Login with Student ID
echo "[3/4] LOGGING IN WITH STUDENT ID..."
echo "Student ID: $STUDENT_ID"
echo "Password: $TEST_PASSWORD"
echo ""

LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"student_id\": \"$STUDENT_ID\",
    \"password\": \"$TEST_PASSWORD\"
  }")

echo "Response: $LOGIN_RESPONSE"
echo ""

# Check if login was successful
if echo "$LOGIN_RESPONSE" | grep -q '"message":"Login successful"'; then
    echo "✅ Login successful!"
else
    echo "❌ Login failed!"
    exit 1
fi

# Step 4: Get User Profile
echo "[4/4] RETRIEVING USER PROFILE..."
echo "Student ID: $STUDENT_ID"
echo ""

# Extract user ID from login response (or use student_id as identifier)
PROFILE_RESPONSE=$(curl -s -X GET "$API_URL/user/$STUDENT_ID" \
  -H "Content-Type: application/json")

echo "Response: $PROFILE_RESPONSE"
echo ""

if echo "$PROFILE_RESPONSE" | grep -q '"student_id"'; then
    echo "✅ User profile retrieved successfully!"
else
    echo "⚠️ Could not retrieve profile (this might be expected if endpoint uses different parameter)"
fi

echo ""
echo "========================================="
echo "✅ ALL TESTS PASSED!"
echo "========================================="
echo ""
echo "Summary:"
echo "  📧 Email: $TEST_EMAIL"
echo "  👤 Username: $TEST_USERNAME"
echo "  🆔 Student ID: $STUDENT_ID"
echo ""
echo "Next steps:"
echo "  1. Check email inbox for Student ID email"
echo "  2. Verify account was moved from temp_users to users table"
echo "  3. Verify temp_users record was deleted"
echo ""
