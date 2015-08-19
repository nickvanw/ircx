package ircx

import (
	"bufio"
	"fmt"
	"net"
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	wantOpts := map[string]bool{
		"rejoin":    true,
		"connected": true,
	}
	cfg := Config{
		Options: wantOpts,
		User:    "test-user",
	}
	b := New("irc.example.org", "test-bot", cfg)
	if b.Server != "irc.example.org" {
		t.Fatalf("Wanted server %s, got %s", "irc.example.org", b.Server)
	}
	if b.OriginalName != "test-bot" {
		t.Fatalf("Wanted name %s, got %s", "test-bot", b.OriginalName)
	}
	if b.Config.User != "test-user" {
		t.Fatalf("Wanted user %s, got %s", "test-user", b.Config.User)
	}
	if !reflect.DeepEqual(b.Config.Options, wantOpts) {
		t.Fatalf("Wanted config options %#v, got %#v", wantOpts, b.Config.Options)
	}
}

func TestClassicHelper(t *testing.T) {
	b := Classic("irc.example.org", "test-bot")
	wantOpts := map[string]bool{
		"connected": true,
	}
	if b.Server != "irc.example.org" {
		t.Fatalf("Wanted server %s, got %s", "irc.example.org", b.Server)
	}
	if b.OriginalName != "test-bot" {
		t.Fatalf("Wanted name %s, got %s", "test-bot", b.OriginalName)
	}
	if b.Config.User != "test-bot" {
		t.Fatalf("Wanted user %s, got %s", "test-bot", b.Config.User)
	}
	if !reflect.DeepEqual(b.Config.Options, wantOpts) {
		t.Fatalf("Wanted config options %#v, got %#v", wantOpts, b.Config.Options)
	}
}

func TestPasswordHelper(t *testing.T) {
	b := WithLogin("irc.example.org", "test-bot", "test-user", "test-password")
	wantOpts := map[string]bool{
		"connected": true,
	}
	if b.Server != "irc.example.org" {
		t.Fatalf("Wanted server %s, got %s", "irc.example.org", b.Server)
	}
	if b.OriginalName != "test-bot" {
		t.Fatalf("Wanted name %s, got %s", "test-bot", b.OriginalName)
	}
	if b.Config.User != "test-user" {
		t.Fatalf("Wanted user %s, got %s", "test-user", b.Config.User)
	}
	if b.Config.Password != "test-password" {
		t.Fatalf("Wanted password %s, got %s", "test-password", b.Config.Password)
	}
	if !reflect.DeepEqual(b.Config.Options, wantOpts) {
		t.Fatalf("Wanted config options %#v, got %#v", wantOpts, b.Config.Options)
	}
}

func TestConnect(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	b := Classic(l.Addr().String(), "test-bot")

	not := make(chan struct{})
	go dummyHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock server: %v", err)
	}

	// block on 2 seconds or recieving that the mock server has been connected to
	select {
	case <-not:
		return
	case <-time.After(2 * time.Second):
		t.Fatal("dummy server did not get connected to after 2 seconds")
	}
}

func TestSendsData(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	b := WithLogin(l.Addr().String(), "test-bot", "test-user", "test-password")

	not := make(chan string)
	go echoHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock server: %v", err)
	}

	// We should get back the connect info. If 500ms has happened and we haven't gotten anything
	// we're either not connected right, or all of the data has been sent.
	data := []string{}
	for {
		select {
		case d := <-not:
			data = append(data, d)
		case <-time.After(250 * time.Millisecond):
			goto DONE
		}
	}
DONE:
	d := b.connectMessages()
	for k, v := range d {
		if v.String() != data[k] {
			t.Fatalf("Should have recieved %s, got %s", d[k], v)
		}
	}
}

func TestSendsDataWithPassword(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	b := Classic(l.Addr().String(), "test-bot")

	not := make(chan string)
	go echoHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock server: %v", err)
	}

	// We should get back the connect info. If 500ms has happened and we haven't gotten anything
	// we're either not connected right, or all of the data has been sent.
	data := []string{}
	for {
		select {
		case d := <-not:
			data = append(data, d)
		case <-time.After(250 * time.Millisecond):
			goto DONE
		}
	}
DONE:
	d := b.connectMessages()
	for k, v := range d {
		if v.String() != data[k] {
			t.Fatalf("Should have recieved %s, got %s", d[k], v)
		}
	}
}

func TestReconnect(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Wanted listener, got err: %v", err)
	}

	b := Classic(l.Addr().String(), "test-bot")

	not := make(chan struct{})
	go dcHelper(l, not)
	err = b.Connect()
	if err != nil {
		t.Fatalf("error connecting to mock server: %v", err)
	}
	tries := 0
	for {
		select {
		case <-not:
			tries++
			if tries > 2 {
				return
			}
		case <-time.After(1 * time.Second):
			t.Fatal("dummy server did not get reconnected in time, reconnect is broken")
		}
	}
}

func dummyHelper(l net.Listener, not chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func(c net.Conn) {
			not <- struct{}{}
		}(conn)
	}
}

func echoHelper(l net.Listener, not chan string) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func(c net.Conn) {
			rdr := bufio.NewReader(c)
			for {
				d, _, _ := rdr.ReadLine()
				not <- string(d)
			}
		}(conn)
	}
}

func dcHelper(l net.Listener, not chan struct{}) {
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go func(c net.Conn) {
			time.Sleep(500 * time.Millisecond)
			c.Close()
			not <- struct{}{}
		}(conn)
	}
}
