package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rafaelsoares/alfredo/internal/app"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

type TxRunner struct {
	db *sql.DB
}

func NewTxRunner(db *sql.DB) *TxRunner {
	return &TxRunner{db: db}
}

func (r *TxRunner) WithinTx(ctx context.Context, fn func(pets *service.PetService, vaccines *service.VaccineService, treatments *service.TreatmentService, doses *service.DoseService) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	petRepo := NewPetRepository(tx)
	vaccineRepo := NewVaccineRepository(tx)
	treatmentRepo := NewTreatmentRepository(tx)
	doseRepo := NewDoseRepository(tx)

	pets := service.NewPetService(petRepo)
	vaccines := service.NewVaccineService(vaccineRepo, petRepo)
	treatments := service.NewTreatmentService(treatmentRepo)
	doses := service.NewDoseService(doseRepo)

	if err := fn(pets, vaccines, treatments, doses); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%w: %v", app.ErrTxCommit, err)
	}
	return nil
}
