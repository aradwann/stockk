# Start with the official Golang image
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and other necessary tools
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o stockk ./cmd/server/main.go

# Start a new stage from scratch
FROM alpine:latest

# Install certificates for HTTPS
RUN apk add --no-cache ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/stockk .

# Copy migration files
COPY --from=builder /app/migrations ./migrations

# Expose port for the application
EXPOSE 8080

# Command to run the executable
CMD ["./stockk"]