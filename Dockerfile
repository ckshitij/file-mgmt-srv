# Start from the official Go base image
FROM golang:1.24-alpine AS builder

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Create app directory
WORKDIR /app

# Copy go.mod and go.sum first for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the Go application
RUN go build -o file-mgmt-srv .

# Use a minimal final image
FROM alpine:latest
WORKDIR /root/

# Copy the binary
COPY --from=builder /app/file-mgmt-srv .
COPY --from=builder /app/resource ./resource

# (Optional) Override default port if you want
ENV SERVICE_PORT=8088

# Expose the same port
EXPOSE 8088

CMD ["./file-mgmt-srv"]