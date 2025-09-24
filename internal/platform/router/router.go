package router

import (
	"multi-client-whatsapp/internal/platform/web/handlers"
	"os"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Create new instance endpoint
	r.POST("/instance/create", handlers.CreateInstance)

	// Connect instance endpoint
	r.POST("/instance/connect", handlers.ConnectInstance)

	// QR code endpoint
	r.GET("/instance/:instanceKey/qr", handlers.GetQRCode)

	// Status endpoint for specific instance
	r.GET("/instance/:instanceKey/status", handlers.GetInstanceStatus)

	// List all instances endpoint
	r.GET("/instances", handlers.ListInstances)

	// Disconnect instance endpoint
	r.POST("/instance/:instanceKey/disconnect", handlers.DisconnectInstance)

	// Delete instance endpoint
	r.DELETE("/instance/:instanceKey", handlers.DeleteInstance)

	// Phone validation endpoint
	r.POST("/phone/validate", handlers.ValidatePhone)
	r.POST("/phone/test-exists", handlers.TestPhoneExists)
	r.POST("/phone/lid-to-phone", handlers.ConvertLIDToPhone)

	// Message sending endpoints
	r.POST("/message/send", handlers.SendTextMessage)
	r.POST("/message/send-media", handlers.SendMediaMessage)
	r.POST("/message/send-contact", handlers.SendContactMessage)
	r.POST("/message/send-voice", handlers.SendVoiceMessage)
	r.POST("/message/send-location", handlers.SendLocationMessage)
	r.POST("/message/send-interactive", handlers.SendInteractiveMessage)

	// Webhook endpoint for incoming messages
	r.POST("/webhook", handlers.HandleWebhook)

	// Static file server for media files
	mediaPath := os.Getenv("MEDIA_PATH")
	if mediaPath == "" {
		mediaPath = "/app/media"
	}
	r.Static("/media", mediaPath)

	return r
}
