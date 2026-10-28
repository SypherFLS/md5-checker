package router

import (
	"ksp/internal/api/handlers"
	"ksp/internal/config"
	"log/slog"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *handlers.Handler, cfg config.Config, appLogger *slog.Logger) *gin.Engine {
	rout := gin.New()

	return rout
}