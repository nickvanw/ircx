package ircx

import "github.com/sorcix/irc"

// connectMessages is a list of IRC messages to send when attempting to
// connect to the IRC server.
func (b *Bot) connectMessages() []*irc.Message {
	messages := []*irc.Message{}
	if b.Config.Password != "" {
		messages = append(messages, &irc.Message{
			Command: irc.PASS,
			Params:  []string{b.Config.Password},
		})
	}
	messages = append(messages, &irc.Message{
		Command: irc.NICK,
		Params:  []string{b.OriginalName},
	})
	messages = append(messages, &irc.Message{
		Command:  irc.USER,
		Params:   []string{b.Config.User, "0", "*"},
		Trailing: b.Config.User,
	})
	return messages
}
