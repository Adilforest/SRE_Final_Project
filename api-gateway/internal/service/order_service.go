package service

import (
    "BikeStoreGolang/api-gateway/proto/order"
    "context"
    "io"
)

type OrderService interface {
    CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrderResponse, error)
    GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrderResponse, error)
    ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) ([]*orderpb.OrderResponse, error)
    CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.OrderResponse, error)
    ApproveOrder(ctx context.Context, req *orderpb.ApproveOrderRequest) (*orderpb.OrderResponse, error)
}

type orderService struct {
    client orderpb.OrderServiceClient
}

func NewOrderService(client orderpb.OrderServiceClient) OrderService {
    return &orderService{client: client}
}

func (s *orderService) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.OrderResponse, error) {
    return s.client.CreateOrder(ctx, req)
}

func (s *orderService) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.OrderResponse, error) {
    return s.client.GetOrder(ctx, req)
}

func (s *orderService) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) ([]*orderpb.OrderResponse, error) {
    stream, err := s.client.ListOrders(ctx, req)
    if err != nil {
        return nil, err
    }
    var orders []*orderpb.OrderResponse
    for {
        order, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }
        orders = append(orders, order)
    }
    return orders, nil
}

func (s *orderService) CancelOrder(ctx context.Context, req *orderpb.CancelOrderRequest) (*orderpb.OrderResponse, error) {
    return s.client.CancelOrder(ctx, req)
}

func (s *orderService) ApproveOrder(ctx context.Context, req *orderpb.ApproveOrderRequest) (*orderpb.OrderResponse, error) {
    return s.client.ApproveOrder(ctx, req)
}