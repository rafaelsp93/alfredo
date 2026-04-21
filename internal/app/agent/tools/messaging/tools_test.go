package messaging

import (
	"context"
	"errors"
	"testing"

	agentcontracts "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/telegram"
	"go.uber.org/zap"
)

func TestSendTelegramHandler(t *testing.T) {
	handlers := Handlers(agentcontracts.MessagingToolsDeps{Telegram: fakeTelegram{}}, zap.NewNop())
	if len(Specs()) != 1 || handlers[0].Spec().Name != "send_telegram" {
		t.Fatalf("unexpected specs")
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{"message": "oi"}); err != nil {
		t.Fatalf("send err = %v", err)
	}
	handlers = Handlers(agentcontracts.MessagingToolsDeps{Telegram: fakeTelegram{err: errors.New("down")}}, zap.NewNop())
	if _, err := handlers[0].Handle(context.Background(), map[string]any{"message": "oi"}); err != nil {
		t.Fatalf("send failure should still be best effort: %v", err)
	}
	handlers = Handlers(agentcontracts.MessagingToolsDeps{}, zap.NewNop())
	if _, err := handlers[0].Handle(context.Background(), map[string]any{"message": "oi"}); err != nil {
		t.Fatalf("nil adapter should still be best effort: %v", err)
	}
	if _, err := handlers[0].Handle(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected missing message error")
	}
}

type fakeTelegram struct{ err error }

func (f fakeTelegram) Send(context.Context, telegram.Message) error { return f.err }
