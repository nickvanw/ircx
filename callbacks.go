package ircx

import (
	"errors"

	"github.com/sorcix/irc"
)

var ErrInvalidHandler = errors.New("invalid handler specified for callback")

// messageCallback is called on every message recieved from the IRC
// server, checking to see if there are any actions that need to be performed
func (b *Bot) messageCallback(m *irc.Message) {
	if data, ok := b.callbacks[m.Command]; ok {
		for _, v := range data {
			go v.Handler.Handle(v.Sender, m)
		}
	}
}

// AddCallback is used to add a callback method for a given action
func (b *Bot) AddCallback(value string, c Callback) error {
	if c.Handler == nil {
		return ErrInvalidHandler
	}
	if c.Sender == nil {
		c.Sender = b.Sender // if no sender is specified, use default
	}
	b.callbacks[value] = append(b.callbacks[value], c)
	return nil
}

// CallbackLoop reads from the ReadLoop channel and initiates a
// callback check for every message it recieves.
func (b *Bot) CallbackLoop() {
	for {
		select {
		case msg, ok := <-b.Data:
			if ok {
				b.messageCallback(msg)
			} else {
				return
			}
		}
	}
}

// Callback represents a Handler and a Sender for
// a specified callback
type Callback struct {
	Handler Handler
	Sender  Sender
}

type Handler interface {
	Handle(Sender, *irc.Message)
}

type HandlerFunc func(s Sender, m *irc.Message)

func (f HandlerFunc) Handle(s Sender, m *irc.Message) {
	f(s, m)
}
