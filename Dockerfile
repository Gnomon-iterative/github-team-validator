# Build stage
FROM golang:1.20-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git build-base

WORKDIR /build

# Copy go.mod and go.sum first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy the rest of the code
COPY . .

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static -w -s' -o validator

# Final stage
FROM alpine:3.18

# Add CA certificates for HTTPS
RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

WORKDIR /app

# Copy only the binary from builder
COPY --from=builder /build/validator /app/validator

# Verify the binary exists and is executable
RUN chmod +x /app/validator && \
    ls -la /app/validator

# Add environment variables
ENV INPUT_GITHUB-TOKEN=""
ENV INPUT_PR-NUMBER=""
ENV INPUT_ORGANIZATION=""

# Add a healthcheck
HEALTHCHECK --interval=5s --timeout=3s \
  CMD ps aux | grep validator || exit 1

ENTRYPOINT ["/app/validator"]
