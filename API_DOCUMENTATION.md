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

## Phone Number Validation

### Validate Phone Number

**POST** `/phone/validate`

Validates and corrects Brazilian phone numbers with smart 9-digit handling.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "551288053918"
}
```

**Response:**

```json
{
  "status": "success",
  "original_phone": "551288053918",
  "valid_phone": "5512988053918@s.whatsapp.net",
  "exists": true,
  "message": "Phone number validated successfully"
}
```

**How it works (CORRECTED LOGIC):**

1. **Always Try Original First**: The system first checks if the number exists exactly as provided
2. **Smart Fallback**: Only if the original number doesn't exist, then try variations:
   - For 8-digit numbers: Try adding "9" prefix
   - For 9-digit numbers starting with "9": Try removing the "9"
   - For other 9-digit numbers: Try adding "9" prefix
3. **Prevents Wrong Modifications**: Won't modify numbers that already exist correctly
4. **Area Code Support**: Supports all Brazilian area codes that use 9-digit mobile numbers
5. **Landline Handling**: Landline numbers are not modified

**Examples:**

```bash
# Case 1: Number exists as provided (no modification needed)
Input: 554191968071
Logic: Try 554191968071 → Found! → Return 554191968071@s.whatsapp.net
Result: No modification (prevents wrong 5541991968071)

# Case 2: Number needs 9 added
Input: 551288053918
Logic: Try 551288053918 → Not found → Try 5512988053918 → Found!
Result: 5512988053918@s.whatsapp.net

# Case 3: Number has 9 but should be removed
Input: 5512988053918
Logic: Try 5512988053918 → Not found → Try 551288053918 → Found!
Result: 551288053918@s.whatsapp.net
```

**Brazilian Area Codes with 9-digit mobile numbers:**

- São Paulo (11), Rio de Janeiro (21), Belo Horizonte (31), Curitiba (41), Porto Alegre (51)
- Brasília (61), Salvador (71), Recife (81), Belém (91), and many others

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
- `audio` - Audio files (MP3, WAV, etc.) - Use `is_ptt: true` for voice recordings
- `video` - Video files (MP4, AVI, etc.)
- `file` - Documents (PDF, DOC, etc.)

**Response:**

```json
{
  "status": "sent",
  "message_id": "3EB0C767D82B3C2E"
}
```

### Send Contact Message

**POST** `/message/send-contact`

Sends a contact message to a specific phone number.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "contact_name": "John Doe",
  "contact_phone": "9876543210@s.whatsapp.net",
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

### Send Voice Recording

**POST** `/message/send-voice`

Sends a voice recording (PTT) message to a specific phone number.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "url": "https://example.com/voice.ogg",
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

### Send Location

**POST** `/message/send-location`

Sends a location message to a specific phone number.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "latitude": -23.5505,
  "longitude": -46.6333,
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

### Send Contact Message (via Node.js)

**POST** `/message/send-contact`

Proxy endpoint that forwards contact message requests to the Go service.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "contact_name": "John Doe",
  "contact_phone": "9876543210@s.whatsapp.net",
  "reply_to": "optional_message_id_to_reply_to"
}
```

### Send Voice Recording (via Node.js)

**POST** `/message/send-voice`

Proxy endpoint that forwards voice recording requests to the Go service.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "url": "https://example.com/voice.ogg",
  "reply_to": "optional_message_id_to_reply_to"
}
```

### Send Location (via Node.js)

**POST** `/message/send-location`

Proxy endpoint that forwards location requests to the Go service.

**Request Body:**

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "latitude": -23.5505,
  "longitude": -46.6333,
  "reply_to": "optional_message_id_to_reply_to"
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
- **Linked ID (LID)**: `1234567890@lid` (for contacts that don't have @s.whatsapp.net)

### LID Support

The API supports sending messages to contacts using Linked IDs (LID) when the contact doesn't have a standard @s.whatsapp.net format. The system will:

1. **Try to resolve LID to phone number** - If possible, convert the LID to the actual phone number
2. **Fallback to LID** - If resolution fails, send the message using the LID directly
3. **Auto-format** - If no suffix is provided, automatically add @s.whatsapp.net

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

### 4. Send a Contact

```bash
curl -X POST http://localhost:4444/message/send-contact \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "abc123def456",
    "phone": "1234567890@s.whatsapp.net",
    "contact_name": "John Doe",
    "contact_phone": "9876543210@s.whatsapp.net"
  }'
```

### 5. Send a Voice Recording

```bash
curl -X POST http://localhost:4444/message/send-voice \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "abc123def456",
    "phone": "1234567890@s.whatsapp.net",
    "url": "https://example.com/voice.ogg"
  }'
```

### 6. Send Audio File (not voice recording)

```bash
curl -X POST http://localhost:4444/message/send-media \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "abc123def456",
    "phone": "1234567890@s.whatsapp.net",
    "url": "https://example.com/music.mp3",
    "type": "audio",
    "is_ptt": false
  }'
```

### 7. Send Location

```bash
curl -X POST http://localhost:4444/message/send-location \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "abc123def456",
    "phone": "1234567890@s.whatsapp.net",
    "latitude": -23.5505,
    "longitude": -46.6333
  }'
```

### 8. Send to LID Contact

```bash
curl -X POST http://localhost:4444/message/send \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "abc123def456",
    "phone": "1234567890@lid",
    "message": "Hello LID contact!"
  }'
```

### 9. Send via Node.js Proxy

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
