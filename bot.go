package ircx

import (
	"crypto/tls"
	"errors"
	"math"
	"net"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"gopkg.in/sorcix/irc.v2"
)

var defaultLogger = log.NewNopLogger()

// Bot contains all of the information necessary to run a single IRC client
type Bot struct {
	Server       string
	OriginalName string
	Config       Config
	Data         chan *irc.Message
	Sender       Sender
	handlers     map[string][]Handler
	reader       *irc.Decoder
	writer       *irc.Encoder
	conn         net.Conn
	tries        int

	log log.Logger
}

// Config contains optional configuration options for an IRC Bot
type Config struct {
	Password   string
	User       string
	TLSConfig  *tls.Config
	MaxRetries int
}

// New creates a new IRC bot with the specified server, name and config
func New(server, name string, config Config) *Bot {
	b := &Bot{
		Server:       server,
		OriginalName: name,
		Config:       config,
		Data:         make(chan *irc.Message, 10), // buffer 10 messages
		handlers:     make(map[string][]Handler),
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
	b.Sender = serverSender{writer: b.writer, logger: b.Logger}
	for _, msg := range b.connectMessages() {
		level.Debug(b.Logger()).Log("action", "send", "message", msg.String())
		if err := b.writer.Encode(msg); err != nil {
			level.Error(b.Logger()).Log("action", "send", "error", err)
			return err
		}
	}
	b.tries = 0
	go b.ReadLoop()
	return nil
}

func (b *Bot) Logger() log.Logger {
	if b.log == nil {
		return defaultLogger
	}
	return b.log
}

func (b *Bot) SetLogger(l log.Logger) {
	b.log = l
}

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
		Command: irc.USER,
		Params:  []string{b.Config.User, "0", "*", b.Config.User},
	})
	return messages
}

// Reconnect checks to make sure we want to, and then attempts to
// reconnect to the server
func (b *Bot) Reconnect() error {
	if b.Config.MaxRetries > 0 {
		b.conn.Close()
		for b.tries < b.Config.MaxRetries {
			level.Info(b.Logger()).Log("action", "reconnect", "server", b.Server)
			if err := b.Connect(); err != nil {
				b.tries++
				duration := time.Duration(math.Pow(2.0, float64(b.tries))*200) * time.Millisecond
				level.Error(b.Logger()).Log("action", "reconnect_error", "delay", duration, "error", err)
				time.Sleep(duration)
			} else {
				break
			}
		}
		return errors.New("Too many reconnects")
	}
	close(b.Data)
	return nil
}

// ReadLoop sets a timeout of 300 seconds, and then attempts to read
// from the IRC server. If there is an error, it calls Reconnect
func (b *Bot) ReadLoop() error {
	for {
		b.conn.SetDeadline(time.Now().Add(300 * time.Second))
		msg, err := b.reader.Decode()
		if err != nil {
			level.Error(b.Logger()).Log("action", "readloop", "error", err)
			return b.Reconnect()
		}
		level.Debug(b.log).Log("action", "read", "message", msg.String())
		b.Data <- msg
	}
}

func (b *Bot) onMessage(m *irc.Message) {
	handlers, ok := b.handlers[m.Command]
	if !ok {
		return
	}
	for _, h := range handlers {
		h.Handle(b.Sender, m)
	}
}

// Handle registers the handler for the given command
func (b *Bot) Handle(cmd string, handler Handler) {
	b.handlers[cmd] = append(b.handlers[cmd], handler)
}

// HandleFunc registers the handler function for the given command
func (b *Bot) HandleFunc(cmd string, handler func(s Sender, m *irc.Message)) {
	b.handlers[cmd] = append(b.handlers[cmd], HandlerFunc(handler))
}

// HandleLoop reads from the ReadLoop channel and initiates a handler check
// for every message it recieves.
func (b *Bot) HandleLoop() {
	for msg := range b.Data {
		b.onMessage(msg)
	}
}

// Handler is an interface to handle IRC messages
type Handler interface {
	Handle(Sender, *irc.Message)
}

// HandlerFunc is a type that represents the method necessary to implement Handler
type HandlerFunc func(s Sender, m *irc.Message)

// Handle calls the HandlerFunc with the sender and irc message
func (f HandlerFunc) Handle(s Sender, m *irc.Message) {
	f(s, m)
}
