# Build stage
FROM golang:1.26.1-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates, tzdata for HTTPS and time support
RUN apk --no-cache add ca-certificates tzdata

# Install Chromium for chromedp (headless browser)
RUN apk --no-cache add chromium

# Install Python3, pip and ffmpeg for yt-dlp (YouTube downloading)
RUN apk --no-cache add python3 py3-pip ffmpeg && \
    pip3 install --break-system-packages yt-dlp && \
    rm -rf /root/.cache/pip

# Copy binary
COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

# Expose port from env
ARG SERVER_PORT
EXPOSE ${SERVER_PORT}

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${SERVER_PORT}/health || exit 1

# Run
CMD ["./main"]
