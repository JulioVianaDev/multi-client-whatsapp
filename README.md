# Multi-Instance Go WhatsApp Bridge

A powerful WhatsApp bridge built with Go that supports multiple WhatsApp instances with message sending capabilities. This system allows you to manage multiple WhatsApp accounts simultaneously and send text, image, audio, video, and file messages through a REST API.

## Features

- **Multi-Instance Support**: Manage multiple WhatsApp accounts simultaneously
- **Message Sending**: Send text, image, audio, video, and file messages
- **Webhook Integration**: Receive real-time WhatsApp events via webhooks
- **Node.js Proxy**: Additional Node.js layer for enhanced functionality
- **Docker Support**: Easy deployment with Docker Compose
- **Auto-Replies**: Automatic response system for incoming messages
- **Media Support**: Download and send media from URLs
- **Thread-Safe**: Concurrent operations with proper locking

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Your App      │    │  Node.js Proxy   │    │  Go WhatsApp    │
│                 │◄──►│   (Port 5555)    │◄──►│   Bridge        │
│                 │    │                  │    │  (Port 4444)    │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │   WhatsApp Web   │
                       │                  │
                       └──────────────────┘
```

## Quick Start

### 1. Clone the Repository

```bash
git clone <repository-url>
cd go-whats
```

### 2. Start with Docker

```bash
docker-compose up -d --build
```

This will start:

- Go WhatsApp Bridge on port 4444
- Node.js Webhook Receiver on port 5555

### 3. Create and Connect an Instance

```bash
# Create a new instance
curl -X POST http://localhost:4444/instance/create

# Connect the instance (replace with actual instance_key)
curl -X POST http://localhost:4444/instance/connect \
  -H "Content-Type: application/json" \
  -d '{"instance_key": "your_instance_key"}'

# Get QR code
curl http://localhost:4444/instance/your_instance_key/qr -o qr.png
```

### 4. Scan QR Code

Open the generated QR code image and scan it with your WhatsApp mobile app.

### 5. Send Messages

```bash
# Send text message
curl -X POST http://localhost:4444/message/send \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "your_instance_key",
    "phone": "1234567890@s.whatsapp.net",
    "message": "Hello from WhatsApp Bridge!"
  }'

# Send image
curl -X POST http://localhost:4444/message/send-media \
  -H "Content-Type: application/json" \
  -d '{
    "instance_key": "your_instance_key",
    "phone": "1234567890@s.whatsapp.net",
    "url": "https://example.com/image.jpg",
    "type": "image",
    "caption": "Check out this image!"
  }'
```

## API Endpoints

### Instance Management

| Method | Endpoint                     | Description                      |
| ------ | ---------------------------- | -------------------------------- |
| POST   | `/instance/create`           | Create new WhatsApp instance     |
| POST   | `/instance/connect`          | Connect instance and generate QR |
| GET    | `/instance/{key}/qr`         | Get QR code image                |
| GET    | `/instance/{key}/status`     | Get instance status              |
| GET    | `/instances`                 | List all instances               |
| POST   | `/instance/{key}/disconnect` | Disconnect instance              |

### Message Sending

| Method | Endpoint              | Description        |
| ------ | --------------------- | ------------------ |
| POST   | `/message/send`       | Send text message  |
| POST   | `/message/send-media` | Send media message |

### Node.js Proxy

| Method | Endpoint              | Description            |
| ------ | --------------------- | ---------------------- |
| POST   | `/message/send`       | Send text via proxy    |
| POST   | `/message/send-media` | Send media via proxy   |
| POST   | `/webhook`            | Receive webhook events |

## Message Types

### Text Messages

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "message": "Hello, this is a text message!",
  "reply_to": "optional_message_id"
}
```

### Media Messages

```json
{
  "instance_key": "abc123def456",
  "phone": "1234567890@s.whatsapp.net",
  "url": "https://example.com/file.jpg",
  "type": "image",
  "caption": "Optional caption"
}
```

**Supported Media Types:**

- `image` - Images (JPG, PNG, etc.)
- `audio` - Audio files (MP3, WAV, etc.)
- `video` - Video files (MP4, AVI, etc.)
- `file` - Documents (PDF, DOC, etc.)

## Phone Number Format

Phone numbers must be in WhatsApp format:

- **Individual**: `1234567890@s.whatsapp.net`
- **Group**: `1234567890-1234567890@g.us`

## Webhook Events

The system sends webhook events for:

- `connected` - WhatsApp connection established
- `disconnected` - WhatsApp connection lost
- `message` - New message received
- `receipt` - Message delivery receipt
- `presence` - User presence update
- `message_sent` - Message sent successfully
- `message_error` - Message sending failed

## Testing

Run the comprehensive test script:

```bash
node test_multi_instance.js
```

This will:

1. Create a new instance
2. Generate QR code
3. Wait for connection
4. Test message sending
5. Test media sending
6. Test webhook functionality

## Configuration

### Environment Variables

The system uses default configurations, but you can customize:

- **Ports**: 4444 (Go), 5555 (Node.js)
- **Database**: SQLite files per instance
- **Media Storage**: Temporary files in `./media/`

### Docker Configuration

The `docker-compose.yml` file includes:

- Volume persistence for WhatsApp sessions
- Network isolation
- Health checks (commented out)
- Automatic restart policies

## Development

### Prerequisites

- Go 1.19+
- Node.js 16+
- Docker & Docker Compose

### Local Development

```bash
# Start Go service
go run main.go

# Start Node.js service (in another terminal)
cd nodejs-project
npm install
node server.js
```

### Building

```bash
# Build Go binary
go build -o whatsapp-bridge main.go

# Build Docker images
docker-compose build
```

## Security Considerations

- Phone numbers should be validated before sending messages
- Media URLs should be from trusted sources
- Consider implementing rate limiting for production use
- WhatsApp session data is persisted in Docker volumes
- Use HTTPS in production environments

## Troubleshooting

### Common Issues

1. **QR Code Not Working**

   - Ensure the instance is created and connected
   - Check if the QR code image is generated correctly
   - Verify WhatsApp mobile app is up to date

2. **Message Sending Fails**

   - Verify the instance is connected and logged in
   - Check phone number format (must include `@s.whatsapp.net`)
   - Ensure media URLs are accessible

3. **Webhook Not Receiving Events**
   - Check if Node.js service is running on port 5555
   - Verify network connectivity between services
   - Check webhook endpoint logs

### Logs

```bash
# View Go service logs
docker-compose logs whatsapp-bridge

# View Node.js service logs
docker-compose logs webhook-receiver

# Follow logs in real-time
docker-compose logs -f
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For issues and questions:

1. Check the troubleshooting section
2. Review the API documentation
3. Open an issue on GitHub

## Changelog

### v2.0.0

- Added message sending capabilities
- Added media message support
- Added Node.js proxy layer
- Added webhook event handling
- Added auto-reply functionality
- Improved error handling and logging
- Added comprehensive testing
- Updated documentation
