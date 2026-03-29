package port

import (
	"context"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
)

// PetRepository persists and retrieves Pet records.
type PetRepository interface {
	List(ctx context.Context) ([]domain.Pet, error)
	Create(ctx context.Context, pet domain.Pet) (*domain.Pet, error)
	GetByID(ctx context.Context, id string) (*domain.Pet, error)
	Update(ctx context.Context, pet domain.Pet) (*domain.Pet, error)
	Delete(ctx context.Context, id string) error
}

// VaccineRepository persists vaccine records.
type VaccineRepository interface {
	ListVaccines(ctx context.Context, petID string) ([]domain.Vaccine, error)
	CreateVaccine(ctx context.Context, v domain.Vaccine) (*domain.Vaccine, error)
	GetVaccine(ctx context.Context, petID, vaccineID string) (*domain.Vaccine, error)
	DeleteVaccine(ctx context.Context, petID, vaccineID string) error
}

// DBHealthChecker verifies the SQLite connection is alive.
type DBHealthChecker interface {
	Ping(ctx context.Context) error
}
