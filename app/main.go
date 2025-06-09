package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/commands"
)

const (
	DEFAULT_TCP_BUFFER           = 4096
	DEFAULT_ERR_CHANNEL_CAPACITY = 4
)

func listenTcp(address, port string) *net.TCPListener {
	ip, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", address, port))
	if err != nil {
		log.Fatalf("address %s is not valid", address)
	}

	list, err := net.ListenTCP("tcp", ip)
	if err != nil {
		log.Fatalf("failed to listen to %s: %s", ip, err)
	}

	return list
}

func dialTcp(address, port string) *net.TCPConn {
	ip, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", address, port))
	if err != nil {
		log.Fatalf("address %s is not valid", address)
	}

	dial, err := net.DialTCP("tcp", nil, ip)
	if err != nil {
		log.Fatalf("failed to dial to %s: %s", ip, err)
	}

	return dial
}

func handleErrors(errCh chan error) {
	for err := range errCh {
		commands.CloseMaps()
		log.Fatalf("connection handler error: %s\n", err)
	}
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

func acceptMasterTCP(master string, errCh chan error) {
	addr, port, _ := strings.Cut(master, " ")
	c := dialTcp(addr, port)

	defer func() {
		if err := c.Close(); err != nil {
			errCh <- err
		}
	}()

	_, err := c.Write(commands.NewCmdPing().Render().Serialize())
	if err != nil {
		errCh <- fmt.Errorf("error during master handshake: %w", err)
	}
}

func main() {
	conf, err := commands.Setup()
	if err != nil {
		log.Fatalf("error during setup: %s", err)
	}

	l := listenTcp("0.0.0.0", conf.Port)

	errCh := make(chan error, DEFAULT_ERR_CHANNEL_CAPACITY)

	go handleErrors(errCh)

	if conf.IsReplicaOf() {
		go acceptMasterTCP(conf.ReplicaOf.String(), errCh)
	}

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Fatalf("error accepting connection: %s\n", err)
		}

		go handleConn(conn, errCh)
	}
}
