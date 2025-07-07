FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o agent cmd/main.go

FROM alpine:latest

RUN apk --no-cache add docker curl

WORKDIR /root/

COPY --from=builder /app/agent .

CMD ["./agent"] 