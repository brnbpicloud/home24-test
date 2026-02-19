# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# Build the API
RUN CGO_ENABLED=0 GOOS=linux go build -o api_server ./api

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/api_server .

EXPOSE 4001

CMD ["./api_server"]
