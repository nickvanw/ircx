package ircx

import "github.com/sorcix/irc"

// ServerSender is a barebones writer used to send messages
type ServerSender struct {
	writer *irc.Encoder
}

// Send implements the irc.Handler Send method, and merely
// sends the given message, returning any errors that may have
// occured
func (m ServerSender) Send(msg *irc.Message) error {
	return m.writer.Encode(msg)
}
