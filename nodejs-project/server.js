const express = require("express");
const cors = require("cors");
const bodyParser = require("body-parser");
const chalk = require("chalk");
const axios = require("axios");

const app = express();
const PORT = 5555;

// Middleware
app.use(cors());
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: true }));

// Store instances and their QR codes
let instances = new Map();

// Routes
app.get("/scan", (req, res) => {
  console.log(chalk.blue("🔍 QR Code Scan Endpoint Accessed"));
  console.log(
    chalk.blue("📱 Please scan the QR code from the Go WhatsApp Bridge")
  );
  console.log(
    chalk.blue(
      "🌐 Visit: http://localhost:4444/instance/{instanceKey}/qr to get the QR code"
    )
  );
  console.log(
    chalk.blue(
      "📋 Or use the /instance/connect endpoint to generate a new QR code"
    )
  );

  res.json({
    message: "QR Code scan endpoint",
    instructions: [
      "1. Create instance: POST http://localhost:4444/instance/create",
      "2. Connect instance: POST http://localhost:4444/instance/connect",
      "3. Get QR code: GET http://localhost:4444/instance/{instanceKey}/qr",
      "4. Scan the QR code with your WhatsApp mobile app",
      "5. All WhatsApp events will be sent to this webhook receiver with instance info",
    ],
    endpoints: {
      create_instance: "POST http://localhost:4444/instance/create",
      connect_instance: "POST http://localhost:4444/instance/connect",
      qr_code: "GET http://localhost:4444/instance/{instanceKey}/qr",
      status: "GET http://localhost:4444/instance/{instanceKey}/status",
      list_instances: "GET http://localhost:4444/instances",
      send_text: "POST http://localhost:4444/message/send",
      send_media: "POST http://localhost:4444/message/send-media",
    },
  });
});

// Multi-instance management endpoints
app.get("/instances", async (req, res) => {
  try {
    console.log(chalk.blue("📋 Fetching all instances..."));
    const response = await axios.get("http://localhost:4444/instances");
    const data = response.data;

    console.log(chalk.green(`✅ Found ${data.count} instances`));
    data.instances.forEach((instance) => {
      console.log(
        chalk.cyan(
          `  - ${instance.instance_key}: Connected=${instance.connected}, Phone=${instance.phone_number}`
        )
      );
    });

    res.json(data);
  } catch (error) {
    console.log(chalk.red(`❌ Failed to fetch instances: ${error.message}`));
    res.status(500).json({ error: "Failed to fetch instances" });
  }
});

app.post("/instance/create", async (req, res) => {
  try {
    console.log(chalk.blue("🆕 Creating new instance..."));
    const response = await axios.post("http://localhost:4444/instance/create");
    const data = response.data;

    console.log(chalk.green(`✅ Instance created: ${data.instance_key}`));
    res.json(data);
  } catch (error) {
    console.log(chalk.red(`❌ Failed to create instance: ${error.message}`));
    res.status(500).json({ error: "Failed to create instance" });
  }
});

app.post("/instance/connect", async (req, res) => {
  try {
    const { instance_key } = req.body;
    if (!instance_key) {
      return res.status(400).json({ error: "instance_key is required" });
    }

    console.log(chalk.blue(`🔗 Connecting instance: ${instance_key}`));
    const response = await axios.post(
      "http://localhost:4444/instance/connect",
      { instance_key }
    );
    const data = response.data;

    console.log(chalk.green(`✅ ${data.message}`));
    res.json(data);
  } catch (error) {
    console.log(chalk.red(`❌ Failed to connect instance: ${error.message}`));
    res.status(500).json({ error: "Failed to connect instance" });
  }
});

app.get("/instance/:instanceKey/qr", async (req, res) => {
  try {
    const { instanceKey } = req.params;
    console.log(chalk.blue(`📱 Getting QR code for instance: ${instanceKey}`));

    const response = await axios.get(
      `http://localhost:4444/instance/${instanceKey}/qr`,
      {
        responseType: "arraybuffer",
      }
    );

    console.log(
      chalk.green(`✅ QR code received for instance: ${instanceKey}`)
    );
    res.set("Content-Type", "image/png");
    res.send(response.data);
  } catch (error) {
    console.log(chalk.red(`❌ Failed to get QR code: ${error.message}`));
    res.status(500).json({ error: "Failed to get QR code" });
  }
});

