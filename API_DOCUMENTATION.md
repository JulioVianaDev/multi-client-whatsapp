# WhatsApp Bridge API Documentation

This document describes the API endpoints for the Multi-Instance Go WhatsApp Bridge with message sending capabilities.

## Base URLs

- **Go WhatsApp Bridge**: `http://localhost:4444`
- **Node.js Webhook Receiver**: `http://localhost:5555`

## Instance Management

### Create Instance

**POST** `/instance/create`

Creates a new WhatsApp instance.

**Response:**

```json
{
  "status": "instance_created",
  "instance_key": "abc123def456",
  "message": "Instance created successfully"
}
```

### Connect Instance

**POST** `/instance/connect`

Connects a WhatsApp instance and generates QR code if needed.

**Request Body:**

```json
{
  "instance_key": "abc123def456"
}
```

**Response:**

```json
{
  "status": "qr_generated",
  "instance_key": "abc123def456",
  "message": "QR code generated, scan to connect"
}
```

### Get QR Code

**GET** `/instance/{instanceKey}/qr`

Returns QR code image for scanning with WhatsApp mobile app.

**Response:** PNG image

### Get Instance Status

**GET** `/instance/{instanceKey}/status`

Returns the connection status of an instance.

**Response:**

```json
{
  "instance_key": "abc123def456",
  "connected": true,
  "logged_in": true,
  "phone_number": "1234567890"
}
```

### List All Instances

**GET** `/instances`

Returns all created instances and their status.

**Response:**

```json
{
  "instances": [
    {
      "instance_key": "abc123def456",
      "connected": true,
      "logged_in": true,
      "phone_number": "1234567890"
    }
  ],
  "count": 1
}
```

### Disconnect Instance

**POST** `/instance/{instanceKey}/disconnect`

Disconnects a WhatsApp instance.

**Response:**

```json
{
  "status": "disconnected",
  "instance_key": "abc123def456",
  "message": "Instance disconnected successfully"
}
```

## Message Sending

### Send Text Message

**POST** `/message/send`

Sends a text message to a specific phone number.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "message": "Hello, this is a test message!",
  "reply_to": "optional_message_id_to_reply_to"
}
```

**Response:**

```json
{
  "status": "sent",
  "message_id": "3EB0C767D82B3C2E"
}
```

### Send Media Message

**POST** `/message/send-media`

Sends a media message (image, audio, video, file) to a specific phone number.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "url": "https://example.com/image.jpg",
  "type": "image",
  "caption": "Optional caption for the media"
}
```

**Media Types:**

- `image` - Images (JPG, PNG, etc.)
- `audio` - Audio files (MP3, WAV, etc.)
- `video` - Video files (MP4, AVI, etc.)
- `file` - Documents (PDF, DOC, etc.)

**Response:**

```json
{
  "status": "sent",
  "message_id": "3EB0C767D82B3C2E"
}
```

## Node.js Webhook Receiver Endpoints

### Send Text Message (via Node.js)

**POST** `/message/send`

Proxy endpoint that forwards text message requests to the Go service.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "message": "Hello, this is a test message!",
  "reply_to": "optional_message_id_to_reply_to"
}
```

### Send Media Message (via Node.js)

**POST** `/message/send-media`

Proxy endpoint that forwards media message requests to the Go service.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "url": "https://example.com/image.jpg",
  "type": "image",
  "caption": "Optional caption for the media"
}
```

### Webhook Receiver

**POST** `/webhook`

Receives webhook events from the Go WhatsApp service.

**Request Body:**

```json
{
  "event": "message",
  "instance": "abc123def456",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "from": "1234567890@s.whatsapp.net",
    "message": "Hello!",
    "type": "text"
  }
}
```

## Phone Number Format

Phone numbers should be in the following format:

- **Individual**: `1234567890@s.whatsapp.net`
- **Group**: `1234567890-1234567890@g.us`

## Error Responses

All endpoints return error responses in the following format:

```json
{
  "error": "Error description"
}
```

Common error codes:

- `400` - Bad Request (invalid parameters)
- `404` - Instance not found
- `500` - Internal server error

## Usage Examples

### 1. Create and Connect an Instance

```bash
# Create instance
curl -X POST http://localhost:4444/instance/create

# Connect instance (replace with actual instance_key)
curl -X POST http://localhost:4444/instance/connect \
  -H "Content-Type: application/json" \
  -d '{"instance_key": "abc123def456"}'

# Get QR code
curl http://localhost:4444/instance/abc123def456/qr -o qr.png
```

### 2. Send a Text Message

```bash
curl -X POST http://localhost:4444/message/send \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "abc123def456",
    "phone": "1234567890@s.whatsapp.net",
    "message": "Hello from the WhatsApp bridge!"
  }'
```

### 3. Send an Image

```bash
curl -X POST http://localhost:4444/message/send-media \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "abc123def456",
    "phone": "1234567890@s.whatsapp.net",
    "url": "https://example.com/image.jpg",
    "type": "image",
    "caption": "Check out this image!"
  }'
```

### 4. Send via Node.js Proxy

```bash
curl -X POST http://localhost:5555/message/send \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "abc123def456",
    "phone": "1234567890@s.whatsapp.net",
    "message": "Hello via Node.js proxy!"
  }'
```

## Webhook Events

The system sends webhook events to the Node.js receiver for:

- `connected` - WhatsApp connection established
- `disconnected` - WhatsApp connection lost
- `message` - New message received
- `receipt` - Message delivery receipt
- `presence` - User presence update
- `message_sent` - Message sent successfully
- `message_error` - Message sending failed

## Auto-Reply Feature

The Node.js webhook receiver includes an auto-reply feature that automatically responds to messages containing "hello" with a predefined response. This can be customized in the `handleIncomingMessage` function.

## Docker Deployment

The system is designed to run with Docker Compose:

```bash
docker-compose up -d --build
```

This will start both the Go WhatsApp bridge (port 4444) and the Node.js webhook receiver (port 5555).

## Security Notes

- Phone numbers should be validated before sending messages
- Media URLs should be from trusted sources
- Consider implementing rate limiting for production use
- WhatsApp session data is persisted in Docker volumes
