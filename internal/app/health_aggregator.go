package app

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/shared/health"
)

// HealthAggregator combines health checks from all registered dependencies
// into a single /api/v1/health response.
type HealthAggregator struct {
	checkers map[string]HealthPinger
}

func NewHealthAggregator(checkers map[string]HealthPinger) *HealthAggregator {
	return &HealthAggregator{checkers: checkers}
}

func (h *HealthAggregator) Check(ctx context.Context) health.HealthResult {
	deps := make(map[string]health.DependencyStatus, len(h.checkers))
	allHealthy := true

	for name, checker := range h.checkers {
		if err := checker.Ping(ctx); err != nil {
			deps[name] = health.DependencyStatus{Status: "down", Error: err.Error()}
			allHealthy = false
		} else {
			deps[name] = health.DependencyStatus{Status: "up"}
		}
	}

	status := "healthy"
	if !allHealthy {
		status = "degraded"
	}
	return health.HealthResult{Status: status, Dependencies: deps}
}

// HealthHandler handles GET /api/v1/health — the unified health endpoint.
type HealthHandler struct {
	aggregator *HealthAggregator
}

func NewHealthHandler(aggregator *HealthAggregator) *HealthHandler {
	return &HealthHandler{aggregator: aggregator}
}

func (h *HealthHandler) Health(c echo.Context) error {
	result := h.aggregator.Check(c.Request().Context())
	if result.Status == "healthy" {
		return c.JSON(http.StatusOK, result)
	}
	return c.JSON(http.StatusServiceUnavailable, result)
}
