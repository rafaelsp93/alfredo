package sqlite

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelsoares/alfredo/internal/health/domain"
)

type WorkoutRepository struct {
	db dbtx
}

func NewWorkoutRepository(db dbtx) *WorkoutRepository {
	return &WorkoutRepository{db: db}
}

func (r *WorkoutRepository) BulkUpsert(ctx context.Context, sessions []domain.WorkoutSession) (int, error) {
	if len(sessions) == 0 {
		return 0, nil
	}

	now := time.Now().UTC()
	count := 0

	for _, s := range sessions {
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO health_workout_sessions
			(activity_type, start_date, end_date, duration_seconds, active_calories_kcal, basal_calories_kcal,
			 hr_avg_bpm, hr_min_bpm, hr_max_bpm, distance_m, source, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(start_date) DO UPDATE SET
				activity_type        = excluded.activity_type,
				end_date             = excluded.end_date,
				duration_seconds     = excluded.duration_seconds,
				active_calories_kcal = excluded.active_calories_kcal,
				basal_calories_kcal  = excluded.basal_calories_kcal,
				hr_avg_bpm           = excluded.hr_avg_bpm,
				hr_min_bpm           = excluded.hr_min_bpm,
				hr_max_bpm           = excluded.hr_max_bpm,
				distance_m           = excluded.distance_m,
				source               = excluded.source,
				updated_at           = excluded.updated_at
		`,
			s.ActivityType,
			s.StartDate.Format(time.RFC3339Nano),
			s.EndDate.Format(time.RFC3339Nano),
			s.DurationSeconds,
			s.ActiveCaloriesKcal,
			s.BasalCaloriesKcal,
			s.HRAvgBPM,
			s.HRMinBPM,
			s.HRMaxBPM,
			s.DistanceM,
			s.Source,
			now.Format(time.RFC3339Nano),
			now.Format(time.RFC3339Nano),
		)
		if err != nil {
			return 0, fmt.Errorf("bulk upsert workouts: %w", err)
		}
		count++
	}

	return count, nil
}

func (r *WorkoutRepository) List(ctx context.Context, from, to time.Time) ([]domain.WorkoutSession, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, activity_type, start_date, end_date, duration_seconds,
		       active_calories_kcal, basal_calories_kcal, hr_avg_bpm, hr_min_bpm, hr_max_bpm,
		       distance_m, source, created_at, updated_at
		FROM health_workout_sessions
		WHERE start_date >= ? AND start_date <= ?
		ORDER BY start_date DESC
	`,
		from.Format(time.RFC3339Nano),
		to.Format(time.RFC3339Nano),
	)
	if err != nil {
		return nil, fmt.Errorf("query workouts: %w", err)
	}
	defer rows.Close()

	var sessions []domain.WorkoutSession
	for rows.Next() {
		var s domain.WorkoutSession
		var startDate, endDate, createdAt, updatedAt string

		err := rows.Scan(
			&s.ID,
			&s.ActivityType,
			&startDate,
			&endDate,
			&s.DurationSeconds,
			&s.ActiveCaloriesKcal,
			&s.BasalCaloriesKcal,
			&s.HRAvgBPM,
			&s.HRMinBPM,
			&s.HRMaxBPM,
			&s.DistanceM,
			&s.Source,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan workout: %w", err)
		}

		s.StartDate, err = time.Parse(time.RFC3339Nano, startDate)
		if err != nil {
			return nil, fmt.Errorf("parse workout start_date: %w", err)
		}
		s.EndDate, err = time.Parse(time.RFC3339Nano, endDate)
		if err != nil {
			return nil, fmt.Errorf("parse workout end_date: %w", err)
		}
		s.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse workout created_at: %w", err)
		}
		s.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse workout updated_at: %w", err)
		}

		sessions = append(sessions, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workouts: %w", err)
	}

	return sessions, nil
}