app.get("/instance/:instanceKey/status", async (req, res) => {
  try {
    const { instanceKey } = req.params;
    console.log(chalk.blue(`📊 Checking status for instance: ${instanceKey}`));

    const response = await axios.get(
      `http://localhost:4444/instance/${instanceKey}/status`
    );
    const data = response.data;

    console.log(
      chalk.green(
        `✅ Status for ${instanceKey}: Connected=${data.connected}, Phone=${data.phone_number}`
      )
    );
    res.json(data);
  } catch (error) {
    console.log(chalk.red(`❌ Failed to check status: ${error.message}`));
    res.status(500).json({ error: "Failed to check status" });
  }
});

// Message sending endpoints
app.post("/message/send", async (req, res) => {
  try {
    const { instance_key, phone, message, reply_to } = req.body;

    if (!instance_key || !phone || !message) {
      return res.status(400).json({
        error: "instance_key, phone, and message are required",
      });
    }

    console.log(
      chalk.blue(
        `💬 Sending text message to ${phone} via instance ${instance_key}`
      )
    );

    const response = await axios.post("http://localhost:4444/message/send", {
      instance_key,
      phone,
      message,
      reply_to,
    });

    console.log(
      chalk.green(
        `✅ Text message sent successfully: ${response.data.message_id}`
      )
    );
    res.json(response.data);
  } catch (error) {
    console.log(chalk.red(`❌ Failed to send text message: ${error.message}`));
    res.status(500).json({ error: "Failed to send text message" });
  }
});

app.post("/message/send-media", async (req, res) => {
  try {
    const { instance_key, phone, url, type, caption } = req.body;

    if (!instance_key || !phone || !url || !type) {
      return res.status(400).json({
        error: "instance_key, phone, url, and type are required",
      });
    }

    // Validate media type
    const validTypes = ["image", "audio", "video", "file"];
    if (!validTypes.includes(type)) {
      return res.status(400).json({
        error: "type must be one of: image, audio, video, file",
      });
    }

    console.log(
      chalk.blue(
        `📎 Sending ${type} message to ${phone} via instance ${instance_key}`
      )
    );

    const response = await axios.post(
      "http://localhost:4444/message/send-media",
      {
        instance_key,
        phone,
        url,
        type,
        caption,
      }
    );

    console.log(
      chalk.green(
        `✅ ${type} message sent successfully: ${response.data.message_id}`
      )
    );
    res.json(response.data);
  } catch (error) {
    console.log(
      chalk.red(
        `❌ Failed to send ${req.body.type || "media"} message: ${
          error.message
        }`
      )
    );
    res.status(500).json({ error: "Failed to send media message" });
  }
});

app.post("/message/send-contact", async (req, res) => {
  try {
    const { instance_key, phone, contact_name, contact_phone, reply_to } =
      req.body;

    if (!instance_key || !phone || !contact_name || !contact_phone) {
      return res.status(400).json({
        error:
          "instance_key, phone, contact_name, and contact_phone are required",
      });
    }

    console.log(
      chalk.blue(
        `👤 Sending contact message to ${phone} via instance ${instance_key}`
      )
    );

    const response = await axios.post(
      "http://localhost:4444/message/send-contact",
      {
        instance_key,
        phone,
        contact_name,
        contact_phone,
        reply_to,
      }
    );

    console.log(
      chalk.green(
        `✅ Contact message sent successfully: ${response.data.message_id}`
      )
    );
    res.json(response.data);
  } catch (error) {
    console.log(
      chalk.red(`❌ Failed to send contact message: ${error.message}`)
    );
    res.status(500).json({ error: "Failed to send contact message" });
  }
});

app.post("/message/send-voice", async (req, res) => {
  try {
    const { instance_key, phone, url, reply_to } = req.body;

    if (!instance_key || !phone || !url) {
      return res.status(400).json({
        error: "instance_key, phone, and url are required",
      });
    }

    console.log(
      chalk.blue(
        `🎤 Sending voice recording to ${phone} via instance ${instance_key}`
      )
    );

    const response = await axios.post(
      "http://localhost:4444/message/send-voice",
      {
        instance_key,
        phone,
        url,
        reply_to,
      }
    );

    console.log(
      chalk.green(
        `✅ Voice recording sent successfully: ${response.data.message_id}`
      )
    );
    res.json(response.data);
  } catch (error) {
    console.log(
      chalk.red(`❌ Failed to send voice recording: ${error.message}`)
    );
    res.status(500).json({ error: "Failed to send voice recording" });
  }
});

