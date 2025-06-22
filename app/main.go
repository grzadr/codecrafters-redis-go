package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/commands"
	"github.com/codecrafters-io/redis-starter-go/connection"
)

const (
	defaultTcpBuffer            = 64 * 1024
	defaultErrorChannelCapacity = 4
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

func clean() {
	commands.CloseMaps()
	connection.GetConnectionPool().CloseAlls()
}

func handleErrors(errCh chan error) {
	for err := range errCh {
		clean()
		log.Fatalf("connection handler error: %s\n", err)
	}
}

func sendResponse(
	conn *net.TCPConn,
	result *commands.CommandResult,
) (err error) {
	if err = result.Err; err != nil {
		err = fmt.Errorf("error during cmd execution: %w", err)
	} else if _, err = conn.Write(result.Serialize()); err != nil {
		err = fmt.Errorf("error sending data: %w", err)
	}

	return
}

func readCommand(conn *net.TCPConn, errCh chan error) (cmd []byte, end bool) {
	buf := make([]byte, defaultTcpBuffer)
	// n := 0

	if n, err := conn.Read(buf); err != nil {
		end = true

		if !errors.Is(err, io.EOF) {
			errCh <- err
		}

		return
	} else {
		cmd = buf[:n]
	}

	log.Printf("read cmd: %q\n", cmd)

	return
}

func handleConn(conn *net.TCPConn, errCh chan error) {
	var keep bool

	defer func() {
		if keep {
			return
		}

		if err := conn.Close(); err != nil {
			errCh <- err
		}
	}()

	for {
		cmd, end := readCommand(conn, errCh)
		if end {
			return
		}

		for result := range commands.ExecuteCommand(cmd) {
			if err := result.Err; err != nil {
				errCh <- err

				return
			} else if err := sendResponse(conn, result); err != nil {
				errCh <- err

				return
			}

			pool := connection.GetConnectionPool()

			if keep = result.KeepConnection; keep {
				pool.Add(conn)
			}

			if result.Resend {
				pool.Resend(cmd, errCh)
			}
		}
	}
}

func sendHandshake(c *net.TCPConn, port string) (err error) {
	handshakeCommands := []struct {
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
			label: "replfconf capa",
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

	for _, cmd := range handshakeCommands {
		log.Printf("sending %q: %q", cmd.label, cmd.cmd)

		_, err = c.Write(cmd.cmd)
		if err != nil {
			return fmt.Errorf("failed to send %s: %w", cmd.label, err)
		}

		n, err := c.Read(response)
		if err != nil {
			return fmt.Errorf("failed to read %s response: %w", cmd.label, err)
		}

		log.Printf("received %q (%d): %q", cmd.label, n, response[:n])
	}

	return err
}

func acceptMasterTCP(master, port string, errCh chan error) {
	addrMaster, portMaster, _ := strings.Cut(master, " ")
	conn := dialTcp(addrMaster, portMaster)

	defer func() {
		if err := conn.Close(); err != nil {
			errCh <- err
		}
	}()

	if err := sendHandshake(conn, port); err != nil {
		errCh <- fmt.Errorf("error during master handshake: %w", err)

		return
	}

	log.Println("handshake complete")

	for {
		cmd, done := readCommand(conn, errCh)
		if done {
			return
		}

		for result := range commands.ExecuteCommand(cmd) {
			log.Println(result)

			if err := result.Err; err != nil {
				errCh <- err

				return
			}

			if !result.ReplicaRespond {
				log.Println("no replica response needed")

				continue
			}

			log.Println("responding")

			if err := sendResponse(conn, result); err != nil {
				errCh <- err

				return
			}

			connection.GetOffsetTracker().Add(result.Size)
		}
	}
}

func main() {
	conf, err := commands.Setup()

	defer clean()

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
			errCh <- fmt.Errorf("error accepting connection: %s\n", err)
		}

		go handleConn(conn, errCh)
	}
}
