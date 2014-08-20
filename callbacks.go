package ircx

import (
	"log"

	"github.com/sorcix/irc"
)

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
func (b *Bot) AddCallback(value string, c Callback) {
	if c.Handler == nil {
		log.Println("Ignoring nil handler for callback ", value)
		return
	}
	if c.Sender == nil {
		c.Sender = b.sender // if no sender is specified, use default
	}
	b.callbacks[value] = append(b.callbacks[value], c)
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
	Handler irc.Handler
	Sender  irc.Sender
}
