# Go WhatsApp Bridge

A multi-instance WhatsApp bridge built with Go and whatsmeow, featuring media download capabilities and webhook integration.

## Features

- **Multi-instance Support**: Create and manage multiple WhatsApp instances
- **Media Download**: Automatically download and store media files (images, videos, audio, documents, stickers)
- **Webhook Integration**: Send detailed webhooks to Node.js with event type identification
- **Persistent Storage**: Media files are stored in Docker volumes for persistence
- **Static File Server**: Access downloaded media files via HTTP endpoints

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
- `GET /media/*` - Access downloaded media files
- `POST /webhook` - Receive webhooks from Node.js
