package main

import (
	"context"
	"log"
	"notification-service/internal/infrastructure/rabbitmq"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rabbitMQURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	queueName := getEnv("QUEUE_NAME", "payment.completed")
	exchangeName := getEnv("EXCHANGE_NAME", "payment.exchange")
	dlxName := getEnv("DLX_NAME", "payment.completed.dlx")
	dlqName := getEnv("DLQ_NAME", "payment.completed.dlq")

	consumer, err := rabbitmq.NewConsumer(rabbitMQURL, exchangeName, queueName, dlxName, dlqName)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ consumer: %v", err)
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := consumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}

	log.Printf("Notification service started. Listening for events on %s...", queueName)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down notification service...")
	
	cancel()
	
	time.Sleep(2 * time.Second)
	log.Println("Notification service exited gracefully")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
