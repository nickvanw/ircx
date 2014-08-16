ircx
====

ircx is a very basic IRC bot written on top of the wonderfully small [sorcix/irc](https://github.com/sorcix/irc) library. It's designed to be a small building block up, as a small example of one way to use the library.

Using it is very simple:

```
package main

import (
	"log"

	"github.com/nickvanw/ircx"
	"github.com/sorcix/irc"
)

func main() {
	bot := ircx.Classic("chat.freenode.org:6667", "ircx")
	if err := bot.Connect(); err != nil {
		log.Panicln("Unable to dial IRC Server ", err)
	}
	RegisterHandlers(bot)
	bot.CallbackLoop()
	log.Println("Exiting..")
}

func RegisterHandlers(bot *ircx.Bot) {
	bot.AddCallback(irc.RPL_WELCOME, irc.HandlerFunc(RegisterConnect))
}

func RegisterConnect(s irc.Sender, m *irc.Message) {
	s.Send(&irc.Message{
		Command: irc.JOIN,
		Params:  []string{"#test,#othertest"},
	})
}
```


This example doesn't do anything other than join a couple of channels, but registering more callbacks is easy and fun!
