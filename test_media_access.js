const axios = require("axios");
const fs = require("fs");
const path = require("path");

const BASE_URL = "http://localhost:4444";

async function testMediaAccess() {
  console.log("🧪 Testing Media File Access...\n");

  try {
    // 1. Create a new instance
    console.log("1. Creating new WhatsApp instance...");
    const createResponse = await axios.post(`${BASE_URL}/instance/create`);
    const instanceKey = createResponse.data.instance_key;
    console.log(`✅ Instance created: ${instanceKey}\n`);

    // 2. Check if media directory exists on host
    console.log("2. Checking media directory on host...");
    const mediaDir = path.join(__dirname, "media");
    if (fs.existsSync(mediaDir)) {
      console.log(`✅ Media directory exists: ${mediaDir}`);

      // List contents
      const files = fs.readdirSync(mediaDir, { recursive: true });
      console.log(`📁 Media directory contents:`, files);
    } else {
      console.log(`❌ Media directory not found: ${mediaDir}`);
    }

    // 3. Test media endpoint
    console.log("\n3. Testing media endpoint...");
    try {
      const mediaResponse = await axios.get(`${BASE_URL}/media/test.jpg`);
      console.log("✅ Media endpoint accessible");
    } catch (error) {
      if (error.response?.status === 404) {
        console.log(
          "ℹ️  Media endpoint working (404 expected for non-existent file)"
        );
      } else {
        console.log("❌ Media endpoint error:", error.message);
      }
    }

    // 4. Check container logs for media download
    console.log("\n4. Checking container logs...");
    console.log("📋 To see media download logs, run:");
    console.log("   docker-compose logs -f whatsapp-bridge");

    console.log("\n🎉 Media access test completed!");
    console.log("\n📋 Next steps:");
    console.log("1. Connect your WhatsApp instance");
    console.log("2. Send a media message to the connected number");
    console.log("3. Check the ./media directory for downloaded files");
    console.log(
      "4. Access files via: http://localhost:4444/media/{instance_key}/{date}/{filename}"
    );
  } catch (error) {
    console.error("❌ Test failed:", error.response?.data || error.message);
  }
}

testMediaAccess().catch(console.error);
