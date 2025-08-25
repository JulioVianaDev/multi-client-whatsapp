const axios = require('axios');
const chalk = require('chalk');

const GO_BRIDGE_URL = 'http://localhost:4444';
const NODEJS_PROXY_URL = 'http://localhost:5555';

// Test configuration
const TEST_INSTANCE_KEY = 'abc123def456'; // Replace with actual instance key
const TEST_PHONE = '1234567890@s.whatsapp.net'; // Replace with actual phone number
const TEST_CONTACT_NAME = 'John Doe';
const TEST_CONTACT_PHONE = '9876543210@s.whatsapp.net';

async function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

async function testContactSending() {
  console.log("🚀 Starting Contact Sending Test\n");

  // Step 1: Test contact sending via Go service
  console.log("👤 Step 1: Testing contact sending via Go service...");
  try {
    const contactResponse = await axios.post(`${GO_BRIDGE_URL}/message/send-contact`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: TEST_PHONE,
      contact_name: TEST_CONTACT_NAME,
      contact_phone: TEST_CONTACT_PHONE,
    });
    console.log(
      chalk.green(
        `✅ Contact sent via Go service! ID: ${contactResponse.data.message_id}\n`
      )
    );
  } catch (error) {
    console.log(
      chalk.red(
        `❌ Contact sending via Go service failed: ${
          error.response?.data?.error || error.message
        }\n`
      )
    );
  }

  // Step 2: Test contact sending via Node.js proxy
  console.log("👤 Step 2: Testing contact sending via Node.js proxy...");
  try {
    const contactProxyResponse = await axios.post(`${NODEJS_PROXY_URL}/message/send-contact`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: TEST_PHONE,
      contact_name: 'Jane Smith',
      contact_phone: '5551234567@s.whatsapp.net',
    });
    console.log(
      chalk.green(
        `✅ Contact sent via Node.js proxy! ID: ${contactProxyResponse.data.message_id}\n`
      )
    );
  } catch (error) {
    console.log(
      chalk.red(
        `❌ Contact sending via Node.js proxy failed: ${
          error.response?.data?.error || error.message
        }\n`
      )
    );
  }

  // Step 3: Test contact sending with reply
  console.log("👤 Step 3: Testing contact sending with reply...");
  try {
    const contactReplyResponse = await axios.post(`${GO_BRIDGE_URL}/message/send-contact`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: TEST_PHONE,
      contact_name: 'Reply Contact',
      contact_phone: '1112223333@s.whatsapp.net',
      reply_to: '3EB0C767D82B3C2E', // Replace with actual message ID
    });
    console.log(
      chalk.green(
        `✅ Contact with reply sent! ID: ${contactReplyResponse.data.message_id}\n`
      )
    );
  } catch (error) {
    console.log(
      chalk.red(
        `❌ Contact with reply failed: ${
          error.response?.data?.error || error.message
        }\n`
      )
    );
  }

  // Step 4: Test error handling - invalid instance key
  console.log("👤 Step 4: Testing error handling - invalid instance key...");
  try {
    await axios.post(`${GO_BRIDGE_URL}/message/send-contact`, {
      instance_key: 'invalid_key',
      phone: TEST_PHONE,
      contact_name: TEST_CONTACT_NAME,
      contact_phone: TEST_CONTACT_PHONE,
    });
    console.log(chalk.red(`❌ Should have failed with invalid instance key\n`));
  } catch (error) {
    if (error.response?.status === 404) {
      console.log(
        chalk.green(
          `✅ Correctly handled invalid instance key: ${error.response.data.error}\n`
        )
      );
    } else {
      console.log(
        chalk.red(
          `❌ Unexpected error for invalid instance key: ${
            error.response?.data?.error || error.message
          }\n`
        )
      );
    }
  }

  // Step 5: Test error handling - missing required fields
  console.log("👤 Step 5: Testing error handling - missing required fields...");
  try {
    await axios.post(`${GO_BRIDGE_URL}/message/send-contact`, {
      instance_key: TEST_INSTANCE_KEY,
      phone: TEST_PHONE,
      // Missing contact_name and contact_phone
    });
    console.log(chalk.red(`❌ Should have failed with missing fields\n`));
  } catch (error) {
    if (error.response?.status === 400) {
      console.log(
        chalk.green(
          `✅ Correctly handled missing fields: ${error.response.data.error}\n`
        )
      );
    } else {
      console.log(
        chalk.red(
          `❌ Unexpected error for missing fields: ${
            error.response?.data?.error || error.message
          }\n`
        )
      );
    }
  }

  console.log("🎉 Contact sending test completed!");
}

// Run the test
testContactSending().catch(console.error);
