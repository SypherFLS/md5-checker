package router

import (
	"ksp/internal/api/handlers"
	"ksp/internal/api/middlewares"
	"ksp/internal/config"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handlers.Handler, cfg config.Config, appLogger *slog.Logger) *gin.Engine {
	rout := gin.New()

	rout.Use(
		middlewares.TimeOutMiddleware(cfg.Server.Timeout),
	)
	rout.GET("/info/{md5}", h.GetHandler)
	rout.POST("/info", h.PostHandler)
	rout.PUT("/info/{md5}", h.PutHandler)

	return rout
}
