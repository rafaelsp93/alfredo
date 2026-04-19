package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/health/domain"
)

type workoutUseCaseStub struct {
	importFn func(context.Context, []domain.WorkoutSession, string, time.Time) (int, error)
	listFn   func(context.Context, time.Time, time.Time) ([]domain.WorkoutSession, error)
}

func (s *workoutUseCaseStub) Import(ctx context.Context, sessions []domain.WorkoutSession, payload string, importedAt time.Time) (int, error) {
	if s.importFn != nil {
		return s.importFn(ctx, sessions, payload, importedAt)
	}
	return len(sessions), nil
}

func (s *workoutUseCaseStub) List(ctx context.Context, from, to time.Time) ([]domain.WorkoutSession, error) {
	if s.listFn != nil {
		return s.listFn(ctx, from, to)
	}
	return nil, nil
}

func doWorkoutRequest(t *testing.T, method, path, body string, uc WorkoutUseCaser) *httptest.ResponseRecorder {
	t.Helper()
	e := echo.New()
	var req *http.Request
	if body == "" {
		req = httptest.NewRequest(method, path, nil)
	} else {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath(path)

	// Parse query params for GET requests
	if method == http.MethodGet {
		u := req.URL
		q := u.Query()
		for k, v := range q {
			if len(v) > 0 {
				_ = k
			}
		}
	}

	h := NewWorkoutHandler(uc)
	switch method {
	case http.MethodPost:
		if err := h.ImportWorkouts(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case http.MethodGet:
		if err := h.ListWorkouts(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	default:
		t.Fatalf("unsupported method: %s", method)
	}
	return rec
}

func TestWorkoutHandlerRejectsMalformedStartDate(t *testing.T) {
	rec := doWorkoutRequest(t, http.MethodPost, "/api/v1/health/workouts/import", `{
		"workouts": [
			{
				"activityName": "Running",
				"startDate": "not-a-date",
				"endDate": "2026-04-18T10:30:00Z",
				"duration": 1800
			}
		]
	}`, &workoutUseCaseStub{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for malformed startDate", rec.Code)
	}
}

func TestWorkoutHandlerRejectsMalformedEndDate(t *testing.T) {
	rec := doWorkoutRequest(t, http.MethodPost, "/api/v1/health/workouts/import", `{
		"workouts": [
			{
				"activityName": "Running",
				"startDate": "2026-04-18T10:00:00Z",
				"endDate": "bad-end",
				"duration": 1800
			}
		]
	}`, &workoutUseCaseStub{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for malformed endDate", rec.Code)
	}
}

func TestWorkoutHandlerListToDateIsInclusive(t *testing.T) {
	var capturedTo time.Time

	stub := &workoutUseCaseStub{
		listFn: func(_ context.Context, _, to time.Time) ([]domain.WorkoutSession, error) {
			capturedTo = to
			return nil, nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health/workouts?from=2026-04-01&to=2026-04-30", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/v1/health/workouts")
	c.QueryParams().Set("from", "2026-04-01")
	c.QueryParams().Set("to", "2026-04-30")

	h := NewWorkoutHandler(stub)
	if err := h.ListWorkouts(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	// to should be end of 2026-04-30, not midnight (start of day)
	if capturedTo.Day() != 30 || capturedTo.Month() != 4 || capturedTo.Year() != 2026 {
		t.Fatalf("to date = %v, want 2026-04-30", capturedTo)
	}
	if capturedTo.Hour() == 0 && capturedTo.Minute() == 0 && capturedTo.Second() == 0 {
		t.Fatalf("to = %v is midnight (start of day), want end of day so workouts on April 30 are included", capturedTo)
	}
}
