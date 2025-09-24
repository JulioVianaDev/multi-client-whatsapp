package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"multi-client-whatsapp/internal/instance"
	"multi-client-whatsapp/internal/platform/database"
	"multi-client-whatsapp/internal/services"
	"multi-client-whatsapp/internal/types"
	"multi-client-whatsapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	whatsappTypes "go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

func CreateInstance(c *gin.Context) {
	instanceKey := utils.GenerateInstanceKey()

	container, err := database.CreateDatabaseContainer(instanceKey)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create instance database"})
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
	inst := &types.Instance{
		ID:          instanceKey,
		Client:      client,
		PhoneNumber: "",
		IsConnected: false,
		QRCodeChan:  make(chan string, 1),
		Container:   container,
	}

	// Add event handler
	client.AddEventHandler(func(evt interface{}) {
		services.HandleInstanceEvents(instanceKey, evt)
	})

	// Add to instance manager
	instance.Manager.Mutex.Lock()
	instance.Manager.Instances[instanceKey] = inst
	instance.Manager.Mutex.Unlock()

	log.Printf("Created new instance: %s", instanceKey)

	c.JSON(200, types.ConnectResponse{
		Status:      "instance_created",
		InstanceKey: instanceKey,
		Message:     "Instance created successfully",
	})
}

func ConnectInstance(c *gin.Context) {
	var req types.ConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.Lock()
	defer inst.Mutex.Unlock()

	if inst.IsConnected {
		c.JSON(200, types.ConnectResponse{
			Status:      "already_connected",
			InstanceKey: req.InstanceKey,
			Message:     "Instance is already connected",
		})
		return
	}

	// Check if already logged in
	if inst.Client.IsLoggedIn() {
		inst.IsConnected = true
		// Get phone number
		if inst.Client.Store.ID != nil {
			inst.PhoneNumber = inst.Client.Store.ID.User
		}

		// Send manual connection webhook
		connectionData := map[string]interface{}{
			"instance_key": req.InstanceKey,
			"phone_number": inst.PhoneNumber,
			"status":       "manually_connected",
			"timestamp":    time.Now(),
		}
		services.SendWebhook("instance_manually_connected", connectionData, req.InstanceKey)

		c.JSON(200, types.ConnectResponse{
			Status:      "already_logged_in",
			InstanceKey: req.InstanceKey,
			Message:     "Instance is already logged in",
		})
		return
	}

	// Get QR channel
	qrChan, _ := inst.Client.GetQRChannel(context.Background())
	err := inst.Client.Connect()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// Wait for QR code
	go func() {
		for evt := range qrChan {
			if evt.Event == "code" {
				inst.QRCodeChan <- evt.Code
				break
			}
		}
	}()

	c.JSON(200, types.ConnectResponse{
		Status:      "qr_generated",
		InstanceKey: req.InstanceKey,
		Message:     "QR code generated, scan to connect",
	})
}

