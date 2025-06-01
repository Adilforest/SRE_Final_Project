package usecase

import (
	"context"
	"errors"
	"time"

	"BikeStoreGolang/services/auth-service/internal/domain"
	"BikeStoreGolang/services/auth-service/internal/logger"
	"BikeStoreGolang/services/auth-service/internal/mail_sender"
	pb "BikeStoreGolang/services/auth-service/proto/gen"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthUsecase struct {
	users  *mongo.Collection
	logger logger.Logger
	sender mail_sender.Sender
	redis  *redis.Client
}

func (u *AuthUsecase) RedisClient() *redis.Client {
    return u.redis
}

func NewAuthUsecase(mongoClient *mongo.Client, dbName string, l logger.Logger, sender mail_sender.Sender, redisClient *redis.Client) *AuthUsecase {
	return &AuthUsecase{
		users:  mongoClient.Database(dbName).Collection("users"),
		logger: l,
		sender: sender,
		redis:  redisClient,
	}
}

func (u *AuthUsecase) Logger() logger.Logger {
    return u.logger
}

func (a *AuthUsecase) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	a.logger.Infof("Register attempt for email: %s", req.Email)

	count, err := a.users.CountDocuments(ctx, bson.M{"email": req.Email})
	if err != nil {
		a.logger.Errorf("Error checking existing user: %v", err)
		return nil, err
	}
	if count > 0 {
		a.logger.Warnf("User already exists: %s", req.Email)
		return nil, errors.New("user already exists")
	}

	hash, err := domain.HashPassword(req.Password)
	if err != nil {
		a.logger.Errorf("Password hash error: %v", err)
		return nil, err
	}

	activationToken, err := domain.GenerateToken()
	if err != nil {
		a.logger.Errorf("Activation token generation error: %v", err)
		return nil, err
	}

	user := domain.User{
		Name:              req.Name,
		Email:             req.Email,
		PasswordHash:      hash,
		Role:              domain.RoleCustomer,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		IsActive:          false,
		ActivationToken:   activationToken,
		ActivationExpires: time.Now().Add(24 * time.Hour),
	}

	res, err := a.users.InsertOne(ctx, user)
	if err != nil {
		a.logger.Errorf("Insert user error: %v", err)
		return nil, err
	}

	id := ""
	if oid, ok := res.InsertedID.(interface{ Hex() string }); ok {
		id = oid.Hex()
	}

	// Формируем ссылку для активации (замените на свой frontend/domain)
	activationLink := "http://localhost:8080/activate?token=" + activationToken

	// Отправляем письмо
	subject := "Activate your account"
	body := "Hello, " + req.Name + "!<br><br>Please activate your account by clicking the link below:<br><a href=\"" + activationLink + "\">Activate Account</a>"
	if err := a.sender.Send(req.Email, subject, body); err != nil {
		a.logger.Errorf("Failed to send activation email: %v", err)
		// Можно вернуть ошибку или продолжить, если email не критичен
	}

	a.logger.Infof("User registered successfully: %s (id: %s)", req.Email, id)
	return &pb.RegisterResponse{
		Id:      id,
		Message: "User registered successfully. Please check your email to activate your account.",
	}, nil
}

