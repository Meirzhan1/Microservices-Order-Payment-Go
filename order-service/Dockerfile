FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy the generated contracts first (as they are external)
COPY contracts-generated ./contracts-generated/
COPY contracts-proto ./contracts-proto/

# Copy order-service
COPY order-service/go.mod order-service/go.sum ./order-service/
RUN cd order-service && go mod download

COPY order-service ./order-service

# Build
RUN cd order-service && go build -o order-service ./cmd/order-service

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/order-service/order-service .
COPY --from=builder /app/order-service/migrations ./migrations

CMD ["./order-service"]
