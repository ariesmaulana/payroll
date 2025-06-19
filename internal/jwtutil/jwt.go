package jwtutil

import (
	"errors"
	"time"

	"github.com/ariesmaulana/payroll/data"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecretKey []byte

func SetSecret(secret string) {
	jwtSecretKey = []byte(secret)
}

type Claims struct {
	UserID   int           `json:"user_id"`
	Username string        `json:"username"`
	Role     data.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// GenerateJWT bikin token baru
func GenerateJWT(userID int, username string, role data.UserRole) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

// ValidateJWT parsing token dan balikin claim
func ValidateJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
