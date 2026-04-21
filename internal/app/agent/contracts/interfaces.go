package contracts

import (
	"context"
	"time"

	healthdomain "github.com/rafaelsoares/alfredo/internal/health/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
	"github.com/rafaelsoares/alfredo/internal/telegram"
)

type PetCareServicer interface {
	List(ctx context.Context) ([]domain.Pet, error)
	GetByID(ctx context.Context, id string) (*domain.Pet, error)
}

type VaccineUseCaser interface {
	ListVaccines(ctx context.Context, petID string) ([]domain.Vaccine, error)
	RecordVaccine(ctx context.Context, in service.RecordVaccineInput) (*domain.Vaccine, error)
}

type TreatmentUseCaser interface {
	Create(ctx context.Context, in service.CreateTreatmentInput) (*domain.Treatment, []domain.Dose, error)
	List(ctx context.Context, petID string) ([]domain.Treatment, map[string][]domain.Dose, error)
}

type ObservationServicer interface {
	Create(ctx context.Context, in service.CreateObservationInput) (*domain.Observation, error)
	ListByPet(ctx context.Context, petID string) ([]domain.Observation, error)
}

type AppointmentServicer interface {
	Create(ctx context.Context, in service.CreateAppointmentInput) (*domain.Appointment, error)
	List(ctx context.Context, petID string) ([]domain.Appointment, error)
	Update(ctx context.Context, petID, appointmentID string, in service.UpdateAppointmentInput) (*domain.Appointment, error)
}

type SupplyServicer interface {
	Create(ctx context.Context, in service.CreateSupplyInput) (*domain.Supply, error)
	GetByID(ctx context.Context, petID, supplyID string) (*domain.Supply, error)
	List(ctx context.Context, petID string) ([]domain.Supply, error)
	Update(ctx context.Context, petID, supplyID string, in service.UpdateSupplyInput) (*domain.Supply, error)
}

type SummaryUseCaser interface {
	AllPets(ctx context.Context) (domain.AllPetsSummary, error)
}

type HealthProfileQuerier interface {
	Get(ctx context.Context) (healthdomain.HealthProfile, error)
}

type HealthMetricsQuerier interface {
	List(ctx context.Context, metricType string, from, to time.Time) ([]healthdomain.DailyMetric, error)
}

type HealthWorkoutsQuerier interface {
	List(ctx context.Context, from, to time.Time) ([]healthdomain.WorkoutSession, error)
}

type HealthInsightComputer interface {
	Compute(ctx context.Context, days int) (healthdomain.HealthInsight, error)
}

type TelegramPort = telegram.Port

type PetToolsDeps struct {
	Pets PetCareServicer
}

type VaccineToolsDeps struct {
	Vaccines VaccineUseCaser
	Location *time.Location
}

type TreatmentToolsDeps struct {
	Treatments TreatmentUseCaser
	Location   *time.Location
}

type ObservationToolsDeps struct {
	Observations ObservationServicer
	Location     *time.Location
}

type AppointmentToolsDeps struct {
	Appointments AppointmentServicer
	Location     *time.Location
}

type SupplyToolsDeps struct {
	Supplies SupplyServicer
}

type SummaryToolsDeps struct {
	Summary SummaryUseCaser
}

type MessagingToolsDeps struct {
	Telegram TelegramPort
}

type HealthToolsDeps struct {
	Profile  HealthProfileQuerier
	Metrics  HealthMetricsQuerier
	Workouts HealthWorkoutsQuerier
	Insight  HealthInsightComputer
}
