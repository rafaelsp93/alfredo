package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rafaelsoares/alfredo/internal/petcare/service"
)

type mockDBChecker struct{ err error }

func (m *mockDBChecker) Ping(_ context.Context) error { return m.err }

func TestSystemService_Healthy(t *testing.T) {
	svc := service.NewSystemService(&mockDBChecker{})
	result := svc.Check(context.Background())
	if result.Status != "healthy" {
		t.Errorf("got %q, want healthy", result.Status)
	}
}

func TestSystemService_Degraded(t *testing.T) {
	svc := service.NewSystemService(&mockDBChecker{err: errors.New("db down")})
	result := svc.Check(context.Background())
	if result.Status != "degraded" {
		t.Errorf("got %q, want degraded", result.Status)
	}
	if result.Dependencies["sqlite"].Error == "" {
		t.Error("expected sqlite error message")
	}
}
