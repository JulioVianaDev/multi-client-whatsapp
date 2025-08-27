const axios = require("axios");

// Test interactive message with buttons
async function testInteractiveMessage() {
  try {
    console.log("🧪 Testing Interactive Message with Buttons...");

    const response = await axios.post(
      "http://localhost:4444/message/send-interactive",
      {
        instance_key: "f744b809555fbdfa217aa79f857e3c14", // Replace with your instance key
        phone: "5511999999999@s.whatsapp.net", // Replace with target phone number
        title: "Choose an option",
        body: "Please select one of the following options:",
        footer: "Powered by WhatsApp Bridge",
        buttons: [
          {
            id: "option_1",
            title: "Option 1",
          },
          {
            id: "option_2",
            title: "Option 2",
          },
          {
            id: "option_3",
            title: "Option 3",
          },
        ],
      }
    );

    console.log("✅ Interactive message sent successfully!");
    console.log("Message ID:", response.data.message_id);
    console.log("Response:", response.data);
  } catch (error) {
    console.error(
      "❌ Error sending interactive message:",
      error.response?.data || error.message
    );
  }
}

// Test interactive message with 2 buttons
async function testInteractiveMessageTwoButtons() {
  try {
    console.log("🧪 Testing Interactive Message with 2 Buttons...");

    const response = await axios.post(
      "http://localhost:4444/message/send-interactive",
      {
        instance_key: "f744b809555fbdfa217aa79f857e3c14", // Replace with your instance key
        phone: "5511999999999@s.whatsapp.net", // Replace with target phone number
        title: "Customer Support",
        body: "How can we help you today?",
        footer: "Select an option below",
        buttons: [
          {
            id: "support_general",
            title: "General Support",
          },
          {
            id: "support_technical",
            title: "Technical Support",
          },
        ],
      }
    );

    console.log("✅ Interactive message with 2 buttons sent successfully!");
    console.log("Message ID:", response.data.message_id);
    console.log("Response:", response.data);
  } catch (error) {
    console.error(
      "❌ Error sending interactive message:",
      error.response?.data || error.message
    );
  }
}

// Test interactive message with single button
async function testInteractiveMessageSingleButton() {
  try {
    console.log("🧪 Testing Interactive Message with Single Button...");

    const response = await axios.post(
      "http://localhost:4444/message/send-interactive",
      {
        instance_key: "f744b809555fbdfa217aa79f857e3c14", // Replace with your instance key
        phone: "5511999999999@s.whatsapp.net", // Replace with target phone number
        title: "Welcome!",
        body: "Thank you for using our service. Click the button below to get started.",
        footer: "Get started now",
        buttons: [
          {
            id: "get_started",
            title: "Get Started",
          },
        ],
      }
    );

    console.log("✅ Interactive message with single button sent successfully!");
    console.log("Message ID:", response.data.message_id);
    console.log("Response:", response.data);
  } catch (error) {
    console.error(
      "❌ Error sending interactive message:",
      error.response?.data || error.message
    );
  }
}

// Test interactive message with reply
async function testInteractiveMessageWithReply() {
  try {
    console.log("🧪 Testing Interactive Message with Reply...");

    const response = await axios.post(
      "http://localhost:4444/message/send-interactive",
      {
        instance_key: "f744b809555fbdfa217aa79f857e3c14", // Replace with your instance key
        phone: "5511999999999@s.whatsapp.net", // Replace with target phone number
        title: "Survey Response",
        body: "How would you rate our service?",
        footer: "Your feedback helps us improve",
        reply_to: "3EB0C767D82B3C2E", // Replace with actual message ID to reply to
        buttons: [
          {
            id: "rating_excellent",
            title: "Excellent",
          },
          {
            id: "rating_good",
            title: "Good",
          },
          {
            id: "rating_fair",
            title: "Fair",
          },
        ],
      }
    );

    console.log("✅ Interactive message with reply sent successfully!");
    console.log("Message ID:", response.data.message_id);
    console.log("Response:", response.data);
  } catch (error) {
    console.error(
      "❌ Error sending interactive message:",
      error.response?.data || error.message
    );
  }
}

// Run tests
async function runTests() {
  console.log("🚀 Starting Interactive Message Tests...\n");

  await testInteractiveMessageSingleButton();
  console.log("\n" + "=".repeat(50) + "\n");

  await testInteractiveMessageTwoButtons();
  console.log("\n" + "=".repeat(50) + "\n");

  await testInteractiveMessage();
  console.log("\n" + "=".repeat(50) + "\n");

  await testInteractiveMessageWithReply();

  console.log("\n🎉 All tests completed!");
}

// Run the tests
runTests().catch(console.error);
