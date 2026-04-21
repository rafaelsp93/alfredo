package appointments

import (
	"context"
	"fmt"
	"time"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/app/agent/args"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("list_appointments", "List a pet's appointment history, including vet visits and grooming sessions such as banho, banho e tosa, tosa, or grooming. Use it for questions like quando foi o banho or quando foi a última consulta.", registry.ObjectSchema(registry.Properties("pet_id", "string"), []string{"pet_id"})),
		registry.Tool("schedule_appointment", "Schedule a pet appointment. Use for vet visits and grooming sessions such as banho, banho e tosa, tosa, or grooming.", registry.ObjectSchema(map[string]any{
			"pet_id": map[string]any{"type": "string"},
			"type": map[string]any{
				"type":        "string",
				"description": "Use vet for consulta veterinária, grooming for banho, banho e tosa, tosa, or grooming, and other for any other appointment.",
			},
			"scheduled_at": map[string]any{"type": "string"},
			"provider":     map[string]any{"type": "string"},
			"location":     map[string]any{"type": "string"},
			"notes":        map[string]any{"type": "string"},
		}, []string{"pet_id", "type", "scheduled_at"})),
		registry.Tool("reschedule_appointment", "Move an existing appointment to a new time.", registry.ObjectSchema(registry.Properties("pet_id", "string", "appointment_id", "string", "scheduled_at", "string"), []string{"pet_id", "appointment_id", "scheduled_at"})),
	}
}

func Handlers(deps appagent.AppointmentToolsDeps) []registry.ToolHandler {
	return []registry.ToolHandler{
		listAppointmentsHandler{appointments: deps.Appointments},
		scheduleAppointmentHandler{appointments: deps.Appointments, location: deps.Location},
		rescheduleAppointmentHandler{appointments: deps.Appointments, location: deps.Location},
	}
}

type listAppointmentsHandler struct{ appointments appagent.AppointmentServicer }

func (h listAppointmentsHandler) Spec() agentdomain.Tool { return Specs()[0] }

func (h listAppointmentsHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return nil, err
	}
	return h.appointments.List(ctx, petID)
}

type scheduleAppointmentHandler struct {
	appointments appagent.AppointmentServicer
	location     *time.Location
}

func (h scheduleAppointmentHandler) Spec() agentdomain.Tool { return Specs()[1] }

func (h scheduleAppointmentHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	in, err := decodeScheduleAppointment(values, h.location)
	if err != nil {
		return nil, err
	}
	return h.appointments.Create(ctx, in)
}

type rescheduleAppointmentHandler struct {
	appointments appagent.AppointmentServicer
	location     *time.Location
}

func (h rescheduleAppointmentHandler) Spec() agentdomain.Tool { return Specs()[2] }

func (h rescheduleAppointmentHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, appointmentID, err := args.RequireTwoStrings(values, "pet_id", "appointment_id")
	if err != nil {
		return nil, err
	}
	scheduledAt, err := args.RequireUserTime(values, "scheduled_at", h.location)
	if err != nil {
		return nil, err
	}
	return h.appointments.Update(ctx, petID, appointmentID, service.UpdateAppointmentInput{ScheduledAt: &scheduledAt})
}

func decodeScheduleAppointment(values map[string]any, location *time.Location) (service.CreateAppointmentInput, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return service.CreateAppointmentInput{}, err
	}
	typeText, err := args.RequireString(values, "type")
	if err != nil {
		return service.CreateAppointmentInput{}, err
	}
	appointmentType := domain.AppointmentType(typeText)
	switch appointmentType {
	case domain.AppointmentTypeVet, domain.AppointmentTypeGrooming, domain.AppointmentTypeOther:
	default:
		return service.CreateAppointmentInput{}, fmt.Errorf("type must be one of: vet, grooming, other")
	}
	scheduledAt, err := args.RequireUserTime(values, "scheduled_at", location)
	if err != nil {
		return service.CreateAppointmentInput{}, err
	}
	return service.CreateAppointmentInput{
		PetID:       petID,
		Type:        appointmentType,
		ScheduledAt: scheduledAt,
		Provider:    args.OptionalString(values, "provider"),
		Location:    args.OptionalString(values, "location"),
		Notes:       args.OptionalString(values, "notes"),
	}, nil
}
