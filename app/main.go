package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/commands"
)

const (
	defaultTcpBuffer            = 4096
	defaultErrorChannelCapacity = 4
)

var (
	connections []net.Conn
	connMutex   sync.Mutex
)

func addConnection(conn net.Conn) {
	connMutex.Lock()
	defer connMutex.Unlock()

	connections = append(connections, conn)
}

func removeConnection(conn net.Conn) {
	connMutex.Lock()
	defer connMutex.Unlock()

	for i, c := range connections {
		if c == conn {
			connections = append(connections[:i], connections[i+1:]...)

			break
		}
	}
}

func closeAllConnections() {
	connMutex.Lock()
	defer connMutex.Unlock()

	for _, conn := range connections {
		conn.Close()
	}

	connections = nil
}

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

func sendResponse(
	conn *net.TCPConn,
	result commands.CommandResult,
) (keep bool, err error) {
	if err = result.Err; err != nil {
		err = fmt.Errorf("error during cmd execution: %w", err)
	} else if _, err = conn.Write(result.Serialize()); err != nil {
		err = fmt.Errorf("error sending data: %w", err)
	} else if result.KeepConnection {
		addConnection(conn)

		keep = true
	}

	return
}

func handleConn(conn *net.TCPConn, errCh chan error) {
	var keepConnection bool

	var err error

	defer func() {
		if keepConnection {
			return
		}

		if err := conn.Close(); err != nil {
			errCh <- err
		}
	}()

	buf := make([]byte, defaultTcpBuffer)

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

		for result := range commands.ExecuteCommand(buf[:n]) {
			keepConnection, err = sendResponse(conn, result)
			if err != nil {
				errCh <- err
			}

			if keepConnection {
				addConnection(conn)
			}
		}
	}
}

func sendHandshake(c *net.TCPConn, port string) (err error) {
	commands := []struct {
		label string
		cmd   []byte
	}{
		{label: "ping", cmd: commands.NewCmdPing().Render().Serialize()},
		{
			label: "replfconf port",
			cmd: commands.NewCmdReplconf().
				Render("listening-port", port).
				Serialize(),
		},
		{
			label: "replfconf port",
			cmd: commands.NewCmdReplconf().
				Render("capa", "psync2").
				Serialize(),
		},
		{
			label: "psync",
			cmd:   commands.NewCmdPsync().Render("?", "-1").Serialize(),
		},
	}

	response := make([]byte, defaultTcpBuffer)

	for _, cmd := range commands {
		_, err = c.Write(cmd.cmd)
		if err != nil {
			return fmt.Errorf("failed to send %s: %w", cmd.label, err)
		}

		_, err = c.Read(response)
		if err != nil {
			return fmt.Errorf("failed to read %s response: %w", cmd.label, err)
		}
	}

	return err
}

func acceptMasterTCP(master, port string, errCh chan error) {
	addrMaster, portMaster, _ := strings.Cut(master, " ")
	c := dialTcp(addrMaster, portMaster)

	defer func() {
		if err := c.Close(); err != nil {
			errCh <- err
		}
	}()

	if err := sendHandshake(c, port); err != nil {
		errCh <- fmt.Errorf("error during master handshake: %w", err)
	}
}

func main() {
	conf, err := commands.Setup()
	if err != nil {
		log.Fatalf("error during setup: %s", err)
	}

	l := listenTcp("0.0.0.0", conf.Port)

	errCh := make(chan error, defaultErrorChannelCapacity)

	go handleErrors(errCh)

	if conf.IsReplicaOf() {
		go acceptMasterTCP(conf.ReplicaOf.String(), conf.Port, errCh)
	}

	for {
		conn, err := l.AcceptTCP()
		if err != nil {
			log.Fatalf("error accepting connection: %s\n", err)
		}

		// Set keepConnection to true when you want to keep the connection
		// for later closure, false for immediate closure
		keepConnection := false // Change this based on your conditions
		go handleConn(conn, errCh, keepConnection)
	}
}
