package grpc

import (
    "google.golang.org/grpc"
)

func NewServer() *grpc.Server {
    return grpc.NewServer()
}
// DialProxy прокси для grpc.Dial, чтобы main.go не импортировал grpc напрямую
func DialProxy(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
    return grpc.Dial(target, opts...)
}

func WithInsecure() grpc.DialOption {
    return grpc.WithInsecure()
}