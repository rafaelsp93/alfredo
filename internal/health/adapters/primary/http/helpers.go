package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/health/domain"
)

type errorResponse struct {
	Error   string       `json:"error"`
	Message string       `json:"message"`
	Fields  []fieldError `json:"fields,omitempty"`
}

type fieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

func newErrorResponse(code, message string, fields []fieldError) errorResponse {
	return errorResponse{Error: code, Message: message, Fields: fields}
}

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return c.JSON(http.StatusNotFound, newErrorResponse("not_found", "Resource not found", nil))
	case errors.Is(err, domain.ErrValidation):
		return c.JSON(http.StatusBadRequest, newErrorResponse("validation_failed", "Request validation failed", nil))
	default:
		return c.JSON(http.StatusInternalServerError, newErrorResponse("internal_error", "An unexpected error occurred", nil))
	}
}
