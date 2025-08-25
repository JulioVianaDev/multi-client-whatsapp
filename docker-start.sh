#!/bin/bash

echo "🐳 Starting WhatsApp Bridge Project with Docker Compose..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Error: Docker is not running. Please start Docker first."
    exit 1
fi

# Check if Docker Compose is available
if ! command -v docker-compose &> /dev/null; then
    echo "❌ Error: Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

echo "🔨 Building and starting services..."
echo ""

# Build and start services
docker-compose up --build -d

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Services started successfully!"
    echo ""
    echo "📊 Service Status:"
    docker-compose ps
    echo ""
    echo "🌐 Access URLs:"
    echo "   Go WhatsApp Bridge: http://localhost:4444"
    echo "   Node.js Webhook Receiver: http://localhost:5555"
    echo ""
    echo "📋 Instructions:"
    echo "   1. Visit http://localhost:4444/qr to get QR code"
    echo "   2. Scan QR code with WhatsApp mobile app"
    echo "   3. Visit http://localhost:5555/scan for instructions"
    echo "   4. Monitor events: docker-compose logs -f webhook-receiver"
    echo ""
    echo "🔧 Useful Commands:"
    echo "   View logs: docker-compose logs -f"
    echo "   Stop services: docker-compose down"
    echo "   Restart services: docker-compose restart"
    echo "   View service status: docker-compose ps"
else
    echo "❌ Error: Failed to start services"
    exit 1
fi
