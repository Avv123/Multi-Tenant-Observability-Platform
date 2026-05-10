package auth

import (
	"errors"
	"time"

	commonauth "github.com/omniful/pulselens-common/auth"

	"github.com/golang-jwt/jwt/v5"
)

func NewClaims(userID, tenantID, email, role string, expiry time.Duration) commonauth.Claims {
	return commonauth.NewClaims(userID, tenantID, email, role, expiry)
}

func Generate(secret string, claims commonauth.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Parse(secret, tokenString string) (*commonauth.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &commonauth.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*commonauth.Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
