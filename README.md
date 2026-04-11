# AP2 Assignment 2 — gRPC Migration & Contract-First

This repository contains the solution for Assignment 2, demonstrating the migration from REST inter-service communication to gRPC using a Contract-First approach.

## 🔗 Contract-First Repositories
The gRPC contracts (`.proto` files) and their generated Go code are maintained in external repositories as per clean architecture principles:
- **Proto definitions:** [microservices-order-payment-contracts-proto](https://github.com/Meirzhan1/microservices-order-payment-contracts-proto)
- **Generated Go code:** [microservices-order-payment-contracts-generated](https://github.com/Meirzhan1/microservices-order-payment-contracts-generated)

## 🚀 Features & Requirements Covered
- **Order Service:** Keeps the external REST API but now calls Payment Service internally via a gRPC client.
- **Payment Service:** Exposes a gRPC server (`ProcessPayment`) instead of REST.
- **Server-Side Streaming:** The Order service hosts a gRPC stream (`SubscribeToOrderUpdates`) that pushes real-time status updates based on actual database changes.
- **Interceptors:** Implemented a Unary Interceptor on the Payment server to log method names and execution duration.
- **Proto Standards:** Used `go_package` and appropriate types like `google.protobuf.Timestamp`.
- **Environment Variables:** All ports and DSNs are configurable.

## 🏛 Architecture
```mermaid
flowchart LR
  Client --> OREST[Order REST :8081]
  OREST --> OUC[Order UseCase]
  OUC --> ODB[(order_db)]
  OUC -->|gRPC ProcessPayment| PGRPC[Payment gRPC :50051]
  PGRPC --> PUC[Payment UseCase]
  PUC --> PDB[(payment_db)]

  StreamClient --> OGRPC[Order gRPC :50052 Stream Updates]
  OGRPC --> OUC
```

## 🛠 How To Run

### 1. Database Migrations
```bash
psql "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable" -f order-service/migrations/001_create_orders.sql
psql "postgres://postgres:postgres@localhost:5432/payment_db?sslmode=disable" -f payment-service/migrations/001_create_payments.sql
```

### 2. Start Services
**Payment Service:**
```bash
cd payment-service
set PAYMENT_SERVICE_PORT=8082
set PAYMENT_SERVICE_GRPC_PORT=50051
set PAYMENT_DB_DSN=postgres://postgres:postgres@localhost:5432/payment_db?sslmode=disable
go run ./cmd/payment-service
```

**Order Service:**
```bash
cd order-service
set ORDER_SERVICE_PORT=8081
set ORDER_SERVICE_GRPC_PORT=50052
set ORDER_DB_DSN=postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable
set PAYMENT_SERVICE_GRPC_ADDR=localhost:50051
go run ./cmd/order-service
```

### 3. Testing
**Create an Order (REST):**
```bash
curl -X POST http://localhost:8081/orders ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: demo-1" ^
  -d "{\"customer_id\":\"cust-1\",\"item_name\":\"Book\",\"amount\":15000}"
```

**Listen to Real-Time Updates (gRPC Stream):**
```bash
cd order-service
go run ./cmd/order-updates-client --addr localhost:50052 --order <ORDER_ID>
```
