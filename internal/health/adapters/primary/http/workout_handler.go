package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rafaelsoares/alfredo/internal/health/domain"
)

type WorkoutUseCaser interface {
	Import(ctx context.Context, sessions []domain.WorkoutSession, payload string, importedAt time.Time) (int, error)
	List(ctx context.Context, from, to time.Time) ([]domain.WorkoutSession, error)
}

type WorkoutHandler struct {
	uc WorkoutUseCaser
}

func NewWorkoutHandler(uc WorkoutUseCaser) *WorkoutHandler {
	return &WorkoutHandler{uc: uc}
}

func (h *WorkoutHandler) Register(g *echo.Group) {
	g.POST("/health/workouts/import", h.ImportWorkouts)
	g.GET("/health/workouts", h.ListWorkouts)
}

type workoutImportResponse struct {
	Imported int `json:"imported"`
}

type workoutSessionResponse struct {
	ID                 int      `json:"id"`
	ActivityType       string   `json:"activity_type"`
	StartDate          string   `json:"start_date"`
	EndDate            string   `json:"end_date"`
	DurationSeconds    float64  `json:"duration_seconds"`
	ActiveCaloriesKcal *float64 `json:"active_calories_kcal,omitempty"`
	BasalCaloriesKcal  *float64 `json:"basal_calories_kcal,omitempty"`
	HRAvgBPM           *float64 `json:"hr_avg_bpm,omitempty"`
	HRMinBPM           *float64 `json:"hr_min_bpm,omitempty"`
	HRMaxBPM           *float64 `json:"hr_max_bpm,omitempty"`
	DistanceM          *float64 `json:"distance_m,omitempty"`
	Source             string   `json:"source"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

// Apple Health Exporter workouts format. The exporter has emitted both
// activityName/value/minimum/maximum and activityType/sum/min/max shapes.
type healthExporterWorkout struct {
	ActivityName string                               `json:"activityName"`
	ActivityType string                               `json:"activityType"`
	StartDate    string                               `json:"startDate"`
	EndDate      string                               `json:"endDate"`
	Duration     float64                              `json:"duration"`
	Source       string                               `json:"source"`
	Statistics   map[string]healthExporterWorkoutStat `json:"statistics"`
}

type healthExporterWorkoutStat struct {
	Value   *float64 `json:"value"`
	Sum     *float64 `json:"sum"`
	Average *float64 `json:"average"`
	Minimum *float64 `json:"minimum"`
	Min     *float64 `json:"min"`
	Maximum *float64 `json:"maximum"`
	Max     *float64 `json:"max"`
}

func (h *WorkoutHandler) ImportWorkouts(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Parse workouts export JSON
	var payload struct {
		Workouts []healthExporterWorkout `json:"workouts"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid JSON format"})
	}

	var sessions []domain.WorkoutSession
	importedAt := time.Now().UTC()

	for _, w := range payload.Workouts {
		startDate, err := time.Parse(time.RFC3339Nano, w.StartDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "workout startDate must be RFC3339 (e.g. 2026-04-18T10:00:00Z)"})
		}
		endDate, err := time.Parse(time.RFC3339Nano, w.EndDate)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "workout endDate must be RFC3339 (e.g. 2026-04-18T10:30:00Z)"})
		}

		session := domain.WorkoutSession{
			ActivityType:    firstNonEmpty(w.ActivityType, w.ActivityName),
			StartDate:       startDate,
			EndDate:         endDate,
			DurationSeconds: w.Duration,
			Source:          firstNonEmpty(w.Source, "Apple Watch"),
		}

		// Map HKQuantityTypeIdentifier keys to domain fields
		if stats := w.Statistics; stats != nil {
			if stat, ok := stats["HKQuantityTypeIdentifierActiveEnergyBurned"]; ok {
				session.ActiveCaloriesKcal = firstFloatPtr(stat.Sum, stat.Value)
			}
			if stat, ok := stats["HKQuantityTypeIdentifierBasalEnergyBurned"]; ok {
				session.BasalCaloriesKcal = firstFloatPtr(stat.Sum, stat.Value)
			}
			if stat, ok := stats["HKQuantityTypeIdentifierHeartRate"]; ok {
				session.HRAvgBPM = stat.Average
				session.HRMinBPM = firstFloatPtr(stat.Min, stat.Minimum)
				session.HRMaxBPM = firstFloatPtr(stat.Max, stat.Maximum)
			}
			if stat, ok := stats["HKQuantityTypeIdentifierDistanceWalkingRunning"]; ok {
				session.DistanceM = firstFloatPtr(stat.Sum, stat.Value)
			}
		}

		sessions = append(sessions, session)
	}

	count, err := h.uc.Import(c.Request().Context(), sessions, string(body), importedAt)
	if err != nil {
		return mapError(c, err)
	}

	return c.JSON(http.StatusOK, workoutImportResponse{Imported: count})
}

func (h *WorkoutHandler) ListWorkouts(c echo.Context) error {
	fromStr := c.QueryParam("from")
	toStr := c.QueryParam("to")

	if fromStr == "" || toStr == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "from and to query params required"})
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "from date format must be YYYY-MM-DD"})
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "to date format must be YYYY-MM-DD"})
	}
	to = to.Add(24*time.Hour - time.Nanosecond)

	sessions, err := h.uc.List(c.Request().Context(), from, to)
	if err != nil {
		return mapError(c, err)
	}

	responses := make([]workoutSessionResponse, 0, len(sessions))
	for _, s := range sessions {
		responses = append(responses, workoutSessionResponse{
			ID:                 s.ID,
			ActivityType:       s.ActivityType,
			StartDate:          s.StartDate.Format(time.RFC3339Nano),
			EndDate:            s.EndDate.Format(time.RFC3339Nano),
			DurationSeconds:    s.DurationSeconds,
			ActiveCaloriesKcal: s.ActiveCaloriesKcal,
			BasalCaloriesKcal:  s.BasalCaloriesKcal,
			HRAvgBPM:           s.HRAvgBPM,
			HRMinBPM:           s.HRMinBPM,
			HRMaxBPM:           s.HRMaxBPM,
			DistanceM:          s.DistanceM,
			Source:             s.Source,
			CreatedAt:          s.CreatedAt.Format(time.RFC3339Nano),
			UpdatedAt:          s.UpdatedAt.Format(time.RFC3339Nano),
		})
	}

	return c.JSON(http.StatusOK, responses)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func firstFloatPtr(values ...*float64) *float64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
