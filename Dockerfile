# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o monitor

# Final stage
FROM alpine:latest
RUN apk add --no-cache docker-cli busybox
WORKDIR /app
COPY --from=builder /app/monitor .
EXPOSE 6969
VOLUME /var/run/docker.sock
CMD ["./monitor"]
