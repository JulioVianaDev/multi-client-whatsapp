package main

import (
	"log"

	"multi-client-whatsapp/internal/instance"
	"multi-client-whatsapp/internal/platform/router"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables")
	}

	// Initialize instance manager
	instance.InitializeManager()

	// Setup and run router
	r := router.SetupRouter()

	// Start server
	log.Println("Starting Multi-Instance Go WhatsApp Bridge on port 4444")
	if err := r.Run(":4444"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
