package utils

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("development_key_123")
	}
	return []byte(secret)
}

// 24 hour exp time
func GenerateJWT(userID uint) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix() // Token expires in 24 hours
	claims["iat"] = time.Now().Unix()                     // Issued at timestamp
	tokenString, err := token.SignedString(getJWTSecret())
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	return tokenString, nil
}

func ValidateJWT(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return getJWTSecret(), nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userIDFloat, ok := claims["user_id"].(float64); ok {
			return uint(userIDFloat), nil
		}
		return 0, errors.New("invalid user_id format in token string")
	}
	return 0, errors.New("invalid token payload")
}
