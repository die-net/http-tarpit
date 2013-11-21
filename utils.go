package main

import "bufio"
import "net"
import "net/http"
import "math/rand"

func number(rng *rand.Rand, low int, high int) int {
	high -= low
	if high < 0 {
		high = 0
	}
	return low + rng.Int()%(high+1)
}

func robotsDisallowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("User-agent: *\nDisallow: /\n"))
}

func doHijack(w http.ResponseWriter) (net.Conn, *bufio.ReadWriter, bool) {
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return nil, nil, false
	}

	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, nil, false
	}

	return conn, bufrw, true
}
