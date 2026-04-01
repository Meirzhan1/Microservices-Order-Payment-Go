# AP2 Assignment 1 — Clean Architecture Microservices (Order & Payment)

This repository contains a full solution for **Assignment 1** with two bounded-context microservices in Go:

- **Order Service** (`order-service`)
- **Payment Service** (`payment-service`)

Both services follow Clean Architecture and communicate **only via REST** (Gin framework), with real relational persistence (PostgreSQL).

---

## 1) Architecture Overview

### Bounded Contexts & Ownership

- **Order Context** owns order lifecycle (`Pending`, `Paid`, `Failed`, `Cancelled`), cancellation rules, and idempotent order creation.
- **Payment Context** owns payment authorization/decline decision and transaction records.
- **No shared entities/models package** is used between services.
- **Database per service** is enforced (separate DSN/database).

### Clean Architecture Layers (inside each service)

- `internal/domain` — entities + domain constants
- `internal/usecase` — business rules, orchestration, state transitions
- `internal/repository` — persistence adapters (PostgreSQL)
- `internal/transport/http` — thin Gin handlers
- `internal/app` — composition helpers (router/DB)
- `cmd/<service-name>/main.go` — composition root (manual DI)

### Dependency Direction

Outer layers depend inward; use cases depend on interfaces (ports).

```mermaid
flowchart LR
    A[HTTP Handler] --> B[Use Case]
    B --> C[Repository Interface]
    C --> D[(PostgreSQL)]
```

### Service-to-Service Interaction

```mermaid
flowchart LR
    Client --> OrderAPI[Order Service :8081]
    OrderAPI -->|POST /payments, timeout=2s| PaymentAPI[Payment Service :8082]
    OrderAPI --> OrderDB[(order_db)]
    PaymentAPI --> PaymentDB[(payment_db)]
```

---

## 2) Project Structure

```text
.
├── README.md
├── order-service/
│   ├── cmd/order-service/main.go
│   ├── internal/
│   │   ├── app/
│   │   ├── domain/
│   │   ├── repository/
│   │   ├── transport/http/
│   │   ├── transport/httpclient/
│   │   └── usecase/
│   └── migrations/
└── payment-service/
    ├── cmd/payment-service/main.go
    ├── internal/
    │   ├── app/
    │   ├── domain/
    │   ├── repository/
    │   ├── transport/http/
    │   └── usecase/
    └── migrations/
```

---

## 3) Domain Models

### Order
- `ID string`
- `CustomerID string`
- `ItemName string`
- `Amount int64` (cents)
- `Status string` (`Pending`, `Paid`, `Failed`, `Cancelled`)
- `CreatedAt`, `UpdatedAt`

### Payment
- `ID string`
- `OrderID string`
- `TransactionID string`
- `Amount int64` (cents)
- `Status string` (`Authorized`, `Declined`)
- `DeclineReason string`
- `CreatedAt`

---

## 4) Business Rules Implemented

1. **Money type is `int64`** (no floating point).
2. **Order amount must be > 0**.
3. **Paid orders cannot be cancelled** (only `Pending` can be cancelled).
4. **Payment decline threshold:** if `amount > 100000`, payment status is `Declined`.
5. **Order -> Payment synchronous call** via REST using custom `http.Client` timeout **2 seconds**.
6. **Failure scenario (Payment unavailable):**
   - Order service request does not hang.
   - Returns **503 Service Unavailable**.
   - Order is marked **Failed** (chosen strategy, documented for defense).

---

## 5) Bonus: Idempotency (+10%)

- Implemented in **Order Service** using `Idempotency-Key` header.
- Persisted in `orders.idempotency_key` (unique).
- Repeated `POST /orders` with same key returns existing order instead of creating duplicates.
- Payment side also avoids duplicate payment per order via unique `order_id` and usecase check.

---

## 6) Database Setup (PostgreSQL)

Create two databases:

- `order_db`
- `payment_db`

Apply migrations:

- `order-service/migrations/001_create_orders.sql`
- `payment-service/migrations/001_create_payments.sql`

Example using `psql`:

```bash
psql "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable" -f order-service/migrations/001_create_orders.sql
psql "postgres://postgres:postgres@localhost:5432/payment_db?sslmode=disable" -f payment-service/migrations/001_create_payments.sql
```

---

## 7) Run Services

### Payment Service

```bash
cd payment-service
go mod tidy
set PAYMENT_SERVICE_PORT=8082
set PAYMENT_DB_DSN=postgres://postgres:postgres@localhost:5432/payment_db?sslmode=disable
go run ./cmd/payment-service
```

### Order Service

```bash
cd order-service
go mod tidy
set ORDER_SERVICE_PORT=8081
set ORDER_DB_DSN=postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable
set PAYMENT_SERVICE_URL=http://localhost:8082
go run ./cmd/order-service
```

---

## 8) API Examples (curl)

### 8.1 Create Order (authorized payment)

```bash
curl -X POST http://localhost:8081/orders ^
  -H "Content-Type: application/json" ^
  -H "Idempotency-Key: demo-key-1" ^
  -d "{\"customer_id\":\"cust-1\",\"item_name\":\"Laptop\",\"amount\":15000}"
```

Expected: `201 Created`, order status becomes `Paid`.

### 8.2 Create Order (declined payment)

```bash
curl -X POST http://localhost:8081/orders ^
  -H "Content-Type: application/json" ^
  -d "{\"customer_id\":\"cust-2\",\"item_name\":\"Premium Bundle\",\"amount\":150001}"
```

Expected: `201 Created`, order status becomes `Failed` (payment declined).

### 8.3 Get Order

```bash
curl http://localhost:8081/orders/{order_id}
```

### 8.4 Cancel Order

```bash
curl -X PATCH http://localhost:8081/orders/{order_id}/cancel
```

Expected:
- `200 OK` only for `Pending`
- `409 Conflict` for `Paid`/`Failed`/`Cancelled`

### 8.5 Create Payment directly

```bash
curl -X POST http://localhost:8082/payments ^
  -H "Content-Type: application/json" ^
  -d "{\"order_id\":\"11111111-1111-1111-1111-111111111111\",\"amount\":1000}"
```

### 8.6 Get Payment by Order ID

```bash
curl http://localhost:8082/payments/{order_id}
```

---

## 9) Failure Handling Explanation (for defense)

When `POST /orders` triggers payment authorization:

1. Order is first persisted as `Pending`.
2. Order service calls payment service with shared `http.Client{Timeout: 2 * time.Second}`.
3. If timeout/network failure/unavailability occurs:
   - Order status transitions to `Failed`.
   - API returns `503 Service Unavailable` with order snapshot.

This behavior demonstrates resilience and prevents hanging requests.

---

## 10) Manual DI (Composition Root)

- `cmd/order-service/main.go` and `cmd/payment-service/main.go` manually wire:
  - database connection
  - repository implementation
  - use case
  - transport handlers
  - Gin router startup

No DI framework is used.
