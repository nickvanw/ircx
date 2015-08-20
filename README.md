ircx
====
[![Build Status](https://travis-ci.org/nickvanw/ircx.svg?branch=master)](https://travis-ci.org/nickvanw/ircx)

ircx is a very basic IRC bot written on top of the wonderfully small [sorcix/irc](https://github.com/sorcix/irc) library. It's designed to be a small building block, a small example of one way to use the library.

Using it is very simple, see [the example](example/main.go).

This example doesn't do anything other than connect to specified channels and idle, but it's trivial to add additional handlers for any IRC event you want.

Context can be passed around by creating custom Handlers and Senders and using them, versus the default sender created, and an empty handler struct.
