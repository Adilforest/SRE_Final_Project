package client

import (
    "BikeStoreGolang/api-gateway/proto/auth"
    "google.golang.org/grpc"
)

type AuthClient struct {
    Client authpb.AuthServiceClient
}

func NewAuthClient(conn *grpc.ClientConn) *AuthClient {
    return &AuthClient{
        Client: authpb.NewAuthServiceClient(conn),
    }
}