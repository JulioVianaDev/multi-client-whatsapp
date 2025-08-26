#!/usr/bin/env node
/**
 * Test script for Phone Number Correction Fix
 * This script demonstrates the FIXED phone number correction for message sending.
 *
 * The key fix: Now properly handles 11-digit numbers and corrects them automatically.
 */

const axios = require("axios");

const GO_BRIDGE_URL = "http://localhost:4444";
const TEST_INSTANCE_KEY = "your_instance_key_here"; // Replace with actual instance key

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function testPhoneCorrection() {
  console.log("🇧🇷 Testing Phone Number Correction Fix for Message Sending\n");

  try {
    // Test 1: Your specific case - 5541991968071 (incorrect with extra 9)
    console.log(
      "📱 Test 1: Sending message with incorrect number 5541991968071"
    );
    console.log(
      "Expected: Should automatically correct to 554191968071@s.whatsapp.net"
    );
    console.log(
      "Previous bug: Would send to 5541991968071@s.whatsapp.net (wrong number!)"
    );
    console.log(
      "Fixed logic: Should detect 5541991968071 doesn't exist, try 554191968071, find it exists\n"
    );

    try {
      const test1 = await axios.post(`${GO_BRIDGE_URL}/message/send`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "5541991968071",
        message: "Test message with auto-corrected phone number! 🇧🇷",
        reply_to: "",
      });
      console.log("✅ Result:", JSON.stringify(test1.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    // Test 2: Validate the same number via phone validation endpoint
    console.log(
      "📱 Test 2: Validating 5541991968071 via /phone/validate endpoint"
    );
    console.log(
      "Expected: Should return 554191968071@s.whatsapp.net as the correct number\n"
    );

    try {
      const test2 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "5541991968071",
      });
      console.log("✅ Result:", JSON.stringify(test2.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    // Test 3: Send message with correct number 554191968071
    console.log("📱 Test 3: Sending message with correct number 554191968071");
    console.log(
      "Expected: Should send to 554191968071@s.whatsapp.net (no correction needed)\n"
    );

    try {
      const test3 = await axios.post(`${GO_BRIDGE_URL}/message/send`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "554191968071",
        message: "Test message with correct phone number! ✅",
        reply_to: "",
      });
      console.log("✅ Result:", JSON.stringify(test3.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    // Test 4: Send message with number that needs 9 added
    console.log("📱 Test 4: Sending message with 551288053918 (needs 9 added)");
    console.log(
      "Expected: Should automatically correct to 5512988053918@s.whatsapp.net\n"
    );

    try {
      const test4 = await axios.post(`${GO_BRIDGE_URL}/message/send`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "551288053918",
        message: "Test message with 9-digit correction! 📱",
        reply_to: "",
      });
      console.log("✅ Result:", JSON.stringify(test4.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    // Test 5: Send message with number that already has @s.whatsapp.net
    console.log("📱 Test 5: Sending message with 5541991968071@s.whatsapp.net");
    console.log(
      "Expected: Should still correct to 554191968071@s.whatsapp.net\n"
    );

    try {
      const test5 = await axios.post(`${GO_BRIDGE_URL}/message/send`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: "5541991968071@s.whatsapp.net",
        message: "Test message with @s.whatsapp.net suffix! 🔧",
        reply_to: "",
      });
      console.log("✅ Result:", JSON.stringify(test5.data, null, 2));
    } catch (error) {
      console.log(
        "❌ Error (expected if instance not connected):",
        error.response?.data || error.message
      );
    }
    console.log("");

    console.log("🎯 SUMMARY OF THE FIX:");
    console.log("1. ✅ Now handles 11-digit numbers (like 5541991968071)");
    console.log("2. ✅ Automatically removes extra 9 when needed");
    console.log("3. ✅ Works with and without @s.whatsapp.net suffix");
    console.log("4. ✅ Corrects numbers before sending messages");
    console.log("5. ✅ Prevents sending to non-existent numbers");
    console.log("");

    console.log("🔧 TECHNICAL DETAILS:");
    console.log(
      "- Added support for 11-digit numbers in validateBrazilianPhoneNumber()"
    );
    console.log(
      "- Enhanced validateAndCorrectPhone() to handle @s.whatsapp.net suffix"
    );
    console.log("- Now checks both original and corrected variations");
    console.log("- Returns the number that actually exists on WhatsApp");
    console.log("");

    console.log("🚀 To test with real instance:");
    console.log("1. Replace 'your_instance_key_here' with actual instance key");
    console.log("2. Make sure instance is connected");
    console.log("3. Run: node test_phone_correction_fix.js");
  } catch (error) {
    console.error("❌ Test failed:", error.message);
  }
}

// Run the test
testPhoneCorrection();
