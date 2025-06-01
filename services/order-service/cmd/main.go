package main

import (
	"context"
	"net"
	"os"
	productpb "BikeStoreGolang/services/product-service/proto/gen"
	ordergrpc "BikeStoreGolang/services/order-service/internal/delivery/grpc"
	"BikeStoreGolang/services/order-service/internal/delivery/nats"
	"BikeStoreGolang/services/order-service/internal/logger"
	"BikeStoreGolang/services/order-service/internal/usecase"
	pb "BikeStoreGolang/services/order-service/proto/gen"

	"github.com/joho/godotenv"
	natsio "github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Логгер
	logFile := "order-service.log"
	logr, err := logger.NewLogrusLoggerToFile(logFile)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// .env
	if err := godotenv.Load(".env"); err != nil {
		logr.Warn("Warning: .env file not found, using system environment variables")
	}

	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	if mongoURI == "" || mongoDB == "" {
		logr.Fatal("MONGO_URI or MONGO_DB not set in environment")
	}

	// MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		logr.Fatal("MongoDB connection error: ", err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		logr.Fatal("MongoDB ping error: ", err)
	}
	ordersCollection := client.Database(mongoDB).Collection("orders")

	// NATS
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = natsio.DefaultURL
	}
	nc, err := natsio.Connect(natsURL)
	if err != nil {
		logr.Fatal("NATS connection error: ", err)
	}
	publisher := nats.NewPublisher(nc)

	// Usecase
	productServiceAddr := os.Getenv("PRODUCT_SERVICE_ADDR")
	if productServiceAddr == "" {
		productServiceAddr = "localhost:50052"
	}
	productConn, err := grpc.Dial(productServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logr.Fatal("Failed to connect to ProductService: ", err)
	}
	defer productConn.Close()
	productClient := productpb.NewProductServiceClient(productConn)

	// Передайте productClient в usecase:
	orderUC := usecase.NewOrderUsecase(ordersCollection, publisher, productClient)

	// Подписка на событие "order.processed"
	nats.SubscribeOrderProcessed(nc, func(event nats.OrderProcessedEvent) {
		logr.Infof("Order processed event received: order_id=%s, status=%s, message=%s", event.OrderID, event.Status, event.Message)
	})

	// gRPC server
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		logr.Fatal("Failed to listen: ", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterOrderServiceServer(grpcServer, ordergrpc.NewOrderHandler(orderUC, logr))

	logr.Info("OrderService gRPC server started on :50053")
	if err := grpcServer.Serve(lis); err != nil {
		logr.Fatal("Failed to serve: ", err)
	}
}
