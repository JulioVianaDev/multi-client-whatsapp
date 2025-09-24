package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"multi-client-whatsapp/internal/instance"
	"multi-client-whatsapp/internal/types"

	"github.com/google/uuid"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
)

var webhookBaseURL string

func init() {
	webhookBaseURL = os.Getenv("EASY_ZAP_WEBHOOK_URL")
	if webhookBaseURL == "" {
		webhookBaseURL = "http://localhost:4444" // Default fallback
		log.Printf("EASY_ZAP_WEBHOOK_URL not set, using default: %s", webhookBaseURL)
	} else {
		log.Printf("Using webhook URL from environment: %s", webhookBaseURL)
	}
}

// SendWebhook sends webhook data to Node.js with instance information
func SendWebhook(eventType string, data interface{}, instanceKey string) {
	// Check if this is a media message and download if needed
	var enhancedData interface{}

	if msgEvent, ok := data.(*events.Message); ok {
		// Start with the full raw event data
		enhancedData = data

		// Check for media and download if present
		instance.Manager.Mutex.RLock()
		inst, exists := instance.Manager.Instances[instanceKey]
		instance.Manager.Mutex.RUnlock()

		if exists && inst.Client != nil {
			ctx := context.Background()

			// Check for different media types and download them
			if img := msgEvent.Message.GetImageMessage(); img != nil {
				if extractedMedia, err := downloadMedia(ctx, inst.Client, img, instanceKey); err == nil {
					// Add media download info to the raw data
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event":      data,
						"message":        msgEvent.Message,
						"info":           msgEvent.Info,
						"source_string":  msgEvent.Info.SourceString(),
						"push_name":      msgEvent.Info.PushName,
						"is_from_me":     msgEvent.Info.IsFromMe,
						"is_group":       msgEvent.Info.Chat.Server == "g.us",

						// Media download information
						"media_type":     "image",
						"media_url":      extractedMedia.URL,
						"media_path":     extractedMedia.MediaPath,
						"mime_type":      extractedMedia.MimeType,
						"caption":        extractedMedia.Caption,
						"local_file_url": fmt.Sprintf("%s%s", webhookBaseURL, extractedMedia.URL),
					}
				}
			} else if vid := msgEvent.Message.GetVideoMessage(); vid != nil {
				if extractedMedia, err := downloadMedia(ctx, inst.Client, vid, instanceKey); err == nil {
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event":      data,
						"message":        msgEvent.Message,
						"info":           msgEvent.Info,
						"source_string":  msgEvent.Info.SourceString(),
						"push_name":      msgEvent.Info.PushName,
						"is_from_me":     msgEvent.Info.IsFromMe,
						"is_group":       msgEvent.Info.Chat.Server == "g.us",

						// Media download information
						"media_type":     "video",
						"media_url":      extractedMedia.URL,
						"media_path":     extractedMedia.MediaPath,
						"mime_type":      extractedMedia.MimeType,
						"caption":        extractedMedia.Caption,
						"local_file_url": fmt.Sprintf("%s%s", webhookBaseURL, extractedMedia.URL),
					}
				}
			} else if aud := msgEvent.Message.GetAudioMessage(); aud != nil {
				if extractedMedia, err := downloadMedia(ctx, inst.Client, aud, instanceKey); err == nil {
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event":      data,
						"message":        msgEvent.Message,
						"info":           msgEvent.Info,
						"source_string":  msgEvent.Info.SourceString(),
						"push_name":      msgEvent.Info.PushName,
						"is_from_me":     msgEvent.Info.IsFromMe,
						"is_group":       msgEvent.Info.Chat.Server == "g.us",

						// Media download information
						"media_type":     "audio",
						"media_url":      extractedMedia.URL,
						"media_path":     extractedMedia.MediaPath,
						"mime_type":      extractedMedia.MimeType,
						"local_file_url": fmt.Sprintf("%s%s", webhookBaseURL, extractedMedia.URL),
					}
				}
			} else if doc := msgEvent.Message.GetDocumentMessage(); doc != nil {
				if extractedMedia, err := downloadMedia(ctx, inst.Client, doc, instanceKey); err == nil {
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event":      data,
						"message":        msgEvent.Message,
						"info":           msgEvent.Info,
						"source_string":  msgEvent.Info.SourceString(),
						"push_name":      msgEvent.Info.PushName,
						"is_from_me":     msgEvent.Info.IsFromMe,
						"is_group":       msgEvent.Info.Chat.Server == "g.us",

						// Media download information
						"media_type":     "document",
						"media_url":      extractedMedia.URL,
						"media_path":     extractedMedia.MediaPath,
						"mime_type":      extractedMedia.MimeType,
						"caption":        extractedMedia.Caption,
						"filename":       doc.GetFileName(),
						"local_file_url": fmt.Sprintf("%s%s", webhookBaseURL, extractedMedia.URL),
					}
				}
			} else if stk := msgEvent.Message.GetStickerMessage(); stk != nil {
				if extractedMedia, err := downloadMedia(ctx, inst.Client, stk, instanceKey); err == nil {
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event":      data,
						"message":        msgEvent.Message,
						"info":           msgEvent.Info,
						"source_string":  msgEvent.Info.SourceString(),
						"push_name":      msgEvent.Info.PushName,
						"is_from_me":     msgEvent.Info.IsFromMe,
						"is_group":       msgEvent.Info.Chat.Server == "g.us",

						// Media download information
						"media_type":     "sticker",
						"media_url":      extractedMedia.URL,
						"media_path":     extractedMedia.MediaPath,
						"mime_type":      extractedMedia.MimeType,
						"local_file_url": fmt.Sprintf("%s%s", webhookBaseURL, extractedMedia.URL),
					}
				}
			}
		}
	} else {
		enhancedData = data
	}

	payload := types.WebhookPayload{
		Event:     eventType,
		EventType: eventType,
		Instance:  instanceKey,
		Timestamp: time.Now(),
		Data:      enhancedData,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling webhook payload: %v", err)
		return
	}

	// Send to Node.js webhook receiver
	resp, err := http.Post("http://webhook-receiver:5555/webhook", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error sending webhook for instance %s: %v", instanceKey, err)
		return
	}
	defer resp.Body.Close()

	log.Printf("Webhook sent for instance %s: %s", instanceKey, eventType)
}

