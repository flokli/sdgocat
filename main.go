package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

var (
	remoteAddr = flag.String("remote", "localhost:22", "The remote address and port to proxy to")
)

// https://github.com/coreos/go-systemd/blob/7d375ecc2b092916968b5601f74cca28a8de45dd/activation/files_unix.go#L46-L69
// https://github.com/coreos/go-systemd/blob/main/LICENSE
func ActivationFiles() []*os.File {
	const (
		// SD_LISTEN_FDS_START
		listenFdsStart = 3
	)
	pid, err := strconv.Atoi(os.Getenv("LISTEN_PID"))
	if err != nil || pid != os.Getpid() {
		return nil
	}
	nfds, err := strconv.Atoi(os.Getenv("LISTEN_FDS"))
	if err != nil || nfds == 0 {
		return nil
	}
	names := strings.Split(os.Getenv("LISTEN_FDNAMES"), ":")
	files := make([]*os.File, 0, nfds)
	for fd := listenFdsStart; fd < listenFdsStart+nfds; fd++ {
		syscall.CloseOnExec(fd)
		name := "LISTEN_FD_" + strconv.Itoa(fd)
		offset := fd - listenFdsStart
		if offset < len(names) && len(names[offset]) > 0 {
			name = names[offset]
		}
		files = append(files, os.NewFile(uintptr(fd), name))
	}
	return files
}

// https://github.com/coreos/go-systemd/blob/7d375ecc2b092916968b5601f74cca28a8de45dd/activation/listeners.go#L28-L39
func ActivationListeners() []net.Listener {
	files := ActivationFiles()
	listeners := make([]net.Listener, len(files))
	for i, f := range files {
		if pc, err := net.FileListener(f); err == nil {
			listeners[i] = pc
			f.Close()
		}
	}
	return listeners
}

func main() {
	listeners := ActivationListeners()

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
