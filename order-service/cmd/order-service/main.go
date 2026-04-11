package main

import (
	"log"
	"order-service/internal/app"
	"os"
)

func main() {
	port := getEnv("ORDER_SERVICE_PORT", "8081")
	grpcPort := getEnv("ORDER_SERVICE_GRPC_PORT", "50052")
	dsn := getEnv("ORDER_DB_DSN", "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable")
	paymentServiceGRPCAddr := getEnv("PAYMENT_SERVICE_GRPC_ADDR", "localhost:50051")

	db, err := app.OpenDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect order database: %v", err)
	}
	defer db.Close()

	go func() {
		if err := app.RunGRPCServer(db, grpcPort, paymentServiceGRPCAddr); err != nil {
			log.Fatalf("failed to start order gRPC service: %v", err)
		}
	}()

	router, err := app.NewRouter(db, paymentServiceGRPCAddr)
	if err != nil {
		log.Fatalf("failed to initialize order service router: %v", err)
	}
	if err := router.Run(app.Addr(port)); err != nil {
		log.Fatalf("failed to start order service: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
