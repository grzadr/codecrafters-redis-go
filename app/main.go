package main

import (
	"io"
	"log"
	"net"
)

func connectTcp(address string) *net.TCPListener {
	ip, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatalf("address %s is not valid", address)
	}

	list, err := net.ListenTCP("tcp", ip)
	if err != nil {
		log.Fatalf("failed to listen to %s: %s", ip, err)
	}

	return list
}

func pong(conn *net.TCPConn, errCh chan error) {
	defer func() {
		if err := conn.Close(); err != nil {
			errCh <- err
		}
	}()

	buf := make([]byte, 4096)
	var err error
	for {
		n := 0
		if n, err = conn.Read(buf); err != nil {
			switch err {
			case io.EOF:
				return
			default:
				errCh <- err
			}
		}
		log.Println(string(buf[:n]))
		if _, err := conn.Write([]byte("+PONG\r\n")); err != nil {
			errCh <- err
		}
		// conn.CloseWrite()
	}
}

func main() {
	l := connectTcp("0.0.0.0:6379")

	errCh := make(chan error, 4)

	go func() {
		for err := range errCh {
			log.Fatalf("connection handler error: %s\n", err)
		}
	}()

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Fatalf("Error accepting connection: %s\n", err)
		}
		go pong(conn, errCh)
	}
}
