package app

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"order-service/internal/repository"
	grpcTransport "order-service/internal/transport/grpc"
	"order-service/internal/transport/httpclient"
	"order-service/internal/usecase"
	"time"

	orderv1 "github.com/Meirzhan1/microservices-order-payment-contracts-generated/gen/go/proto/order/v1"
	"google.golang.org/grpc"
)

func RunGRPCServer(db *sql.DB, port string, paymentServiceGRPCAddr string) error {
	repo := repository.NewPostgresOrderRepository(db)

	paymentClient, err := httpclient.NewGRPCPaymentClient(paymentServiceGRPCAddr, 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to initialize payment gRPC client: %w", err)
	}

	uc := usecase.NewOrderUseCase(repo, paymentClient)
	serverImpl := grpcTransport.NewOrderUpdatesGRPCServer(uc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderUpdatesServiceServer(grpcServer, serverImpl)

	log.Printf("order gRPC server started on :%s", port)
	return grpcServer.Serve(lis)
}
