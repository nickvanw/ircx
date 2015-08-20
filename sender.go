package ircx

import "github.com/sorcix/irc"

type Sender interface {
	// Send sends the given message and returns any errors.
	Send(*irc.Message) error
}

// serverSender is a barebones writer used
// as the default sender for all callbacks
type serverSender struct {
	writer *irc.Encoder
}

func (m serverSender) Send(msg *irc.Message) error {
	return m.writer.Encode(msg)
}
