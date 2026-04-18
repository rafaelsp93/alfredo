package port

import (
	"context"

	"github.com/rafaelsoares/alfredo/internal/health/domain"
)

type ProfileRepository interface {
	Get(ctx context.Context) (domain.HealthProfile, error)
	Upsert(ctx context.Context, profile domain.HealthProfile) (domain.HealthProfile, error)
}
