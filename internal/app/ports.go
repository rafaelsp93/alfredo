package app

import (
	"context"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"github.com/rafaelsoares/alfredo/internal/shared/health"
)

// PetNameGetter allows use cases to look up a pet's name for event payloads.
// Satisfied by petcare/service.PetService.GetByID.
type PetNameGetter interface {
	GetByID(ctx context.Context, id string) (*domain.Pet, error)
}

// HealthPinger is the narrow health check interface used by HealthAggregator.
type HealthPinger interface {
	Ping(ctx context.Context) error
}

// --- Pet-care service interfaces (used by Use Cases) ---

type PetCareServicer interface {
	List(ctx context.Context) ([]domain.Pet, error)
	Create(ctx context.Context, in service.CreatePetInput) (*domain.Pet, error)
	GetByID(ctx context.Context, id string) (*domain.Pet, error)
	Update(ctx context.Context, id string, in service.UpdatePetInput) (*domain.Pet, error)
	Delete(ctx context.Context, id string) error
}

type VaccineServicer interface {
	ListVaccines(ctx context.Context, petID string) ([]domain.Vaccine, error)
	RecordVaccine(ctx context.Context, in service.RecordVaccineInput) (*domain.Vaccine, error)
	DeleteVaccine(ctx context.Context, petID, vaccineID string) error
}

// HealthResult mirrors shared/health.HealthResult (re-exported for convenience).
type HealthResult = health.HealthResult
