package vaccines

import (
	"context"
	"time"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/app/agent/args"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("list_vaccines", "List vaccine records for one pet.", registry.ObjectSchema(registry.Properties("pet_id", "string"), []string{"pet_id"})),
		registry.Tool("record_vaccine", "Record a vaccine administration for one pet.", registry.ObjectSchema(registry.Properties("pet_id", "string", "name", "string", "date", "string", "recurrence_days", "integer", "vet_name", "string", "batch_number", "string", "notes", "string"), []string{"pet_id", "name", "date"})),
	}
}

func Handlers(deps appagent.VaccineToolsDeps) []registry.ToolHandler {
	return []registry.ToolHandler{
		listVaccinesHandler{vaccines: deps.Vaccines},
		recordVaccineHandler{vaccines: deps.Vaccines, location: deps.Location},
	}
}

type listVaccinesHandler struct{ vaccines appagent.VaccineUseCaser }

func (h listVaccinesHandler) Spec() agentdomain.Tool { return Specs()[0] }

func (h listVaccinesHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return nil, err
	}
	return h.vaccines.ListVaccines(ctx, petID)
}

type recordVaccineHandler struct {
	vaccines appagent.VaccineUseCaser
	location *time.Location
}

func (h recordVaccineHandler) Spec() agentdomain.Tool { return Specs()[1] }

func (h recordVaccineHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	in, err := decodeRecordVaccine(values, h.location)
	if err != nil {
		return nil, err
	}
	return h.vaccines.RecordVaccine(ctx, in)
}

func decodeRecordVaccine(values map[string]any, location *time.Location) (service.RecordVaccineInput, error) {
	petID, err := args.RequireString(values, "pet_id")
	if err != nil {
		return service.RecordVaccineInput{}, err
	}
	name, err := args.RequireString(values, "name")
	if err != nil {
		return service.RecordVaccineInput{}, err
	}
	administeredAt, err := args.RequireUserTime(values, "date", location)
	if err != nil {
		return service.RecordVaccineInput{}, err
	}
	var recurrence *int
	if value, ok := values["recurrence_days"]; ok {
		out, err := args.NumberToInt(value, "recurrence_days")
		if err != nil {
			return service.RecordVaccineInput{}, err
		}
		recurrence = &out
	}
	return service.RecordVaccineInput{
		PetID:          petID,
		Name:           name,
		AdministeredAt: administeredAt,
		RecurrenceDays: recurrence,
		VetName:        args.OptionalString(values, "vet_name"),
		BatchNumber:    args.OptionalString(values, "batch_number"),
		Notes:          args.OptionalString(values, "notes"),
	}, nil
}
