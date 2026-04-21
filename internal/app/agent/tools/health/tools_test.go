package health

import (
	"context"
	"testing"
	"time"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	healthdomain "github.com/rafaelsoares/alfredo/internal/health/domain"
)

func TestHealthHandlers(t *testing.T) {
	handlers := Handlers(agentcontracts.HealthToolsDeps{
		Profile:  fakeHealthProfile{},
		Metrics:  fakeHealthMetrics{},
		Workouts: fakeHealthWorkouts{},
		Insight:  fakeHealthInsight{},
	})
	if len(Specs()) != 4 || handlers[0].Spec().Name != "get_health_profile" || handlers[1].Spec().Name != "get_health_metrics" || handlers[2].Spec().Name != "list_workouts" || handlers[3].Spec().Name != "get_health_summary" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), nil); err != nil {
		t.Fatalf("profile err = %v", err)
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{"metric_type": "weight", "from": "2026-04-20"}); err != nil {
		t.Fatalf("metrics err = %v", err)
	}
	if _, err := handlers[2].Handle(context.Background(), map[string]any{"from": "2026-04-20"}); err != nil {
		t.Fatalf("workouts err = %v", err)
	}
	if _, err := handlers[3].Handle(context.Background(), map[string]any{"days": 7}); err != nil {
		t.Fatalf("summary err = %v", err)
	}
	if _, err := handlers[0].Handle(context.Background(), nil); err != nil {
		t.Fatalf("profile second err = %v", err)
	}
	handlers = Handlers(agentcontracts.HealthToolsDeps{Profile: fakeHealthProfile{err: healthdomain.ErrNotFound}})
	if _, err := handlers[0].Handle(context.Background(), nil); err == nil {
		t.Fatal("expected not found error")
	}
	if _, err := handlers[1].Handle(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected metric_type error")
	}
	if _, err := handlers[2].Handle(context.Background(), map[string]any{"from": "bad"}); err == nil {
		t.Fatal("expected workout date error")
	}
	if _, err := Handlers(agentcontracts.HealthToolsDeps{})[3].Handle(context.Background(), nil); err == nil {
		t.Fatal("expected missing insight error")
	}
}

type fakeHealthProfile struct{ err error }

func (f fakeHealthProfile) Get(context.Context) (healthdomain.HealthProfile, error) {
	return healthdomain.HealthProfile{}, f.err
}

type fakeHealthMetrics struct{}

func (fakeHealthMetrics) List(context.Context, string, time.Time, time.Time) ([]healthdomain.DailyMetric, error) {
	return []healthdomain.DailyMetric{}, nil
}

type fakeHealthWorkouts struct{}

func (fakeHealthWorkouts) List(context.Context, time.Time, time.Time) ([]healthdomain.WorkoutSession, error) {
	return []healthdomain.WorkoutSession{}, nil
}

type fakeHealthInsight struct{}

func (fakeHealthInsight) Compute(context.Context, int) (healthdomain.HealthInsight, error) {
	return healthdomain.HealthInsight{}, nil
}
