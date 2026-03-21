package core

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtFallbackSecret = "my-broker-dev-fallback-secret"

func jwtSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret != "" {
		return secret
	}
	log.Println("WARN: JWT_SECRET is missing, using fallback secret")
	return jwtFallbackSecret
}

func IssueJWT(userID uint) (string, error) {
	secret := jwtSecret()

	claims := jwt.MapClaims{
		"uid": float64(userID),
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseJWT(tokenString string) (uint, error) {
	secret := jwtSecret()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	uid, ok := claims["uid"].(float64)
	if !ok || uid <= 0 {
		return 0, errors.New("invalid uid claim")
	}

	return uint(uid), nil
}
