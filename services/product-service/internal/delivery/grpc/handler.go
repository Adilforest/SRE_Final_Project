package grpc

import (
	authpb "BikeStoreGolang/services/auth-service/proto/gen"
	"BikeStoreGolang/services/product-service/internal/usecase"
	pb "BikeStoreGolang/services/product-service/proto/gen"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProductHandler struct {
	pb.UnimplementedProductServiceServer
	uc         *usecase.ProductUsecase
	authClient authpb.AuthServiceClient
}

func (h *ProductHandler) checkAdmin(ctx context.Context) error {
	md, _ := metadata.FromIncomingContext(ctx)
	var token string
	if authHeaders := md["authorization"]; len(authHeaders) > 0 {
		token = authHeaders[0]
	}
	if token == "" {
		return status.Error(codes.Unauthenticated, "no token provided")
	}
	authCtx := metadata.AppendToOutgoingContext(ctx, "authorization", token)
	authResp, err := h.authClient.GetMe(authCtx, &authpb.GetMeRequest{})
	if err != nil {
		return status.Error(codes.PermissionDenied, "invalid token")
	}
	if authResp.Role != authpb.Role_ROLE_ADMIN {
		return status.Error(codes.PermissionDenied, "only admin can perform this action")
	}
	return nil
}

func NewProductHandler(uc *usecase.ProductUsecase, authClient authpb.AuthServiceClient) *ProductHandler {
	return &ProductHandler{uc: uc, authClient: authClient}
}

func (h *ProductHandler) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	if err := h.checkAdmin(ctx); err != nil {
		return nil, err
	}
	return h.uc.CreateProduct(ctx, req)
}

func (h *ProductHandler) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
	if err := h.checkAdmin(ctx); err != nil {
		return nil, err
	}
	return h.uc.UpdateProduct(ctx, req)
}

func (h *ProductHandler) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := h.checkAdmin(ctx); err != nil {
		return nil, err
	}
	return h.uc.DeleteProduct(ctx, req)
}

func (h *ProductHandler) ChangeProductStock(ctx context.Context, req *pb.ChangeStockRequest) (*pb.ProductResponse, error) {
	return h.uc.ChangeProductStock(ctx, req)
}

func (h *ProductHandler) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
	return h.uc.GetProduct(ctx, req)
}

func (h *ProductHandler) ListProducts(req *pb.ProductFilter, stream pb.ProductService_ListProductsServer) error {
	return h.uc.ListProducts(stream.Context(), req, stream)
}

func (h *ProductHandler) SearchProducts(req *pb.SearchRequest, stream pb.ProductService_SearchProductsServer) error {
	return h.uc.SearchProducts(req, stream)
}
