package main

import (
	"flag"
	"log"

	"github.com/nickvanw/ircx"
	"github.com/sorcix/irc"
)

var (
	name     = flag.String("name", "ircx", "Nick to use in IRC")
	server   = flag.String("server", "chat.freenode.org:6667", "Host:Port to connect to")
	channels = flag.String("chan", "#test", "Channels to join")
)

func init() {
	flag.Parse()
}

func main() {
	bot := ircx.Classic(*server, *name)
	if err := bot.Connect(); err != nil {
		log.Panicln("Unable to dial IRC Server ", err)
	}
	RegisterHandlers(bot)
	bot.HandleLoop()
	log.Println("Exiting..")
}

func RegisterHandlers(bot *ircx.Bot) {
	bot.HandleFunc(irc.RPL_WELCOME, RegisterConnect)
	bot.HandleFunc(irc.PING, PingHandler)
}

func RegisterConnect(s ircx.Sender, m *irc.Message) {
	s.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  []string{*channels},
	})
}

func PingHandler(s ircx.Sender, m *irc.Message) {
	s.Send(&irc.Message{
		Command:  irc.PONG,
		Params:   m.Params,
		Trailing: m.Trailing,
	})
}
