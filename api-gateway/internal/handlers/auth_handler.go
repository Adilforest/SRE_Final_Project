package handlers

import (
	"BikeStoreGolang/api-gateway/internal/logger"
	"BikeStoreGolang/api-gateway/internal/service"
	authpb "BikeStoreGolang/api-gateway/proto/auth"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

type AuthHandler struct {
	Service service.AuthService
	Logger  logger.Logger
}

// NewAuthHandler creates a new AuthHandler with injected AuthService and Logger.
func NewAuthHandler(s service.AuthService, l logger.Logger) *AuthHandler {
	return &AuthHandler{
		Service: s,
		Logger:  l,
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req authpb.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in Login: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.Login(ctx, &req)
	if err != nil {
		h.Logger.Warnf("Login failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("User logged in: %s", req.Email)
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req authpb.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in Register: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.Register(ctx, &req)
	if err != nil {
		h.Logger.Warnf("Register failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("User registered: %s", req.Email)
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Activate(c *gin.Context) {
	token := c.Query("token")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.Activate(ctx, &authpb.ActivateRequest{Token: token})
	if err != nil {
		h.Logger.Warnf("Activation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Info("User activated")
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req authpb.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in ForgotPassword: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.ForgotPassword(ctx, &req)
	if err != nil {
		h.Logger.Warnf("ForgotPassword failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Infof("Forgot password requested for: %s", req.Email)
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req authpb.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in ResetPassword: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.ResetPassword(ctx, &req)
	if err != nil {
		h.Logger.Warnf("ResetPassword failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Info("Password reset")
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req authpb.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in RefreshToken: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.RefreshToken(ctx, &req)
	if err != nil {
		h.Logger.Warnf("RefreshToken failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Info("Token refreshed")
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token := c.GetHeader("Authorization")
	if token == "" {
		h.Logger.Warn("No Authorization header in HTTP request")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no authorization header"})
		return
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "authorization", token)

	resp, err := h.Service.GetMe(ctx, &authpb.GetMeRequest{})
	if err != nil {
		h.Logger.Warnf("GetMe failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Info("GetMe called")
	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req authpb.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.Warn("Invalid JSON in Logout: ", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := h.Service.Logout(ctx, &req)
	if err != nil {
		h.Logger.Warnf("Logout failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "gRPC error: " + err.Error()})
		return
	}
	h.Logger.Info("User logged out")
	c.JSON(http.StatusOK, resp)
}
