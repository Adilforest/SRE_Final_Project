package service

import (
    "context"
    "BikeStoreGolang/api-gateway/proto/auth"
)

type AuthService interface {
    Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error)
	Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error)
	Activate(ctx context.Context, req *authpb.ActivateRequest) (*authpb.ActivateResponse, error)
	ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (*authpb.ForgotPasswordResponse, error)
	ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.ResetPasswordResponse, error)
	RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error)
	GetMe(ctx context.Context, req *authpb.GetMeRequest) (*authpb.UserResponse, error)
	Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error)


}

type authService struct {
    client authpb.AuthServiceClient
}

func NewAuthService(client authpb.AuthServiceClient) AuthService {
    return &authService{client: client}
}

func (s *authService) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
    return s.client.Login(ctx, req)
}

func (s *authService) Register(ctx context.Context, req *authpb.RegisterRequest) (*authpb.RegisterResponse, error) {
    return s.client.Register(ctx, req)
}
func (s *authService) Activate(ctx context.Context, req *authpb.ActivateRequest) (*authpb.ActivateResponse, error) {
    return s.client.Activate(ctx, req)
}
func (s *authService) ForgotPassword(ctx context.Context, req *authpb.	ForgotPasswordRequest) (*authpb.ForgotPasswordResponse, error) {
    return s.client.ForgotPassword(ctx, req)
}
func (s *authService) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) (*authpb.ResetPasswordResponse, error) {
	return s.client.ResetPassword(ctx, req)
}
func (s *authService) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
    return s.client.RefreshToken(ctx, req)
}
func (s *authService) GetMe(ctx context.Context, req *authpb.GetMeRequest) (*authpb.UserResponse, error) {
    return s.client.GetMe(ctx, req)
}
func (s *authService) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
    return s.client.Logout(ctx, req)
}