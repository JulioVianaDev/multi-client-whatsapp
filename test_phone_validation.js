#!/usr/bin/env node
/**
 * Test script for Brazilian Phone Number Validation
 * This script demonstrates the new phone validation functionality for Brazilian numbers.
 */

const axios = require("axios");

const GO_BRIDGE_URL = "http://localhost:4444";
const NODEJS_URL = "http://localhost:5555";

// Test configuration
const TEST_INSTANCE_KEY = "your_instance_key_here"; // Replace with actual instance key

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function testPhoneValidation() {
  console.log("🇧🇷 Starting Brazilian Phone Number Validation Test\n");

  try {
    // Test 1: Validate a number without 9 digits (should add 9)
    console.log("📱 Test 1: Validating number without 9 digits (551288053918)");
    const test1 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: "551288053918",
    });
    console.log("✅ Result:", test1.data);
    console.log("");

    // Test 2: Validate a number with 9 digits (should work as is)
    console.log("📱 Test 2: Validating number with 9 digits (5512988053918)");
    const test2 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: "5512988053918",
    });
    console.log("✅ Result:", test2.data);
    console.log("");

    // Test 3: Validate a landline number (should not add 9)
    console.log("📞 Test 3: Validating landline number (551123456789)");
    const test3 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: "551123456789",
    });
    console.log("✅ Result:", test3.data);
    console.log("");

    // Test 4: Validate a number with @s.whatsapp.net suffix
    console.log("📱 Test 4: Validating number with @s.whatsapp.net suffix");
    const test4 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: "551288053918@s.whatsapp.net",
    });
    console.log("✅ Result:", test4.data);
    console.log("");

    // Test 5: Validate a non-Brazilian number
    console.log("🌍 Test 5: Validating non-Brazilian number (1234567890)");
    const test5 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: "1234567890",
    });
    console.log("✅ Result:", test5.data);
    console.log("");

    // Test 6: Send message to number without 9 digits (should auto-correct)
    console.log("💬 Test 6: Sending message to number without 9 digits");
    const test6 = await axios.post(`${GO_BRIDGE_URL}/message/send`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: "551288053918",
      message: "Test message with auto-corrected phone number! 🇧🇷",
    });
    console.log("✅ Result:", test6.data);
    console.log("");

    // Test 7: Send message to number with 9 digits
    console.log("💬 Test 7: Sending message to number with 9 digits");
    const test7 = await axios.post(`${GO_BRIDGE_URL}/message/send`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: "5512988053918",
      message: "Test message with correct phone number! 📱",
    });
    console.log("✅ Result:", test7.data);
    console.log("");

    // Test 8: Send media message with auto-corrected phone
    console.log("📸 Test 8: Sending media message with auto-corrected phone");
    const test8 = await axios.post(`${GO_BRIDGE_URL}/message/send-media`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: "551288053918",
      url: "https://picsum.photos/400/300",
      type: "image",
      caption: "Test image with auto-corrected phone number! 🖼️",
    });
    console.log("✅ Result:", test8.data);
    console.log("");

    console.log(
      "🎉 All Brazilian phone validation tests completed successfully!"
    );
  } catch (error) {
    console.error("❌ Test failed:", error.response?.data || error.message);
  }
}

// Test different area codes that use 9-digit numbers
async function testDifferentAreaCodes() {
  console.log("\n🏙️ Testing different area codes with 9-digit validation\n");

  const areaCodes = [
    "11", // São Paulo
    "21", // Rio de Janeiro
    "31", // Belo Horizonte
    "41", // Curitiba
    "51", // Porto Alegre
    "61", // Brasília
    "71", // Salvador
    "81", // Recife
    "91", // Belém
  ];

  for (const areaCode of areaCodes) {
    try {
      console.log(`📱 Testing area code ${areaCode}`);

      // Test with 8 digits (should add 9)
      const phone8Digits = `55${areaCode}12345678`;
      const test8 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: phone8Digits,
      });
      console.log(`  8 digits (${phone8Digits}): ${test8.data.valid_phone}`);

      // Test with 9 digits (should work as is)
      const phone9Digits = `55${areaCode}912345678`;
      const test9 = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: phone9Digits,
      });
      console.log(`  9 digits (${phone9Digits}): ${test9.data.valid_phone}`);
      console.log("");
    } catch (error) {
      console.error(
        `❌ Error testing area code ${areaCode}:`,
        error.response?.data || error.message
      );
    }
  }
}

// Test error cases
async function testErrorCases() {
  console.log("\n⚠️ Testing error cases\n");

  const errorTests = [
    {
      name: "Invalid phone number (too short)",
      phone: "551234",
    },
    {
      name: "Invalid phone number (too long)",
      phone: "551234567890123456",
    },
    {
      name: "Non-Brazilian number",
      phone: "1234567890",
    },
    {
      name: "Invalid format",
      phone: "abc123def456",
    },
  ];

  for (const test of errorTests) {
    try {
      console.log(`🔍 ${test.name}: ${test.phone}`);
      const result = await axios.post(`${GO_BRIDGE_URL}/phone/validate`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: test.phone,
      });
      console.log(`  Result: ${result.data.status} - ${result.data.message}`);
    } catch (error) {
      console.log(
        `  Expected error: ${error.response?.data?.error || error.message}`
      );
    }
    console.log("");
  }
}

// Main test function
async function runAllTests() {
  console.log("🚀 Starting comprehensive Brazilian phone validation tests\n");

  await testPhoneValidation();
  await testDifferentAreaCodes();
  await testErrorCases();

  console.log("✨ All tests completed!");
}

// Run tests if this file is executed directly
if (require.main === module) {
  runAllTests().catch(console.error);
}

module.exports = {
  testPhoneValidation,
  testDifferentAreaCodes,
  testErrorCases,
  runAllTests,
};
