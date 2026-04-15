package domain

import "time"

type Treatment struct {
	ID                    string
	PetID                 string
	Name                  string
	DosageAmount          float64
	DosageUnit            string // e.g. "mg", "ml"
	Route                 string // e.g. "oral", "injection", "topical"
	IntervalHours         int    // 24=daily, 12=BID, 8=TID
	StartedAt             time.Time
	EndedAt               *time.Time // nil = open-ended
	StoppedAt             *time.Time // set when stopped early via DELETE
	VetName               *string
	Notes                 *string
	GoogleCalendarEventID string
	CreatedAt             time.Time
}

type Dose struct {
	ID                    string
	TreatmentID           string
	PetID                 string
	ScheduledFor          time.Time
	GoogleCalendarEventID string
}
