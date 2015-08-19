package ircx

import (
	"github.com/sorcix/irc"
)

func (b *Bot) onMessage(m *irc.Message) {
	handlers, ok := b.handlers[m.Command]
	if !ok {
		return
	}
	for _, h := range handlers {
		go h.Handle(b.Sender, m)
	}
}

// Handle registers the handler for the given command
func (b *Bot) Handle(cmd string, handler Handler) {
	b.handlers[cmd] = append(b.handlers[cmd], handler)
}

// Handle registers the handler function for the given command
func (b *Bot) HandleFunc(cmd string, handler func(s Sender, m *irc.Message)) {
	b.handlers[cmd] = append(b.handlers[cmd], HandlerFunc(handler))
}

// HandleLoop reads from the ReadLoop channel and initiates a handler check
// for every message it recieves.
func (b *Bot) HandleLoop() {
	for {
		select {
		case msg, ok := <-b.Data:
			if !ok {
				return
			}
			b.onMessage(msg)
		}
	}
}

type Handler interface {
	Handle(Sender, *irc.Message)
}

type HandlerFunc func(s Sender, m *irc.Message)

func (f HandlerFunc) Handle(s Sender, m *irc.Message) {
	f(s, m)
}