func (a *AuthUsecase) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	a.logger.Infof("Login attempt for email: %s", req.Email)

	var user domain.User
	err := a.users.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		a.logger.Warnf("Login failed - user not found: %s", req.Email)
		return nil, errors.New("invalid email or password")
	}

	if !user.IsActive {
		a.logger.Warnf("Login failed - account not activated for: %s", req.Email)
		return nil, errors.New("account not activated. Please check your email")
	}

	if !domain.CheckPasswordHash(req.Password, user.PasswordHash) {
		a.logger.Warnf("Login failed - wrong password for user: %s", req.Email)
		return nil, errors.New("invalid email or password")
	}

	accessToken, err := domain.GenerateJWT(user.ID.Hex(), string(user.Role), time.Hour)
	if err != nil {
		a.logger.Errorf("Failed to generate access token: %v", err)
		return nil, err
	}
	refreshToken, err := domain.GenerateJWT(user.ID.Hex(), string(user.Role), 24*time.Hour)
	if err != nil {
		a.logger.Errorf("Failed to generate refresh token: %v", err)
		return nil, err
	}

	var pbRole pb.Role
	switch string(user.Role) {
	case "admin":
		pbRole = pb.Role_ROLE_ADMIN
	case "customer":
		pbRole = pb.Role_ROLE_CUSTOMER
	default:
		pbRole = pb.Role_ROLE_CUSTOMER
	}

	a.logger.Infof("Login successful for user: %s", req.Email)
	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &pb.UserResponse{
			Id:    user.ID.Hex(),
			Name:  user.Name,
			Email: user.Email,
			Role:  pbRole,
		},
	}, nil
}

func (a *AuthUsecase) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	a.logger.Infof("Logout attempt")
	token := req.GetAccessToken()
	claims, err := domain.ParseJWT(token)
	if err != nil {
		a.logger.Warnf("Invalid token on logout")
		return nil, errors.New("invalid token")
	}
	exp := time.Unix(claims.ExpiresAt.Unix(), 0)
	ttl := time.Until(exp)
	if ttl > 0 {
		err := a.redis.Set(ctx, "blacklist:"+token, "1", ttl).Err()
		if err != nil {
			a.logger.Errorf("Failed to blacklist token: %v", err)
			return nil, err
		}
	}
	a.logger.Infof("Token blacklisted successfully")
	return &pb.LogoutResponse{Message: "Logged out successfully"}, nil
}

func (a *AuthUsecase) Activate(ctx context.Context, req *pb.ActivateRequest) (*pb.ActivateResponse, error) {
	a.logger.Infof("Activation attempt with token: %s", req.Token)

	filter := bson.M{"activation_token": req.Token, "is_active": false}
	update := bson.M{"$set": bson.M{"is_active": true, "activation_token": ""}}
	res, err := a.users.UpdateOne(ctx, filter, update)
	if err != nil {
		a.logger.Errorf("Activation failed: %v", err)
		return nil, err
	}
	if res.ModifiedCount == 0 {
		a.logger.Warnf("Activation token invalid or already used: %s", req.Token)
		return nil, errors.New("invalid or expired activation token")
	}
	a.logger.Infof("Account successfully activated with token: %s", req.Token)
	return &pb.ActivateResponse{Message: "Account activated"}, nil
}

func (a *AuthUsecase) ForgotPassword(ctx context.Context, req *pb.ForgotPasswordRequest) (*pb.ForgotPasswordResponse, error) {
	a.logger.Infof("Forgot password request for: %s", req.Email)

	resetToken, err := domain.GenerateToken()
	if err != nil {
		a.logger.Errorf("Failed to generate reset token: %v", err)
		return nil, err
	}
	filter := bson.M{"email": req.Email}
	update := bson.M{
		"$set": bson.M{
			"reset_token":   resetToken,
			"reset_expires": time.Now().Add(1 * time.Hour),
		},
	}
	res, err := a.users.UpdateOne(ctx, filter, update)
	if err != nil {
		a.logger.Errorf("Failed to set reset token: %v", err)
		return nil, err
	}
	if res.MatchedCount == 0 {
		a.logger.Warnf("Forgot password failed - user not found: %s", req.Email)
		return nil, errors.New("user not found")
	}
	a.logger.Infof("Reset token generated and saved for user: %s", req.Email)

	// Формируем ссылку для сброса пароля (замените на свой frontend/domain)
	resetLink := "http://localhost:8080/reset-password?token=" + resetToken

	// Отправляем письмо
	subject := "Reset your password"
	body := "Hello!<br><br>To reset your password, click the link below:<br><a href=\"" + resetLink + "\">Reset Password</a>"
	if err := a.sender.Send(req.Email, subject, body); err != nil {
		a.logger.Errorf("Failed to send reset password email: %v", err)
	}

	return &pb.ForgotPasswordResponse{Message: "Reset token sent to email"}, nil
}

