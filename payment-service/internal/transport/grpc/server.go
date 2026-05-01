package grpc

import (
	"context"
	"errors"
	"payment-service/internal/usecase"
	"time"

	paymentv1 "github.com/Meirzhan1/microservices-order-payment-contracts-generated/gen/go/proto/payment/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PaymentGRPCServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	uc *usecase.PaymentUseCase
}

func NewPaymentGRPCServer(uc *usecase.PaymentUseCase) *PaymentGRPCServer {
	return &PaymentGRPCServer{uc: uc}
}

func (s *PaymentGRPCServer) ProcessPayment(ctx context.Context, req *paymentv1.PaymentRequest) (*paymentv1.PaymentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	res, err := s.uc.ProcessPayment(ctx, usecase.CreatePaymentInput{
		OrderID: req.GetOrderId(),
		Amount:  req.GetAmount(),
	}, req.GetCustomerEmail())
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidAmount) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to process payment")
	}

	createdAt := res.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return &paymentv1.PaymentResponse{
		Id:            res.ID,
		OrderId:       res.OrderID,
		TransactionId: res.TransactionID,
		Amount:        res.Amount,
		Status:        res.Status,
		DeclineReason: res.DeclineReason,
		CreatedAt:     timestamppb.New(createdAt),
	}, nil
}

func (s *PaymentGRPCServer) ListPayments(ctx context.Context, req *paymentv1.ListPaymentsRequest) (*paymentv1.ListPaymentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	results, err := s.uc.ListPayments(ctx, req.GetMinAmount(), req.GetMaxAmount())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var pbPayments []*paymentv1.PaymentResponse
	for _, res := range results {
		createdAt := res.CreatedAt
		if createdAt.IsZero() {
			createdAt = time.Now().UTC()
		}

		pbPayments = append(pbPayments, &paymentv1.PaymentResponse{
			Id:            res.ID,
			OrderId:       res.OrderID,
			TransactionId: res.TransactionID,
			Amount:        res.Amount,
			Status:        res.Status,
			DeclineReason: res.DeclineReason,
			CreatedAt:     timestamppb.New(createdAt),
		})
	}

	return &paymentv1.ListPaymentsResponse{
		Payments: pbPayments,
	}, nil
}
