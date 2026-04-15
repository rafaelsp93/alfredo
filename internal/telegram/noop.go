package telegram

import (
	"context"

	"go.uber.org/zap"
)

type NoopAdapter struct {
	logger *zap.Logger
}

func NewNoopAdapter(logger *zap.Logger) *NoopAdapter {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &NoopAdapter{logger: logger}
}

func (a *NoopAdapter) Send(_ context.Context, msg Message) error {
	a.logger.Info("telegram noop send message",
		zap.String("parse_mode", msg.ParseMode),
		zap.String("text", msg.Text),
	)
	return nil
}
