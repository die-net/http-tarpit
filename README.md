http-tarpit [![Build Status](https://travis-ci.org/die-net/http-tarpit.svg?branch=master)](https://travis-ci.org/die-net/http-tarpit)
===========

Golang-based HTTP [Tarpit][]
[Tarpit]: http://en.wikipedia.org/wiki/Tarpit_(networking)

Copyright &copy; 2013 Aaron Hopkins tools@die.net

A simple HTTP server that drags out answering requests as long as possible,
using a minimal amount of CPU, RAM, and network traffic to do so.

It sends one byte of response every few seconds to keep the client from
timing out, with the goal of tying up an attacking client's limited
connection pool and/or thread pool for as long as possible, slowing it down.

To avoid hurting legitimate webcrawlers, requests for /robots.txt are always
answered with a request to not crawl at all.

On my 64-bit CentOS 6 machine, http-tarpit eats roughly 1.5KB of RAM per
connection.  Having been built in Go, there's no specific limit to how many
connections it can handle, but you can specify the number of threads to use
and the upper limit on connections.

With the default of sending one packet per connection every 16
seconds, for 500000 connections it was eating approximately 740 megs of RAM
and one CPU core, and it was sending roughly 20 megabits of responses in
31250 packets per second.

Building:
--------

Install [Go](http://golang.org/doc/install) and git, then:

	git clone https://github.com/die-net/http-tarpit.git
	cd http-tarpit
	go build

And you'll end up with an "http-tarpit" binary in the current directory.

Command-line flags:
------------------

	-listen=":8080": The [IP]:port to listen for incoming connections on.
	-max_connections=4096: The maximum number of incoming connections allowed.
	-period=16s: Time between each byte sent on a connection.
	-response_len=10485760: The number of bytes to send total per connection.
	-timeslice=50ms: How often each thread should wake up to send.
	-workers=4: The number of worker threads to execute.

It defaults to dual-stack IPv4/IPv6.  If you want IPv4-only, specify an IPv4
listen address, like -listen="0.0.0.0:8080".

It will try to raise "ulimit -n" to the max_connections that you specify. 
It defaults to raising the limit as much as it can; if you want it higher
than this, you'll likely need to set the ulimit higher as root.

Each thread will wake up every -timeslice and try to send to a fraction of
the open connections handled by that thread.  This attempts to smooth out
traffic sent, but the shorter the timeslice, the more CPU it will consume
even when idle.  If -timeslice isn't an even multiple of -period, you will
get a slightly inaccurate -period.

The workers count defaults to the number of CPUs you have in /proc/cpuinfo.

Deploying to Heroku:
-------------------

To deploy a tarpit server to [Heroku](http://heroku.com), start with this [guide](http://mmcgrana.github.io/2012/09/getting-started-with-go-on-heroku.html).
Assuming you have the Heroku command line tools installed, you can start with:

	heroku create -b https://github.com/kr/heroku-buildpack-go.git
	git push heroku master
