package domain

import "time"

// Supply is a per-pet consumable with enough structured data to compute reorder timing.
type Supply struct {
	ID                  string
	PetID               string
	Name                string
	LastPurchasedAt     time.Time
	EstimatedDaysSupply int
	Notes               *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (s Supply) NextReorderAt() time.Time {
	return s.LastPurchasedAt.AddDate(0, 0, s.EstimatedDaysSupply)
}
