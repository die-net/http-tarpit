package main

import (
	"net/http"
)

func robotsDisallowHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte("User-agent: *\nDisallow: /\n"))
}
