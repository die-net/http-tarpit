package main

import (
	"container/list"
	"flag"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

type TarpitConn struct {
	conn          net.Conn
	sent          int
	contentLength int
}

var period = flag.Duration("period", 16*time.Second, "Time between each byte sent on a connection.")
var timeslice = flag.Duration("timeslice", 50*time.Millisecond, "How often each thread should wake up to send.")
var responseLen = flag.Int("response_len", 10485760, "The number of bytes to send total per connection.")

var toTimer = make(chan *TarpitConn, 10000)

func tarpitHandler(w http.ResponseWriter, r *http.Request) {
	// Headers must reflect that we don't do chunked encoding.
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", strconv.Itoa(*responseLen))
	w.WriteHeader(http.StatusOK)

	if conn, _, ok := doHijack(w); ok {
		// Pass this connection on to tarpitTimer.
		tc := &TarpitConn{
			conn:          conn,
			sent:          0,
			contentLength: *responseLen,
		}
		toTimer <- tc
	}
}

func tarpitTimer() {
	num_timeslices := (int(*period) + int(*timeslice) - 1) / int(*timeslice)
	timeslices := make([]*list.List, num_timeslices)
	for i := range timeslices {
		timeslices[i] = list.New()
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// At startup, randomize within timeslice to try to avoid thundering herd.
	time.Sleep(time.Duration(rng.Int63n(int64(*timeslice))))

	tick := time.NewTicker(*timeslice)

	nextslice := 0

	for {
		select {
		case tc := <-toTimer:
			timeslices[rng.Int31n(int32(len(timeslices)))].PushBack(tc)

		case <-tick.C:
			// Pick a printable ascii character to send.
			b := make([]byte, 1)
			b[0] = byte(rng.Int31n(95) + 32)

			tarpitWrite(timeslices[nextslice], b)

			nextslice++
			if nextslice >= len(timeslices) {
				nextslice = 0
			}
		}
	}
}

// Write a byte array to all conns in a timeslice.

func tarpitWrite(conns *list.List, b []byte) {
	var en *list.Element
	for e := conns.Front(); e != nil; e = en {
		en = e.Next()

		tc, _ := e.Value.(*TarpitConn)

		// FIXME: This theoretically could block.
		len, err := tc.conn.Write(b)

		tc.sent++
		if tc.sent >= tc.contentLength || len == 0 || err != nil {
			conns.Remove(e)
			tc.conn.Close()
		}
	}
}
