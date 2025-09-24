package services

import (
	"log"
	"time"

	"multi-client-whatsapp/internal/instance"

	whatsappTypes "go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

// HandleInstanceEvents handles events for a specific instance
func HandleInstanceEvents(instanceKey string, evt interface{}) {
	// Determine event type and send appropriate webhook
	eventType := GetEventType(evt)
	SendWebhook(eventType, evt, instanceKey)

	// Handle connection events - check for successful login
	instance.Manager.Mutex.RLock()
	inst, exists := instance.Manager.Instances[instanceKey]
	instance.Manager.Mutex.RUnlock()

	if exists && inst.Client.IsLoggedIn() {
		inst.Mutex.Lock()
		inst.IsConnected = true
		if inst.Client.Store.ID != nil {
			inst.PhoneNumber = inst.Client.Store.ID.User
		}
		inst.Mutex.Unlock()

		log.Printf("Instance %s connected with phone number: %s", instanceKey, inst.PhoneNumber)

		// Send connection webhook
		connectionData := map[string]interface{}{
			"instance_key": instanceKey,
			"phone_number": inst.PhoneNumber,
			"status":       "connected",
			"timestamp":    time.Now(),
		}
		SendWebhook("instance_connected", connectionData, instanceKey)
	}

	// Handle disconnection events
	if _, ok := evt.(*events.Disconnected); ok {
		instance.Manager.Mutex.RLock()
		inst, exists := instance.Manager.Instances[instanceKey]
		instance.Manager.Mutex.RUnlock()

		if exists {
			inst.Mutex.Lock()
			inst.IsConnected = false
			inst.Mutex.Unlock()

			log.Printf("Instance %s disconnected", instanceKey)

			// Send disconnection webhook
			disconnectionData := map[string]interface{}{
				"instance_key": instanceKey,
				"phone_number": inst.PhoneNumber,
				"status":       "disconnected",
				"timestamp":    time.Now(),
			}
			SendWebhook("instance_disconnected", disconnectionData, instanceKey)
		}
	}
}

// GetEventType determines the type of event and returns a descriptive string
func GetEventType(evt interface{}) string {
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
		if e.Message.GetInteractiveMessage() != nil {
			return "interactive_message_received"
		}
		return "message_received"

	case *events.Receipt:
		switch e.Type {
		case whatsappTypes.ReceiptTypeDelivered:
			return "message_delivered"
		case whatsappTypes.ReceiptTypeRead:
			return "message_read"
		case whatsappTypes.ReceiptTypeReadSelf:
			return "message_read_self"
		case whatsappTypes.ReceiptTypePlayed:
			return "message_played"
		case whatsappTypes.ReceiptTypePlayedSelf:
			return "message_played_self"
		case whatsappTypes.ReceiptTypeSender:
			return "message_sender_receipt"
		case whatsappTypes.ReceiptTypeRetry:
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
