package telegram

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNoopAdapterSendLogsAndReturnsNil(t *testing.T) {
	var buf bytes.Buffer
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(&buf),
		zapcore.InfoLevel,
	)
	adapter := NewNoopAdapter(zap.New(core))

	err := adapter.Send(context.Background(), Message{Text: "mensagem", ParseMode: ParseModeHTML})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if !strings.Contains(buf.String(), "telegram noop") {
		t.Fatalf("expected noop log line, got %q", buf.String())
	}
}
