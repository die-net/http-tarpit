package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/die-net/http-tarpit/tarpit"
)

var listenAddr = flag.String("listen", ":8080", "The [IP]:port to listen for incoming connections on.")
var workers = flag.Int("workers", runtime.NumCPU(), "The number of worker threads to execute.")
var period = flag.Duration("period", 16*time.Second, "Time between each byte sent on a connection.")
var timeslice = flag.Duration("timeslice", 50*time.Millisecond, "How often each thread should wake up to send.")
var contentType = flag.String("content_type", "text/html", "The content-type to send with the response.")
var minResponseLen = flag.Int64("min_response_len", 1048576, "The minimum number of bytes to send total per connection.")
var maxResponseLen = flag.Int64("max_response_len", 10485760, "The maximum number of bytes to send total per connection.")
var rcvBuf = flag.Int("rcvbuf", 2048, "Kernel receive buffer size (0=default).")
var sndBuf = flag.Int("sndbuf", 2048, "Kernel send buffer size (0=default).")

func main() {
	flag.Parse()

	setRlimitFromFlags()

	runtime.GOMAXPROCS(*workers)

	tp := tarpit.New(*workers, *contentType, *period, *timeslice, *minResponseLen, *maxResponseLen)
	if tp == nil {
		log.Fatal("Invalid argument")
	}

	http.HandleFunc("/", tp.Handler)
	http.HandleFunc("/robots.txt", robotsDisallowHandler)

	log.Fatal(listenAndServe(*listenAddr))
}

func listenAndServe(addr string) error {
	srv := &http.Server{Addr: addr}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(NewBufSizeListener(*rcvBuf, *sndBuf, ln.(*net.TCPListener)))
}
