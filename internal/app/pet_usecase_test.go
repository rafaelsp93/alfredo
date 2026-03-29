package app_test

import (
	"context"
	"testing"

	"github.com/rafaelsoares/alfredo/internal/app"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

// stubPetService returns a preset pet on Create; panics on unimplemented methods.
type stubPetService struct {
	pet *domain.Pet
	err error
}

func (s *stubPetService) List(_ context.Context) ([]domain.Pet, error)       { return nil, nil }
func (s *stubPetService) Create(_ context.Context, _ service.CreatePetInput) (*domain.Pet, error) {
	return s.pet, s.err
}
func (s *stubPetService) GetByID(_ context.Context, _ string) (*domain.Pet, error) {
	return s.pet, nil
}
func (s *stubPetService) Update(_ context.Context, _ string, _ service.UpdatePetInput) (*domain.Pet, error) {
	return s.pet, nil
}
func (s *stubPetService) Delete(_ context.Context, _ string) error { return nil }

func TestPetUseCase_Create_emitsPetCreated(t *testing.T) {
	spy := &spyEmitter{} // defined in care_usecase_test.go
	svc := &stubPetService{
		pet: &domain.Pet{ID: "p1", Name: "Luna", Species: "dog"},
	}
	uc := app.NewPetUseCase(svc, spy)

	if _, err := uc.Create(context.Background(), service.CreatePetInput{Name: "Luna", Species: "dog"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.events) != 1 || spy.events[0] != "pet.created" {
		t.Errorf("events = %v, want [pet.created]", spy.events)
	}
}

func TestPetUseCase_Create_noEmitOnServiceError(t *testing.T) {
	spy := &spyEmitter{}
	svc := &stubPetService{
		err: context.DeadlineExceeded,
	}
	uc := app.NewPetUseCase(svc, spy)

	_, err := uc.Create(context.Background(), service.CreatePetInput{Name: "Luna", Species: "dog"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if len(spy.events) != 0 {
		t.Errorf("expected no events, got %v", spy.events)
	}
}

func TestPetUseCase_Delete_emitsPetDeleted(t *testing.T) {
	spy := &spyEmitter{}
	svc := &stubPetService{
		pet: &domain.Pet{ID: "p1", Name: "Luna", Species: "dog"},
	}
	uc := app.NewPetUseCase(svc, spy)

	if err := uc.Delete(context.Background(), "p1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.events) != 1 || spy.events[0] != "pet.deleted" {
		t.Errorf("events = %v, want [pet.deleted]", spy.events)
	}
}
