package middleware

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	platformauth "github.com/omniful/pulselens-platform/auth"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/errs"
	platformresponse "github.com/omniful/pulselens-platform/response"
)

func AuthenticateJWT(ctx context.Context) gin.HandlerFunc {
	_ = ctx
	secret := config.GetString("jwt.secret")
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			platformresponse.Error(c, 401, errs.New("UNAUTHORIZED", "missing bearer token"))
			return
		}

		claims, err := platformauth.Parse(secret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			platformresponse.Error(c, 401, errs.New("UNAUTHORIZED", err.Error()))
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
