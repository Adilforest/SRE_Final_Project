package handlers

import (
	"BikeStoreGolang/api-gateway/internal/logger"
	"BikeStoreGolang/api-gateway/internal/service"
	"BikeStoreGolang/api-gateway/proto/product"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

type ProductHandler struct {
	Service service.ProductService
	Logger  logger.Logger
}

func NewProductHandler(s service.ProductService, l logger.Logger) *ProductHandler {
	return &ProductHandler{Service: s, Logger: l}
}

// GET /products
func (h *ProductHandler) ListProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var filter productpb.ProductFilter
	if err := c.ShouldBindJSON(&filter); err != nil && c.Request.Body != http.NoBody {
		h.Logger.Warn("Invalid filter JSON in ListProducts: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter JSON"})
		return
	}

	products, err := h.Service.ListProducts(ctx, &filter)
	if err != nil {
		h.Logger.Warnf("ListProducts failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Info("Products listed")
	c.JSON(http.StatusOK, products)
}

// POST /products
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req productpb.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in CreateProduct: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	token := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}
	resp, err := h.Service.CreateProduct(ctx, &req)
	if err != nil {
		h.Logger.Warnf("CreateProduct failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Product created: %v", resp.Id)
	c.JSON(http.StatusOK, resp)
}

// GET /products/:id
func (h *ProductHandler) GetProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Warn("Product ID required in GetProduct")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID required"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.GetProduct(ctx, &productpb.GetProductRequest{Id: id})
	if err != nil {
		h.Logger.Warnf("GetProduct failed: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Product retrieved: %v", id)
	c.JSON(http.StatusOK, resp)
}

// PUT /products/:id
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Warn("Product ID required in UpdateProduct")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID required"})
		return
	}
	var req productpb.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in UpdateProduct: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	req.Id = id
	token := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}
	resp, err := h.Service.UpdateProduct(ctx, &req)
	if err != nil {
		h.Logger.Warnf("UpdateProduct failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Product updated: %v", id)
	c.JSON(http.StatusOK, resp)
}

// DELETE /products/:id
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Warn("Product ID required in DeleteProduct")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID required"})
		return
	}
	token := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}
	_, err := h.Service.DeleteProduct(ctx, &productpb.DeleteProductRequest{Id: id})
	if err != nil {
		h.Logger.Warnf("DeleteProduct failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Product deleted: %v", id)
	c.Status(http.StatusNoContent)
}

// POST /products/search
func (h *ProductHandler) SearchProducts(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var req productpb.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in SearchProducts: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	products, err := h.Service.SearchProducts(ctx, &req)
	if err != nil {
		h.Logger.Warnf("SearchProducts failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Info("Products searched")
	c.JSON(http.StatusOK, products)
}

// POST /products/:id/stock
func (h *ProductHandler) ChangeProductStock(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.Logger.Warn("Product ID required in ChangeProductStock")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID required"})
		return
	}
	var req productpb.ChangeStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in ChangeProductStock: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	req.ProductId = id
	token := c.GetHeader("Authorization")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)
	}
	resp, err := h.Service.ChangeProductStock(ctx, &req)
	if err != nil {
		h.Logger.Warnf("ChangeProductStock failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Product stock changed: %v", id)
	c.JSON(http.StatusOK, resp)
}
