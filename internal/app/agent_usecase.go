package app

import (
	"context"
	"time"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	agentservice "github.com/rafaelsoares/alfredo/internal/agent/service"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent"
	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	agentregistry "github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	"go.uber.org/zap"
)

var agentSystemPrompt = appagent.BuildSystemPrompt()

type AgentRouter interface {
	Execute(
		ctx context.Context,
		systemPrompt string,
		tools []agentdomain.Tool,
		inputText string,
		dispatch func(ctx context.Context, call agentdomain.ToolCall) (agentdomain.ToolResult, error),
	) (reply string, inv agentdomain.Invocation, err error)
}

type AgentUseCase struct {
	router       AgentRouter
	systemPrompt string
	registry     agentregistry.ToolRegistry
	logger       *zap.Logger
}

func NewAgentUseCase(
	router AgentRouter,
	pets PetCareServicer,
	vaccines AgentVaccineUseCaser,
	treatments AgentTreatmentUseCaser,
	observations ObservationServicer,
	appointments AppointmentServicer,
	supplies SupplyServicer,
	summary SummaryUseCaser,
	telegram TelegramPort,
	healthProfile HealthProfileQuerier,
	healthMetrics HealthMetricsQuerier,
	healthWorkouts HealthWorkoutsQuerier,
	healthInsight HealthInsightComputer,
	timezone *time.Location,
	logger *zap.Logger,
) *AgentUseCase {
	if logger == nil {
		logger = zap.NewNop()
	}
	reg := appagent.BuildRegistry(appagent.Deps{
		Pets:         agentcontracts.PetToolsDeps{Pets: pets},
		Vaccines:     agentcontracts.VaccineToolsDeps{Vaccines: vaccines, Location: timezone},
		Treatments:   agentcontracts.TreatmentToolsDeps{Treatments: treatments, Location: timezone},
		Observations: agentcontracts.ObservationToolsDeps{Observations: observations, Location: timezone},
		Appointments: agentcontracts.AppointmentToolsDeps{Appointments: appointments, Location: timezone},
		Supplies:     agentcontracts.SupplyToolsDeps{Supplies: supplies},
		Summary:      agentcontracts.SummaryToolsDeps{Summary: summary},
		Messaging:    agentcontracts.MessagingToolsDeps{Telegram: telegram},
		Health: agentcontracts.HealthToolsDeps{
			Profile:  healthProfile,
			Metrics:  healthMetrics,
			Workouts: healthWorkouts,
			Insight:  healthInsight,
		},
	}, logger)
	return &AgentUseCase{
		router:       router,
		systemPrompt: agentSystemPrompt,
		registry:     reg,
		logger:       logger,
	}
}

func (uc *AgentUseCase) Handle(ctx context.Context, inputText string) (string, error) {
	reply, _, err := uc.router.Execute(ctx, uc.systemPrompt, uc.registry.Tools(), inputText, uc.DispatchToolCall)
	if err != nil {
		uc.logger.Warn("agent handled request with fallback reply", zap.Error(err))
		return reply, nil
	}
	return reply, nil
}

func (uc *AgentUseCase) DispatchToolCall(ctx context.Context, call agentdomain.ToolCall) (agentdomain.ToolResult, error) {
	return uc.registry.Execute(ctx, call)
}

func buildAgentTools() []agentdomain.Tool {
	return appagent.AllSpecs()
}

var _ AgentRouter = (*agentservice.Router)(nil)