func GetQRCode(c *gin.Context) {
	instanceKey := c.Param("instanceKey")

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[instanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(200, gin.H{"status": "connected", "message": "Instance is already connected"})
		return
	}
	inst.Mutex.RUnlock()

	select {
	case qr := <-inst.QRCodeChan:
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

func GetInstanceStatus(c *gin.Context) {
	instanceKey := c.Param("instanceKey")

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[instanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	defer inst.Mutex.RUnlock()

	c.JSON(200, gin.H{
		"instance_key": instanceKey,
		"connected":    inst.IsConnected,
		"logged_in":    inst.Client.IsLoggedIn(),
		"phone_number": inst.PhoneNumber,
	})
}

func ListInstances(c *gin.Context) {
	instance.Manager.Mutex.RLock()
	defer instance.Manager.Mutex.RUnlock()

	instances := make([]gin.H, 0)
	for key, inst := range instance.Manager.Instances {
		inst.Mutex.RLock()
		instances = append(instances, gin.H{
			"instance_key": key,
			"connected":    inst.IsConnected,
			"logged_in":    inst.Client.IsLoggedIn(),
			"phone_number": inst.PhoneNumber,
		})
		inst.Mutex.RUnlock()
	}

	c.JSON(200, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

func DisconnectInstance(c *gin.Context) {
	instanceKey := c.Param("instanceKey")

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[instanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.Lock()
	defer inst.Mutex.Unlock()

	if inst.Client != nil {
		inst.Client.Disconnect()
	}
	inst.IsConnected = false

	// Send manual disconnection webhook
	disconnectionData := map[string]interface{}{
		"instance_key": instanceKey,
		"phone_number": inst.PhoneNumber,
		"status":       "manually_disconnected",
		"timestamp":    time.Now(),
	}
	services.SendWebhook("instance_manually_disconnected", disconnectionData, instanceKey)

	c.JSON(200, gin.H{
		"status":       "disconnected",
		"instance_key": instanceKey,
		"message":      "Instance disconnected successfully",
	})
}

func DeleteInstance(c *gin.Context) {
	instanceKey := c.Param("instanceKey")

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[instanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	// Disconnect the client first if it's connected
	inst.Mutex.Lock()
	if inst.Client != nil {
		inst.Client.Disconnect()
	}
	if inst.Container != nil {
		inst.Container.Close() // Close the database connection pool
	}
	inst.IsConnected = false
	inst.Mutex.Unlock()

	// Remove from instance manager
	instance.Manager.Mutex.Lock()
	delete(instance.Manager.Instances, instanceKey)
	instance.Manager.Mutex.Unlock()

	// Now, drop the database
	database.DropDatabase(instanceKey)

	// Delete media directory for this instance
	mediaDir := fmt.Sprintf("/app/media/%s", instanceKey)
	if err := os.RemoveAll(mediaDir); err != nil {
		log.Printf("Warning: Error deleting media directory %s: %v", mediaDir, err)
	} else {
		log.Printf("Deleted media directory: %s", mediaDir)
	}

	// Send deletion webhook
	deletionData := map[string]interface{}{
		"instance_key": instanceKey,
		"phone_number": inst.PhoneNumber,
		"status":       "deleted",
		"timestamp":    time.Now(),
	}
	services.SendWebhook("instance_deleted", deletionData, instanceKey)

	c.JSON(200, gin.H{
		"status":       "deleted",
		"instance_key": instanceKey,
		"message":      "Instance and all associated data deleted successfully",
	})
}

func ValidatePhone(c *gin.Context) {
	var req types.PhoneValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Validate and correct the phone number
	validPhone, err := services.ValidateAndCorrectPhone(req.Phone, inst)
	if err != nil {
		c.JSON(400, types.PhoneValidationResponse{
			Status:        "error",
			OriginalPhone: req.Phone,
			ValidPhone:    req.Phone,
			Exists:        false,
			Message:       err.Error(),
		})
		return
	}

	// Check if the validated phone exists
	phoneExists, _ := services.CheckPhoneExists(validPhone, inst)

	c.JSON(200, types.PhoneValidationResponse{
		Status:        "success",
		OriginalPhone: req.Phone,
		ValidPhone:    validPhone,
		Exists:        phoneExists,
		Message:       "Phone number validated successfully",
	})
}

func TestPhoneExists(c *gin.Context) {
	var req types.PhoneValidationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Test the phone number directly with WhatsApp API
	phoneToTest := req.Phone
	if !strings.Contains(phoneToTest, "@s.whatsapp.net") {
		phoneToTest = phoneToTest + "@s.whatsapp.net"
	}

	log.Printf("🧪 Testing phone existence directly: %s", phoneToTest)

	data, err := inst.Client.IsOnWhatsApp([]string{phoneToTest})
	if err != nil {
		c.JSON(500, gin.H{
			"error": fmt.Sprintf("WhatsApp API error: %v", err),
			"phone": phoneToTest,
		})
		return
	}

	log.Printf("📊 Raw WhatsApp API response: %+v", data)

	response := gin.H{
		"phone":        phoneToTest,
		"raw_response": data,
		"exists":       false,
		"details":      []gin.H{},
	}

	for _, v := range data {
		detail := gin.H{
			"jid":   v.JID,
			"is_in": v.IsIn,
		}
		response["details"] = append(response["details"].([]gin.H), detail)

		if v.IsIn {
			response["exists"] = true
		}
	}

	c.JSON(200, response)
}

func ConvertLIDToPhone(c *gin.Context) {
	var req types.LIDToPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Validate LID format
	if !strings.HasSuffix(req.LID, "@lid") {
		c.JSON(400, types.LIDToPhoneResponse{
			Status:  "error",
			LID:     req.LID,
			Exists:  false,
			Message: "Invalid LID format. LID must end with @lid",
		})
		return
	}

	// Parse LID
	lidJID, err := whatsappTypes.ParseJID(req.LID)
	if err != nil {
		c.JSON(400, types.LIDToPhoneResponse{
			Status:  "error",
			LID:     req.LID,
			Exists:  false,
			Message: fmt.Sprintf("Invalid LID format: %v", err),
		})
		return
	}

	// Check if LIDs store is available
	if inst.Client == nil || inst.Client.Store == nil || inst.Client.Store.LIDs == nil {
		c.JSON(500, types.LIDToPhoneResponse{
			Status:  "error",
			LID:     req.LID,
			Exists:  false,
			Message: "LID store not available",
		})
		return
	}

	// Get phone number for this LID
	ctx := context.Background()
	phoneNumber, err := inst.Client.Store.LIDs.GetPNForLID(ctx, lidJID)
	if err != nil {
		log.Printf("Error getting phone number for LID %s: %v", lidJID.String(), err)
		c.JSON(500, types.LIDToPhoneResponse{
			Status:  "error",
			LID:     req.LID,
			Exists:  false,
			Message: fmt.Sprintf("Failed to get phone number for LID: %v", err),
		})
		return
	}

	// Check if phone number exists
	if phoneNumber.IsEmpty() {
		c.JSON(200, types.LIDToPhoneResponse{
			Status:  "success",
			LID:     req.LID,
			Exists:  false,
			Message: "LID found but no phone number associated",
		})
		return
	}

	// Return the phone number
	c.JSON(200, types.LIDToPhoneResponse{
		Status:      "success",
		LID:         req.LID,
		PhoneNumber: phoneNumber.String(),
		Exists:      true,
		Message:     "Phone number found for LID",
	})
}

func SendTextMessage(c *gin.Context) {
	var req types.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Validate and correct phone number
	validPhone, err := services.ValidateAndCorrectPhone(req.Phone, inst)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
		return
	}

	// Parse phone number to JID
	recipient, err := whatsappTypes.ParseJID(validPhone)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
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
	resp, err := inst.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, types.MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

func SendMediaMessage(c *gin.Context) {
	var req types.MediaMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Validate and correct phone number
	validPhone, err := services.ValidateAndCorrectPhone(req.Phone, inst)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
		return
	}

	// Parse phone number to JID
	recipient, err := whatsappTypes.ParseJID(validPhone)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
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
		uploaded, err := inst.Client.Upload(context.Background(), mediaData, whatsmeow.MediaImage)
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
		uploaded, err := inst.Client.Upload(context.Background(), mediaData, whatsmeow.MediaAudio)
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
		uploaded, err := inst.Client.Upload(context.Background(), mediaData, whatsmeow.MediaVideo)
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
		uploaded, err := inst.Client.Upload(context.Background(), mediaData, whatsmeow.MediaDocument)
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
	resp, err := inst.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, types.MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

func SendContactMessage(c *gin.Context) {
	var req types.ContactMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Validate and correct phone number
	validPhone, err := services.ValidateAndCorrectPhone(req.Phone, inst)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
		return
	}

	// Parse phone number to JID
	recipient, err := whatsappTypes.ParseJID(validPhone)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
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
	resp, err := inst.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, types.MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

func SendVoiceMessage(c *gin.Context) {
	var req types.VoiceMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Validate and correct phone number
	validPhone, err := services.ValidateAndCorrectPhone(req.Phone, inst)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
		return
	}

	// Parse phone number to JID
	recipient, err := whatsappTypes.ParseJID(validPhone)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
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
	uploaded, err := inst.Client.Upload(context.Background(), mediaData, whatsmeow.MediaAudio)
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
	resp, err := inst.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, types.MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

func SendLocationMessage(c *gin.Context) {
	var req types.LocationMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Validate and correct phone number
	validPhone, err := services.ValidateAndCorrectPhone(req.Phone, inst)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
		return
	}

	// Parse phone number to JID
	recipient, err := whatsappTypes.ParseJID(validPhone)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
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
	resp, err := inst.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, types.MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
	})
}

func SendInteractiveMessage(c *gin.Context) {
	var req types.InteractiveMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[req.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	// Validate and correct phone number
	validPhone, err := services.ValidateAndCorrectPhone(req.Phone, inst)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
		return
	}

	// Parse phone number to JID
	recipient, err := whatsappTypes.ParseJID(validPhone)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Invalid phone number format: %v", err)})
		return
	}

	// Validate buttons (max 3 buttons allowed)
	if len(req.Buttons) > 3 {
		c.JSON(400, gin.H{"error": "Maximum 3 buttons allowed"})
		return
	}

	if len(req.Buttons) == 0 {
		c.JSON(400, gin.H{"error": "At least one button is required"})
		return
	}

	// For now, send a simple text message with button information
	// TODO: Implement proper interactive message when protobuf structure is confirmed
	buttonText := "Available options:\n"
	for i, button := range req.Buttons {
		buttonText += fmt.Sprintf("%d. %s\n", i+1, button.Title)
	}

	messageText := fmt.Sprintf("**%s**\n\n%s\n\n%s\n\n%s", req.Title, req.Body, buttonText, req.Footer)

	// Create text message with button information
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(messageText),
		},
	}

	// Add reply context if provided
	if req.ReplyTo != "" {
		msg.ExtendedTextMessage.ContextInfo = &waE2E.ContextInfo{
			StanzaID: proto.String(req.ReplyTo),
		}
	}

	// Send message
	resp, err := inst.Client.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, types.MessageResponse{
		Status:    "sent",
		MessageID: resp.ID,
		Error:     "Note: Interactive buttons not yet implemented. Sent as formatted text message.",
	})
}

