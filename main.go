package main

import (
	_ "github.com/mattn/go-sqlite3"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

// Instance represents a WhatsApp client instance
type Instance struct {
	ID           string
	Client       *whatsmeow.Client
	PhoneNumber  string
	IsConnected  bool
	QRCodeChan   chan string
	Container    *sqlstore.Container
	Mutex        sync.RWMutex
}

// InstanceManager manages all WhatsApp instances
type InstanceManager struct {
	Instances map[string]*Instance
	Mutex     sync.RWMutex
}

// WebhookPayload represents the webhook data sent to Node.js
type WebhookPayload struct {
	Event     string      `json:"event"`
	EventType string      `json:"event_type"`
	Instance  string      `json:"instance"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}

// ConnectRequest represents the request to connect a new instance
type ConnectRequest struct {
	InstanceKey string `json:"instance_key"`
}

// ConnectResponse represents the response from connect endpoint
type ConnectResponse struct {
	Status      string `json:"status"`
	InstanceKey string `json:"instance_key"`
	Message     string `json:"message,omitempty"`
}

// MessageRequest represents a message sending request
type MessageRequest struct {
	InstanceKey string `json:"instance_key" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Message     string `json:"message" binding:"required"`
	ReplyTo     string `json:"reply_to,omitempty"`
}

// MediaMessageRequest represents a media message sending request
type MediaMessageRequest struct {
	InstanceKey string `json:"instance_key" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	Caption     string `json:"caption,omitempty"`
	URL         string `json:"url" binding:"required"`
	Type        string `json:"type" binding:"required"` // "image", "audio", "video", "file"
	IsPTT       bool   `json:"is_ptt,omitempty"`        // For audio: true = voice recording, false = audio file
}

// VoiceMessageRequest represents a voice recording message sending request
type VoiceMessageRequest struct {
	InstanceKey string `json:"instance_key" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	URL         string `json:"url" binding:"required"`
	ReplyTo     string `json:"reply_to,omitempty"`
}

// LocationMessageRequest represents a location message sending request
type LocationMessageRequest struct {
	InstanceKey string  `json:"instance_key" binding:"required"`
	Phone       string  `json:"phone" binding:"required"`
	Latitude    float64 `json:"latitude" binding:"required"`
	Longitude   float64 `json:"longitude" binding:"required"`
	ReplyTo     string  `json:"reply_to,omitempty"`
}

// ContactMessageRequest represents a contact message sending request
type ContactMessageRequest struct {
	InstanceKey string `json:"instance_key" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
	ContactName string `json:"contact_name" binding:"required"`
	ContactPhone string `json:"contact_phone" binding:"required"`
	ReplyTo     string `json:"reply_to,omitempty"`
}

// MessageResponse represents the response from message sending
type MessageResponse struct {
	Status    string `json:"status"`
	MessageID string `json:"message_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// IncomingMessage represents an incoming WhatsApp message
type IncomingMessage struct {
	InstanceKey string    `json:"instance_key"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Message     string    `json:"message"`
	MessageType string    `json:"message_type"`
	Timestamp   time.Time `json:"timestamp"`
	MessageID   string    `json:"message_id"`
}

var instanceManager *InstanceManager

func main() {
	// Initialize instance manager
	instanceManager = &InstanceManager{
		Instances: make(map[string]*Instance),
	}

	// Setup Gin router
	r := gin.Default()

	// Create new instance endpoint
	r.POST("/instance/create", createInstance)

	// Connect instance endpoint
	r.POST("/instance/connect", connectInstance)

	// QR code endpoint
	r.GET("/instance/:instanceKey/qr", getQRCode)

	// Status endpoint for specific instance
	r.GET("/instance/:instanceKey/status", getInstanceStatus)

	// List all instances endpoint
	r.GET("/instances", listInstances)

	// Disconnect instance endpoint
	r.POST("/instance/:instanceKey/disconnect", disconnectInstance)

	// Message sending endpoints
	r.POST("/message/send", sendTextMessage)
	r.POST("/message/send-media", sendMediaMessage)
	r.POST("/message/send-contact", sendContactMessage)
	r.POST("/message/send-voice", sendVoiceMessage)
	r.POST("/message/send-location", sendLocationMessage)

	// Webhook endpoint for incoming messages
	r.POST("/webhook", handleWebhook)

	// Static file server for media files
	r.Static("/media", "/app/media")

	// Start server
	go func() {
		log.Println("Starting Multi-Instance Go WhatsApp Bridge on port 4444")
		r.Run(":4444")
	}()

	// Keep the main function running
	select {}
}

// generateInstanceKey generates a unique instance key
func generateInstanceKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// createInstance creates a new WhatsApp instance
func createInstance(c *gin.Context) {
	instanceKey := generateInstanceKey()

	// Setup database for this instance
	dbLog := waLog.Stdout(fmt.Sprintf("Database-%s", instanceKey), "DEBUG", true)
	dbPath := fmt.Sprintf("./whatsapp_%s.db?_foreign_keys=on", instanceKey)
	
	container, err := sqlstore.New(context.Background(), "sqlite3", dbPath, dbLog)
	if err != nil {
		log.Printf("Error creating database container for instance %s: %v", instanceKey, err)
		c.JSON(500, gin.H{"error": "Failed to create database"})
		return
	}

	deviceStore, err := container.GetFirstDevice(context.Background())
	if err != nil {
		log.Printf("Error getting device store for instance %s: %v", instanceKey, err)
		c.JSON(500, gin.H{"error": "Failed to get device store"})
		return
	}

	// Create client
	client := whatsmeow.NewClient(deviceStore, waLog.Stdout(fmt.Sprintf("Client-%s", instanceKey), "DEBUG", true))
	
	// Create instance
	instance := &Instance{
		ID:          instanceKey,
		Client:      client,
		PhoneNumber: "",
		IsConnected: false,
		QRCodeChan:  make(chan string, 1),
		Container:   container,
	}

	// Add event handler
	client.AddEventHandler(func(evt interface{}) {
		handleInstanceEvents(instanceKey, evt)
	})

	// Add to instance manager
	instanceManager.Mutex.Lock()
	instanceManager.Instances[instanceKey] = instance
	instanceManager.Mutex.Unlock()

	log.Printf("Created new instance: %s", instanceKey)

	c.JSON(200, ConnectResponse{
		Status:      "instance_created",
		InstanceKey: instanceKey,
		Message:     "Instance created successfully",
	})
}

// connectInstance connects a specific instance
func connectInstance(c *gin.Context) {
	var req ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[req.InstanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.Lock()
	defer instance.Mutex.Unlock()

	if instance.IsConnected {
		c.JSON(200, ConnectResponse{
			Status:      "already_connected",
			InstanceKey: req.InstanceKey,
			Message:     "Instance is already connected",
		})
		return
	}

	// Check if already logged in
	if instance.Client.IsLoggedIn() {
		instance.IsConnected = true
		// Get phone number
		if instance.Client.Store.ID != nil {
			instance.PhoneNumber = instance.Client.Store.ID.User
		}
		
		c.JSON(200, ConnectResponse{
			Status:      "already_logged_in",
			InstanceKey: req.InstanceKey,
			Message:     "Instance is already logged in",
		})
		return
	}

	// Get QR channel
	qrChan, _ := instance.Client.GetQRChannel(context.Background())
	err := instance.Client.Connect()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Wait for QR code
	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				instance.QRCodeChan <- evt.Code
				break
			}
		}
	}()

	c.JSON(200, ConnectResponse{
		Status:      "qr_generated",
		InstanceKey: req.InstanceKey,
		Message:     "QR code generated, scan to connect",
	})
}

// getQRCode returns the QR code for a specific instance
func getQRCode(c *gin.Context) {
	instanceKey := c.Param("instanceKey")

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[instanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.RLock()
	if instance.IsConnected {
		instance.Mutex.RUnlock()
		c.JSON(200, gin.H{"status": "connected", "message": "Instance is already connected"})
		return
	}
	instance.Mutex.RUnlock()

	select {
	case qr := <-instance.QRCodeChan:
		// Generate QR code
		qrCode, err := qrcode.Encode(qr, qrcode.Medium, 256)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate QR code"})
			return
		}
		c.Data(200, "image/png", qrCode)
	case <-time.After(30 * time.Second):
		c.JSON(408, gin.H{"error": "QR code timeout"})
	}
}

// getInstanceStatus returns the status of a specific instance
func getInstanceStatus(c *gin.Context) {
	instanceKey := c.Param("instanceKey")

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[instanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.RLock()
	defer instance.Mutex.RUnlock()

	c.JSON(200, gin.H{
		"instance_key": instanceKey,
		"connected":    instance.IsConnected,
		"logged_in":    instance.Client.IsLoggedIn(),
		"phone_number": instance.PhoneNumber,
	})
}

// listInstances returns all instances
func listInstances(c *gin.Context) {
	instanceManager.Mutex.RLock()
	defer instanceManager.Mutex.RUnlock()

	instances := make([]gin.H, 0)
	for key, instance := range instanceManager.Instances {
		instance.Mutex.RLock()
		instances = append(instances, gin.H{
			"instance_key": key,
			"connected":    instance.IsConnected,
			"logged_in":    instance.Client.IsLoggedIn(),
			"phone_number": instance.PhoneNumber,
		})
		instance.Mutex.RUnlock()
	}

	c.JSON(200, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

// disconnectInstance disconnects a specific instance
func disconnectInstance(c *gin.Context) {
	instanceKey := c.Param("instanceKey")

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[instanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.Lock()
	defer instance.Mutex.Unlock()

	if instance.Client != nil {
		instance.Client.Disconnect()
	}
	instance.IsConnected = false

	c.JSON(200, gin.H{
		"status":       "disconnected",
		"instance_key": instanceKey,
		"message":      "Instance disconnected successfully",
	})
}

// handleInstanceEvents handles events for a specific instance
func handleInstanceEvents(instanceKey string, evt interface{}) {
	// Determine event type and send appropriate webhook
	eventType := getEventType(evt)
	sendWebhook(eventType, evt, instanceKey)
	
	// Handle connection events - check for successful login
	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[instanceKey]
	instanceManager.Mutex.RUnlock()
	
	if exists && instance.Client.IsLoggedIn() {
		instance.Mutex.Lock()
		instance.IsConnected = true
		if instance.Client.Store.ID != nil {
			instance.PhoneNumber = instance.Client.Store.ID.User
		}
		instance.Mutex.Unlock()
		
		log.Printf("Instance %s connected with phone number: %s", instanceKey, instance.PhoneNumber)
	}
}

// getEventType determines the type of event and returns a descriptive string
func getEventType(evt interface{}) string {
	switch e := evt.(type) {
	case *events.Message:
		// Check for protocol messages (revoke, edit, etc.)
		if protocolMsg := e.Message.GetProtocolMessage(); protocolMsg != nil {
			switch protocolMsg.GetType().String() {
			case "REVOKE":
				return "message_revoked"
			case "MESSAGE_EDIT":
				return "message_edited"
			}
		}
		
		// Check message content type
		if e.Message.GetConversation() != "" || e.Message.GetExtendedTextMessage() != nil {
			return "message_received"
		}
		if e.Message.GetImageMessage() != nil {
			return "image_received"
		}
		if e.Message.GetVideoMessage() != nil {
			return "video_received"
		}
		if e.Message.GetAudioMessage() != nil {
			return "audio_received"
		}
		if e.Message.GetDocumentMessage() != nil {
			return "document_received"
		}
		if e.Message.GetStickerMessage() != nil {
			return "sticker_received"
		}
		if e.Message.GetContactMessage() != nil {
			return "contact_received"
		}
		if e.Message.GetLocationMessage() != nil {
			return "location_received"
		}
		if e.Message.GetLiveLocationMessage() != nil {
			return "live_location_received"
		}
		if e.Message.GetListMessage() != nil {
			return "list_received"
		}
		if e.Message.GetOrderMessage() != nil {
			return "order_received"
		}
		return "message_received"
		
	case *events.Receipt:
		switch e.Type {
		case types.ReceiptTypeDelivered:
			return "message_delivered"
		case types.ReceiptTypeRead:
			return "message_read"
		case types.ReceiptTypeReadSelf:
			return "message_read_self"
		case types.ReceiptTypePlayed:
			return "message_played"
		case types.ReceiptTypePlayedSelf:
			return "message_played_self"
		case types.ReceiptTypeSender:
			return "message_sender_receipt"
		case types.ReceiptTypeRetry:
			return "message_retry"
		default:
			return "message_receipt"
		}
		
	case *events.DeleteForMe:
		return "message_deleted"
		
	case *events.Presence:
		return "presence_update"
		
	case *events.ChatPresence:
		return "chat_presence_update"
		
	case *events.Connected:
		return "connected"
		
	case *events.Disconnected:
		return "disconnected"
		
	case *events.LoggedOut:
		return "logged_out"
		
	case *events.PairSuccess:
		return "pair_success"
		
	case *events.PushNameSetting:
		return "push_name_setting"
		
	case *events.StreamReplaced:
		return "stream_replaced"
		
	case *events.HistorySync:
		return "history_sync"
		
	case *events.AppState:
		return "app_state_update"
		
	case *events.AppStateSyncComplete:
		return "app_state_sync_complete"
		
	case *events.GroupInfo:
		return "group_info_update"
		
	case *events.PushName:
		return "push_name_update"
		
	default:
		return "unknown_event"
	}
}

// ExtractedMedia represents extracted media information
type ExtractedMedia struct {
	MediaPath string `json:"media_path"`
	MimeType  string `json:"mime_type"`
	Caption   string `json:"caption"`
	URL       string `json:"url"`
}

// downloadMedia downloads media from WhatsApp and saves it to the media volume
func downloadMedia(ctx context.Context, client *whatsmeow.Client, mediaFile whatsmeow.DownloadableMessage, instanceKey string) (*ExtractedMedia, error) {
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

	extractedMedia := &ExtractedMedia{}

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

// sendWebhook sends webhook data to Node.js with instance information
func sendWebhook(eventType string, data interface{}, instanceKey string) {
	// Check if this is a media message and download if needed
	var enhancedData interface{}
	
	if msgEvent, ok := data.(*events.Message); ok {
		// Start with the full raw event data
		enhancedData = data
		
		// Check for media and download if present
		instanceManager.Mutex.RLock()
		instance, exists := instanceManager.Instances[instanceKey]
		instanceManager.Mutex.RUnlock()
		
		if exists && instance.Client != nil {
			ctx := context.Background()
			
			// Check for different media types and download them
			if img := msgEvent.Message.GetImageMessage(); img != nil {
				if extractedMedia, err := downloadMedia(ctx, instance.Client, img, instanceKey); err == nil {
					// Add media download info to the raw data
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event": data,
						"message": msgEvent.Message,
						"info": msgEvent.Info,
						"source_string": msgEvent.Info.SourceString(),
						"push_name": msgEvent.Info.PushName,
						"is_from_me": msgEvent.Info.IsFromMe,
						"is_group": msgEvent.Info.Chat.Server == "g.us",
						
						// Media download information
						"media_type": "image",
						"media_url": extractedMedia.URL,
						"media_path": extractedMedia.MediaPath,
						"mime_type": extractedMedia.MimeType,
						"caption": extractedMedia.Caption,
						"local_file_url": fmt.Sprintf("http://localhost:4444%s", extractedMedia.URL),
					}
				}
			} else if vid := msgEvent.Message.GetVideoMessage(); vid != nil {
				if extractedMedia, err := downloadMedia(ctx, instance.Client, vid, instanceKey); err == nil {
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event": data,
						"message": msgEvent.Message,
						"info": msgEvent.Info,
						"source_string": msgEvent.Info.SourceString(),
						"push_name": msgEvent.Info.PushName,
						"is_from_me": msgEvent.Info.IsFromMe,
						"is_group": msgEvent.Info.Chat.Server == "g.us",
						
						// Media download information
						"media_type": "video",
						"media_url": extractedMedia.URL,
						"media_path": extractedMedia.MediaPath,
						"mime_type": extractedMedia.MimeType,
						"caption": extractedMedia.Caption,
						"local_file_url": fmt.Sprintf("http://localhost:4444%s", extractedMedia.URL),
					}
				}
			} else if aud := msgEvent.Message.GetAudioMessage(); aud != nil {
				if extractedMedia, err := downloadMedia(ctx, instance.Client, aud, instanceKey); err == nil {
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event": data,
						"message": msgEvent.Message,
						"info": msgEvent.Info,
						"source_string": msgEvent.Info.SourceString(),
						"push_name": msgEvent.Info.PushName,
						"is_from_me": msgEvent.Info.IsFromMe,
						"is_group": msgEvent.Info.Chat.Server == "g.us",
						
						// Media download information
						"media_type": "audio",
						"media_url": extractedMedia.URL,
						"media_path": extractedMedia.MediaPath,
						"mime_type": extractedMedia.MimeType,
						"local_file_url": fmt.Sprintf("http://localhost:4444%s", extractedMedia.URL),
					}
				}
			} else if doc := msgEvent.Message.GetDocumentMessage(); doc != nil {
				if extractedMedia, err := downloadMedia(ctx, instance.Client, doc, instanceKey); err == nil {
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event": data,
						"message": msgEvent.Message,
						"info": msgEvent.Info,
						"source_string": msgEvent.Info.SourceString(),
						"push_name": msgEvent.Info.PushName,
						"is_from_me": msgEvent.Info.IsFromMe,
						"is_group": msgEvent.Info.Chat.Server == "g.us",
						
						// Media download information
						"media_type": "document",
						"media_url": extractedMedia.URL,
						"media_path": extractedMedia.MediaPath,
						"mime_type": extractedMedia.MimeType,
						"caption": extractedMedia.Caption,
						"filename": doc.GetFileName(),
						"local_file_url": fmt.Sprintf("http://localhost:4444%s", extractedMedia.URL),
					}
				}
			} else if stk := msgEvent.Message.GetStickerMessage(); stk != nil {
				if extractedMedia, err := downloadMedia(ctx, instance.Client, stk, instanceKey); err == nil {
					enhancedData = map[string]interface{}{
						// Full raw event data
						"raw_event": data,
						"message": msgEvent.Message,
						"info": msgEvent.Info,
						"source_string": msgEvent.Info.SourceString(),
						"push_name": msgEvent.Info.PushName,
						"is_from_me": msgEvent.Info.IsFromMe,
						"is_group": msgEvent.Info.Chat.Server == "g.us",
						
						// Media download information
						"media_type": "sticker",
						"media_url": extractedMedia.URL,
						"media_path": extractedMedia.MediaPath,
						"mime_type": extractedMedia.MimeType,
						"local_file_url": fmt.Sprintf("http://localhost:4444%s", extractedMedia.URL),
					}
				}
			}
		}
	} else {
		enhancedData = data
	}

	payload := WebhookPayload{
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



// sendTextMessage sends a text message to a specific phone number
func sendTextMessage(c *gin.Context) {
	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[req.InstanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.RLock()
	if !instance.IsConnected {
		instance.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	instance.Mutex.RUnlock()

	// Parse phone number to JID
	recipient, err := types.ParseJID(req.Phone)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid phone number format"})
		return
	}

	// Create text message
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(req.Message),
		},
	}

	// Add reply context if provided
	if req.ReplyTo != "" {
		msg.ExtendedTextMessage.ContextInfo = &waE2E.ContextInfo{
			StanzaID: proto.String(req.ReplyTo),
		}
	}

	// Send message
	resp, err := instance.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

// sendMediaMessage sends a media message (image, audio, video, file) to a specific phone number
func sendMediaMessage(c *gin.Context) {
	var req MediaMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[req.InstanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.RLock()
	if !instance.IsConnected {
		instance.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	instance.Mutex.RUnlock()

	// Parse phone number to JID
	recipient, err := types.ParseJID(req.Phone)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid phone number format"})
		return
	}

	// Download media from URL
	httpResp, err := http.Get(req.URL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to download media from URL"})
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to download media: %d", httpResp.StatusCode)})
		return
	}

	mediaData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read media data"})
		return
	}

	// Create media directory if it doesn't exist
	mediaDir := "./media"
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create media directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s_%d", req.Type, time.Now().Unix())
	var filepath string
	var msg *waE2E.Message

	switch req.Type {
	case "image":
		filepath = fmt.Sprintf("%s/%s.jpg", mediaDir, filename)
		if err := os.WriteFile(filepath, mediaData, 0644); err != nil {
			c.JSON(500, gin.H{"error": "Failed to save image"})
			return
		}
		defer os.Remove(filepath) // Clean up after sending

		// Upload image to WhatsApp
		uploaded, err := instance.Client.Upload(context.Background(), mediaData, whatsmeow.MediaImage)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to upload image"})
			return
		}

		msg = &waE2E.Message{
			ImageMessage: &waE2E.ImageMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				Caption:       proto.String(req.Caption),
			},
		}

	case "audio":
		filepath = fmt.Sprintf("%s/%s.mp3", mediaDir, filename)
		if err := os.WriteFile(filepath, mediaData, 0644); err != nil {
			c.JSON(500, gin.H{"error": "Failed to save audio"})
			return
		}
		defer os.Remove(filepath) // Clean up after sending

		// Upload audio to WhatsApp
		uploaded, err := instance.Client.Upload(context.Background(), mediaData, whatsmeow.MediaAudio)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to upload audio"})
			return
		}

		// Detect mimetype
		mimeType := http.DetectContentType(mediaData)
		
		// Check if this is a PTT (voice recording) or regular audio file
		isPTT := req.IsPTT
		
		msg = &waE2E.Message{
			AudioMessage: &waE2E.AudioMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				Mimetype:      proto.String(mimeType),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				PTT:           proto.Bool(isPTT),
			},
		}

	case "video":
		filepath = fmt.Sprintf("%s/%s.mp4", mediaDir, filename)
		if err := os.WriteFile(filepath, mediaData, 0644); err != nil {
			c.JSON(500, gin.H{"error": "Failed to save video"})
			return
		}
		defer os.Remove(filepath) // Clean up after sending

		// Upload video to WhatsApp
		uploaded, err := instance.Client.Upload(context.Background(), mediaData, whatsmeow.MediaVideo)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to upload video"})
			return
		}

		msg = &waE2E.Message{
			VideoMessage: &waE2E.VideoMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				Caption:       proto.String(req.Caption),
			},
		}

	case "file":
		filepath = fmt.Sprintf("%s/%s", mediaDir, filename)
		if err := os.WriteFile(filepath, mediaData, 0644); err != nil {
			c.JSON(500, gin.H{"error": "Failed to save file"})
			return
		}
		defer os.Remove(filepath) // Clean up after sending

		// Upload document to WhatsApp
		uploaded, err := instance.Client.Upload(context.Background(), mediaData, whatsmeow.MediaDocument)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to upload document"})
			return
		}

		msg = &waE2E.Message{
			DocumentMessage: &waE2E.DocumentMessage{
				URL:           proto.String(uploaded.URL),
				DirectPath:    proto.String(uploaded.DirectPath),
				MediaKey:      uploaded.MediaKey,
				FileEncSHA256: uploaded.FileEncSHA256,
				FileSHA256:    uploaded.FileSHA256,
				FileLength:    proto.Uint64(uploaded.FileLength),
				FileName:      proto.String(filename),
			},
		}

	default:
		c.JSON(400, gin.H{"error": "Invalid media type"})
		return
	}

	// Send message
	resp, err := instance.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

// sendContactMessage sends a contact message to a specific phone number
func sendContactMessage(c *gin.Context) {
	var req ContactMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[req.InstanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.RLock()
	if !instance.IsConnected {
		instance.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	instance.Mutex.RUnlock()

	// Parse phone number to JID
	recipient, err := types.ParseJID(req.Phone)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid phone number format"})
		return
	}

	// Parse contact phone number to extract the number without @s.whatsapp.net
	contactPhoneNumber := req.ContactPhone
	if strings.Contains(contactPhoneNumber, "@s.whatsapp.net") {
		contactPhoneNumber = strings.Split(contactPhoneNumber, "@")[0]
	}

	// Create contact message with proper vCard format
	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nN:;%s;;;\nFN:%s\nTEL;type=CELL;type=VOICE;waid=%s:+%s\nEND:VCARD",
		req.ContactName,
		req.ContactName,
		contactPhoneNumber,
		contactPhoneNumber)

	msg := &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: proto.String(req.ContactName),
			Vcard:       proto.String(vcard),
		},
	}

	// Add reply context if provided
	if req.ReplyTo != "" {
		msg.ContactMessage.ContextInfo = &waE2E.ContextInfo{
			StanzaID: proto.String(req.ReplyTo),
		}
	}

	// Send message
	resp, err := instance.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

// sendVoiceMessage sends a voice recording (PTT) message to a specific phone number
func sendVoiceMessage(c *gin.Context) {
	var req VoiceMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[req.InstanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.RLock()
	if !instance.IsConnected {
		instance.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	instance.Mutex.RUnlock()

	// Parse phone number to JID
	recipient, err := types.ParseJID(req.Phone)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid phone number format"})
		return
	}

	// Download audio from URL
	httpResp, err := http.Get(req.URL)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to download audio from URL"})
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to download audio: %d", httpResp.StatusCode)})
		return
	}

	mediaData, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to read audio data"})
		return
	}

	// Create media directory if it doesn't exist
	mediaDir := "./media"
	if err := os.MkdirAll(mediaDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create media directory"})
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("voice_%d", time.Now().Unix())
	filepath := fmt.Sprintf("%s/%s.ogg", mediaDir, filename)
	if err := os.WriteFile(filepath, mediaData, 0644); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save voice recording"})
		return
	}
	defer os.Remove(filepath) // Clean up after sending

	// Upload voice recording to WhatsApp
	uploaded, err := instance.Client.Upload(context.Background(), mediaData, whatsmeow.MediaAudio)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to upload voice recording"})
		return
	}

	// Detect mimetype
	mimeType := http.DetectContentType(mediaData)
	
	// Create voice message (PTT = true for voice recordings)
	msg := &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           proto.String(uploaded.URL),
			DirectPath:    proto.String(uploaded.DirectPath),
			Mimetype:      proto.String(mimeType),
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    proto.Uint64(uploaded.FileLength),
			PTT:           proto.Bool(true), // Always true for voice recordings
		},
	}

	// Add reply context if provided
	if req.ReplyTo != "" {
		msg.AudioMessage.ContextInfo = &waE2E.ContextInfo{
			StanzaID: proto.String(req.ReplyTo),
		}
	}

	// Send message
	resp, err := instance.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

// sendLocationMessage sends a location message to a specific phone number
func sendLocationMessage(c *gin.Context) {
	var req LocationMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[req.InstanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.RLock()
	if !instance.IsConnected {
		instance.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	instance.Mutex.RUnlock()

	// Parse phone number to JID
	recipient, err := types.ParseJID(req.Phone)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid phone number format"})
		return
	}

	// Create location message
	msg := &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  proto.Float64(req.Latitude),
			DegreesLongitude: proto.Float64(req.Longitude),
		},
	}

	// Add reply context if provided
	if req.ReplyTo != "" {
		msg.LocationMessage.ContextInfo = &waE2E.ContextInfo{
			StanzaID: proto.String(req.ReplyTo),
		}
	}

	// Send message
	resp, err := instance.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

// handleWebhook handles incoming webhooks from Node.js
func handleWebhook(c *gin.Context) {
	var msg IncomingMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(400, gin.H{"error": "Invalid webhook payload"})
		return
	}

	instanceManager.Mutex.RLock()
	instance, exists := instanceManager.Instances[msg.InstanceKey]
	instanceManager.Mutex.RUnlock()

	if !exists {
		log.Printf("Received webhook for unknown instance: %s", msg.InstanceKey)
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	instance.Mutex.RLock()
	if !instance.IsConnected {
		instance.Mutex.RUnlock()
		log.Printf("Received webhook for disconnected instance: %s", msg.InstanceKey)
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	instance.Mutex.RUnlock()

	instance.Mutex.Lock()
	defer instance.Mutex.Unlock()

	// Handle different message types
	switch msg.MessageType {
	case "text":
		handleTextMessage(instance, msg)
	case "image":
		handleImageMessage(instance, msg)
	case "audio":
		handleAudioMessage(instance, msg)
	case "video":
		handleVideoMessage(instance, msg)
	case "document":
		handleDocumentMessage(instance, msg)
	default:
		log.Printf("Received unknown message type: %s", msg.MessageType)
		c.JSON(200, gin.H{"status": "received", "message": "Message received", "type": msg.MessageType})
	}
}

// handleTextMessage handles incoming text messages
func handleTextMessage(instance *Instance, msg IncomingMessage) {
	// Parse recipient JID
	recipient, err := types.ParseJID(msg.From)
	if err != nil {
		log.Printf("Error parsing recipient JID %s: %v", msg.From, err)
		return
	}

	// Create text message
	waMsg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(msg.Message),
		},
	}

	// Send message
	resp, err := instance.Client.SendMessage(context.Background(), recipient, waMsg)
	if err != nil {
		log.Printf("Error sending text message to %s: %v", msg.From, err)
		sendWebhook("message_error", gin.H{"instance_key": msg.InstanceKey, "phone": msg.From, "error": err.Error()}, msg.InstanceKey)
		return
	}
	
	log.Printf("Sent text message to %s: %s (ID: %s)", msg.From, msg.Message, resp.ID)
	sendWebhook("message_sent", gin.H{"instance_key": msg.InstanceKey, "phone": msg.From, "message_id": resp.ID}, msg.InstanceKey)
}

// handleImageMessage handles incoming image messages
func handleImageMessage(instance *Instance, msg IncomingMessage) {
	// For simplicity, we'll just log the image URL.
	// In a real scenario, you'd download the image and send it as a media message.
	log.Printf("Received image message from %s: %s", msg.From, msg.Message)
	sendWebhook("message_received", gin.H{"instance_key": msg.InstanceKey, "from": msg.From, "message": msg.Message, "type": msg.MessageType}, msg.InstanceKey)
}

// handleAudioMessage handles incoming audio messages
func handleAudioMessage(instance *Instance, msg IncomingMessage) {
	// For simplicity, we'll just log the audio URL.
	// In a real scenario, you'd download the audio and send it as a media message.
	log.Printf("Received audio message from %s: %s", msg.From, msg.Message)
	sendWebhook("message_received", gin.H{"instance_key": msg.InstanceKey, "from": msg.From, "message": msg.Message, "type": msg.MessageType}, msg.InstanceKey)
}

// handleVideoMessage handles incoming video messages
func handleVideoMessage(instance *Instance, msg IncomingMessage) {
	// For simplicity, we'll just log the video URL.
	// In a real scenario, you'd download the video and send it as a media message.
	log.Printf("Received video message from %s: %s", msg.From, msg.Message)
	sendWebhook("message_received", gin.H{"instance_key": msg.InstanceKey, "from": msg.From, "message": msg.Message, "type": msg.MessageType}, msg.InstanceKey)
}

// handleDocumentMessage handles incoming document messages
func handleDocumentMessage(instance *Instance, msg IncomingMessage) {
	// For simplicity, we'll just log the document URL.
	// In a real scenario, you'd download the document and send it as a media message.
	log.Printf("Received document message from %s: %s", msg.From, msg.Message)
	sendWebhook("message_received", gin.H{"instance_key": msg.InstanceKey, "from": msg.From, "message": msg.Message, "type": msg.MessageType}, msg.InstanceKey)
}
