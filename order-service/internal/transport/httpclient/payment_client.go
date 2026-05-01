package httpclient

import (
	"context"
	"fmt"
	"time"

	paymentv1 "github.com/Meirzhan1/microservices-order-payment-contracts-generated/gen/go/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCPaymentClient struct {
	client  paymentv1.PaymentServiceClient
	timeout time.Duration
}

func NewGRPCPaymentClient(addr string, timeout time.Duration) (*GRPCPaymentClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial payment gRPC service: %w", err)
	}

	return &GRPCPaymentClient{
		client:  paymentv1.NewPaymentServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *GRPCPaymentClient) CreatePayment(ctx context.Context, orderID string, amount int64, customerEmail string) (string, error) {
	callCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.client.ProcessPayment(callCtx, &paymentv1.PaymentRequest{
		OrderId:       orderID,
		Amount:        amount,
		CustomerEmail: customerEmail,
	})
	if err != nil {
		return "", err
	}

	return resp.GetStatus(), nil
}
