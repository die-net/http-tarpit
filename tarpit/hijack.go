package tarpit

import (
	"bufio"
	"net"
	"net/http"
)

func hijack(w http.ResponseWriter) (net.Conn, *bufio.ReadWriter, bool) {
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
