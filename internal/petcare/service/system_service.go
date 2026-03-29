package service

import (
	"context"

	"github.com/rafaelsoares/alfredo/internal/petcare/port"
)

type HealthResult struct {
	Status       string
	Dependencies map[string]DependencyStatus
}

type DependencyStatus struct {
	Status string
	Error  string
}

// SystemService checks infrastructure health.
type SystemService struct {
	checker port.DBHealthChecker
}

func NewSystemService(checker port.DBHealthChecker) *SystemService {
	return &SystemService{checker: checker}
}

func (s *SystemService) Check(ctx context.Context) HealthResult {
	depStatus := DependencyStatus{Status: "healthy"}
	if err := s.checker.Ping(ctx); err != nil {
		depStatus = DependencyStatus{Status: "unhealthy", Error: err.Error()}
	}

	overall := "healthy"
	if depStatus.Status != "healthy" {
		overall = "degraded"
	}

	return HealthResult{
		Status:       overall,
		Dependencies: map[string]DependencyStatus{"sqlite": depStatus},
	}
}
