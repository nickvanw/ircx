package ircx

import (
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	irc "gopkg.in/sorcix/irc.v2"
)

// Sender is an interface for sending IRC messages
type Sender interface {
	// Send sends the given message and returns any errors.
	Send(*irc.Message) error
}

// serverSender is a barebones writer used
// as the default sender for all callbacks
type serverSender struct {
	writer *irc.Encoder

	logger func() log.Logger
}

// Send sends the specified message
func (m serverSender) Send(msg *irc.Message) error {
	level.Debug(m.logger()).Log("action", "send", "message", msg.String())
	return m.writer.Encode(msg)
}
