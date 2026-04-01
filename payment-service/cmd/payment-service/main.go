package main

import (
	"log"
	"os"
	"payment-service/internal/app"
)

func main() {
	port := getEnv("PAYMENT_SERVICE_PORT", "8082")
	dsn := getEnv("PAYMENT_DB_DSN", "postgres://postgres:postgres@localhost:5432/payment_db?sslmode=disable")

	db, err := app.OpenDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect payment database: %v", err)
	}
	defer db.Close()

	router := app.NewRouter(db)
	if err := router.Run(app.Addr(port)); err != nil {
		log.Fatalf("failed to start payment service: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
