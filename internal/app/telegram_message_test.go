package app

import (
	"strings"
	"testing"
	"time"

	"github.com/rafaelsoares/alfredo/internal/petcare/domain"
	"github.com/rafaelsoares/alfredo/internal/telegram"
)

func TestTelegramMessageFormattingEscapesUserControlledText(t *testing.T) {
	administeredAt := time.Date(2026, 5, 10, 9, 30, 0, 0, time.FixedZone("BRT", -3*60*60))
	vaccine := &domain.Vaccine{
		Name:           `Raiva <script>`,
		AdministeredAt: administeredAt,
	}
	pet := &domain.Pet{Name: `Luna & Bob`}

	msg := formatVaccineCreatedMessage(pet, vaccine, "America/Sao_Paulo")

	if msg.ParseMode != telegram.ParseModeHTML {
		t.Fatalf("parse mode = %q, want %q", msg.ParseMode, telegram.ParseModeHTML)
	}
	if !strings.Contains(msg.Text, "Luna &amp; Bob") {
		t.Fatalf("pet name was not escaped: %q", msg.Text)
	}
	if !strings.Contains(msg.Text, "Raiva &lt;script&gt;") {
		t.Fatalf("vaccine name was not escaped: %q", msg.Text)
	}
	if strings.Contains(msg.Text, "<script>") {
		t.Fatalf("message contains raw HTML input: %q", msg.Text)
	}
}
