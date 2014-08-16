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
			go v.Handle(b.sender, m)
		}
	}
}

// AddCallback is used to add a callback method for a given action
func (b *Bot) AddCallback(value string, f irc.Handler) {
	b.callbacks[value] = append(b.callbacks[value], f)
	log.Println("Added callback for", value)
}

// CallbackLoop reads from the ReadLoop channel and initiates a
// callback check for every message it recieves.
func (b *Bot) CallbackLoop() {
	// Creates the default transport mechanism for all replies
	b.sender = ServerSender{writer: b.writer}
	for {
		select {
		case msg, ok := <-b.Data:
			if ok {
				go b.messageCallback(msg)
			} else {
				return
			}
		}
	}
}
