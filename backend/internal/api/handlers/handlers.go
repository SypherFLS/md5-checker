package handlers

import (
	"log/slog"
	"net/http"

	"ksp/internal/api/httperr"
	"ksp/internal/dto"
	"ksp/internal/services"
	"ksp/internal/validation"

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
	md5 := c.Param("md5")
	if !validation.IsValidMD5(md5) {
		status, body := httperr.BadRequest("Invalid MD5", "MD5 must be 32 hexadecimal characters")
		c.JSON(status, body)
		return
	}

	threat, err := h.service.GetByMD5(c.Request.Context(), md5)
	if err != nil {
		h.writeServiceError(c, err, md5)
		return
	}

	c.JSON(http.StatusOK, threat)
}

func (h *Handler) PostHandler(c *gin.Context) {
	var req dto.ThreatCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := httperr.BadRequest("Invalid request body", err.Error())
		c.JSON(status, body)
		return
	}

	if err := validation.Validate(req); err != nil {
		status, body := httperr.BadRequest("Validation failed", err.Error())
		c.JSON(status, body)
		return
	}

	threat, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		h.writeServiceError(c, err, req.MD5)
		return
	}

	c.JSON(http.StatusCreated, threat)
}

func (h *Handler) PutHandler(c *gin.Context) {
	md5 := c.Param("md5")
	if !validation.IsValidMD5(md5) {
		status, body := httperr.BadRequest("Invalid MD5", "MD5 must be 32 hexadecimal characters")
		c.JSON(status, body)
		return
	}

	var req dto.ThreatUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		status, body := httperr.BadRequest("Invalid request body", err.Error())
		c.JSON(status, body)
		return
	}

	if err := validation.Validate(req); err != nil {
		status, body := httperr.BadRequest("Validation failed", err.Error())
		c.JSON(status, body)
		return
	}

	threat, err := h.service.Update(c.Request.Context(), md5, req)
	if err != nil {
		h.writeServiceError(c, err, md5)
		return
	}

	c.JSON(http.StatusOK, threat)
}

func (h *Handler) writeServiceError(c *gin.Context, err error, md5 string) {
	status, body := httperr.FromService(err, md5)
	if status >= http.StatusInternalServerError {
		h.appLogger.Error("request failed", "error", err, "md5", md5)
	}
	c.JSON(status, body)
}
