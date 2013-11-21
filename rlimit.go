package main

import "flag"
import "log"
import "syscall"

var maxConnections = flag.Int("maxConnections", getRlimitMax(syscall.RLIMIT_NOFILE), "The number of incoming connections to allow.")

func getRlimitMax(resource int) int {
	var rlimit syscall.Rlimit

	err := syscall.Getrlimit(resource, &rlimit)

	if err == nil {
		return int(rlimit.Max)
	} else {
		return 0
	}
}

func setRlimit(resource int, value int) {
	rlimit := &syscall.Rlimit{Cur: uint64(value), Max: uint64(value)}

	err := syscall.Setrlimit(resource, rlimit)
	if err != nil {
		log.Fatalln("Error Setting Rlimit ", err)
	}
}

func setRlimitFromFlags() {
	setRlimit(syscall.RLIMIT_NOFILE, *maxConnections)
}
