package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// EventEmitter is the interface use cases depend on to emit domain events.
type EventEmitter interface {
	Emit(ctx context.Context, event string, payload any)
}

type envelope struct {
	Event      string `json:"event"`
	OccurredAt string `json:"occurred_at"`
	Domain     string `json:"domain"`
	Payload    any    `json:"payload"`
}

// Emitter sends fire-and-forget webhook events to n8n.
// Emit is a no-op if baseURL is empty.
type Emitter struct {
	baseURL string
	domain  string
	client  *http.Client
	logger  *zap.Logger
}

// New returns an Emitter pointed at baseURL. Pass "" to disable emission.
func New(baseURL, domain string, logger *zap.Logger) *Emitter {
	return &Emitter{
		baseURL: baseURL,
		domain:  domain,
		client:  &http.Client{Timeout: 5 * time.Second},
		logger:  logger,
	}
}

// Emit fires a domain event to n8n as a fire-and-forget POST request.
// ctx is accepted for interface compatibility but is not used to cancel
// the goroutine — callers cannot cancel the HTTP request after Emit returns.
// Emit is a no-op if the Emitter was created with an empty baseURL.
func (e *Emitter) Emit(_ context.Context, event string, payload any) {
	if e.baseURL == "" {
		return
	}
	env := envelope{
		Event:      event,
		OccurredAt: time.Now().UTC().Format(time.RFC3339),
		Domain:     e.domain,
		Payload:    payload,
	}
	body, err := json.Marshal(env)
	if err != nil {
		e.logger.Warn("webhook: marshal failed", zap.String("event", event), zap.Error(err))
		return
	}
	e.logger.Debug("webhook: emitting", zap.String("event", event), zap.Any("payload", env))
	go func() {
		resp, err := e.client.Post(e.baseURL+"/events", "application/json", bytes.NewReader(body))
		if err != nil {
			e.logger.Warn("webhook: emit failed", zap.String("event", event), zap.Error(err))
			return
		}
		defer resp.Body.Close() //nolint:errcheck
		if resp.StatusCode >= 400 {
			respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
			e.logger.Warn("webhook: n8n returned error",
				zap.String("event", event),
				zap.Int("status", resp.StatusCode),
				zap.String("n8n_response", string(respBody)),
			)
		}
	}()
}
