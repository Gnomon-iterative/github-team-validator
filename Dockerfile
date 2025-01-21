# Build stage
FROM golang:1.20-alpine AS builder

# Install git and build dependencies
RUN apk add --no-cache git build-base

WORKDIR /build

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags='-extldflags=-static' -o validator

# Final stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy only the binary from builder
COPY --from=builder /build/validator /app/validator

ENTRYPOINT ["/app/validator"]
