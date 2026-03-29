package http

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

// SystemChecker is the consumer-defined interface for the petcare system health check.
// Used internally — the monolith exposes a unified /api/v1/health via the app layer.
type SystemChecker interface {
	Check(ctx context.Context) service.HealthResult
}

type SystemHandler struct{ checker SystemChecker }

func NewSystemHandler(checker SystemChecker) *SystemHandler {
	return &SystemHandler{checker: checker}
}

type depStatusResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type healthResponse struct {
	Status       string                       `json:"status"`
	Dependencies map[string]depStatusResponse `json:"dependencies"`
}

// Health handles GET /api/v1/health (petcare-only health check, not exposed in monolith)
func (h *SystemHandler) Health(c echo.Context) error {
	result := h.checker.Check(c.Request().Context())

	deps := make(map[string]depStatusResponse, len(result.Dependencies))
	for k, v := range result.Dependencies {
		deps[k] = depStatusResponse{Status: v.Status, Error: v.Error}
	}

	resp := healthResponse{Status: result.Status, Dependencies: deps}

	if result.Status == "healthy" {
		return c.JSON(http.StatusOK, resp)
	}
	return c.JSON(http.StatusServiceUnavailable, resp)
}
