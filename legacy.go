package ircx

import (
	"crypto/tls"
)

// Classic creates an instance of ircx poised to connect to the given server
// with the given IRC name.
func Classic(server string, name string) *Bot {
	cfg := Config{
		Options: map[string]bool{"connected": true},
		User:    name,
	}
	return New(server, name, cfg)
}

func WithLogin(server string, name string, user string, password string) *Bot {
	cfg := Config{
		Options:  map[string]bool{"connected": true},
		User:     user,
		Password: password,
	}

	return New(server, name, cfg)
}

// WithTLS creates an instance of ircx poised to connect to the given server
// using TLS with the given IRC name.
func WithTLS(server string, name string, tlsConfig *tls.Config) *Bot {
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	cfg := Config{
		Options:   map[string]bool{"connected": true},
		TLSConfig: tlsConfig,
		User:      name,
	}
	return New(server, name, cfg)
}

func WithLoginTLS(server string, name string, user string, password string, tlsConfig *tls.Config) *Bot {
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	cfg := Config{
		Options:   map[string]bool{"connected": true},
		TLSConfig: tlsConfig,
		User:      user,
		Password:  password,
	}
	return New(server, name, cfg)
}
