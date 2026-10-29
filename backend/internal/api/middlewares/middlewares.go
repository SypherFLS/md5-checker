package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"ksp/internal/dto"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type ctxKey string

const TraceIDKey ctxKey = "trace_id"

func TimeOutMiddleware(timeout int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(timeout)*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := c.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = newTraceID()
		}

		ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
		c.Request = c.Request.WithContext(ctx)
		c.Set(string(TraceIDKey), traceID)
		c.Header("X-Trace-Id", traceID)
		c.Next()
	}
}

func LoggingMiddleware(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		traceID, _ := c.Get(string(TraceIDKey))
		logger.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
			"trace_id", traceID,
		)
	}
}

func RateLimitMiddleware(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 {
		rps = 20
	}
	if burst <= 0 {
		burst = 40
	}

	var limiters sync.Map

	return func(c *gin.Context) {
		ip := c.ClientIP()
		val, _ := limiters.LoadOrStore(ip, rate.NewLimiter(rate.Limit(rps), burst))
		limiter := val.(*rate.Limiter)

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Message: "Rate limit exceeded",
				Details: "Too many requests from " + ip,
			})
			return
		}

		c.Next()
	}
}

func newTraceID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000000")))
	}
	return hex.EncodeToString(b)
}
