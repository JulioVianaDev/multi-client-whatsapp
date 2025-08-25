#!/usr/bin/env node
/**
 * Test script for Multi-Instance WhatsApp API
 * This script demonstrates how to create, connect, and manage multiple WhatsApp instances.
 */

const axios = require("axios");

const GO_BRIDGE_URL = "http://localhost:4444";
const NODEJS_URL = "http://localhost:5555";

// Test configuration
const TEST_PHONE = "1234567890@s.whatsapp.net"; // Replace with actual phone number
const TEST_IMAGE_URL = "https://picsum.photos/400/300"; // Random test image
const TEST_AUDIO_URL =
  "https://www.soundjay.com/misc/sounds/bell-ringing-05.wav"; // Test audio

async function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

async function testMultiInstanceWhatsApp() {
  console.log("🚀 Starting Multi-Instance WhatsApp Bridge Test\n");

  try {
    // Step 1: Create a new instance
    console.log("📋 Step 1: Creating new WhatsApp instance...");
    const createResponse = await axios.post(`${GO_BRIDGE_URL}/instance/create`);
    const instanceKey = createResponse.data.instance_key;
    console.log(`✅ Instance created: ${instanceKey}\n`);

    // Step 2: Connect the instance
    console.log("🔗 Step 2: Connecting instance...");
    const connectResponse = await axios.post(
      `${GO_BRIDGE_URL}/instance/connect`,
      {
        instance_key: instanceKey,
      }
    );
    console.log(`✅ ${connectResponse.data.message}\n`);

    // Step 3: Get QR code
    console.log("📱 Step 3: Getting QR code...");
    const qrResponse = await axios.get(
      `${GO_BRIDGE_URL}/instance/${instanceKey}/qr`,
      {
        responseType: "arraybuffer",
      }
    );
    console.log(`✅ QR code received (${qrResponse.data.length} bytes)`);
    console.log("📱 Please scan the QR code with your WhatsApp mobile app\n");

    // Step 4: Wait for connection (poll status)
    console.log("⏳ Step 4: Waiting for WhatsApp connection...");
    let connected = false;
    let attempts = 0;
    const maxAttempts = 30; // 30 seconds timeout

    while (!connected && attempts < maxAttempts) {
      try {
        const statusResponse = await axios.get(
          `${GO_BRIDGE_URL}/instance/${instanceKey}/status`
        );
        if (statusResponse.data.connected && statusResponse.data.logged_in) {
          connected = true;
          console.log(
            `✅ WhatsApp connected! Phone: ${statusResponse.data.phone_number}\n`
          );
        } else {
          console.log(
            `⏳ Waiting for connection... (attempt ${
              attempts + 1
            }/${maxAttempts})`
          );
          await sleep(1000);
          attempts++;
        }
      } catch (error) {
        console.log(
          `⏳ Status check failed, retrying... (attempt ${
            attempts + 1
          }/${maxAttempts})`
        );
        await sleep(1000);
        attempts++;
      }
    }

    if (!connected) {
      console.log(
        "❌ Connection timeout. Please check if you scanned the QR code correctly.\n"
      );
      return;
    }

    // Step 5: Test message sending via Go service
    console.log("💬 Step 5: Testing text message sending via Go service...");
    try {
      const textResponse = await axios.post(`${GO_BRIDGE_URL}/message/send`, {
        instance_key: instanceKey,
        phone: TEST_PHONE,
        message: "Hello from Go WhatsApp Bridge! 🚀",
      });
      console.log(
        `✅ Text message sent! ID: ${textResponse.data.message_id}\n`
      );
    } catch (error) {
      console.log(
        `❌ Text message failed: ${
          error.response?.data?.error || error.message
        }\n`
      );
    }

    // Step 6: Test media message sending via Go service
    console.log("📎 Step 6: Testing image message sending via Go service...");
    try {
      const imageResponse = await axios.post(
        `${GO_BRIDGE_URL}/message/send-media`,
        {
          instance_key: instanceKey,
          phone: TEST_PHONE,
          url: TEST_IMAGE_URL,
          type: "image",
          caption: "Test image from WhatsApp Bridge! 📸",
        }
      );
      console.log(
        `✅ Image message sent! ID: ${imageResponse.data.message_id}\n`
      );
    } catch (error) {
      console.log(
        `❌ Image message failed: ${
          error.response?.data?.error || error.message
        }\n`
      );
    }

    // Step 7: Test message sending via Node.js proxy
    console.log("🔄 Step 7: Testing text message sending via Node.js proxy...");
    try {
      const proxyTextResponse = await axios.post(`${NODEJS_URL}/message/send`, {
        instance_key: instanceKey,
        phone: TEST_PHONE,
        message: "Hello from Node.js proxy! 🔄",
      });
      console.log(
        `✅ Proxy text message sent! ID: ${proxyTextResponse.data.message_id}\n`
      );
    } catch (error) {
      console.log(
        `❌ Proxy text message failed: ${
          error.response?.data?.error || error.message
        }\n`
      );
    }

    // Step 8: Test media message sending via Node.js proxy
    console.log(
      "🔄 Step 8: Testing audio message sending via Node.js proxy..."
    );
    try {
      const proxyAudioResponse = await axios.post(
        `${NODEJS_URL}/message/send-media`,
        {
          instance_key: instanceKey,
          phone: TEST_PHONE,
          url: TEST_AUDIO_URL,
          type: "audio",
        }
      );
      console.log(
        `✅ Proxy audio message sent! ID: ${proxyAudioResponse.data.message_id}\n`
      );
    } catch (error) {
      console.log(
        `❌ Proxy audio message failed: ${
          error.response?.data?.error || error.message
        }\n`
      );
    }

    // Step 9: List all instances
    console.log("📋 Step 9: Listing all instances...");
    try {
      const instancesResponse = await axios.get(`${GO_BRIDGE_URL}/instances`);
      console.log(`✅ Found ${instancesResponse.data.count} instances:`);
      instancesResponse.data.instances.forEach((instance) => {
        console.log(
          `   - ${instance.instance_key}: Connected=${instance.connected}, Phone=${instance.phone_number}`
        );
      });
      console.log();
    } catch (error) {
      console.log(
        `❌ Failed to list instances: ${
          error.response?.data?.error || error.message
        }\n`
      );
    }

    // Step 10: Test webhook functionality
    console.log("📨 Step 10: Testing webhook functionality...");
    try {
      const webhookResponse = await axios.post(`${NODEJS_URL}/webhook`, {
        event: "test_message",
        instance: instanceKey,
        timestamp: new Date().toISOString(),
        data: {
          from: TEST_PHONE,
          message: "Test webhook message",
          type: "text",
        },
      });
      console.log(
        `✅ Webhook test successful: ${webhookResponse.data.status}\n`
      );
    } catch (error) {
      console.log(
        `❌ Webhook test failed: ${
          error.response?.data?.error || error.message
        }\n`
      );
    }

    console.log("🎉 Multi-Instance WhatsApp Bridge Test Completed!");
    console.log("\n📝 Summary:");
    console.log(`   - Instance Key: ${instanceKey}`);
    console.log(`   - Connected: ${connected}`);
    console.log(`   - Test Phone: ${TEST_PHONE}`);
    console.log("\n🔗 Available endpoints:");
    console.log(`   - Go Bridge: ${GO_BRIDGE_URL}`);
    console.log(`   - Node.js Proxy: ${NODEJS_URL}`);
    console.log("\n📚 See API_DOCUMENTATION.md for complete API reference");
  } catch (error) {
    console.error("❌ Test failed:", error.message);
    if (error.response) {
      console.error("Response data:", error.response.data);
    }
  }
}

// Run the test
if (require.main === module) {
  testMultiInstanceWhatsApp();
}

module.exports = { testMultiInstanceWhatsApp };
