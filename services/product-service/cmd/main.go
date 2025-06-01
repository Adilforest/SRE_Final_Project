package main

import (
	"context"
	"net"
	"os"

	"BikeStoreGolang/services/product-service/internal/delivery/grpc"
	deliverynats "BikeStoreGolang/services/product-service/internal/delivery/nats"
	"BikeStoreGolang/services/product-service/internal/logger"
	"BikeStoreGolang/services/product-service/internal/usecase"
	authpb "BikeStoreGolang/services/auth-service/proto/gen"
	pb "BikeStoreGolang/services/product-service/proto/gen"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Логгер
	logFile := "product-service.log"
	log, err := logger.NewLogrusLoggerToFile(logFile)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// .env
	if err := godotenv.Load(".env"); err != nil {
		log.Warn("Warning: .env file not found, using system environment variables")
	}

	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	if mongoURI == "" || mongoDB == "" {
		log.Fatal("MONGO_URI or MONGO_DB not set in environment")
	}

	// MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB connection error: ", err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal("MongoDB ping error: ", err)
	}
	productsCollection := client.Database(mongoDB).Collection("products")

	// NATS Publisher
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	nc, err := nats.Connect(natsURL)
	if err != nil {
		log.Fatal("NATS connection error: ", err)
	}
	publisher := deliverynats.NewPublisher(nc, log)

	deliverynats.SubscribeOrderCreated(nc, log, func(event deliverynats.OrderCreatedEvent) {
    log.Infof("Handle order.created: order_id=%s", event.OrderID)
    // Здесь уменьшайте stock, проверяйте наличие товаров и т.д.
    // После обработки отправьте событие order.processed:
    processedEvent := deliverynats.OrderProcessedEvent{
        OrderID: event.OrderID,
        Status:  "processed",
        Message: "Order processed by product-service",
    }
    if err := publisher.PublishOrderProcessed(processedEvent); err != nil {
        log.Warnf("Failed to publish order.processed: %v", err)
    }
})

	// Usecase
	productUC := usecase.NewProductUsecase(productsCollection, log, publisher)

	// AuthService gRPC client
	authServiceAddr := os.Getenv("AUTH_SERVICE_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "localhost:50051"
	}
	authConn, err := grpc.DialProxy(authServiceAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal("Failed to connect to AuthService: ", err)
	}
	defer authConn.Close()
	authClient := authpb.NewAuthServiceClient(authConn)

	// gRPC server
	lis, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatal("Failed to listen: ", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProductServiceServer(grpcServer, grpc.NewProductHandler(productUC, authClient))

	log.Info("ProductService gRPC server started on :50052")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve: ", err)
	}
}
