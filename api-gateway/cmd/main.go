package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"BikeStoreGolang/api-gateway/internal/client"
	"BikeStoreGolang/api-gateway/internal/handlers"
	"BikeStoreGolang/api-gateway/internal/logger"
	"BikeStoreGolang/api-gateway/internal/service"
)

// Метрики Prometheus
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 1.5, 2, 3, 5},
		},
		[]string{"method", "path"},
	)

	grpcClientConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_client_connections_total",
			Help: "Current number of gRPC client connections",
		},
		[]string{"service"},
	)
)

func init() {
	// Регистрируем метрики
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(grpcClientConnections)
}

// PrometheusMiddleware добавляет сбор метрик для Gin
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		// Пропускаем метрики самого Prometheus
		if path == "/metrics" {
			c.Next()
			return
		}

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			string(status),
		).Inc()

		httpRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)
	}
}

func main() {
	logFile := "api-gateway.log"
	log, err := logger.NewLogrusLoggerToFile(logFile)
	godotenv.Load(".env")
	if err != nil {
		log.Warn(".env file not found or failed to load")
	}

	// gRPC connections
	authConn, err := grpc.NewClient(os.Getenv("AUTH_SERVICE_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к auth-service: %v", err)
	}
	defer authConn.Close()
	grpcClientConnections.WithLabelValues("auth").Inc()

	productConn, err := grpc.NewClient(os.Getenv("PRODUCT_SERVICE_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к product-service: %v", err)
	}
	defer productConn.Close()
	grpcClientConnections.WithLabelValues("product").Inc()

	orderConn, err := grpc.NewClient(os.Getenv("ORDER_SERVICE_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Не удалось подключиться к order-service: %v", err)
	}
	defer orderConn.Close()
	grpcClientConnections.WithLabelValues("order").Inc()

	// Clients
	authClient := client.NewAuthClient(authConn)
	productClient := client.NewProductClient(productConn)
	orderClient := client.NewOrderClient(orderConn)

	// Services
	authService := service.NewAuthService(authClient.Client)
	productService := service.NewProductService(productClient.Client)
	orderService := service.NewOrderService(orderClient.Client)

	// Handlers
	authHandler := handlers.NewAuthHandler(authService, log)
	productHandler := handlers.NewProductHandler(productService, log)
	orderHandler := handlers.NewOrderHandler(orderService, log)

	// Gin router
	router := gin.Default()

	// Добавляем middleware для сбора метрик
	router.Use(PrometheusMiddleware())

	// Добавляем endpoint для Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Auth routes
	router.POST("/login", authHandler.Login)
	router.POST("/register", authHandler.Register)
	router.GET("/activate", authHandler.Activate)
	router.POST("/forgot-password", authHandler.ForgotPassword)
	router.POST("/reset-password", authHandler.ResetPassword)
	router.POST("/refresh-token", authHandler.RefreshToken)
	router.GET("/me", authHandler.GetMe)
	router.POST("/logout", authHandler.Logout)

	// Product routes
	router.GET("/products", productHandler.ListProducts)
	router.POST("/products/search", productHandler.SearchProducts)
	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products/:id", productHandler.GetProduct)
	router.PUT("/products/:id", productHandler.UpdateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)
	router.POST("/products/:id/stock", productHandler.ChangeProductStock)

	// Order routes
	router.POST("/orders", orderHandler.CreateOrder)
	router.GET("/orders/:id", orderHandler.GetOrder)
	router.GET("/orders/user/:user_id", orderHandler.ListOrders)
	router.POST("/orders/:id/cancel", orderHandler.CancelOrder)
	router.POST("/orders/:id/approve", orderHandler.ApproveOrder)

	log.Info("API Gateway запущен на :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Ошибка запуска API Gateway: %v", err)
	}
}