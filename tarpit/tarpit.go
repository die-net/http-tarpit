package tarpit

import (
	"container/list"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"
)

type tarpit struct {
	contentType    string
	period         time.Duration
	timeslice      time.Duration
	minResponseLen int64
	maxResponseLen int64
	rng            *rand.Rand
	toTimer        chan *tarpitConn
}

type tarpitConn struct {
	conn      net.Conn
	remaining int64
}

func New(workers int, contentType string, period, timeslice time.Duration, minResponseLen, maxResponseLen int64) *tarpit {
	if workers <= 0 || contentType == "" || period.Nanoseconds() <= 0 || timeslice.Nanoseconds() <= 0 || minResponseLen <= 0 || maxResponseLen < minResponseLen {
		return nil
	}

	t := &tarpit{
		contentType:    contentType,
		period:         period,
		timeslice:      timeslice,
		minResponseLen: minResponseLen,
		maxResponseLen: maxResponseLen,
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
		toTimer:        make(chan *tarpitConn, 10000),
	}

	for i := 0; i < workers; i++ {
		go t.timer()
	}

	return t
}

func (t *tarpit) Handler(w http.ResponseWriter, r *http.Request) {
	responseLen := t.rng.Int63n(t.maxResponseLen-t.minResponseLen) + t.minResponseLen

	// Headers must reflect that we don't do chunked encoding.
	w.Header().Set("Content-Type", t.contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(responseLen, 10))
	w.WriteHeader(http.StatusOK)

	if conn, _, ok := hijack(w); ok {
		// Pass this connection on to tarpit.timer().
		tc := &tarpitConn{
			conn:      conn,
			remaining: responseLen,
		}
		t.toTimer <- tc
	}
}

func (t *tarpit) Close() {
	close(t.toTimer)
}

func (t *tarpit) timer() {
	numTimeslices := (int(t.period) + int(t.timeslice) - 1) / int(t.timeslice)
	timeslices := make([]*list.List, numTimeslices)
	for i := range timeslices {
		timeslices[i] = list.New()
	}

	// At startup, randomize within timeslice to try to avoid thundering herd.
	time.Sleep(time.Duration(t.rng.Int63n(int64(t.timeslice))))

	tick := time.NewTicker(t.timeslice)

	nextslice := 0

	for {
		select {
		case tc, ok := <-t.toTimer:
			if !ok {
				break
			}
			timeslices[t.rng.Intn(len(timeslices))].PushBack(tc)

		case <-tick.C:
			// Pick a printable ascii character to send.
			b := make([]byte, 1)
			b[0] = byte(t.rng.Int31n(95) + 32)

			writeConns(timeslices[nextslice], b)

			nextslice++
			if nextslice >= len(timeslices) {
				nextslice = 0
			}
		}
	}

	tick.Stop()

	for slice := 0; slice < len(timeslices); slice++ {
		closeConns(timeslices[slice])
	}
}

// Write a byte array to all conns in a timeslice.

func writeConns(conns *list.List, b []byte) {
	var en *list.Element
	for e := conns.Front(); e != nil; e = en {
		en = e.Next()

		tc, _ := e.Value.(*tarpitConn)

		// This theoretically could block.
		len, err := tc.conn.Write(b)

		tc.remaining--
		if tc.remaining <= 0 || len == 0 || err != nil {
			conns.Remove(e)
			tc.conn.Close()
		}
	}
}

// Close all conns in a timeslice.

func closeConns(conns *list.List) {
	var en *list.Element
	for e := conns.Front(); e != nil; e = en {
		en = e.Next()

		tc, _ := e.Value.(*tarpitConn)
		conns.Remove(e)
		tc.conn.Close()
	}
}
