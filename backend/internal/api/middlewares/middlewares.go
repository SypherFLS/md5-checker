package middlewares

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

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

	}
}
