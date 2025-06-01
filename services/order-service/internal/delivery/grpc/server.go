package grpc

import (
    "google.golang.org/grpc"
)

func NewServer() *grpc.Server {
    return grpc.NewServer()
}