func (a *AuthUsecase) ResetPassword(ctx context.Context, req *pb.ResetPasswordRequest) (*pb.ResetPasswordResponse, error) {
	a.logger.Infof("Reset password attempt with token: %s", req.Token)

	filter := bson.M{
		"reset_token":   req.Token,
		"reset_expires": bson.M{"$gt": time.Now()},
	}
	hash, err := domain.HashPassword(req.NewPassword)
	if err != nil {
		a.logger.Errorf("Password hash error during reset: %v", err)
		return nil, err
	}
	update := bson.M{
		"$set": bson.M{
			"password_hash": hash,
			"reset_token":   "",
			"reset_expires": time.Time{},
		},
	}
	res, err := a.users.UpdateOne(ctx, filter, update)
	if err != nil {
		a.logger.Errorf("Password reset error: %v", err)
		return nil, err
	}
	if res.ModifiedCount == 0 {
		a.logger.Warnf("Reset password failed - invalid or expired token: %s", req.Token)
		return nil, errors.New("invalid or expired reset token")
	}
	a.logger.Infof("Password reset successful for token: %s", req.Token)
	return &pb.ResetPasswordResponse{Message: "Password reset successful"}, nil
}

func (a *AuthUsecase) RefreshToken(ctx context.Context, req *pb.RefreshTokenRequest) (*pb.RefreshTokenResponse, error) {
	a.logger.Infof("Refresh token request")

	claims, err := domain.ParseJWT(req.RefreshToken)
	if err != nil {
		a.logger.Warnf("Invalid refresh token")
		return nil, errors.New("invalid refresh token")
	}
	accessToken, err := domain.GenerateJWT(claims.UserID, claims.Role, time.Hour)
	if err != nil {
		a.logger.Errorf("Error generating access token: %v", err)
		return nil, err
	}
	refreshToken, err := domain.GenerateJWT(claims.UserID, claims.Role, 24*time.Hour)
	if err != nil {
		a.logger.Errorf("Error generating refresh token: %v", err)
		return nil, err
	}
	a.logger.Infof("Tokens refreshed successfully for userID: %s", claims.UserID)
	return &pb.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (a *AuthUsecase) GetMe(ctx context.Context, userID string) (*pb.UserResponse, error) {
	a.logger.Infof("GetMe called for userID: %s", userID)

	var user domain.User
	objID, err := domain.ParseObjectID(userID)
	if err != nil {
		a.logger.Errorf("Invalid user ID: %v", err)
		return nil, errors.New("invalid user id")
	}
	err = a.users.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		a.logger.Warnf("User not found: %s", userID)
		return nil, errors.New("user not found")
	}

	// Явное сопоставление роли из базы с proto enum
	var pbRole pb.Role
	switch string(user.Role) {
	case "admin":
		pbRole = pb.Role_ROLE_ADMIN
	case "customer":
		pbRole = pb.Role_ROLE_CUSTOMER
	default:
		pbRole = pb.Role_ROLE_CUSTOMER
	}

	a.logger.Infof("User info retrieved for userID: %s", userID)
	return &pb.UserResponse{
		Id:        user.ID.Hex(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      pbRole,
		CreatedAt: domain.TimeToProtoTimestamp(user.CreatedAt),
	}, nil
}

func (a *AuthUsecase) ParseUserIDFromToken(tokenStr string) (string, error) {
    claims, err := domain.ParseJWT(tokenStr)
    if err != nil {
        a.logger.Errorf("Failed to parse JWT: %v", err)
        return "", err
    }
    if claims.UserID == "" {
        a.logger.Error("UserID not found in token claims")
        return "", errors.New("userID not found in token")
    }
    return claims.UserID, nil
}