package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID injects a correlation ID into every request. If the caller provides
// X-Request-Id it is honoured; otherwise a UUID is generated. The ID is set on
// the response header so callers can correlate logs end-to-end. (B9)
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

// Logger emits a structured log line after every request. (B9)
// Format: [METHOD] /path status=NNN latency=Nms request_id=XXX
func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		latency := time.Since(start)
		requestID, _ := ctx.Get("request_id")
		fmt.Printf("[%s] %s status=%d latency=%dms request_id=%v\n",
			ctx.Request.Method,
			ctx.Request.URL.Path,
			ctx.Writer.Status(),
			latency.Milliseconds(),
			requestID,
		)
	}
}
