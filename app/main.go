package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/commands"
)

const (
	DEFAULT_TCP_BUFFER           = 4096
	DEFAULT_ERR_CHANNEL_CAPACITY = 4
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

	buf := make([]byte, DEFAULT_TCP_BUFFER)

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

		output, err := commands.ExecuteCommand(buf[:n])
		if err != nil {
			errCh <- fmt.Errorf("error during cmd execution: %w", err)
		} else if output == nil {
			errCh <- fmt.Errorf("missing command output")
		} else if _, err := conn.Write(output.Serialize()); err != nil {
			errCh <- fmt.Errorf("error sending data: %w", err)
		} else {
			continue
		}

		break
	}
}

func main() {
	l := connectTcp("0.0.0.0:6379")

	errCh := make(chan error, DEFAULT_ERR_CHANNEL_CAPACITY)

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
