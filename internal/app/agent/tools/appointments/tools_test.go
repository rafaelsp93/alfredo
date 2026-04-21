package appointments

import (
	"context"
	"testing"
	"time"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func TestAppointmentHandlers(t *testing.T) {
	handlers := Handlers(agentcontracts.AppointmentToolsDeps{Appointments: fakeAppointmentService{}, Location: time.UTC})
	if len(Specs()) != 3 || handlers[0].Spec().Name != "list_appointments" || handlers[1].Spec().Name != "schedule_appointment" || handlers[2].Spec().Name != "reschedule_appointment" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err != nil {
		t.Fatalf("list err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "type": "vet", "scheduled_at": "2026-04-21T10:00:00"}); err != nil {
		t.Fatalf("schedule err = %v", err)
	}
	if _, err := handlers[2].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "appointment_id": "appt-1", "scheduled_at": "2026-04-21T10:00:00"}); err != nil {
		t.Fatalf("reschedule err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"pet_id": "pet-1", "type": "bath", "scheduled_at": "2026-04-21T10:00:00"}); err == nil {
		t.Fatal("expected type error")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected list decode error")
	}
	if _, err := handlers[2].Handle(context.Background(), map[string]any{"pet_id": "pet-1"}); err == nil {
		t.Fatal("expected reschedule decode error")
	}
}

type fakeAppointmentService struct{}

func (fakeAppointmentService) Create(context.Context, service.CreateAppointmentInput) (*domain.Appointment, error) {
	return &domain.Appointment{}, nil
}
func (fakeAppointmentService) List(context.Context, string) ([]domain.Appointment, error) {
	return nil, nil
}
func (fakeAppointmentService) Update(context.Context, string, string, service.UpdateAppointmentInput) (*domain.Appointment, error) {
	return &domain.Appointment{}, nil
}
