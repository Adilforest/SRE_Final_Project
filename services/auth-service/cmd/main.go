package main

import (
	"context"
	"net"
	"os"

	"BikeStoreGolang/services/auth-service/internal/delivery/grpc"
	"BikeStoreGolang/services/auth-service/internal/logger"
	"BikeStoreGolang/services/auth-service/internal/mail_sender"
	"BikeStoreGolang/services/auth-service/internal/usecase"
	pb "BikeStoreGolang/services/auth-service/proto/gen"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Инициализация логгера
	logFile := "auth-service.log"
	log, err := logger.NewLogrusLoggerToFile(logFile)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// Загрузка переменных окружения
	if err := godotenv.Load(".env"); err != nil {
		log.Warn("Warning: .env file not found, using system environment variables")
	}

	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	if mongoURI == "" || mongoDB == "" {
		log.Fatal("MONGO_URI or MONGO_DB not set in environment")
	}

	// Подключение к MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("MongoDB connection error: ", err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatal("MongoDB ping error: ", err)
	}

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		log.Fatal("SMTP configuration missing in environment")
	}
	sender := mail_sender.NewSMTPMailer(smtpHost, smtpPort, smtpUser, smtpPass, log)

	// Подключение к Redis
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		log.Fatal("REDIS_ADDR not set in environment")
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Redis connection error: ", err)
	}

	// Инициализация бизнес-логики
	authUC := usecase.NewAuthUsecase(client, mongoDB, log, sender, redisClient)

	// Запуск gRPC сервера
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("Failed to listen: ", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, grpc.NewAuthHandler(authUC))

	log.Info("AuthService gRPC server started on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("Failed to serve: ", err)
	}
}
