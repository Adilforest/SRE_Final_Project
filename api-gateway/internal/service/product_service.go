package service

import (
	"BikeStoreGolang/api-gateway/proto/product"
	"context"
	"io"

	"google.golang.org/protobuf/types/known/emptypb"
)

type ProductService interface {
	ListProducts(ctx context.Context, filter *productpb.ProductFilter) ([]*productpb.ProductResponse, error)
	CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.ProductResponse, error)
	GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.ProductResponse, error)
	UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductResponse, error)
	DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*emptypb.Empty, error)
	SearchProducts(ctx context.Context, req *productpb.SearchRequest) ([]*productpb.ProductResponse, error)
	ChangeProductStock(ctx context.Context, req *productpb.ChangeStockRequest) (*productpb.ProductResponse, error)
}

type productService struct {
	client productpb.ProductServiceClient
}

func NewProductService(client productpb.ProductServiceClient) ProductService {
	return &productService{client: client}
}

func (s *productService) ListProducts(ctx context.Context, filter *productpb.ProductFilter) ([]*productpb.ProductResponse, error) {
	stream, err := s.client.ListProducts(ctx, filter)
	if err != nil {
		return nil, err
	}
	var products []*productpb.ProductResponse
	for {
		product, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

func (s *productService) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.ProductResponse, error) {
	return s.client.CreateProduct(ctx, req)
}

func (s *productService) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.ProductResponse, error) {
	return s.client.GetProduct(ctx, req)
}

func (s *productService) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*productpb.ProductResponse, error) {
	return s.client.UpdateProduct(ctx, req)
}

func (s *productService) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*emptypb.Empty, error) {
	return s.client.DeleteProduct(ctx, req)
}
func (s *productService) SearchProducts(ctx context.Context, req *productpb.SearchRequest) ([]*productpb.ProductResponse, error) {
	stream, err := s.client.SearchProducts(ctx, req)
	if err != nil {
		return nil, err
	}
	var products []*productpb.ProductResponse
	for {
		product, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}
func (s *productService) ChangeProductStock(ctx context.Context, req *productpb.ChangeStockRequest) (*productpb.ProductResponse, error) {
	return s.client.ChangeProductStock(ctx, req)
}