// downloadMedia downloads media from WhatsApp and saves it to the media volume
func downloadMedia(ctx context.Context, client *whatsmeow.Client, mediaFile whatsmeow.DownloadableMessage, instanceKey string) (*types.ExtractedMedia, error) {
	if mediaFile == nil {
		return nil, fmt.Errorf("media file is nil")
	}

	// Download the media data
	data, err := client.Download(ctx, mediaFile)
	if err != nil {
		return nil, fmt.Errorf("failed to download media: %v", err)
	}

	// Create media directory if it doesn't exist
	mediaDir := "/app/media"
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create media directory: %v", err)
	}

	// Create instance-specific directory
	instanceDir := filepath.Join(mediaDir, instanceKey)
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create instance directory: %v", err)
	}

	// Create date-based directory
	dateDir := filepath.Join(instanceDir, time.Now().Format("2006-01-02"))
	if err := os.MkdirAll(dateDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create date directory: %v", err)
	}

	extractedMedia := &types.ExtractedMedia{}

	// Determine mime type and caption based on media type
	switch media := mediaFile.(type) {
	case *waE2E.ImageMessage:
		extractedMedia.MimeType = media.GetMimetype()
		extractedMedia.Caption = media.GetCaption()
	case *waE2E.AudioMessage:
		extractedMedia.MimeType = media.GetMimetype()
	case *waE2E.VideoMessage:
		extractedMedia.MimeType = media.GetMimetype()
		extractedMedia.Caption = media.GetCaption()
	case *waE2E.StickerMessage:
		extractedMedia.MimeType = media.GetMimetype()
	case *waE2E.DocumentMessage:
		extractedMedia.MimeType = media.GetMimetype()
		extractedMedia.Caption = media.GetCaption()
	}

	// Determine file extension with better mapping
	var extension string
	switch extractedMedia.MimeType {
	case "image/jpeg", "image/jpg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	case "image/jfif":
		extension = ".jpg" // Convert JFIF to JPG
	case "video/mp4":
		extension = ".mp4"
	case "video/avi":
		extension = ".avi"
	case "video/mov":
		extension = ".mov"
	case "audio/mp3":
		extension = ".mp3"
	case "audio/ogg":
		extension = ".ogg"
	case "audio/wav":
		extension = ".wav"
	case "audio/m4a":
		extension = ".m4a"
	case "application/pdf":
		extension = ".pdf"
	case "application/msword":
		extension = ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		extension = ".docx"
	case "application/vnd.ms-excel":
		extension = ".xls"
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		extension = ".xlsx"
	case "application/vnd.ms-powerpoint":
		extension = ".ppt"
	case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
		extension = ".pptx"
	case "text/plain":
		extension = ".txt"
	default:
		// Fallback to MIME type detection
		if ext, err := mime.ExtensionsByType(extractedMedia.MimeType); err == nil && len(ext) > 0 {
			extension = ext[0]
		} else if parts := strings.Split(extractedMedia.MimeType, "/"); len(parts) > 1 {
			extension = "." + parts[len(parts)-1]
		} else {
			extension = ".bin" // Default fallback
		}
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d-%s%s", time.Now().Unix(), uuid.NewString(), extension)
	filePath := filepath.Join(dateDir, filename)

	// Write file to disk
	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return nil, fmt.Errorf("failed to write media file: %v", err)
	}

	extractedMedia.MediaPath = filePath
	extractedMedia.URL = fmt.Sprintf("/media/%s/%s/%s", instanceKey, time.Now().Format("2006-01-02"), filename)

	log.Printf("Media downloaded successfully: %s (MIME: %s, Extension: %s)", extractedMedia.URL, extractedMedia.MimeType, extension)

	return extractedMedia, nil
}
