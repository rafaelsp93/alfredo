package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/rafaelsoares/alfredo/internal/database"
	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
)

func TestAppointmentRepository_CRUDAndPetScope(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	repo := NewAppointmentRepository(db)
	insertTestPet(t, db, "pet-1")
	insertTestPet(t, db, "pet-2")

	provider := "Clinica VetCare"
	location := "Rua das Flores, 123"
	notes := "checkup anual"
	scheduledAt := time.Date(2026, 6, 15, 10, 0, 0, 0, time.FixedZone("BRT", -3*60*60))
	createdAt := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)

	created, err := repo.Create(ctx, domain.Appointment{
		ID:                    "appt-1",
		PetID:                 "pet-1",
		Type:                  domain.AppointmentTypeVet,
		ScheduledAt:           scheduledAt,
		Provider:              &provider,
		Location:              &location,
		Notes:                 &notes,
		GoogleCalendarEventID: "event-1",
		CreatedAt:             createdAt,
	})
	if err != nil {
		t.Fatalf("create appointment: %v", err)
	}
	if created.ID != "appt-1" {
		t.Fatalf("created id = %q, want appt-1", created.ID)
	}

	got, err := repo.GetByID(ctx, "pet-1", "appt-1")
	if err != nil {
		t.Fatalf("get appointment: %v", err)
	}
	assertAppointment(t, got, scheduledAt, provider, location, notes)

	_, err = repo.GetByID(ctx, "pet-2", "appt-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("wrong pet get error = %v, want ErrNotFound", err)
	}

	listed, err := repo.List(ctx, "pet-1")
	if err != nil {
		t.Fatalf("list appointments: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != "appt-1" {
		t.Fatalf("listed appointments = %#v, want appt-1", listed)
	}

	updatedNotes := "retorno em 30 dias"
	got.Notes = &updatedNotes
	updated, err := repo.Update(ctx, *got)
	if err != nil {
		t.Fatalf("update appointment: %v", err)
	}
	if updated.Notes == nil || *updated.Notes != updatedNotes {
		t.Fatalf("updated notes = %#v, want %q", updated.Notes, updatedNotes)
	}

	err = repo.Delete(ctx, "pet-2", "appt-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("wrong pet delete error = %v, want ErrNotFound", err)
	}

	if err := repo.Delete(ctx, "pet-1", "appt-1"); err != nil {
		t.Fatalf("delete appointment: %v", err)
	}
	_, err = repo.GetByID(ctx, "pet-1", "appt-1")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("deleted get error = %v, want ErrNotFound", err)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(filepath.Join(t.TempDir(), "alfredo.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close test db: %v", err)
		}
	})
	return db
}

func insertTestPet(t *testing.T, db *sql.DB, id string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO pets (id, name, species, created_at, google_calendar_id) VALUES (?, ?, ?, ?, ?)`,
		id,
		"Luna",
		"dog",
		time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"cal-"+id,
	)
	if err != nil {
		t.Fatalf("insert pet %q: %v", id, err)
	}
}

func assertAppointment(t *testing.T, got *domain.Appointment, scheduledAt time.Time, provider, location, notes string) {
	t.Helper()
	if got.ID != "appt-1" || got.PetID != "pet-1" || got.Type != domain.AppointmentTypeVet {
		t.Fatalf("appointment identity = %#v", got)
	}
	if !got.ScheduledAt.Equal(scheduledAt) {
		t.Fatalf("scheduled_at = %v, want %v", got.ScheduledAt, scheduledAt)
	}
	if got.Provider == nil || *got.Provider != provider {
		t.Fatalf("provider = %#v, want %q", got.Provider, provider)
	}
	if got.Location == nil || *got.Location != location {
		t.Fatalf("location = %#v, want %q", got.Location, location)
	}
	if got.Notes == nil || *got.Notes != notes {
		t.Fatalf("notes = %#v, want %q", got.Notes, notes)
	}
	if got.GoogleCalendarEventID != "event-1" {
		t.Fatalf("event id = %q, want event-1", got.GoogleCalendarEventID)
	}
}
