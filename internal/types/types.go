package types

import (
	"sync"
	"time"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

// Instance represents a WhatsApp client instance
type Instance struct {
	ID          string
	Client      *whatsmeow.Client
	PhoneNumber string
	IsConnected bool
	QRCodeChan  chan string
	Container   *sqlstore.Container
	Mutex       sync.RWMutex
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
	InstanceKey  string `json:"instance_key" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	ContactName  string `json:"contact_name" binding:"required"`
	ContactPhone string `json:"contact_phone" binding:"required"`
	ReplyTo      string `json:"reply_to,omitempty"`
}

// InteractiveMessageRequest represents an interactive message sending request
type InteractiveMessageRequest struct {
	InstanceKey string   `json:"instance_key" binding:"required"`
	Phone       string   `json:"phone" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Body        string   `json:"body" binding:"required"`
	Footer      string   `json:"footer,omitempty"`
	Buttons     []Button `json:"buttons" binding:"required"`
	ReplyTo     string   `json:"reply_to,omitempty"`
}

// Button represents a button in an interactive message
type Button struct {
	ID    string `json:"id" binding:"required"`
	Title string `json:"title" binding:"required"`
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

// PhoneValidationRequest represents a phone validation request
type PhoneValidationRequest struct {
	InstanceKey string `json:"instance_key" binding:"required"`
	Phone       string `json:"phone" binding:"required"`
}

// PhoneValidationResponse represents the response from phone validation
type PhoneValidationResponse struct {
	Status        string `json:"status"`
	OriginalPhone string `json:"original_phone"`
	ValidPhone    string `json:"valid_phone"`
	Exists        bool   `json:"exists"`
	Message       string `json:"message,omitempty"`
}

// LIDToPhoneRequest represents a LID to phone conversion request
type LIDToPhoneRequest struct {
	InstanceKey string `json:"instance_key" binding:"required"`
	LID         string `json:"lid" binding:"required"`
}

// LIDToPhoneResponse represents the response from LID to phone conversion
type LIDToPhoneResponse struct {
	Status      string `json:"status"`
	LID         string `json:"lid"`
	PhoneNumber string `json:"phone_number,omitempty"`
	Exists      bool   `json:"exists"`
	Message     string `json:"message,omitempty"`
}

// ExtractedMedia represents extracted media information
type ExtractedMedia struct {
	MediaPath string `json:"media_path"`
	MimeType  string `json:"mime_type"`
	Caption   string `json:"caption"`
	URL       string `json:"url"`
}
