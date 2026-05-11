package main

import (
	"context"
	"flag"
	"log"
	"time"

	orderv1 "github.com/Meirzhan1/microservices-order-payment-contracts-generated/gen/go/proto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	addr := flag.String("addr", "localhost:50052", "order gRPC server address")
	orderID := flag.String("order", "", "order id to subscribe")
	flag.Parse()

	if *orderID == "" {
		log.Fatal("--order is required")
	}

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := orderv1.NewOrderUpdatesServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stream, err := client.SubscribeToOrderUpdates(ctx, &orderv1.OrderRequest{OrderId: *orderID})
	if err != nil {
		log.Fatalf("failed to subscribe: %v", err)
	}

	log.Printf("subscribed to order %s", *orderID)
	for {
		update, err := stream.Recv()
		if err != nil {
			log.Fatalf("stream ended: %v", err)
		}

		updatedAt := ""
		if update.GetUpdatedAt() != nil {
			updatedAt = update.GetUpdatedAt().AsTime().Format(time.RFC3339Nano)
		}

		log.Printf("order=%s status=%s updated_at=%s", update.GetOrderId(), update.GetStatus(), updatedAt)
	}
}
