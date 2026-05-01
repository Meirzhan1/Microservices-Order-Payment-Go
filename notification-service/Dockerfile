FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY notification-service/go.mod notification-service/go.sum ./notification-service/
RUN cd notification-service && go mod download

COPY notification-service ./notification-service

RUN cd notification-service && go build -o notification-service ./cmd/notification-service

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/notification-service/notification-service .

CMD ["./notification-service"]
