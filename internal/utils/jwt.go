package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nahuelmarianolosada/el-campeon-web/internal/config"
)

type JWTClaims struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"` // "access" o "refresh"
	jwt.RegisteredClaims
}

// GenerateAccessToken genera un JWT de acceso
func GenerateAccessToken(userID uint, email, role string, cfg *config.Config) (string, error) {
	expiryTime := time.Now().Add(time.Duration(cfg.JWTExpiryHours) * time.Hour)
	claims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "el-campeon-web",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecretKey))
}

// GenerateRefreshToken genera un JWT de refresco
func GenerateRefreshToken(userID uint, email, role string, cfg *config.Config) (string, error) {
	expiryTime := time.Now().Add(7 * 24 * time.Hour) // 7 días
	claims := &JWTClaims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiryTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "el-campeon-web",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTRefreshSecret))
}

// ValidateToken valida y parsea un JWT
func ValidateToken(tokenString string, cfg *config.Config) (*JWTClaims, error) {
	claims := &JWTClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Seleccionar la clave basada en el tipo de token
		if claims, ok := token.Claims.(*JWTClaims); ok {
			if claims.TokenType == "refresh" {
				return []byte(cfg.JWTRefreshSecret), nil
			}
		}
		return []byte(cfg.JWTSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// ValidateAccessToken valida específicamente un token de acceso
func ValidateAccessToken(tokenString string, cfg *config.Config) (*JWTClaims, error) {
	claims, err := ValidateToken(tokenString, cfg)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

// ValidateRefreshToken valida específicamente un token de refresco
func ValidateRefreshToken(tokenString string, cfg *config.Config) (*JWTClaims, error) {
	claims, err := ValidateToken(tokenString, cfg)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}
