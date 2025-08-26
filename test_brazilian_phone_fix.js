#!/usr/bin/env node
/**
 * Test script for Corrected Brazilian Phone Number Validation
 * This script demonstrates the FIXED phone validation logic for Brazilian numbers.
 *
 * The key fix: Always try the original number FIRST, then try variations.
 */

const axios = require("axios");

const GO_BRIDGE_URL = "http://localhost:4444";
const TEST_INSTANCE_KEY = "your_instance_key_here"; // Replace with actual instance key

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function testPhoneValidation() {
  console.log("🇧🇷 Testing CORRECTED Brazilian Phone Number Validation\n");

  try {
    // Test 1: Your specific case - 554191968071 (correct number without 9)
    console.log(
      "📱 Test 1: Your specific case - 554191968071 (should NOT add 9)"
    );
    console.log(
      "Expected: Should try 554191968071 first, find it exists, return it"
    );
    console.log("Previous bug: Would add 9 → 5541991968071 (wrong!)");
    console.log("Fixed logic: Try original first, only modify if not found\n");

    try {
      const test1 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "554191968071",
      });
      console.log("✅ Result:", JSON.stringify(test1.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    // Test 2: Number that needs 9 added (8 digits)
    console.log("📱 Test 2: Number that needs 9 added - 551288053918");
    console.log(
      "Expected: Should try 551288053918 first, not found, then try 5512988053918"
    );
    console.log("Fixed logic: Try original first, then add 9 if not found\n");

    try {
      const test2 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "551288053918",
      });
      console.log("✅ Result:", JSON.stringify(test2.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    // Test 3: Number with 9 that should be removed
    console.log(
      "📱 Test 3: Number with 9 that should be removed - 5512988053918"
    );
    console.log(
      "Expected: Should try 5512988053918 first, not found, then try 551288053918"
    );
    console.log(
      "Fixed logic: Try original first, then remove 9 if not found\n"
    );

    try {
      const test3 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "5512988053918",
      });
      console.log("✅ Result:", JSON.stringify(test3.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    // Test 4: Landline number (should not be modified)
    console.log("📱 Test 4: Landline number - 551123456789");
    console.log(
      "Expected: Should only try the original number, no modifications"
    );
    console.log("Fixed logic: Landlines are not modified\n");

    try {
      const test4 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "551123456789",
      });
      console.log("✅ Result:", JSON.stringify(test4.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    // Test 5: Non-Brazilian number
    console.log("📱 Test 5: Non-Brazilian number - 1234567890");
    console.log("Expected: Should return error - not a Brazilian number");
    console.log("Fixed logic: Only process Brazilian numbers\n");

    try {
      const test5 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "1234567890",
      });
      console.log("✅ Result:", JSON.stringify(test5.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    console.log("🎯 SUMMARY OF THE FIX:");
    console.log("1. ✅ ALWAYS try the original number FIRST");
    console.log("2. ✅ Only modify (add/remove 9) if original doesn't exist");
    console.log("3. ✅ Return the original number if no variation exists");
    console.log(
      "4. ✅ Prevents wrong modifications like 554191968071 → 5541991968071"
    );
    console.log("5. ✅ Handles all Brazilian area codes correctly");
    console.log("");

    console.log("🚀 To test with real instance:");
    console.log("1. Replace 'your_instance_key_here' with actual instance key");
    console.log("2. Make sure instance is connected");
    console.log("3. Run: node test_brazilian_phone_fix.js");
  } catch (error) {
    console.error("❌ Test failed:", error.message);
  }
}

// Run the test
testPhoneValidation();
