package ircx

import "github.com/sorcix/irc"

type Sender interface {
	// Send sends the given message and returns any errors.
	Send(*irc.Message) error
}

// ServerSender is a barebones writer used
// as the default sender for all callbacks
type ServerSender struct {
	writer *irc.Encoder
}

func (m ServerSender) Send(msg *irc.Message) error {
	return m.writer.Encode(msg)
}
