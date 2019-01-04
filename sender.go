package ircx

import (
	log "github.com/sirupsen/logrus"
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

	logger func() log.FieldLogger
}

// Send sends the specified message
func (m serverSender) Send(msg *irc.Message) error {
	m.logger().WithField("message", msg.String()).Debug("sending message")
	return m.writer.Encode(msg)
}
