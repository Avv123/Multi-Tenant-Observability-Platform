package auth

import (
	"time"

	commonauth "github.com/omniful/pulselens-common/auth"
	platformauth "github.com/omniful/pulselens-platform/auth"
)

func GenerateToken(secret string, claims commonauth.Claims) (string, error) {
	return platformauth.Generate(secret, claims)
}

func ParseToken(secret, tokenString string) (*commonauth.Claims, error) {
	return platformauth.Parse(secret, tokenString)
}

func NewClaims(userID, tenantID, email, role string, expiryMinutes int) commonauth.Claims {
	return platformauth.NewClaims(userID, tenantID, email, role, time.Duration(expiryMinutes)*time.Minute)
}
