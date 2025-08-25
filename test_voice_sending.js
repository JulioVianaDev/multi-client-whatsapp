const axios = require("axios");

// Configuration
const GO_SERVICE_URL = "http://localhost:4444";
const NODEJS_PROXY_URL = "http://localhost:5555";
const TEST_INSTANCE_KEY = "bfc8baadb83578f7a427e0e4209d4d05"; // Replace with your actual instance key
const TEST_PHONE = "16174658484@s.whatsapp.net"; // Replace with your actual phone number

// Test audio URLs
const TEST_AUDIO_URLS = {
  voiceRecording:
    "https://base3.easychannel.online/assets/medias/attendances/cae16460-71e0-4ad3-9d4f-43a55855fc07.mp3",
  musicFile: "https://www.soundjay.com/misc/sounds/bell-ringing-05.wav",
  voiceOgg: "https://example.com/voice.ogg",
};

async function testVoiceRecording() {
  console.log("🎤 Testing Voice Recording Functionality\n");

  try {
    // Test 1: Send voice recording via Go service
    console.log("1️⃣ Testing voice recording via Go service...");
    const voiceResponse = await axios.post(
      `${GO_SERVICE_URL}/message/send-voice`,
      {
        instance_key: TEST_INSTANCE_KEY,
        phone: TEST_PHONE,
        url: TEST_AUDIO_URLS.voiceRecording,
      }
    );
    console.log("✅ Voice recording sent via Go service:", voiceResponse.data);

    // Wait a bit between tests
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // Test 2: Send voice recording via Node.js proxy
    console.log("\n2️⃣ Testing voice recording via Node.js proxy...");
    const proxyVoiceResponse = await axios.post(
      `${NODEJS_PROXY_URL}/message/send-voice`,
      {
        instance_key: TEST_INSTANCE_KEY,
        phone: TEST_PHONE,
        url: TEST_AUDIO_URLS.voiceRecording,
      }
    );
    console.log(
      "✅ Voice recording sent via Node.js proxy:",
      proxyVoiceResponse.data
    );

    // Wait a bit between tests
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // Test 3: Send regular audio file (not voice recording) via Go service
    console.log("\n3️⃣ Testing regular audio file via Go service...");
    const audioResponse = await axios.post(
      `${GO_SERVICE_URL}/message/send-media`,
      {
        instance_key: TEST_INSTANCE_KEY,
        phone: TEST_PHONE,
        url: TEST_AUDIO_URLS.musicFile,
        type: "audio",
        is_ptt: false,
      }
    );
    console.log("✅ Regular audio sent via Go service:", audioResponse.data);

    // Wait a bit between tests
    await new Promise((resolve) => setTimeout(resolve, 2000));

    // Test 4: Send voice recording as audio with PTT flag via Go service
    console.log("\n4️⃣ Testing voice recording as audio with PTT flag...");
    const pttAudioResponse = await axios.post(
      `${GO_SERVICE_URL}/message/send-media`,
      {
        instance_key: TEST_INSTANCE_KEY,
        phone: TEST_PHONE,
        url: TEST_AUDIO_URLS.voiceRecording,
        type: "audio",
        is_ptt: true,
      }
    );
    console.log("✅ PTT audio sent via Go service:", pttAudioResponse.data);

    console.log("\n🎉 All voice recording tests completed successfully!");
  } catch (error) {
    console.error(
      "❌ Error during voice recording tests:",
      error.response?.data || error.message
    );
  }
}

async function testErrorScenarios() {
  console.log("\n🔍 Testing Error Scenarios\n");

  try {
    // Test 1: Invalid instance key
    console.log("1️⃣ Testing invalid instance key...");
    try {
      await axios.post(`${GO_SERVICE_URL}/message/send-voice`, {
        instance_key: "invalid_key",
        phone: TEST_PHONE,
        url: TEST_AUDIO_URLS.voiceRecording,
      });
    } catch (error) {
      console.log(
        "✅ Invalid instance key handled correctly:",
        error.response?.data?.error
      );
    }

    // Test 2: Missing required fields
    console.log("\n2️⃣ Testing missing required fields...");
    try {
      await axios.post(`${GO_SERVICE_URL}/message/send-voice`, {
        instance_key: TEST_INSTANCE_KEY,
        // Missing phone and url
      });
    } catch (error) {
      console.log(
        "✅ Missing fields handled correctly:",
        error.response?.data?.error
      );
    }

    // Test 3: Invalid audio URL
    console.log("\n3️⃣ Testing invalid audio URL...");
    try {
      await axios.post(`${GO_SERVICE_URL}/message/send-voice`, {
        instance_key: TEST_INSTANCE_KEY,
        phone: TEST_PHONE,
        url: "https://invalid-url-that-does-not-exist.com/audio.mp3",
      });
    } catch (error) {
      console.log(
        "✅ Invalid URL handled correctly:",
        error.response?.data?.error
      );
    }

    console.log("\n🎉 All error scenario tests completed!");
  } catch (error) {
    console.error(
      "❌ Error during error scenario tests:",
      error.response?.data || error.message
    );
  }
}

// Run tests
async function runAllTests() {
  console.log("🚀 Starting Voice Recording API Tests\n");
  console.log("📋 Test Configuration:");
  console.log(`   Go Service: ${GO_SERVICE_URL}`);
  console.log(`   Node.js Proxy: ${NODEJS_PROXY_URL}`);
  console.log(`   Instance Key: ${TEST_INSTANCE_KEY}`);
  console.log(`   Test Phone: ${TEST_PHONE}\n`);

  await testVoiceRecording();
  await testErrorScenarios();

  console.log("\n✨ All tests completed!");
}

// Run the tests
runAllTests().catch(console.error);
