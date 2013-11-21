package main

import "container/list"
import "flag"
import "net"
import "net/http"
import "math/rand"
import "strconv"
import "time"

type TarpitConn struct {
	conn          net.Conn
	sent          int
	contentLength int
}

var period = flag.Duration("period", 16*time.Second, "The approximate time between each byte sent.")
var timeslice = flag.Duration("timeslice", 50*time.Millisecond, "How often to send something.")
var responseLen = flag.Int("responseLen", 10485760, "The number of bytes to send in each response.")

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
	time.Sleep(time.Duration(number(rng, 0, int(*timeslice))))

	tick := time.NewTicker(*timeslice)

	nextslice := 0

	for {
		select {
		case tc := <-toTimer:
			timeslices[number(rng, 0, len(timeslices)-1)].PushBack(tc)

		case <-tick.C:
			// Pick a printable ascii character to send.
			b := make([]byte, 1)
			b[0] = byte(number(rng, 32, 126))

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
