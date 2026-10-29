package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
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

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func TimeOutMiddleware(timeout int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if timeout <= 0 {
			c.Next()
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(timeout)*time.Second)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		if errors.Is(ctx.Err(), context.DeadlineExceeded) && !c.Writer.Written() {
			c.AbortWithStatusJSON(http.StatusGatewayTimeout, dto.ErrorResponse{
				Message: "Request timeout",
				Details: "Handler exceeded configured timeout",
			})
		}
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

func RecoveryMiddleware(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		logger.Error("panic recovered", "error", recovered, "path", c.Request.URL.Path)
		c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal server error",
			Details: "Unexpected panic while processing request",
		})
	})
}

func RateLimitMiddleware(rps float64, burst int) gin.HandlerFunc {
	if rps <= 0 {
		rps = 20
	}
	if burst <= 0 {
		burst = 40
	}

	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
		once     sync.Once
	)

	cleanup := func(maxIdle time.Duration) {
		mu.Lock()
		defer mu.Unlock()
		now := time.Now()
		for ip, v := range visitors {
			if now.Sub(v.lastSeen) > maxIdle {
				delete(visitors, ip)
			}
		}
	}

	once.Do(func() {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				cleanup(10 * time.Minute)
			}
		}()
	})

	return func(c *gin.Context) {
		ip := c.ClientIP()

		mu.Lock()
		v, ok := visitors[ip]
		if !ok {
			v = &visitor{limiter: rate.NewLimiter(rate.Limit(rps), burst)}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		limiter := v.limiter
		mu.Unlock()

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
