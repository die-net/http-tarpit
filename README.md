http-tarpit
===========

Golang-based HTTP [Tarpit][]
[Tarpit]: http://en.wikipedia.org/wiki/Tarpit_(networking)

Copyright &copy; 2013 Aaron Hopkins tools@die.net

A simple HTTP server that drags out answering requests as long as possible,
using a minimal amount of CPU, RAM, and network traffic to do so.

It sends one byte of response every few seconds to keep the client from
timing out, with the goal of tying up an attacking client's limited
connection pool and/or thread pool for as long as possible, slowing it down.

On my 64-bit CentOS 6 machine, http-tarpit eats roughly 1.5KB of RAM per
connection.  With the default of sending one packet per connection every 16
seconds, for 500000 connections it was eating approximately 740 megs of RAM
and one CPU core, and it was sending roughly 20 megabits of responses in
31250 packets per second.

Usage:
-----

Install [Go](http://golang.org/doc/install) and git, then:

> git clone https://github.com/die-net/http-tarpit.git
> cd http-tarpit
> go build
> ./http-tarpit --help
