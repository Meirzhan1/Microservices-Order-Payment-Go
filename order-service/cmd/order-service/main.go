package main

import (
	"context"
	"log"
	"net/http"
	"order-service/internal/app"
	"os"
	"os/signal"
	"syscall"
	"time"
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

	srv := &http.Server{
		Addr:    app.Addr(port),
		Handler: router,
	}

	go func() {
		log.Printf("order service started on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start order service: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down order service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Order service forced to shutdown: %v", err)
	}

	log.Println("Order service exited gracefully")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
