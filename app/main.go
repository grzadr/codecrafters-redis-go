package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/commands"
	"github.com/codecrafters-io/redis-starter-go/connection"
	"github.com/codecrafters-io/redis-starter-go/pubsub"
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
	pubsub.GetStreamManager().Close()
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

	if n, err := conn.Read(buf); err != nil {
		end = true

		if !errors.Is(err, io.EOF) {
			errCh <- err
		}

		return
	} else {
		cmd = buf[:n]
	}

	return
}

func masterExecuteCommand(
	conn *net.TCPConn,
	cmd []byte,
	transaction **commands.Transaction,
) (keepConn bool, err error) {
	for result := range commands.ExecuteCommand(cmd, transaction) {
		if err = sendResponse(conn, result); err != nil {
			return keepConn, err
		}

		pool := connection.GetConnectionPool()

		if keep := result.KeepConnection; keep && !keepConn {
			keepConn = true

			pool.Add(conn)
		}

		if result.Resend {
			if err = pool.Resend(cmd, false); err != nil {
				return keepConn, err
			}
		}

		if result.Ack != 0 {
			pool.Ack(conn.RemoteAddr().String(), result.Ack)
		}
	}

	return keepConn, err
}

func handleConn(conn *net.TCPConn, errCh chan error) {
	var keep bool

	var err error

	defer func() {
		if keep {
			return
		}

		if err := conn.Close(); err != nil {
			errCh <- err
		}
	}()

	var transaction *commands.Transaction

	for {
		cmd, end := readCommand(conn, errCh)
		if end {
			return
		}

		log.Println(hex.Dump(cmd))
		log.Printf("%v", transaction)

		if keep, err = masterExecuteCommand(conn, cmd, &transaction); err != nil {
			errCh <- err

			return
		}
	}
}

func sendHandshake(conn *net.TCPConn, port string) (err error) {
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
		_, err = conn.Write(cmd.cmd)
		if err != nil {
			return fmt.Errorf("failed to send %s: %w", cmd.label, err)
		}

		n, err := conn.Read(response)
		if err != nil {
			return fmt.Errorf("failed to read %s response: %w", cmd.label, err)
		}

		if cmd.label != "psync" {
			continue
		}

		if err = replicaExecuteCommand(conn, response[:n]); err != nil {
			return err
		}
	}

	return err
}

func replicaExecuteCommand(conn *net.TCPConn, cmd []byte) error {
	var transaction *commands.Transaction
	for result := range commands.ExecuteCommand(cmd, &transaction) {
		if result.Err != nil {
			return result.Err
		}

		connection.GetOffsetTracker().Add(result.Size)

		if !result.ReplicaRespond {
			continue
		}

		if err := sendResponse(conn, result); err != nil {
			return err
		}
	}

	return nil
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

	for {
		cmd, done := readCommand(conn, errCh)
		if done {
			return
		}

		if err := replicaExecuteCommand(conn, cmd); err != nil {
			errCh <- err

			return
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
