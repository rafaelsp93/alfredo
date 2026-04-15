package telegram

import "context"

const ParseModeHTML = "HTML"

// Message is a domain-agnostic Telegram message.
type Message struct {
	Text      string
	ParseMode string
}

// Port is the outbound Telegram interface used by app use cases.
type Port interface {
	Send(ctx context.Context, msg Message) error
}
