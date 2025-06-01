package usecase

import (
	"context"
	"strings"
	"time"

	"BikeStoreGolang/services/product-service/internal/domain"
	"BikeStoreGolang/services/product-service/internal/logger"
	pb "BikeStoreGolang/services/product-service/proto/gen"
    natsPublisher "BikeStoreGolang/services/product-service/internal/delivery/nats"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductUsecase struct {
    products *mongo.Collection
    logger   logger.Logger
    publisher  natsPublisher.Publisher
}

func NewProductUsecase(products *mongo.Collection, logger logger.Logger, publisher natsPublisher.Publisher) *ProductUsecase {
	return &ProductUsecase{
		products: products,
		logger:   logger,
        publisher: publisher,
	}
}

func isTokenBlacklisted(ctx context.Context, redisClient *redis.Client, token string) (bool, error) {
    exists, err := redisClient.Exists(ctx, "blacklist:"+token).Result()
    return exists == 1, err
}

func (u *ProductUsecase) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.ProductResponse, error) {
	u.logger.Infof("CreateProduct called for name: %s", req.GetName())

	product := domain.Product{
		ID:          primitive.NewObjectID(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Price:       req.GetPrice(),
		Quantity:    int(req.GetQuantity()),
		Type:        domain.BikeType(req.GetType().String()),
		Brand:       req.GetBrand(),
		Size:        req.GetSize(),
		Color:       req.GetColor(),
		Weight:      req.GetWeight(),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	for _, f := range req.GetFeatures() {
		product.Features = append(product.Features, domain.Feature{
			Name:  f.GetName(),
			Value: f.GetValue(),
		})
	}

	_, err := u.products.InsertOne(ctx, product)
	if err != nil {
		u.logger.Errorf("Failed to insert product: %v", err)
		return nil, err
	}

	u.logger.Infof("Product created successfully: %s", product.ID.Hex())

	return &pb.ProductResponse{
		Id:          product.ID.Hex(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Quantity:    int32(product.Quantity),
		Type:        req.GetType(),
		Brand:       product.Brand,
		Size:        product.Size,
		Color:       product.Color,
		Weight:      product.Weight,
		IsActive:    product.IsActive,
		CreatedAt:   nil,
		UpdatedAt:   nil,
		Features:    req.GetFeatures(),
	}, nil
}

func (u *ProductUsecase) UpdateProduct(ctx context.Context, req *pb.UpdateProductRequest) (*pb.ProductResponse, error) {
    u.logger.Infof("UpdateProduct called for id: %s", req.GetId())

    objID, err := primitive.ObjectIDFromHex(req.GetId())
    if err != nil {
        u.logger.Warnf("Invalid product id: %s", req.GetId())
        return nil, err
    }

    update := bson.M{}
    set := bson.M{}

    if req.Name != nil {
        set["name"] = req.GetName()
    }
    if req.Description != nil {
        set["description"] = req.GetDescription()
    }
    if req.Price != nil {
        set["price"] = req.GetPrice()
    }
    if req.Quantity != nil {
        set["quantity"] = req.GetQuantity()
    }
    if req.Type != nil {
        set["type"] = pb.BikeType_name[int32(req.GetType().Number())]
    }
    if req.Brand != nil {
        set["brand"] = req.GetBrand()
    }
    if req.Size != nil {
        set["size"] = req.GetSize()
    }
    if req.Color != nil {
        set["color"] = req.GetColor()
    }
    if req.Weight != nil {
        set["weight"] = req.GetWeight()
    }
    if req.Features != nil {
        features := make([]domain.Feature, 0, len(req.GetFeatures()))
        for _, f := range req.GetFeatures() {
            features = append(features, domain.Feature{
                Name:  f.GetName(),
                Value: f.GetValue(),
            })
        }
        set["features"] = features
    }
    set["updated_at"] = time.Now()
    update["$set"] = set

    res := u.products.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
    if res.Err() != nil {
        u.logger.Errorf("Failed to update product: %v", res.Err())
        return nil, res.Err()
    }

    var updated domain.Product
    if err := res.Decode(&updated); err != nil {
        u.logger.Errorf("Failed to decode updated product: %v", err)
        return nil, err
    }

    u.logger.Infof("Product updated successfully: %s", updated.ID.Hex())
    return productToProto(&updated), nil
}

func (u *ProductUsecase) DeleteProduct(ctx context.Context, req *pb.DeleteProductRequest) (*emptypb.Empty, error) {
    u.logger.Infof("DeleteProduct called for id: %s", req.GetId())

    objID, err := primitive.ObjectIDFromHex(req.GetId())
    if err != nil {
        u.logger.Warnf("Invalid product id: %s", req.GetId())
        return nil, err
    }

    res, err := u.products.DeleteOne(ctx, bson.M{"_id": objID})
    if err != nil {
        u.logger.Errorf("Failed to delete product: %v", err)
        return nil, err
    }
    if res.DeletedCount == 0 {
        u.logger.Warnf("Product not found for delete: %s", req.GetId())
        return nil, mongo.ErrNoDocuments
    }

    u.logger.Infof("Product deleted successfully: %s", req.GetId())
    return &emptypb.Empty{}, nil
}

func (u *ProductUsecase) ChangeProductStock(ctx context.Context, req *pb.ChangeStockRequest) (*pb.ProductResponse, error) {
    u.logger.Infof("ChangeProductStock called for product_id: %s, quantity_change: %d", req.GetProductId(), req.GetQuantityChange())

    objID, err := primitive.ObjectIDFromHex(req.GetProductId())
    if err != nil {
        u.logger.Warnf("Invalid product id: %s", req.GetProductId())
        return nil, err
    }

    update := bson.M{
        "$inc": bson.M{"quantity": req.GetQuantityChange()},
        "$set": bson.M{"updated_at": time.Now()},
    }

    res := u.products.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, options.FindOneAndUpdate().SetReturnDocument(options.After))
    if res.Err() != nil {
        u.logger.Errorf("Failed to change product stock: %v", res.Err())
        return nil, res.Err()
    }

    var updated domain.Product
    if err := res.Decode(&updated); err != nil {
        u.logger.Errorf("Failed to decode updated product: %v", err)
        return nil, err
    }

    if u.publisher != nil {
        event := natsPublisher.OrderProcessedEvent{
            OrderID: req.GetOrderId(),
            Status:  "processed",
            Message: "Order processed and stock updated",
        }
        _ = u.publisher.PublishOrderProcessed(event)
    }

    u.logger.Infof("Product stock changed successfully: %s, new quantity: %d", updated.ID.Hex(), updated.Quantity)
    return productToProto(&updated), nil
}


func (u *ProductUsecase) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.ProductResponse, error) {
    u.logger.Infof("GetProduct called for id: %s", req.GetId())

    objID, err := primitive.ObjectIDFromHex(req.GetId())
    if err != nil {
        u.logger.Warnf("Invalid product id: %s", req.GetId())
        return nil, err
    }

    var product domain.Product
    err = u.products.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
    if err != nil {
        u.logger.Warnf("Product not found: %s", req.GetId())
        return nil, err
    }

    return productToProto(&product), nil
}

func (u *ProductUsecase) ListProducts(ctx context.Context, req *pb.ProductFilter, stream pb.ProductService_ListProductsServer) error {
    u.logger.Infof("ListProducts called")

    filter := bson.M{}
    if len(req.GetTypes()) > 0 {
        types := make([]string, 0, len(req.GetTypes()))
        for _, t := range req.GetTypes() {
            types = append(types, t.String())
        }
        filter["type"] = bson.M{"$in": types}
    }
    if req.GetMinPrice() > 0 {
        filter["price"] = bson.M{"$gte": req.GetMinPrice()}
    }
    if req.GetMaxPrice() > 0 {
        if v, ok := filter["price"].(bson.M); ok {
            v["$lte"] = req.GetMaxPrice()
        } else {
            filter["price"] = bson.M{"$lte": req.GetMaxPrice()}
        }
    }
    if len(req.GetBrands()) > 0 {
        filter["brand"] = bson.M{"$in": req.GetBrands()}
    }
    if len(req.GetSizes()) > 0 {
        filter["size"] = bson.M{"$in": req.GetSizes()}
    }
    if req.GetOnlyActive() {
        filter["is_active"] = true
    }

    opts := options.Find()
    if req.GetSortBy() != "" {
        order := 1
        if req.GetSortOrder() < 0 {
            order = -1
        }
        opts.SetSort(bson.D{{Key: req.GetSortBy(), Value: order}})
    }

    cursor, err := u.products.Find(ctx, filter, opts)
    if err != nil {
        u.logger.Errorf("Failed to list products: %v", err)
        return err
    }
    defer cursor.Close(ctx)

    for cursor.Next(ctx) {
        var product domain.Product
        if err := cursor.Decode(&product); err != nil {
            u.logger.Warnf("Failed to decode product: %v", err)
            continue
        }
        if err := stream.Send(productToProto(&product)); err != nil {
            u.logger.Errorf("Failed to send product: %v", err)
            return err
        }
    }
    return nil
}

// Вспомогательная функция для преобразования domain.Product в pb.ProductResponse
func productToProto(p *domain.Product) *pb.ProductResponse {
    features := make([]*pb.Feature, 0, len(p.Features))
    for _, f := range p.Features {
        features = append(features, &pb.Feature{
            Name:  f.Name,
            Value: f.Value,
        })
    }
    return &pb.ProductResponse{
        Id:          p.ID.Hex(),
        Name:        p.Name,
        Description: p.Description,
        Price:       p.Price,
        Quantity:    int32(p.Quantity),
        Type:        pb.BikeType(pb.BikeType_value[strings.ToUpper(string(p.Type))]),
        Brand:       p.Brand,
        Size:        p.Size,
        Color:       p.Color,
        Weight:      p.Weight,
        Rating:      p.Rating,
        IsActive:    p.IsActive,
        CreatedAt:   timestamppb.New(p.CreatedAt),
        UpdatedAt:   timestamppb.New(p.UpdatedAt),
        Features:    features,
    }
}

func (u *ProductUsecase) SearchProducts(req *pb.SearchRequest, stream pb.ProductService_SearchProductsServer) error {
    u.logger.Infof("SearchProducts called. Query: %s", req.GetQuery())

    filter := bson.M{}

    // Поиск по тексту (например, по name и description)
    if q := strings.TrimSpace(req.GetQuery()); q != "" {
        filter["$or"] = []bson.M{
            {"name": bson.M{"$regex": q, "$options": "i"}},
            {"description": bson.M{"$regex": q, "$options": "i"}},
        }
    }

    // Применяем фильтр, если он есть
    if req.GetFilter() != nil {
        f := req.GetFilter()
        if len(f.GetTypes()) > 0 {
            types := make([]string, 0, len(f.GetTypes()))
            for _, t := range f.GetTypes() {
                types = append(types, t.String())
            }
            filter["type"] = bson.M{"$in": types}
        }
        if f.GetMinPrice() > 0 {
            filter["price"] = bson.M{"$gte": f.GetMinPrice()}
        }
        if f.GetMaxPrice() > 0 {
            if v, ok := filter["price"].(bson.M); ok {
                v["$lte"] = f.GetMaxPrice()
            } else {
                filter["price"] = bson.M{"$lte": f.GetMaxPrice()}
            }
        }
        if len(f.GetBrands()) > 0 {
            filter["brand"] = bson.M{"$in": f.GetBrands()}
        }
        if len(f.GetSizes()) > 0 {
            filter["size"] = bson.M{"$in": f.GetSizes()}
        }
        if f.GetOnlyActive() {
            filter["is_active"] = true
        }
    }

    opts := options.Find()
    if req.GetFilter() != nil && req.GetFilter().GetSortBy() != "" {
        order := 1
        if req.GetFilter().GetSortOrder() < 0 {
            order = -1
        }
        opts.SetSort(bson.D{{Key: req.GetFilter().GetSortBy(), Value: order}})
    }

    cursor, err := u.products.Find(stream.Context(), filter, opts)
    if err != nil {
        u.logger.Errorf("Failed to search products: %v", err)
        return err
    }
    defer cursor.Close(stream.Context())

    for cursor.Next(stream.Context()) {
        var product domain.Product
        if err := cursor.Decode(&product); err != nil {
            u.logger.Warnf("Failed to decode product: %v", err)
            continue
        }
        if err := stream.Send(productToProto(&product)); err != nil {
            u.logger.Errorf("Failed to send product: %v", err)
            return err
        }
    }
    return nil
}