// Webhook endpoint for receiving messages from Go service
app.post("/webhook", (req, res) => {
  const { event, instance, timestamp, data } = req.body;

  // Log the webhook with instance information
  console.log(chalk.blue("📨 Webhook Received:"));
  console.log(chalk.blue(`   Instance: ${instance}`));
  console.log(chalk.blue(`   Event: ${event}`));
  console.log(chalk.blue(`   Timestamp: ${timestamp}`));
  console.log(chalk.blue(`   Data: ${JSON.stringify(data, null, 2)}`));
  console.log(chalk.blue("─".repeat(50)));

  // Handle specific events with instance context
  switch (event) {
    case "connected":
      console.log(
        chalk.green(
          `✅ WhatsApp Connected Successfully for instance: ${instance}!`
        )
      );
      break;
    case "disconnected":
      console.log(
        chalk.red(`❌ WhatsApp Disconnected for instance: ${instance}!`)
      );
      break;
    case "message":
      console.log(
        chalk.yellow(`💬 New Message Received for instance: ${instance}!`)
      );
      // Forward message to Go service if needed
      handleIncomingMessage(instance, data);
      break;
    case "receipt":
      console.log(chalk.cyan(`📋 Message Receipt for instance: ${instance}!`));
      break;
    case "presence":
      console.log(
        chalk.magenta(`👤 Presence Update for instance: ${instance}!`)
      );
      break;
    case "message_sent":
      console.log(
        chalk.green(`✅ Message Sent Successfully for instance: ${instance}!`)
      );
      break;
    case "message_error":
      console.log(
        chalk.red(`❌ Message Error for instance: ${instance}: ${data.error}`)
      );
      break;
    default:
      console.log(
        chalk.gray(`ℹ️  Other event: ${event} for instance: ${instance}`)
      );
  }

  res.json({ status: "received" });
});

// Handle incoming messages from WhatsApp
async function handleIncomingMessage(instanceKey, messageData) {
  try {
    // You can add custom logic here to process incoming messages
    // For example, auto-reply, message filtering, etc.

    console.log(
      chalk.cyan(`📥 Processing incoming message for instance ${instanceKey}:`)
    );
    console.log(chalk.cyan(`   From: ${messageData.from}`));
    console.log(chalk.cyan(`   Message: ${messageData.message}`));
    console.log(chalk.cyan(`   Type: ${messageData.type}`));

    // Example: Auto-reply to text messages
    if (
      messageData.type === "text" &&
      messageData.message.toLowerCase().includes("hello")
    ) {
      console.log(
        chalk.yellow(
          `🤖 Auto-replying to hello message from ${messageData.from}`
        )
      );

      // Send auto-reply
      await axios.post("http://localhost:4444/message/send", {
        instance_key: instanceKey,
        phone: messageData.from,
        message: "Hello! This is an auto-reply from the WhatsApp bridge.",
      });

      console.log(chalk.green(`✅ Auto-reply sent to ${messageData.from}`));
    }
  } catch (error) {
    console.log(
      chalk.red(`❌ Error handling incoming message: ${error.message}`)
    );
  }
}

// Health check endpoint
app.get("/health", (req, res) => {
  res.json({
    status: "healthy",
    service: "WhatsApp Multi-Instance Webhook Receiver",
    port: PORT,
    timestamp: new Date().toISOString(),
    features: [
      "Multi-instance WhatsApp management",
      "QR code generation and retrieval",
      "Instance status monitoring",
      "Webhook event processing with instance context",
    ],
  });
});

// Start server
app.listen(PORT, () => {
  console.log(
    chalk.green(
      `🚀 Node.js Multi-Instance Webhook Receiver started on port ${PORT}`
    )
  );
  console.log(chalk.blue("📡 Ready to receive WhatsApp events from Go bridge"));
  console.log(chalk.blue("🔗 Webhook URL: http://localhost:5555/webhook"));
  console.log(chalk.blue("🔍 Scan endpoint: http://localhost:5555/scan"));
  console.log(
    chalk.blue("📋 Instances endpoint: http://localhost:5555/instances")
  );
  console.log(chalk.blue("❤️  Health check: http://localhost:5555/health"));
  console.log(chalk.blue("─".repeat(50)));
});

// Handle graceful shutdown
process.on("SIGINT", () => {
  console.log(chalk.yellow("\n🛑 Shutting down webhook receiver..."));
  process.exit(0);
});

process.on("SIGTERM", () => {
  console.log(chalk.yellow("\n🛑 Shutting down webhook receiver..."));
  process.exit(0);
});
