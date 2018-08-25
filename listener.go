package main

import (
	"net"
	"time"
)

type BufSizeListener struct {
	*net.TCPListener
	ReadBufferSize  int
	WriteBufferSize int
}

func NewBufSizeListener(readBufferSize, writeBufferSize int, l *net.TCPListener) *BufSizeListener {
	return &BufSizeListener{
		TCPListener:     l,
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
	}
}

func (l *BufSizeListener) Accept() (net.Conn, error) {
	tc, err := l.TCPListener.AcceptTCP()
	if err != nil {
		return nil, err
	}
	_ = tc.SetKeepAlive(true)
	_ = tc.SetKeepAlivePeriod(3 * time.Minute)
	if l.ReadBufferSize > 0 {
		_ = tc.SetReadBuffer(l.ReadBufferSize)
	}
	if l.WriteBufferSize > 0 {
		_ = tc.SetWriteBuffer(l.WriteBufferSize)
	}

	return tc, nil
}
