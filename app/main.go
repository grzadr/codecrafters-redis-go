package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/commands"
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

func handleConn(conn *net.TCPConn, errCh chan error) {
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
		output, err := commands.ExecuteCommand(buf[:n])
		if err != nil {
			errCh <- fmt.Errorf("error during cmd execution: %w", err)
		}

		if output == nil {
			errCh <- fmt.Errorf("missing command output")
		}

		if _, err := conn.Write(output.Serialize()); err != nil {
			errCh <- fmt.Errorf("error sending data: %w", err)
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
		go handleConn(conn, errCh)
	}
}
