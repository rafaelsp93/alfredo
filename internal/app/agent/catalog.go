package agent

import (
	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/policy"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/appointments"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/health"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/messaging"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/observations"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/pets"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/summary"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/supplies"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/treatments"
	"github.com/rafaelsoares/alfredo/internal/app/agent/tools/vaccines"
	"go.uber.org/zap"
)

type Deps struct {
	Pets         agentcontracts.PetToolsDeps
	Vaccines     agentcontracts.VaccineToolsDeps
	Treatments   agentcontracts.TreatmentToolsDeps
	Observations agentcontracts.ObservationToolsDeps
	Appointments agentcontracts.AppointmentToolsDeps
	Supplies     agentcontracts.SupplyToolsDeps
	Summary      agentcontracts.SummaryToolsDeps
	Messaging    agentcontracts.MessagingToolsDeps
	Health       agentcontracts.HealthToolsDeps
}

func BuildSystemPrompt() string {
	return policy.Build()
}

func AllSpecs() []agentdomain.Tool {
	tools := make([]agentdomain.Tool, 0,
		len(pets.Specs())+
			len(vaccines.Specs())+
			len(treatments.Specs())+
			len(appointments.Specs())+
			len(observations.Specs())+
			len(supplies.Specs())+
			len(summary.Specs())+
			len(messaging.Specs())+
			len(health.Specs()),
	)
	tools = append(tools, pets.Specs()...)
	tools = append(tools, vaccines.Specs()...)
	tools = append(tools, treatments.Specs()...)
	tools = append(tools, appointments.Specs()...)
	tools = append(tools, observations.Specs()...)
	tools = append(tools, supplies.Specs()...)
	tools = append(tools, summary.Specs()...)
	tools = append(tools, messaging.Specs()...)
	tools = append(tools, health.Specs()...)
	return tools
}

func BuildRegistry(deps Deps, logger *zap.Logger) registry.ToolRegistry {
	handlers := make([]registry.ToolHandler, 0,
		len(pets.Handlers(deps.Pets))+
			len(vaccines.Handlers(deps.Vaccines))+
			len(treatments.Handlers(deps.Treatments))+
			len(appointments.Handlers(deps.Appointments))+
			len(observations.Handlers(deps.Observations))+
			len(supplies.Handlers(deps.Supplies))+
			len(summary.Handlers(deps.Summary))+
			len(messaging.Handlers(deps.Messaging, logger))+
			len(health.Handlers(deps.Health)),
	)
	handlers = append(handlers, pets.Handlers(deps.Pets)...)
	handlers = append(handlers, vaccines.Handlers(deps.Vaccines)...)
	handlers = append(handlers, treatments.Handlers(deps.Treatments)...)
	handlers = append(handlers, appointments.Handlers(deps.Appointments)...)
	handlers = append(handlers, observations.Handlers(deps.Observations)...)
	handlers = append(handlers, supplies.Handlers(deps.Supplies)...)
	handlers = append(handlers, summary.Handlers(deps.Summary)...)
	handlers = append(handlers, messaging.Handlers(deps.Messaging, logger)...)
	handlers = append(handlers, health.Handlers(deps.Health)...)
	return registry.MustNew(handlers...)
}
