package grpc

import (
	"errors"
	"order-service/internal/usecase"
	"time"

	orderv1 "github.com/Meirzhan1/microservices-order-payment-contracts-generated/gen/go/proto/order/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderUpdatesGRPCServer struct {
	orderv1.UnimplementedOrderUpdatesServiceServer
	uc *usecase.OrderUseCase
}

func NewOrderUpdatesGRPCServer(uc *usecase.OrderUseCase) *OrderUpdatesGRPCServer {
	return &OrderUpdatesGRPCServer{uc: uc}
}

func (s *OrderUpdatesGRPCServer) SubscribeToOrderUpdates(
	req *orderv1.OrderRequest,
	stream grpc.ServerStreamingServer[orderv1.OrderStatusUpdate],
) error {
	if req == nil || req.GetOrderId() == "" {
		return status.Error(codes.InvalidArgument, "order_id is required")
	}

	ctx := stream.Context()
	orderID := req.GetOrderId()

	order, err := s.uc.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, usecase.ErrOrderNotFound) {
			return status.Error(codes.NotFound, "order not found")
		}
		return status.Error(codes.Internal, "failed to load order")
	}

	lastStatus := order.Status
	lastUpdatedAt := order.UpdatedAt

	if err := stream.Send(&orderv1.OrderStatusUpdate{
		OrderId:   order.ID,
		Status:    order.Status,
		UpdatedAt: timestamppb.New(order.UpdatedAt),
	}); err != nil {
		return status.Error(codes.Unavailable, "failed to send initial order update")
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			fresh, err := s.uc.GetOrder(ctx, orderID)
			if err != nil {
				if errors.Is(err, usecase.ErrOrderNotFound) {
					return status.Error(codes.NotFound, "order not found")
				}
				return status.Error(codes.Internal, "failed to refresh order")
			}

			if fresh.Status != lastStatus || !fresh.UpdatedAt.Equal(lastUpdatedAt) {
				if err := stream.Send(&orderv1.OrderStatusUpdate{
					OrderId:   fresh.ID,
					Status:    fresh.Status,
					UpdatedAt: timestamppb.New(fresh.UpdatedAt),
				}); err != nil {
					return status.Error(codes.Unavailable, "failed to send order update")
				}
				lastStatus = fresh.Status
				lastUpdatedAt = fresh.UpdatedAt
			}
		}
	}
}
