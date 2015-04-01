package ircx

import (
	"crypto/tls"
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
	tlsConfig    *tls.Config
	sender       ServerSender
	callbacks    map[string][]Callback
	reader       *irc.Decoder
	writer       *irc.Encoder
	conn         net.Conn
	tries        float64
}

func NewBot(f ...func(*Bot)) *Bot {
	defaultOpts := map[string]bool{
		"rejoin":    true,
		"connected": true,
	}
	b := &Bot{
		Options:   defaultOpts,
		Data:      make(chan *irc.Message),
		callbacks: make(map[string][]Callback),
		tries:     0,
	}
	for _, v := range f {
		v(b)
	}
	return b
}

// Classic creates an instance of ircx poised to connect to the given server
// with the given IRC name.
func Classic(server string, name string) *Bot {
	configFunc := func(b *Bot) {
		b.Server = server
		b.OriginalName = name
		b.User = name
	}
	return NewBot(configFunc)
}

func WithLogin(server string, name string, user string, password string) *Bot {
	configFunc := func(b *Bot) {
		b.Server = server
		b.OriginalName = name
		b.User = user
		b.Password = password
	}
	return NewBot(configFunc)
}

// WithTLS creates an instance of ircx poised to connect to the given server
// using TLS with the given IRC name.
func WithTLS(server string, name string, tlsConfig *tls.Config) *Bot {
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	configFunc := func(b *Bot) {
		b.Server = server
		b.OriginalName = name
		b.User = name
		b.tlsConfig = tlsConfig
	}
	return NewBot(configFunc)
}

func WithLoginTLS(server string, name string, user string, password string, tlsConfig *tls.Config) *Bot {
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	configFunc := func(b *Bot) {
		b.Server = server
		b.OriginalName = name
		b.User = user
		b.Password = password
		b.tlsConfig = tlsConfig
	}
	return NewBot(configFunc)
}

// Connect attempts to connect to the given IRC server
func (b *Bot) Connect() error {
	var conn net.Conn
	var err error
	if b.tlsConfig == nil {
		conn, err = net.Dial("tcp", b.Server)
	} else {
		conn, err = tls.Dial("tcp", b.Server, b.tlsConfig)
	}
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
			log.Printf("Unable to connect to %s, waiting %s", b.Server, duration.String())
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
			log.Println("Error:", err)
			b.Reconnect()
			return
		}
		b.Data <- msg
	}
}
