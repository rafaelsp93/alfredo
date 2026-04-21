package health

import (
	"context"
	"errors"
	"fmt"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/app/agent/args"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	healthdomain "github.com/rafaelsoares/alfredo/internal/health/domain"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("get_health_profile", "Get Rafael's personal health profile (height, birth date, sex). Use ONLY for data about Rafael himself — not for any pet.", registry.ObjectSchema(nil, nil)),
		registry.Tool("get_health_metrics", "Query Rafael's personal daily health metrics by metric type (e.g. weight, bodyFat, restingHeartRate, stepCount, sleepTime, walkingDistance, vo2Max). Optional from and to dates in YYYY-MM-DD format narrow the result. Use ONLY for data about Rafael himself — not for any pet.", registry.ObjectSchema(registry.Properties("metric_type", "string", "from", "string", "to", "string"), []string{"metric_type"})),
		registry.Tool("list_workouts", "List Rafael's workout sessions from Apple Watch (activity type, duration, calories burned, heart rate). Optional from and to dates in YYYY-MM-DD format narrow the result. Use ONLY for Rafael's own workouts — not for pet activities.", registry.ObjectSchema(registry.Properties("from", "string", "to", "string"), nil)),
		registry.Tool("get_health_summary", "Computa um resumo derivado de saúde (tendência de peso, frequência cardíaca, sono, treinos, IMC, VO2Max) para uma janela de dias. Retorna dados estruturados. Use para perguntas como 'como estou me saindo na saúde?', 'estou engordando?', 'meu sono está ruim?', 'qual meu IMC?', 'minha frequência cardíaca está alta?'. O parâmetro days é opcional (padrão: 14).", registry.ObjectSchema(map[string]any{"days": map[string]any{"type": "integer"}}, nil)),
	}
}

func Handlers(deps appagent.HealthToolsDeps) []registry.ToolHandler {
	return []registry.ToolHandler{
		getHealthProfileHandler{profile: deps.Profile},
		getHealthMetricsHandler{metrics: deps.Metrics},
		listWorkoutsHandler{workouts: deps.Workouts},
		getHealthSummaryHandler{insight: deps.Insight},
	}
}

type getHealthProfileHandler struct{ profile appagent.HealthProfileQuerier }

func (h getHealthProfileHandler) Spec() agentdomain.Tool { return Specs()[0] }
func (h getHealthProfileHandler) Handle(ctx context.Context, _ map[string]any) (any, error) {
	profile, err := h.profile.Get(ctx)
	if err != nil {
		if errors.Is(err, healthdomain.ErrNotFound) {
			return nil, fmt.Errorf("nenhum perfil de saúde cadastrado")
		}
		return nil, err
	}
	return profile, nil
}

type getHealthMetricsHandler struct{ metrics appagent.HealthMetricsQuerier }

func (h getHealthMetricsHandler) Spec() agentdomain.Tool { return Specs()[1] }
func (h getHealthMetricsHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	metricType, err := args.RequireString(values, "metric_type")
	if err != nil {
		return nil, err
	}
	from, err := args.OptionalDate(values, "from")
	if err != nil {
		return nil, err
	}
	to, err := args.OptionalDate(values, "to")
	if err != nil {
		return nil, err
	}
	return h.metrics.List(ctx, metricType, from, to)
}

type listWorkoutsHandler struct {
	workouts appagent.HealthWorkoutsQuerier
}

func (h listWorkoutsHandler) Spec() agentdomain.Tool { return Specs()[2] }
func (h listWorkoutsHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	from, err := args.OptionalDate(values, "from")
	if err != nil {
		return nil, err
	}
	to, err := args.OptionalDate(values, "to")
	if err != nil {
		return nil, err
	}
	return h.workouts.List(ctx, from, to)
}

type getHealthSummaryHandler struct {
	insight appagent.HealthInsightComputer
}

func (h getHealthSummaryHandler) Spec() agentdomain.Tool { return Specs()[3] }
func (h getHealthSummaryHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	if h.insight == nil {
		return nil, fmt.Errorf("health insight service is not configured")
	}
	days := 14
	if value, ok := values["days"]; ok && value != nil {
		out, err := args.NumberToInt(value, "days")
		if err == nil && out > 0 {
			days = out
		}
	}
	return h.insight.Compute(ctx, days)
}
