package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

// --- mock repo ---

type mockPetRepo struct {
	pets   []domain.Pet
	pet    *domain.Pet // for GetByID
	stored domain.Pet
	err    error
}

func (m *mockPetRepo) List(_ context.Context) ([]domain.Pet, error) { return m.pets, m.err }
func (m *mockPetRepo) Create(_ context.Context, p domain.Pet) (*domain.Pet, error) {
	m.stored = p
	return &p, m.err
}
func (m *mockPetRepo) GetByID(_ context.Context, id string) (*domain.Pet, error) {
	if m.pet != nil {
		return m.pet, m.err
	}
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.pets {
		if m.pets[i].ID == id {
			return &m.pets[i], nil
		}
	}
	return nil, domain.ErrNotFound
}
func (m *mockPetRepo) Update(_ context.Context, p domain.Pet) (*domain.Pet, error) { return &p, m.err }
func (m *mockPetRepo) Delete(_ context.Context, _ string) error                    { return m.err }

// --- tests ---

func TestPetService_Create_AssignsIDAndCreatedAt(t *testing.T) {
	repo := &mockPetRepo{}
	svc := service.NewPetService(repo)

	pet, err := svc.Create(context.Background(), service.CreatePetInput{Name: "Rex", Species: "dog"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pet.ID == "" {
		t.Error("expected ID to be set")
	}
	if pet.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestPetService_Create_ValidationError_EmptyName(t *testing.T) {
	svc := service.NewPetService(&mockPetRepo{})
	_, err := svc.Create(context.Background(), service.CreatePetInput{Name: "", Species: "dog"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("got %v, want ErrValidation", err)
	}
}

func TestPetService_Create_ValidationError_EmptySpecies(t *testing.T) {
	svc := service.NewPetService(&mockPetRepo{})
	_, err := svc.Create(context.Background(), service.CreatePetInput{Name: "Rex", Species: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("got %v, want ErrValidation", err)
	}
}

func TestPetService_List(t *testing.T) {
	repo := &mockPetRepo{pets: []domain.Pet{{ID: "p1", Name: "Rex", Species: "dog", CreatedAt: time.Now()}}}
	svc := service.NewPetService(repo)

	pets, err := svc.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pets) != 1 {
		t.Errorf("got %d pets, want 1", len(pets))
	}
}

func TestPetService_Update_ValidationError_EmptyName(t *testing.T) {
	repo := &mockPetRepo{pets: []domain.Pet{{ID: "p1", Name: "Rex", Species: "dog", CreatedAt: time.Now()}}}
	svc := service.NewPetService(repo)
	_, err := svc.Update(context.Background(), "p1", service.UpdatePetInput{Name: "", Species: "dog"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("got %v, want ErrValidation", err)
	}
}

func TestPetService_Update_ValidationError_EmptySpecies(t *testing.T) {
	repo := &mockPetRepo{pets: []domain.Pet{{ID: "p1", Name: "Rex", Species: "dog", CreatedAt: time.Now()}}}
	svc := service.NewPetService(repo)
	_, err := svc.Update(context.Background(), "p1", service.UpdatePetInput{Name: "Rex", Species: ""})
	if !errors.Is(err, domain.ErrValidation) {
		t.Errorf("got %v, want ErrValidation", err)
	}
}

func TestPetService_Update_Success(t *testing.T) {
	repo := &mockPetRepo{pets: []domain.Pet{{ID: "p1", Name: "Rex", Species: "dog", CreatedAt: time.Now()}}}
	svc := service.NewPetService(repo)
	updated, err := svc.Update(context.Background(), "p1", service.UpdatePetInput{Name: "Max", Species: "dog"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "Max" {
		t.Errorf("got %q, want Max", updated.Name)
	}
}

func TestPetService_Delete_PropagatesError(t *testing.T) {
	repo := &mockPetRepo{err: domain.ErrNotFound}
	svc := service.NewPetService(repo)
	err := svc.Delete(context.Background(), "missing")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("got %v, want ErrNotFound", err)
	}
}
