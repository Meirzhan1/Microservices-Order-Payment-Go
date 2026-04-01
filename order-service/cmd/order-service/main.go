package main

import (
	"log"
	"order-service/internal/app"
	"os"
)

func main() {
	port := getEnv("ORDER_SERVICE_PORT", "8081")
	dsn := getEnv("ORDER_DB_DSN", "postgres://postgres:postgres@localhost:5432/order_db?sslmode=disable")
	paymentServiceURL := getEnv("PAYMENT_SERVICE_URL", "http://localhost:8082")

	db, err := app.OpenDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect order database: %v", err)
	}
	defer db.Close()

	router := app.NewRouter(db, paymentServiceURL)
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
