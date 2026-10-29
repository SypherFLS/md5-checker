package router

import (
	"ksp/internal/api/handlers"
	"ksp/internal/api/httperr"
	"ksp/internal/api/middlewares"
	"ksp/internal/config"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handlers.Handler, cfg config.Config, appLogger *slog.Logger) *gin.Engine {
	rout := gin.New()
	rout.HandleMethodNotAllowed = true

	rout.Use(
		middlewares.RecoveryMiddleware(appLogger),
		middlewares.TraceMiddleware(),
		middlewares.LoggingMiddleware(appLogger),
		middlewares.RateLimitMiddleware(cfg.Server.RateLimitRPS, cfg.Server.RateLimitBurst),
		middlewares.TimeOutMiddleware(cfg.Server.Timeout),
	)

	rout.NoMethod(func(c *gin.Context) {
		status, body := httperr.MethodNotAllowed()
		c.JSON(status, body)
	})

	rout.GET("/info/:md5", h.GetHandler)
	rout.POST("/info", h.PostHandler)
	rout.PUT("/info/:md5", h.PutHandler)

	return rout
}
