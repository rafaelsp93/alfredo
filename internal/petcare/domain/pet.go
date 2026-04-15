package domain

import "time"

// Pet is the root entity — all health and care records belong to a pet.
type Pet struct {
	ID               string
	Name             string
	Species          string
	Breed            *string
	BirthDate        *time.Time
	WeightKg         *float64
	DailyFoodGrams   *float64
	PhotoPath        *string
	GoogleCalendarID string
	CreatedAt        time.Time
}
