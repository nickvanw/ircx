package ircx

import (
	"log"
	"net"
	"time"

	"github.com/sorcix/irc"
)

type Bot struct {
	Server       string
	OriginalName string
	Options      map[string]bool
	Data         chan *irc.Message
	sender       ServerSender
	callbacks    map[string][]irc.Handler
	reader       *irc.Decoder
	writer       *irc.Encoder
	conn         net.Conn
}

// Classic creates an instance of ircx poised to connect to the given server
// with the given IRC name.
func Classic(server string, name string) *Bot {
	bot := &Bot{
		Server:       server,
		OriginalName: name,
		Options:      make(map[string]bool),
		Data:         make(chan *irc.Message),
		callbacks:    make(map[string][]irc.Handler),
	}
	bot.Options["rejoin"] = true    //Rejoin on kick
	bot.Options["connected"] = true //we are intending to connect
	return bot
}

// Connect attempts to connect to the given IRC server
func (b *Bot) Connect() error {
	log.Println("Connecting..")
	conn, err := net.Dial("tcp", b.Server)
	if err != nil {
		return err
	}
	b.conn = conn
	b.reader = irc.NewDecoder(conn)
	b.writer = irc.NewEncoder(conn)
	for _, msg := range b.connectMessages() {
		err := b.writer.Encode(msg)
		if err != nil {
			return err
		}
	}
	go b.ReadLoop()
	return nil
}

// Reconnect checks to make sure we want to, and then attempts to
// reconnect to the server
func (b *Bot) Reconnect() {
	data, ok := b.Options["connected"]
	if data || !ok {
		b.conn.Close()
		log.Println("Reconnecting..")
		b.Connect()
	} else {
		log.Println("Leaving, bye.")
		close(b.Data)
	}
}

// ReadLoop sets a timeout of 300 seconds, and then attempts to read
// from the IRC server. If there is an error, it calls Reconnect
func (b *Bot) ReadLoop() {
	for {
		b.conn.SetDeadline(time.Now().Add(300 * time.Second))
		msg, err := b.reader.Decode()
		if err != nil {
			b.Reconnect()
			return
		}
		go func() { b.Data <- msg }() // Send the data, let it queue if it wants to
	}
}
