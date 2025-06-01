package usecase

import (
	"context"
	"time"

	natsPublisher "BikeStoreGolang/services/order-service/internal/delivery/nats"
	"BikeStoreGolang/services/order-service/internal/domain"
	pb "BikeStoreGolang/services/order-service/proto/gen"
	productpb "BikeStoreGolang/services/product-service/proto/gen"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderUsecase struct {
    orders    *mongo.Collection
    publisher *natsPublisher.Publisher
	productSvc productpb.ProductServiceClient
}

func NewOrderUsecase(orders *mongo.Collection, publisher *natsPublisher.Publisher, productSvc productpb.ProductServiceClient) *OrderUsecase {
    return &OrderUsecase{orders: orders, publisher: publisher, productSvc: productSvc,}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, req *pb.CreateOrderRequest) (*pb.OrderResponse, error) {
    order := domain.Order{
        ID:        primitive.NewObjectID().Hex(),
        UserID:    req.GetUserId(),
        Items:     toDomainItems(req.GetItems()),
        Total:     req.GetTotal(),
        Address:   req.GetAddress(),
        Status:    "created",
        CreatedAt: time.Now(),
    }
    _, err := u.orders.InsertOne(ctx, order)
    if err != nil {
        return nil, err
    }

    // --- NATS: публикуем событие ---
    if u.publisher != nil {
        event := natsPublisher.OrderCreatedEvent{
            OrderID: order.ID,
            UserID:  order.UserID,
            Items:   order.Items,
            Total:   order.Total,
            Address: order.Address,
            Status:  order.Status,
        }
        _ = u.publisher.PublishOrderCreated(event) // обработай ошибку по необходимости
    }

    return toProtoOrder(&order), nil
}

func (u *OrderUsecase) GetOrder(ctx context.Context, req *pb.GetOrderRequest) (*pb.OrderResponse, error) {
	var order domain.Order
	err := u.orders.FindOne(ctx, bson.M{"_id": req.GetId()}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return toProtoOrder(&order), nil
}

func (u *OrderUsecase) ListOrders(req *pb.ListOrdersRequest, stream pb.OrderService_ListOrdersServer) error {
	filter := bson.M{}
	if req.GetUserId() != "" {
		filter["user_id"] = req.GetUserId()
	}
	cursor, err := u.orders.Find(stream.Context(), filter)
	if err != nil {
		return err
	}
	defer cursor.Close(stream.Context())

	for cursor.Next(stream.Context()) {
		var order domain.Order
		if err := cursor.Decode(&order); err != nil {
			continue
		}
		if err := stream.Send(toProtoOrder(&order)); err != nil {
			return err
		}
	}
	return nil
}

func (u *OrderUsecase) CancelOrder(ctx context.Context, req *pb.CancelOrderRequest) (*pb.OrderResponse, error) {
	filter := bson.M{"_id": req.GetId()}
	update := bson.M{"$set": bson.M{"status": "cancelled"}}
	res := u.orders.FindOneAndUpdate(ctx, filter, update)
	if res.Err() != nil {
		return nil, res.Err()
	}
	var order domain.Order
	if err := res.Decode(&order); err != nil {
		return nil, err
	}
	order.Status = "cancelled"
	return toProtoOrder(&order), nil
}

// --- helpers ---

func toDomainItems(items []*pb.OrderItem) []domain.OrderItem {
    result := make([]domain.OrderItem, 0, len(items))
    for _, i := range items {
        result = append(result, domain.OrderItem{
            ProductID: i.GetProductId(),
            Quantity:  i.GetQuantity(),
        })
    }
    return result
}

func toProtoOrder(o *domain.Order) *pb.OrderResponse {
	items := make([]*pb.OrderItem, 0, len(o.Items))
	for _, i := range o.Items {
		items = append(items, &pb.OrderItem{
			ProductId: i.ProductID,
			Quantity:  i.Quantity,
		})
	}
	return &pb.OrderResponse{
		Id:        o.ID,
		UserId:    o.UserID,
		Items:     items,
		Total:     o.Total,
		Address:   o.Address,
		Status:    o.Status,
		CreatedAt: timestamppb.New(o.CreatedAt),
	}
}

func (u *OrderUsecase) ApproveOrder(ctx context.Context, req *pb.ApproveOrderRequest) (*pb.OrderResponse, error) {
    filter := bson.M{"_id": req.GetId()}
    update := bson.M{"$set": bson.M{"status": "approved"}}
    res := u.orders.FindOneAndUpdate(ctx, filter, update)
    if res.Err() != nil {
        return nil, res.Err()
    }
    var order domain.Order
    if err := res.Decode(&order); err != nil {
        return nil, err
    }
    order.Status = "approved"

    // --- ВАЖНО: уменьшить stock по каждому товару ---
    for _, item := range order.Items {
		if item.ProductID == "" || item.Quantity <= 0 {
        continue
    	}
        _, err := u.productSvc.ChangeProductStock(ctx, &productpb.ChangeStockRequest{
            ProductId:      item.ProductID,
            QuantityChange: -item.Quantity, // минус, чтобы уменьшить
            OrderId:        order.ID,
        })
        if err != nil {
            // Можно обработать ошибку (например, отменить заказ или залогировать)
            // return nil, fmt.Errorf("failed to change stock for product %s: %w", item.ProductID, err)
        }
    }

    // --- NATS: публикуем событие ---
    if u.publisher != nil {
        event := natsPublisher.OrderCreatedEvent{
            OrderID: order.ID,
            UserID:  order.UserID,
            Items:   order.Items,
            Total:   order.Total,
            Address: order.Address,
            Status:  order.Status,
        }
        _ = u.publisher.PublishOrderApproved(event)
    }

    return toProtoOrder(&order), nil
}