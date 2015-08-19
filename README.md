ircx
====
[![Build Status](https://travis-ci.org/nickvanw/ircx.svg?branch=master)](https://travis-ci.org/nickvanw/ircx)

ircx is a very basic IRC bot written on top of the wonderfully small [sorcix/irc](https://github.com/sorcix/irc) library. It's designed to be a small building block, a small example of one way to use the library.

Using it is very simple:

```
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
```


This example doesn't do anything other than connect to specified channels and idle, but it's trivial to add additional handlers for any IRC event you want.

Context can be passed around by creating custom Handlers and Senders and using them, versus the default sender created, and an empty handler struct.
