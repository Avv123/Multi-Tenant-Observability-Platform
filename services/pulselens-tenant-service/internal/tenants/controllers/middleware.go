package controllers

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omniful/pulselens-platform/config"
	platformerrs "github.com/omniful/pulselens-platform/errs"
	serviceauth "github.com/omniful/pulselens-tenant-service/pkg/auth"
	tenanterror "github.com/omniful/pulselens-tenant-service/pkg/error"
)

func AuthenticateInternalToken(ctx context.Context) gin.HandlerFunc {
	_ = ctx
	expected := config.GetString("internal.token")
	return func(c *gin.Context) {
		if c.GetHeader("X-Internal-Token") != expected {
			tenanterror.NewErrorResponse(c, platformerrs.New(tenanterror.Forbidden, "invalid internal token"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func AuthenticateJWT(ctx context.Context) gin.HandlerFunc {
	_ = ctx
	secret := config.GetString("jwt.secret")
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			tenanterror.NewErrorResponse(c, platformerrs.New(tenanterror.Unauthorized, "missing bearer token"))
			c.Abort()
			return
		}
		claims, err := serviceauth.ParseToken(secret, strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			tenanterror.NewErrorResponse(c, platformerrs.New(tenanterror.Unauthorized, err.Error()))
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
