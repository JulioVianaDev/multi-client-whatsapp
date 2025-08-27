# Go WhatsApp Bridge

A multi-instance WhatsApp bridge built with Go and whatsmeow, featuring media download capabilities and webhook integration.

## Features

- **Multi-instance Support**: Create and manage multiple WhatsApp instances
- **Brazilian Phone Validation**: Automatic 9-digit handling for Brazilian mobile numbers
- **Media Download**: Automatically download and store media files (images, videos, audio, documents, stickers)
- **Webhook Integration**: Send detailed webhooks to Node.js with event type identification
- **Persistent Storage**: Media files are stored in Docker volumes for persistence
- **Static File Server**: Access downloaded media files via HTTP endpoints

## Brazilian Phone Number Validation

The system includes automatic validation and correction for Brazilian phone numbers, handling the 9-digit mobile number requirement.

### How it Works

1. **Automatic 9-digit Addition**: For Brazilian mobile numbers without the 9-digit prefix, the system automatically adds it
2. **Smart Validation**: Checks if the number exists with and without the 9-digit prefix
3. **Area Code Support**: Supports all Brazilian area codes that use 9-digit mobile numbers
4. **Landline Handling**: Landline numbers are not modified

### Example

```bash
# Input: 551288053918 (without 9)
# Output: 5512988053918@s.whatsapp.net (with 9 added)

# Input: 5512988053918 (with 9)
# Output: 5512988053918@s.whatsapp.net (unchanged)
```

### API Endpoint

```bash
POST /phone/validate
{
  "instance_key": "abc123def456",
  "phone": "551288053918"
}
```

### Supported Area Codes

All major Brazilian area codes that use 9-digit mobile numbers:

- São Paulo (11), Rio de Janeiro (21), Belo Horizonte (31)
- Curitiba (41), Porto Alegre (51), Brasília (61)
- Salvador (71), Recife (81), Belém (91), and many others

## Media Download Functionality

When a media message is received, the system will:

1. **Download the media** from WhatsApp servers
2. **Store it in organized directories**:
   ```
   /app/media/{instance_key}/{date}/{timestamp-uuid}.{extension}
   ```
3. **Send a webhook** with media information instead of raw data
4. **Provide HTTP access** to the media files

### Media Storage Structure

```
/app/media/
├── {instance_key_1}/
│   ├── 2024-01-15/
│   │   ├── 1705312800-abc123.jpg
│   │   └── 1705312900-def456.mp4
│   └── 2024-01-16/
│       └── 1705399200-ghi789.pdf
└── {instance_key_2}/
    └── 2024-01-15/
        └── 1705313000-jkl012.mp3
```

### Webhook Structure for Media Messages

When a media message is received, the webhook will contain:

```json
{
  "event": "image_received",
  "event_type": "image_received",
  "instance": "abc123",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "message_id": "3EB0C767D82B8A6E",
    "chat_id": "5511999999999",
    "sender_id": "5511999999999",
    "from": "5511999999999@s.whatsapp.net",
    "timestamp": "2024-01-15T10:30:00Z",
    "push_name": "John Doe",
    "is_from_me": false,
    "is_group": false,
    "media_type": "image",
    "media_url": "/media/abc123/2024-01-15/1705312800-abc123.jpg",
    "media_path": "/app/media/abc123/2024-01-15/1705312800-abc123.jpg",
    "mime_type": "image/jpeg",
    "caption": "Check out this image!"
  }
}
```

### Accessing Media Files

Media files can be accessed via HTTP:

```
GET http://localhost:4444/media/{instance_key}/{date}/{filename}
```

Example:

```
GET http://localhost:4444/media/abc123/2024-01-15/1705312800-abc123.jpg
```

## Supported Media Types

- **Images**: JPEG, PNG, GIF, WebP
- **Videos**: MP4, MOV, AVI
- **Audio**: MP3, OGG, WAV, M4A
- **Documents**: PDF, DOC, DOCX, XLS, XLSX, PPT, PPTX, TXT
- **Stickers**: WebP format

## Event Types

The system identifies and categorizes various WhatsApp events:

### Message Events

- `message_received` - Text messages
- `image_received` - Image messages
- `video_received` - Video messages
- `audio_received` - Audio messages
- `document_received` - Document messages
- `sticker_received` - Sticker messages
- `contact_received` - Contact sharing
- `location_received` - Location sharing
- `message_revoked` - Deleted messages
- `message_edited` - Edited messages

### Receipt Events

- `message_delivered` - Message delivery confirmation
- `message_read` - Message read confirmation
- `message_played` - Media played confirmation

### Connection Events

- `connected` - Successfully connected
- `disconnected` - Disconnected
- `logged_out` - User logged out
- `pair_success` - Device pairing successful

## Environment Variables

The application supports the following environment variables:

### EASY_ZAP_WEBHOOK_URL

This environment variable controls the base URL used for constructing local file URLs in webhook responses. It's used when the system downloads media files and needs to provide a publicly accessible URL for those files.

**Default**: `http://localhost:4444`

**Examples**:

```bash
# Development
EASY_ZAP_WEBHOOK_URL=http://localhost:4444

# Production with custom domain
EASY_ZAP_WEBHOOK_URL=https://your-domain.com

# Docker with custom port
EASY_ZAP_WEBHOOK_URL=http://localhost:8080

# HTTPS
EASY_ZAP_WEBHOOK_URL=https://localhost:4444
```

**Usage**:

- Set as environment variable: `export EASY_ZAP_WEBHOOK_URL=https://your-domain.com`
- Add to `.env` file: `EASY_ZAP_WEBHOOK_URL=https://your-domain.com`
- Set in `docker-compose.yml` environment section

## Docker Setup

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f whatsapp-bridge

# Access media files
curl http://localhost:4444/media/{instance_key}/{date}/{filename}
```

## Volumes

- `whatsapp_data`: Stores WhatsApp session data
- `media_storage`: Stores downloaded media files

## API Endpoints

- `POST /instance/create` - Create new WhatsApp instance
- `POST /instance/connect` - Connect to WhatsApp instance
- `GET /instance/{instanceKey}/qr` - Get QR code for connection
- `POST /message/send` - Send text message
- `POST /message/send-media` - Send media message
- `POST /message/send-contact` - Send contact message
- `POST /message/send-voice` - Send voice recording (PTT)
- `POST /message/send-location` - Send location coordinates
- `GET /media/*` - Access downloaded media files
- `POST /webhook` - Receive webhooks from Node.js
