FROM golang:1.24

# Install build and runtime dependencies
RUN apt-get update && apt-get install -y \
    git \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

ENV CGO_ENABLED=1

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Update go.mod and build the application
RUN go mod tidy && CGO_ENABLED=1 go build -o main .

# Create directories for database and media
RUN mkdir -p /app/data && chmod 755 /app/data
RUN mkdir -p /app/media && chmod 755 /app/media

# Expose port
EXPOSE 4444

# Run the application
CMD ["./main"]
