package ircx

import "github.com/sorcix/irc"

type Sender interface {
	Send(*irc.Message) error
}

// ServerSender is a barebones writer used
// as the default sender for all callbacks
type ServerSender struct {
	writer **irc.Encoder
}

// Send implements the irc.Handler Send method, and merely
// sends the given message, returning any errors that may have
// occured
func (m ServerSender) Send(msg *irc.Message) error {
	writer := *m.writer
	return writer.Encode(msg)
}
