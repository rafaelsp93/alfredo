package messaging

import (
	"context"

	agentdomain "github.com/rafaelsoares/alfredo/internal/agent/domain"
	"github.com/rafaelsoares/alfredo/internal/app/agent/args"
	appagent "github.com/rafaelsoares/alfredo/internal/app/agent/contracts"
	"github.com/rafaelsoares/alfredo/internal/app/agent/registry"
	"github.com/rafaelsoares/alfredo/internal/telegram"
	"go.uber.org/zap"
)

func Specs() []agentdomain.Tool {
	return []agentdomain.Tool{
		registry.Tool("send_telegram", "Send a plain-text Portuguese Telegram message to Rafael. Use after rendering the daily digest from get_pet_summary.", registry.ObjectSchema(registry.Properties("message", "string"), []string{"message"})),
	}
}

func Handlers(deps appagent.MessagingToolsDeps, logger *zap.Logger) []registry.ToolHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return []registry.ToolHandler{sendTelegramHandler{telegram: deps.Telegram, logger: logger}}
}

type sendTelegramHandler struct {
	telegram appagent.TelegramPort
	logger   *zap.Logger
}

func (h sendTelegramHandler) Spec() agentdomain.Tool { return Specs()[0] }

func (h sendTelegramHandler) Handle(ctx context.Context, values map[string]any) (any, error) {
	message, err := args.RequireString(values, "message")
	if err != nil {
		return nil, err
	}
	if h.telegram == nil {
		h.logger.Warn("telegram tool skipped because adapter is not configured")
		return map[string]string{"status": "erro", "message": "Não consegui enviar a mensagem no Telegram."}, nil
	}
	if err := h.telegram.Send(ctx, telegram.Message{Text: message}); err != nil {
		h.logger.Warn("telegram tool send failed", zap.Error(err))
		return map[string]string{"status": "erro", "message": "Não consegui enviar a mensagem no Telegram."}, nil
	}
	return map[string]string{"status": "enviado", "message": "Mensagem enviada no Telegram."}, nil
}
