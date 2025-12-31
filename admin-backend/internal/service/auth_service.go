package service

import (
	"context"
	"errors"
	"yflow/internal/config"
	"yflow/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaim 定义JWT的claim
type JWTClaim struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthService 认证服务实现
type AuthService struct {
	jwtConfig config.JWTConfig
}

// NewAuthService 创建认证服务实例
func NewAuthService(jwtConfig config.JWTConfig) *AuthService {
	return &AuthService{
		jwtConfig: jwtConfig,
	}
}

// GenerateToken 生成JWT token
func (s *AuthService) GenerateToken(ctx context.Context, user *domain.User) (string, error) {
	// 设置token有效期
	expirationTime := time.Now().Add(time.Hour * time.Duration(s.jwtConfig.ExpirationHours))

	// 创建claims
	claims := &JWTClaim{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "yflow-admin",
			Subject:   user.Username,
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成刷新token
func (s *AuthService) GenerateRefreshToken(ctx context.Context, user *domain.User) (string, error) {
	// 设置refresh token有效期(更长)
	expirationTime := time.Now().Add(time.Hour * time.Duration(s.jwtConfig.RefreshExpirationHours))

	// 创建claims
	claims := &JWTClaim{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "yflow-admin-refresh",
			Subject:   user.Username,
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(s.jwtConfig.RefreshSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT token
func (s *AuthService) ValidateToken(ctx context.Context, tokenString string) (*domain.User, error) {
	claims, err := s.parseToken(tokenString, s.jwtConfig.Secret)
	if err != nil {
		return nil, err
	}

	// 返回用户信息
	return &domain.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}, nil
}

// ValidateRefreshToken 验证刷新token
func (s *AuthService) ValidateRefreshToken(ctx context.Context, tokenString string) (*domain.User, error) {
	claims, err := s.parseToken(tokenString, s.jwtConfig.RefreshSecret)
	if err != nil {
		return nil, err
	}

	// 返回用户信息
	return &domain.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}, nil
}

// parseToken 解析token的通用方法
func (s *AuthService) parseToken(tokenString, secret string) (*JWTClaim, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	// 验证token是否有效
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 验证并返回claims
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		return nil, errors.New("couldn't parse claims")
	}

	// 检查token是否过期
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token expired")
	}

	return claims, nil
}
