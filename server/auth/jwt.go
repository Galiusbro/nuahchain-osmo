package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Note: jwt.RegisteredClaims is from v5
// If using v4, use jwt.StandardClaims instead

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID  int64  `json:"user_id"`
	Address string `json:"address"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secret        []byte
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, tokenExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		tokenExpiry:   tokenExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateToken generates a new JWT token
func (jm *JWTManager) GenerateToken(userID int64, address string) (string, error) {
	now := time.Now()
	claims := JWTClaims{
		UserID:  userID,
		Address: address,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.tokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jm.secret)
}

// GenerateRefreshToken generates a refresh token
func (jm *JWTManager) GenerateRefreshToken(userID int64) (string, error) {
	now := time.Now()

	// Use MapClaims for refresh token to store custom fields
	customClaims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"exp":     now.Add(jm.refreshExpiry).Unix(),
		"iat":     now.Unix(),
		"nbf":     now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, customClaims)
	return token.SignedString(jm.secret)
}

// ValidateToken validates a JWT token and returns claims
func (jm *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jm.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (jm *JWTManager) ValidateRefreshToken(tokenString string) (int64, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jm.secret, nil
	})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	if claims["type"] != "refresh" {
		return 0, errors.New("not a refresh token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("invalid user_id in token")
	}

	return int64(userID), nil
}
