package telegram

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdapterSendPostsTelegramForm(t *testing.T) {
	var gotPath, gotChatID, gotText, gotParseMode string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form: %v", err)
		}
		gotChatID = r.Form.Get("chat_id")
		gotText = r.Form.Get("text")
		gotParseMode = r.Form.Get("parse_mode")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	adapter, err := newAdapter(AdapterConfig{BotToken: "secret-token", ChatID: "chat-1"}, server.URL, server.Client())
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	err = adapter.Send(context.Background(), Message{Text: "<b>Olá</b>", ParseMode: ParseModeHTML})
	if err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotPath != "/botsecret-token/sendMessage" {
		t.Fatalf("path = %q, want /botsecret-token/sendMessage", gotPath)
	}
	if gotChatID != "chat-1" {
		t.Fatalf("chat_id = %q, want chat-1", gotChatID)
	}
	if gotText != "<b>Olá</b>" {
		t.Fatalf("text = %q, want formatted message", gotText)
	}
	if gotParseMode != ParseModeHTML {
		t.Fatalf("parse_mode = %q, want %q", gotParseMode, ParseModeHTML)
	}
}

func TestAdapterSendReturnsErrorForTelegramFailures(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "non-2xx", status: http.StatusBadRequest, body: `{"ok":false,"description":"bad request"}`},
		{name: "ok false", status: http.StatusOK, body: `{"ok":false,"description":"chat not found"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			adapter, err := newAdapter(AdapterConfig{BotToken: "secret-token", ChatID: "chat-1"}, server.URL, server.Client())
			if err != nil {
				t.Fatalf("new adapter: %v", err)
			}

			err = adapter.Send(context.Background(), Message{Text: "x"})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if strings.Contains(err.Error(), "secret-token") {
				t.Fatalf("error leaked bot token: %v", err)
			}
		})
	}
}

func TestAdapterSendTransportErrorDoesNotLeakToken(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network down")
	})}
	adapter, err := newAdapter(AdapterConfig{BotToken: "secret-token", ChatID: "chat-1"}, defaultBaseURL, client)
	if err != nil {
		t.Fatalf("new adapter: %v", err)
	}

	err = adapter.Send(context.Background(), Message{Text: "x"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if strings.Contains(err.Error(), "secret-token") {
		t.Fatalf("error leaked bot token: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
