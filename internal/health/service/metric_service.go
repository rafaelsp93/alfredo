package service

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelsoares/alfredo/internal/health/domain"
	"github.com/rafaelsoares/alfredo/internal/health/port"
)

type MetricService struct {
	repo           port.MetricRepository
	rawImportRepo  port.RawImportRepository
}

func NewMetricService(repo port.MetricRepository, rawImportRepo port.RawImportRepository) *MetricService {
	return &MetricService{
		repo:          repo,
		rawImportRepo: rawImportRepo,
	}
}

func (s *MetricService) Import(ctx context.Context, metrics []domain.DailyMetric, payload string, importedAt time.Time) (int, error) {
	count, err := s.repo.BulkUpsert(ctx, metrics)
	if err != nil {
		return 0, fmt.Errorf("import metrics: %w", err)
	}

	// Store raw payload for audit trail — best-effort, does not fail the import
	_ = s.rawImportRepo.Store(ctx, "metrics", payload, importedAt)

	return count, nil
}

func (s *MetricService) List(ctx context.Context, metricType string, from, to time.Time) ([]domain.DailyMetric, error) {
	metrics, err := s.repo.List(ctx, metricType, from, to)
	if err != nil {
		return nil, fmt.Errorf("list metrics: %w", err)
	}
	return metrics, nil
}
