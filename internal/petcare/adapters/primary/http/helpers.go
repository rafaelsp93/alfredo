package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/logger"
)

func mapError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		logger.SetError(c, "not_found")
		return c.JSON(http.StatusNotFound, errorBody("not_found"))
	case errors.Is(err, domain.ErrValidation):
		logger.SetError(c, "validation_failed")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "validation_failed", "detail": err.Error()})
	default:
		logger.SetError(c, "internal_error")
		return c.JSON(http.StatusInternalServerError, errorBody("internal_error"))
	}
}

func errorBody(msg string) map[string]string { return map[string]string{"error": msg} }
