package client

import (
	"BikeStoreGolang/api-gateway/proto/order"
	"google.golang.org/grpc"
)

type OrderClient struct {
	Client orderpb.OrderServiceClient
}

func NewOrderClient(conn *grpc.ClientConn) *OrderClient {
	return &OrderClient{
		Client: orderpb.NewOrderServiceClient(conn),
	}
}