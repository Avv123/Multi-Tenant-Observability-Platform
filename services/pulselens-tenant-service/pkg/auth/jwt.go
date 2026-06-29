package auth

import (
	"time"

	commonauth "github.com/Avv123/pulselens-common/auth"
	platformauth "github.com/Avv123/pulselens-platform/auth"
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
