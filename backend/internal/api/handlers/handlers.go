package handlers

import (
	"ksp/internal/services"
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service   *services.Service
	appLogger *slog.Logger
}

func NewHandler(service *services.Service, appLogger *slog.Logger) *Handler {
	return &Handler{
		service:   service,
		appLogger: appLogger,
	}
}

func (h *Handler) GetHandler(c *gin.Context) {

}

func (h *Handler) PostHandler(c *gin.Context) {

}

func (h *Handler) PutHandler(c *gin.Context) {

}
