package ircx

import (
	"fmt"

	"github.com/sorcix/irc"
)

// connectMessages is a list of IRC messages to send when attempting to
// connect to the IRC server.
func (b *Bot) connectMessages() []*irc.Message {
	return []*irc.Message{
		irc.ParseMessage(fmt.Sprintf("USER %s 8 * :%s", b.OriginalName, b.OriginalName)),
		irc.ParseMessage(fmt.Sprintf("NICK %s", b.OriginalName)),
	}
}
