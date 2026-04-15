package timeutil_test

import (
	"testing"
	"time"

	"github.com/rafaelsoares/alfredo/internal/timeutil"
)

func TestParseUserTime(t *testing.T) {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		t.Fatalf("load location: %v", err)
	}

	t.Run("preserves explicit offset", func(t *testing.T) {
		got, err := timeutil.ParseUserTime("2026-04-12T12:00:00-05:00", loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Format(time.RFC3339) != "2026-04-12T12:00:00-05:00" {
			t.Fatalf("got %s", got.Format(time.RFC3339))
		}
	})

	t.Run("applies configured location to naive datetime", func(t *testing.T) {
		got, err := timeutil.ParseUserTime("2026-04-12T12:00:00", loc)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Location().String() != "America/Sao_Paulo" {
			t.Fatalf("got location %s", got.Location())
		}
		if got.Hour() != 12 {
			t.Fatalf("got hour %d", got.Hour())
		}
	})

	t.Run("rejects date only", func(t *testing.T) {
		if _, err := timeutil.ParseUserTime("2026-04-12", loc); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
