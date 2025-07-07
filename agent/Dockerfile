FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o agent cmd/main.go

FROM alpine:latest

# Install docker client and curl
RUN apk --no-cache add docker curl

WORKDIR /root/

COPY --from=builder /app/agent .

CMD ["./agent"] 