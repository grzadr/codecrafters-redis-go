package main

import (
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

func main() {
	l := connectTcp("0.0.0.0:6379")
	conn, err := l.Accept()
	if err != nil {
		log.Fatalf("Error accepting connection: %s", err)
	}

	conn.Write([]byte("+PONG\r\n"))
}
