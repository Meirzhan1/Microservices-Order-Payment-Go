package main

import (
	"context"
	"fmt"
	"log"
	"time"

	paymentv1 "github.com/Meirzhan1/microservices-order-payment-contracts-generated/gen/go/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	println("Connecting to Payment Service on localhost:50051...")
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := paymentv1.NewPaymentServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	req := &paymentv1.ListPaymentsRequest{
		MinAmount: 5000,
		MaxAmount: 0,
	}

	println("\nCalling ListPayments with min_amount=5000, max_amount=0")
	start := time.Now()
	res, err := c.ListPayments(ctx, req)
	if err != nil {
		log.Fatalf("could not list payments: %v", err)
	}

	fmt.Printf("Received %d payments (took %v):\n", len(res.Payments), time.Since(start))
	for i, p := range res.Payments {
		fmt.Printf("  [%d] Amount: %d, OrderID: %s, Status: %s, Time: %v\n", i+1, p.Amount, p.OrderId, p.Status, p.CreatedAt.AsTime().Format(time.RFC3339))
	}
}
