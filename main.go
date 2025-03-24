package main

import (
	"flag"
	"io"
	"log"
	"net"
	"sync"

	"github.com/coreos/go-systemd/activation"
)

var (
	remoteAddr = flag.String("remote", "localhost:22", "The remote address and port to proxy to")
)

func main() {
	listeners, err := activation.Listeners()
	if err != nil {
		log.Panicf("cannot get listeners: %s", err)
	}

	if len(listeners) != 1 {
		log.Panicf("unexpected number of listeners, got %d, expected 1", len(listeners))
	}
	listener := listeners[0]

	for {
		inConn, err := listener.Accept()
		if err != nil {
			log.Panicf("unable to accept connection: %v", err)
		}

		net.Dial("tcp", ":5000")
		outConn, err := net.Dial("tcp", *remoteAddr)
		if err != nil {
			log.Panicf("unable to create outgoing connection: %v", err)
		}

		// TODO: maybe we only need to close one side
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			io.Copy(inConn, outConn)
			wg.Done()
		}()
		go func() {
			io.Copy(outConn, inConn)
			wg.Done()
		}()

		wg.Wait()

		inConn.Close()
		outConn.Close()
	}
}
