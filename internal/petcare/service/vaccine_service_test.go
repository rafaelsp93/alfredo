package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

// --- mock ---

type mockVaccineRepo struct {
	vaccines []domain.Vaccine
	err      error
}

func (m *mockVaccineRepo) ListVaccines(_ context.Context, _ string) ([]domain.Vaccine, error) {
	return m.vaccines, m.err
}
func (m *mockVaccineRepo) CreateVaccine(_ context.Context, v domain.Vaccine) (*domain.Vaccine, error) {
	return &v, m.err
}
func (m *mockVaccineRepo) GetVaccine(_ context.Context, _, _ string) (*domain.Vaccine, error) {
	if len(m.vaccines) == 0 {
		return nil, domain.ErrNotFound
	}
	return &m.vaccines[0], m.err
}
func (m *mockVaccineRepo) DeleteVaccine(_ context.Context, _, _ string) error { return m.err }

// --- tests ---

func TestVaccineService_RecordVaccine_AssignsID(t *testing.T) {
	svc := service.NewVaccineService(&mockVaccineRepo{}, &mockPetRepo{})
	v, err := svc.RecordVaccine(context.Background(), service.RecordVaccineInput{
		PetID: "p1", Name: "Rabies", AdministeredAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.ID == "" {
		t.Error("expected ID to be set")
	}
}

func TestVaccineService_RecordVaccine_ValidationError(t *testing.T) {
	svc := service.NewVaccineService(&mockVaccineRepo{}, &mockPetRepo{})
	_, err := svc.RecordVaccine(context.Background(), service.RecordVaccineInput{PetID: "p1", Name: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("got %v, want ErrValidation", err)
	}
}

func TestVaccineService_DeleteVaccine_NotFound(t *testing.T) {
	svc := service.NewVaccineService(&mockVaccineRepo{}, &mockPetRepo{})
	err := svc.DeleteVaccine(context.Background(), "p1", "v1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}
