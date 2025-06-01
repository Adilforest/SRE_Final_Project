package client

import (
    "BikeStoreGolang/api-gateway/proto/product"
    "google.golang.org/grpc"
)

type ProductClient struct {
    Client productpb.ProductServiceClient
}

func NewProductClient(conn *grpc.ClientConn) *ProductClient {
    return &ProductClient{
        Client: productpb.NewProductServiceClient(conn),
    }
}