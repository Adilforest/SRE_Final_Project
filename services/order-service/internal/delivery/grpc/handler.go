package grpc

import (
	"context"

	"BikeStoreGolang/services/order-service/internal/logger"
	"BikeStoreGolang/services/order-service/internal/usecase"
	pb "BikeStoreGolang/services/order-service/proto/gen"
)

type OrderHandler struct {
	pb.UnimplementedOrderServiceServer
	uc     *usecase.OrderUsecase
	logger logger.Logger
}

func NewOrderHandler(uc *usecase.OrderUsecase, logger logger.Logger) *OrderHandler {
	return &OrderHandler{uc: uc, logger: logger}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
	h.logger.Infof("CreateOrder called for user_id: %s", req.GetUserId())
	return h.uc.CreateOrder(ctx, req)
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	h.logger.Infof("GetOrder called for id: %s", req.GetId())
	return h.uc.GetOrder(ctx, req)
}

func (h *OrderHandler) ListOrders(req *pb.ListOrdersRequest, stream pb.OrderService_ListOrdersServer) error {
	h.logger.Infof("ListOrders called for user_id: %s", req.GetUserId())
	return h.uc.ListOrders(req, stream)
}

func (h *OrderHandler) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	h.logger.Infof("CancelOrder called for id: %s", req.GetId())
	return h.uc.CancelOrder(ctx, req)
}

func (h *OrderHandler) ApproveOrder(ctx context.Context, req *pb.ApproveOrderRequest) (*pb.OrderResponse, error) {
    h.logger.Infof("ApproveOrder called for id: %s", req.GetId())
    return h.uc.ApproveOrder(ctx, req)
}