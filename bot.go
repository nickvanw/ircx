package ircx

import (
	"crypto/tls"
	"math"
	"net"
	"time"

	"github.com/sorcix/irc"
)

type Bot struct {
	Server       string
	OriginalName string
	Config       Config
	Data         chan *irc.Message
	Sender       ServerSender
	callbacks    map[string][]Callback
	reader       *irc.Decoder
	writer       *irc.Encoder
	conn         net.Conn
	tries        int
}

type Config struct {
	Password   string
	User       string
	Options    map[string]bool
	TLSConfig  *tls.Config
	MaxRetries int
}

func New(server, name string, config Config) *Bot {
	b := &Bot{
		Server:       server,
		OriginalName: name,
		Config:       config,
		Data:         make(chan *irc.Message),
		callbacks:    make(map[string][]Callback),
		tries:        0,
	}
	return b
}

// Connect attempts to connect to the given IRC server
func (b *Bot) Connect() error {
	var conn net.Conn
	var err error
	if b.Config.TLSConfig == nil {
		conn, err = net.Dial("tcp", b.Server)
	} else {
		conn, err = tls.Dial("tcp", b.Server, b.Config.TLSConfig)
	}
	if err != nil {
		return err
	}
	b.conn = conn
	b.reader = irc.NewDecoder(conn)
	b.writer = irc.NewEncoder(conn)
	b.Sender = ServerSender{writer: &b.writer}
	for _, msg := range b.connectMessages() {
		if err := b.writer.Encode(msg); err != nil {
			return err
		}
	}
	b.tries = 0
	go b.ReadLoop()
	return nil
}

// Reconnect checks to make sure we want to, and then attempts to
// reconnect to the server
func (b *Bot) Reconnect() error {
	if b.Config.MaxRetries > 0 {
		b.conn.Close()
		var err error
		for err = b.Connect(); err != nil && b.tries < b.Config.MaxRetries; b.tries++ {
			duration := time.Duration(math.Pow(2.0, float64(b.tries))*200) * time.Millisecond
			time.Sleep(duration)
		}
		return err
	} else {
		close(b.Data)
	}
	return nil
}

// ReadLoop sets a timeout of 300 seconds, and then attempts to read
// from the IRC server. If there is an error, it calls Reconnect
func (b *Bot) ReadLoop() error {
	for {
		b.conn.SetDeadline(time.Now().Add(300 * time.Second))
		msg, err := b.reader.Decode()
		if err != nil {
			return b.Reconnect()
		}
		b.Data <- msg
	}
}
