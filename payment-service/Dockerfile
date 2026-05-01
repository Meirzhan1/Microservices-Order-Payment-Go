FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the generated contracts first
COPY contracts-generated ./contracts-generated/
COPY contracts-proto ./contracts-proto/

# Copy payment-service
COPY payment-service/go.mod payment-service/go.sum ./payment-service/
RUN cd payment-service && go mod download

COPY payment-service ./payment-service

# Build
RUN cd payment-service && go build -o payment-service ./cmd/payment-service

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/payment-service/payment-service .
COPY --from=builder /app/payment-service/migrations ./migrations

CMD ["./payment-service"]
