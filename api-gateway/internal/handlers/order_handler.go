package handlers

import (
	"BikeStoreGolang/api-gateway/internal/logger"
	"BikeStoreGolang/api-gateway/internal/service"
	"BikeStoreGolang/api-gateway/proto/order"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	Service service.OrderService
	Logger  logger.Logger
}

func NewOrderHandler(s service.OrderService, l logger.Logger) *OrderHandler {
	return &OrderHandler{Service: s, Logger: l}
}

// POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req orderpb.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in CreateOrder: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.CreateOrder(ctx, &req)
	if err != nil {
		h.Logger.Warnf("CreateOrder failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Order created: %v", resp.Id)
	c.JSON(http.StatusOK, resp)
}

// GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Warn("Order ID required in GetOrder")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.GetOrder(ctx, &orderpb.GetOrderRequest{Id: id})
	if err != nil {
		h.Logger.Warnf("GetOrder failed: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Order retrieved: %v", resp.Id)
	c.JSON(http.StatusOK, resp)
}

// GET /orders/user/:user_id
func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.Logger.Warn("User ID required in ListOrders")
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	orders, err := h.Service.ListOrders(ctx, &orderpb.ListOrdersRequest{UserId: userID})
	if err != nil {
		h.Logger.Warnf("ListOrders failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Orders listed for user: %v", userID)
	c.JSON(http.StatusOK, orders)
}

// POST /orders/:id/cancel
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Warn("Order ID required in CancelOrder")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.CancelOrder(ctx, &orderpb.CancelOrderRequest{Id: id})
	if err != nil {
		h.Logger.Warnf("CancelOrder failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Order cancelled: %v", id)
	c.JSON(http.StatusOK, resp)
}

// POST /orders/:id/approve
func (h *OrderHandler) ApproveOrder(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Warn("Order ID required in ApproveOrder")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.ApproveOrder(ctx, &orderpb.ApproveOrderRequest{Id: id})
	if err != nil {
		h.Logger.Warnf("ApproveOrder failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Order approved: %v", id)
	c.JSON(http.StatusOK, resp)
}
