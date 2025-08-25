const axios = require("axios");

const BASE_URL = "http://localhost:4444";
const WEBHOOK_URL = "http://localhost:5555/webhook";

async function testMediaDownload() {
  console.log("🧪 Testing Media Download Functionality...\n");

  try {
    // 1. Create a new instance
    console.log("1. Creating new WhatsApp instance...");
    const createResponse = await axios.post(`${BASE_URL}/instance/create`);
    const instanceKey = createResponse.data.instance_key;
    console.log(`✅ Instance created: ${instanceKey}\n`);

    // 2. Connect the instance
    console.log("2. Connecting instance...");
    await axios.post(`${BASE_URL}/instance/connect`, {
      instance_key: instanceKey,
    });
    console.log("✅ Instance connection initiated\n");

    // 3. Get QR code
    console.log("3. Getting QR code...");
    const qrResponse = await axios.get(
      `${BASE_URL}/instance/${instanceKey}/qr`,
      {
        responseType: "arraybuffer",
      }
    );
    console.log("✅ QR code received\n");

    // 4. Check instance status
    console.log("4. Checking instance status...");
    const statusResponse = await axios.get(
      `${BASE_URL}/instance/${instanceKey}/status`
    );
    console.log(
      `✅ Instance status: ${JSON.stringify(statusResponse.data, null, 2)}\n`
    );

    // 5. Test media file access (this will work once media is downloaded)
    console.log("5. Testing media file access...");
    try {
      const mediaResponse = await axios.get(
        `${BASE_URL}/media/test/2024-01-15/test.jpg`
      );
      console.log("✅ Media file accessible");
    } catch (error) {
      console.log(
        "ℹ️  Media file not found (expected if no media has been downloaded yet)"
      );
    }

    console.log("\n🎉 Media download functionality test completed!");
    console.log("\n📋 Next steps:");
    console.log("1. Scan the QR code with your WhatsApp mobile app");
    console.log(
      "2. Send a media message (image, video, audio, document) to the connected number"
    );
    console.log(
      "3. Check the webhook receiver at http://localhost:5555 for media download events"
    );
    console.log(
      "4. Access downloaded media files at http://localhost:4444/media/{instance_key}/{date}/{filename}"
    );
  } catch (error) {
    console.error("❌ Test failed:", error.response?.data || error.message);
  }
}

// Test webhook receiver
async function testWebhookReceiver() {
  console.log("\n🧪 Testing Webhook Receiver...\n");

  try {
    const testWebhook = {
      event: "test_event",
      event_type: "test_event",
      instance: "test_instance",
      timestamp: new Date().toISOString(),
      data: {
        message_id: "test_message_id",
        media_type: "image",
        media_url: "/media/test/2024-01-15/test.jpg",
        caption: "Test image",
      },
    };

    const response = await axios.post(WEBHOOK_URL, testWebhook);
    console.log("✅ Webhook received successfully");
    console.log("Response:", response.data);
  } catch (error) {
    console.error(
      "❌ Webhook test failed:",
      error.response?.data || error.message
    );
  }
}

// Run tests
async function runTests() {
  console.log("🚀 Starting Media Download Tests\n");

  await testWebhookReceiver();
  await testMediaDownload();
}

runTests().catch(console.error);
