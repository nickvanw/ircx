package ircx

import (
	"log"
	"math"
	"net"
	"time"

	"github.com/sorcix/irc"
)

type Bot struct {
	Server       string
	OriginalName string
	Password     string
	User         string
	Options      map[string]bool
	Data         chan *irc.Message
	sender       ServerSender
	callbacks    map[string][]Callback
	reader       *irc.Decoder
	writer       *irc.Encoder
	conn         net.Conn
	tries        float64
}

// Classic creates an instance of ircx poised to connect to the given server
// with the given IRC name.
func Classic(server string, name string) *Bot {
	bot := &Bot{
		Server:       server,
		OriginalName: name,
		User:         name,
		Options:      make(map[string]bool),
		Data:         make(chan *irc.Message, 10),
		callbacks:    make(map[string][]Callback),
		tries:        0,
	}
	bot.Options["rejoin"] = true    //Rejoin on kick
	bot.Options["connected"] = true //we are intending to connect
	return bot
}

func WithLogin(server string, name string, user string, password string) *Bot {
	bot := &Bot{
		Server:       server,
		OriginalName: name,
		User:         user,
		Password:     password,
		Options:      make(map[string]bool),
		Data:         make(chan *irc.Message, 10),
		callbacks:    make(map[string][]Callback),
		tries:        0,
	}
	bot.Options["rejoin"] = true    //Rejoin on kick
	bot.Options["connected"] = true //we are intending to connect
	return bot
}

// Connect attempts to connect to the given IRC server
func (b *Bot) Connect() error {
	conn, err := net.Dial("tcp", b.Server)
	if err != nil {
		return err
	}
	b.conn = conn
	b.reader = irc.NewDecoder(conn)
	b.writer = irc.NewEncoder(conn)
	b.sender = ServerSender{writer: &b.writer}
	for _, msg := range b.connectMessages() {
		err := b.writer.Encode(msg)
		if err != nil {
			return err
		}
	}
	log.Println("Connected to", b.Server)
	b.tries = 0
	go b.ReadLoop()
	return nil
}

// Reconnect checks to make sure we want to, and then attempts to
// reconnect to the server
func (b *Bot) Reconnect() {
	data, ok := b.Options["connected"]
	if data || !ok {
		b.conn.Close()
		for err := b.Connect(); err != nil; err = b.Connect() {
			duration := time.Duration(math.Pow(2.0, b.tries)*200) * time.Millisecond
			log.Println("Unable to connect to", b.Server, "- waiting", duration)
			time.Sleep(duration)
			b.tries++
		}
	} else {
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
		b.Data <- msg
	}
}
