# Build stage
FROM golang:1.20-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git build-base

WORKDIR /build

# Copy go.mod first
COPY go.mod ./
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

ENTRYPOINT ["/app/validator"]