func HandleWebhook(c *gin.Context) {
	var msg types.IncomingMessage
	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(400, gin.H{"error": "Invalid webhook payload"})
		return
	}

	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[msg.InstanceKey]
	instance.Manager.Mutex.RUnlock()

	if !exists {
		log.Printf("Received webhook for unknown instance: %s", msg.InstanceKey)
		c.JSON(404, gin.H{"error": "Instance not found"})
		return
	}

	inst.Mutex.RLock()
	if !inst.IsConnected {
		inst.Mutex.RUnlock()
		log.Printf("Received webhook for disconnected instance: %s", msg.InstanceKey)
		c.JSON(400, gin.H{"error": "Instance is not connected"})
		return
	}
	inst.Mutex.RUnlock()

	inst.Mutex.Lock()
	defer inst.Mutex.Unlock()

	// Handle different message types
	switch msg.MessageType {
	case "text":
		handleTextMessage(inst, msg)
	case "image":
		handleImageMessage(inst, msg)
	case "audio":
		handleAudioMessage(inst, msg)
	case "video":
		handleVideoMessage(inst, msg)
	case "document":
		handleDocumentMessage(inst, msg)
	default:
		log.Printf("Received unknown message type: %s", msg.MessageType)
		c.JSON(200, gin.H{"status": "received", "message": "Message received", "type": msg.MessageType})
	}
}

