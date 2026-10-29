package httperr

import (
	"errors"
	"net/http"
	"strings"

	"ksp/internal/dto"
	"ksp/internal/repository"
)

func FromService(err error, md5 string) (int, dto.ErrorResponse) {
	md5 = strings.ToLower(md5)

	switch {
	case errors.Is(err, repository.ErrNotFound):
		return http.StatusNotFound, dto.ErrorResponse{
			Message: "Threat not found",
			Details: "No threat with MD5 " + md5 + " found",
		}
	case errors.Is(err, repository.ErrConflict):
		return http.StatusConflict, dto.ErrorResponse{
			Message: "Threat already exists",
			Details: "Threat with MD5 " + md5 + " already exists",
		}
	default:
		return http.StatusInternalServerError, dto.ErrorResponse{
			Message: "Internal server error",
			Details: "Unexpected error while processing request",
		}
	}
}

func BadRequest(message, details string) (int, dto.ErrorResponse) {
	return http.StatusBadRequest, dto.ErrorResponse{
		Message: message,
		Details: details,
	}
}
