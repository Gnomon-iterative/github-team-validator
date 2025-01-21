FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /validator

FROM alpine:3.19

COPY --from=builder /validator /validator
ENTRYPOINT ["/validator"]