// handleTextMessage handles incoming text messages
func handleTextMessage(instance *types.Instance, msg types.IncomingMessage) {
	// Validate and correct recipient JID
	recipient, err := services.ValidateAndCorrectPhone(msg.From, instance)
	if err != nil {
		log.Printf("Error parsing recipient JID %s: %v", msg.From, err)
		return
	}

	// Parse recipient JID
	jid, err := whatsappTypes.ParseJID(recipient)
	if err != nil {
		log.Printf("Error parsing recipient JID %s: %v", recipient, err)
		return
	}

	// Create text message
	waMsg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text: proto.String(msg.Message),
		},
	}

	// Send message
	resp, err := instance.Client.SendMessage(context.Background(), jid, waMsg)
	if err != nil {
		log.Printf("Error sending text message to %s: %v", msg.From, err)
		services.SendWebhook("message_error", gin.H{"instance_key": msg.InstanceKey, "phone": msg.From, "error": err.Error()}, msg.InstanceKey)
		return
	}

	log.Printf("Sent text message to %s: %s (ID: %s)", msg.From, msg.Message, resp.ID)
	services.SendWebhook("message_sent", gin.H{"instance_key": msg.InstanceKey, "phone": msg.From, "message_id": resp.ID}, msg.InstanceKey)
}

// handleImageMessage handles incoming image messages
func handleImageMessage(instance *types.Instance, msg types.IncomingMessage) {
	// For simplicity, we'll just log the image URL.
	// In a real scenario, you'd download the image and send it as a media message.
	log.Printf("Received image message from %s: %s", msg.From, msg.Message)
	services.SendWebhook("message_received", gin.H{"instance_key": msg.InstanceKey, "from": msg.From, "message": msg.Message, "type": msg.MessageType}, msg.InstanceKey)
}

// handleAudioMessage handles incoming audio messages
func handleAudioMessage(instance *types.Instance, msg types.IncomingMessage) {
	// For simplicity, we'll just log the audio URL.
	// In a real scenario, you'd download the audio and send it as a media message.
	log.Printf("Received audio message from %s: %s", msg.From, msg.Message)
	services.SendWebhook("message_received", gin.H{"instance_key": msg.InstanceKey, "from": msg.From, "message": msg.Message, "type": msg.MessageType}, msg.InstanceKey)
}

// handleVideoMessage handles incoming video messages
func handleVideoMessage(instance *types.Instance, msg types.IncomingMessage) {
	// For simplicity, we'll just log the video URL.
	// In a real scenario, you'd download the video and send it as a media message.
	log.Printf("Received video message from %s: %s", msg.From, msg.Message)
	services.SendWebhook("message_received", gin.H{"instance_key": msg.InstanceKey, "from": msg.From, "message": msg.Message, "type": msg.MessageType}, msg.InstanceKey)
}

// handleDocumentMessage handles incoming document messages
func handleDocumentMessage(instance *types.Instance, msg types.IncomingMessage) {
	// For simplicity, we'll just log the document URL.
	// In a real scenario, you'd download the document and send it as a media message.
	log.Printf("Received document message from %s: %s", msg.From, msg.Message)
	services.SendWebhook("message_received", gin.H{"instance_key": msg.InstanceKey, "from": msg.From, "message": msg.Message, "type": msg.MessageType}, msg.InstanceKey)
}
