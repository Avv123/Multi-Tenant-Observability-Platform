package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := ctx.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx.Set("request_id", requestID)
		ctx.Writer.Header().Set("X-Request-Id", requestID)
		ctx.Next()
	}
}

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		_ = start
	}
}
