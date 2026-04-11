package app

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"payment-service/internal/repository"
	grpcTransport "payment-service/internal/transport/grpc"
	"payment-service/internal/usecase"

	paymentv1 "github.com/Meirzhan1/microservices-order-payment-contracts-generated/gen/go/proto/payment/v1"
	"google.golang.org/grpc"
)

func RunGRPCServer(db *sql.DB, port string) error {
	repo := repository.NewPostgresPaymentRepository(db)
	uc := usecase.NewPaymentUseCase(repo)
	serverImpl := grpcTransport.NewPaymentGRPCServer(uc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return fmt.Errorf("failed to listen gRPC: %w", err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcTransport.LoggingUnaryInterceptor()))
	paymentv1.RegisterPaymentServiceServer(grpcServer, serverImpl)

	log.Printf("payment gRPC server started on :%s", port)
	return grpcServer.Serve(lis)
}
