package grpc

import (
	//"BikeStoreGolang/services/auth-service/internal/domain"
	"context"
	"strings"

	"BikeStoreGolang/services/auth-service/internal/usecase"
	pb "BikeStoreGolang/services/auth-service/proto/gen"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	pb.UnimplementedAuthServiceServer
	uc *usecase.AuthUsecase
}

func NewAuthHandler(uc *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{uc: uc}
}
// Removed erroneous RedisClient method for undefined AuthUsecase type.

func (h *AuthHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return h.uc.Register(ctx, req)
}

func (h *AuthHandler) Activate(ctx context.Context, req *pb.ActivateRequest) (*pb.ActivateResponse, error) {
	return h.uc.Activate(ctx, req)
}

func (h *AuthHandler) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return h.uc.Login(ctx, req)
}

func (h *AuthHandler) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return h.uc.Logout(ctx, req)
}

func (h *AuthHandler) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	return h.uc.ForgotPassword(ctx, req)
}

func (h *AuthHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	return h.uc.ResetPassword(ctx, req)
}

func (h *AuthHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	return h.uc.RefreshToken(ctx, req)
}

func (h *AuthHandler) GetMe(ctx context.Context, req *pb.GetMeRequest) (*pb.UserResponse, error) {
    h.uc.Logger().Info("GetMe called")

    md, ok := metadata.FromIncomingContext(ctx)
    if !ok {
        h.uc.Logger().Warn("No metadata provided in context")
        return nil, status.Error(codes.Unauthenticated, "no metadata provided")
    }
    authHeaders := md["authorization"]
    if len(authHeaders) == 0 {
        h.uc.Logger().Warn("No authorization header in metadata")
        return nil, status.Error(codes.Unauthenticated, "no authorization header")
    }
    token := strings.TrimPrefix(authHeaders[0], "Bearer ")
    h.uc.Logger().Infof("Authorization token received: %s...", token[:10])

    // Проверка токена в Redis blacklist
    exists, err := h.uc.RedisClient().Exists(ctx, "blacklist:"+token).Result()
    if err != nil {
        h.uc.Logger().Errorf("Redis error: %v", err)
        return nil, status.Error(codes.Internal, "redis error")
    }
    if exists == 1 {
        h.uc.Logger().Warn("Token is blacklisted")
        return nil, status.Error(codes.Unauthenticated, "token is blacklisted")
    }

    userID, err := h.uc.ParseUserIDFromToken(token)
    if err != nil {
        return nil, status.Error(codes.Unauthenticated, "invalid token")
    }

    return h.uc.GetMe(ctx, userID)
}

