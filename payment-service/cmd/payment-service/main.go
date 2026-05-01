package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"payment-service/internal/app"
	"payment-service/internal/infrastructure/rabbitmq"
	"syscall"
	"time"
)

func main() {
	port := getEnv("PAYMENT_SERVICE_PORT", "8082")
	grpcPort := getEnv("PAYMENT_SERVICE_GRPC_PORT", "50051")
	dsn := getEnv("PAYMENT_DB_DSN", "postgres://postgres:postgres@localhost:5432/payment_db?sslmode=disable")
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	db, err := app.OpenDB(dsn)
	if err != nil {
		log.Fatalf("failed to connect payment database: %v", err)
	}
	defer db.Close()

	queueName := getEnv("QUEUE_NAME", "payment.completed")
	exchangeName := getEnv("EXCHANGE_NAME", "payment.exchange")
	dlxName := getEnv("DLX_NAME", "payment.completed.dlx")
	dlqName := getEnv("DLQ_NAME", "payment.completed.dlq")
	producer, err := rabbitmq.NewProducer(rabbitMQURL, exchangeName, queueName, dlxName, dlqName)
	if err != nil {
		log.Printf("Warning: failed to initialize RabbitMQ producer: %v", err)
	} else {
		defer producer.Close()
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.RunGRPCServer(db, producer, grpcPort); err != nil {
			log.Fatalf("failed to start payment gRPC service: %v", err)
		}
	}()

	router := app.NewRouter(db, producer)
	srv := &http.Server{
		Addr:    app.Addr(port),
		Handler: router,
	}

	go func() {
		log.Printf("payment service started on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start payment service: %v", err)
		}
	}()

	<-sigChan
	log.Println("Shutting down payment service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Payment service forced to shutdown: %v", err)
	}

	log.Println("Payment service exited gracefully")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
