package agent

import (
	"context"
	"strings"
	"testing"
	"time"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	healthdomain "github.com/rafaelsoares/alfredo/internal/health/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"go.uber.org/zap"
)

func TestCatalogBuildsPromptSpecsAndRegistry(t *testing.T) {
	if !strings.Contains(BuildSystemPrompt(), "PETS:") {
		t.Fatal("missing pets section")
	}
	tools := AllSpecs()
	if len(tools) == 0 {
		t.Fatal("expected tools")
	}
	reg := BuildRegistry(Deps{
		Pets:     agentcontracts.PetToolsDeps{Pets: catalogPetService{}},
		Vaccines: agentcontracts.VaccineToolsDeps{Vaccines: catalogVaccineService{}, Location: time.UTC},
		Treatments: agentcontracts.TreatmentToolsDeps{
			Treatments: catalogTreatmentService{},
			Location:   time.UTC,
		},
		Observations: agentcontracts.ObservationToolsDeps{Observations: catalogObservationService{}, Location: time.UTC},
		Appointments: agentcontracts.AppointmentToolsDeps{Appointments: catalogAppointmentService{}, Location: time.UTC},
		Supplies:     agentcontracts.SupplyToolsDeps{Supplies: catalogSupplyService{}},
		Summary:      agentcontracts.SummaryToolsDeps{Summary: catalogSummaryService{}},
		Messaging:    agentcontracts.MessagingToolsDeps{},
		Health:       agentcontracts.HealthToolsDeps{Profile: catalogHealthProfile{}, Metrics: catalogHealthMetrics{}, Workouts: catalogHealthWorkouts{}, Insight: catalogHealthInsight{}},
	}, zap.NewNop())
	result, err := reg.Execute(context.Background(), agentdomain.ToolCall{ID: "1", Name: "list_pets"})
	if err != nil || result.IsError {
		t.Fatalf("Execute = %#v %v", result, err)
	}
}

type catalogPetService struct{}

func (catalogPetService) List(context.Context) ([]domain.Pet, error) {
	return []domain.Pet{{ID: "pet-1"}}, nil
}
func (catalogPetService) GetByID(context.Context, string) (*domain.Pet, error) {
	return &domain.Pet{ID: "pet-1"}, nil
}

type catalogVaccineService struct{}

func (catalogVaccineService) ListVaccines(context.Context, string) ([]domain.Vaccine, error) {
	return nil, nil
}
func (catalogVaccineService) RecordVaccine(context.Context, service.RecordVaccineInput) (*domain.Vaccine, error) {
	return &domain.Vaccine{}, nil
}

type catalogTreatmentService struct{}

func (catalogTreatmentService) Create(context.Context, service.CreateTreatmentInput) (*domain.Treatment, []domain.Dose, error) {
	return &domain.Treatment{}, nil, nil
}
func (catalogTreatmentService) List(context.Context, string) ([]domain.Treatment, map[string][]domain.Dose, error) {
	return nil, nil, nil
}

type catalogObservationService struct{}

func (catalogObservationService) Create(context.Context, service.CreateObservationInput) (*domain.Observation, error) {
	return &domain.Observation{}, nil
}
func (catalogObservationService) ListByPet(context.Context, string) ([]domain.Observation, error) {
	return nil, nil
}

type catalogAppointmentService struct{}

func (catalogAppointmentService) Create(context.Context, service.CreateAppointmentInput) (*domain.Appointment, error) {
	return &domain.Appointment{}, nil
}
func (catalogAppointmentService) List(context.Context, string) ([]domain.Appointment, error) {
	return nil, nil
}
func (catalogAppointmentService) Update(context.Context, string, string, service.UpdateAppointmentInput) (*domain.Appointment, error) {
	return &domain.Appointment{}, nil
}

type catalogSupplyService struct{}

func (catalogSupplyService) Create(context.Context, service.CreateSupplyInput) (*domain.Supply, error) {
	return &domain.Supply{}, nil
}
func (catalogSupplyService) GetByID(context.Context, string, string) (*domain.Supply, error) {
	return &domain.Supply{}, nil
}
func (catalogSupplyService) List(context.Context, string) ([]domain.Supply, error) { return nil, nil }
func (catalogSupplyService) Update(context.Context, string, string, service.UpdateSupplyInput) (*domain.Supply, error) {
	return &domain.Supply{}, nil
}

type catalogSummaryService struct{}

func (catalogSummaryService) AllPets(context.Context) (domain.AllPetsSummary, error) {
	return domain.AllPetsSummary{}, nil
}

type catalogHealthProfile struct{}

func (catalogHealthProfile) Get(context.Context) (healthdomain.HealthProfile, error) {
	return healthdomain.HealthProfile{}, nil
}

type catalogHealthMetrics struct{}

func (catalogHealthMetrics) List(context.Context, string, time.Time, time.Time) ([]healthdomain.DailyMetric, error) {
	return nil, nil
}

type catalogHealthWorkouts struct{}

func (catalogHealthWorkouts) List(context.Context, time.Time, time.Time) ([]healthdomain.WorkoutSession, error) {
	return nil, nil
}

type catalogHealthInsight struct{}

func (catalogHealthInsight) Compute(context.Context, int) (healthdomain.HealthInsight, error) {
	return healthdomain.HealthInsight{}, nil
}
