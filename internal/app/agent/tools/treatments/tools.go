package treatments

import (
	"context"
	"fmt"
	"strings"
	"time"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/app/agent/args"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("list_treatments", "List treatments and dose events for one pet.", registry.ObjectSchema(registry.Properties("pet_id", "string"), []string{"pet_id"})),
		registry.Tool("start_treatment", "Start a pet treatment and create its dose schedule.", registry.ObjectSchema(registry.Properties("pet_id", "string", "name", "string", "dosage_amount", "number", "dosage_unit", "string", "route", "string", "interval_hours", "integer", "started_at", "string", "ended_at", "string", "vet_name", "string", "notes", "string"), []string{"pet_id", "name", "dosage_amount", "dosage_unit", "route", "interval_hours", "started_at"})),
	}
}

func Handlers(deps appagent.TreatmentToolsDeps) []registry.ToolHandler {
	return []registry.ToolHandler{
		listTreatmentsHandler{treatments: deps.Treatments},
		startTreatmentHandler{treatments: deps.Treatments, location: deps.Location},
	}
}

type listTreatmentsHandler struct{ treatments appagent.TreatmentUseCaser }

func (h listTreatmentsHandler) Spec() agentdomain.Tool { return Specs()[0] }

func (h listTreatmentsHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return nil, err
	}
	treatments, doses, err := h.treatments.List(ctx, petID)
	if err != nil {
		return nil, err
	}
	return map[string]any{"treatments": treatments, "doses": doses}, nil
}

type startTreatmentHandler struct {
	treatments appagent.TreatmentUseCaser
	location   *time.Location
}

func (h startTreatmentHandler) Spec() agentdomain.Tool { return Specs()[1] }

func (h startTreatmentHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	in, err := decodeCreateTreatment(values, h.location)
	if err != nil {
		return nil, err
	}
	treatment, doses, err := h.treatments.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return map[string]any{"treatment": treatment, "doses": doses}, nil
}

func decodeCreateTreatment(values map[string]any, location *time.Location) (service.CreateTreatmentInput, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return service.CreateTreatmentInput{}, err
	}
	name, err := args.RequireString(values, "name")
	if err != nil {
		return service.CreateTreatmentInput{}, err
	}
	amount, err := args.RequireFloat(values, "dosage_amount")
	if err != nil {
		return service.CreateTreatmentInput{}, err
	}
	unit, err := args.RequireString(values, "dosage_unit")
	if err != nil {
		return service.CreateTreatmentInput{}, err
	}
	route, err := args.RequireString(values, "route")
	if err != nil {
		return service.CreateTreatmentInput{}, err
	}
	interval, err := args.RequireInt(values, "interval_hours")
	if err != nil {
		return service.CreateTreatmentInput{}, err
	}
	startedAt, err := args.RequireUserTime(values, "started_at", location)
	if err != nil {
		return service.CreateTreatmentInput{}, err
	}
	var endedAt *time.Time
	if value, ok := values["ended_at"]; ok && value != nil && strings.TrimSpace(fmt.Sprint(value)) != "" {
		out, err := args.ParseUserTime(fmt.Sprint(value), "ended_at", location)
		if err != nil {
			return service.CreateTreatmentInput{}, err
		}
		endedAt = &out
	}
	return service.CreateTreatmentInput{
		PetID:         petID,
		Name:          name,
		DosageAmount:  amount,
		DosageUnit:    unit,
		Route:         route,
		IntervalHours: interval,
		StartedAt:     startedAt,
		EndedAt:       endedAt,
		VetName:       args.OptionalString(values, "vet_name"),
		Notes:         args.OptionalString(values, "notes"),
	}, nil
}
