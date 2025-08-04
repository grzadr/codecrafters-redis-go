package connection

import (
	"fmt"
	"net"
	"sync"
)

const (
	defaultConnectionCapacity = 16
)

type ConnectionPool struct {
	connections []net.Conn
	mutex       sync.Mutex
	ack         map[string]int
	resend      int
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make([]net.Conn, 0, defaultConnectionCapacity),
		ack:         make(map[string]int, defaultConnectionCapacity),
	}
}

var (
	pool     *ConnectionPool
	poolOnce sync.Once
)

func (p *ConnectionPool) Add(conn net.Conn) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.connections = append(p.connections, conn)
	p.ack[conn.RemoteAddr().String()] = 0
}

func (p *ConnectionPool) CloseAlls() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, conn := range p.connections {
		conn.Close()
	}

	p = NewConnectionPool()
}

func (p *ConnectionPool) ResetAck() {
	p.ack = make(map[string]int, defaultConnectionCapacity)
}

func (p *ConnectionPool) Ack(addr string, ack int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.ack[addr] = ack
}

func (p *ConnectionPool) Resend(cmd []byte, ack bool) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, conn := range p.connections {
		if _, err := conn.Write(cmd); err != nil {
			return fmt.Errorf("error resending cmd %q: %w", cmd, err)
		}
	}

	p.resend++

	if ack {
		p.ResetAck()
	}

	return nil
}

func (p *ConnectionPool) NumResend() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.resend
}

func (p *ConnectionPool) NumAcknowledged() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return len(p.ack)
}

func GetConnectionPool() *ConnectionPool {
	poolOnce.Do(func() {
		pool = NewConnectionPool()
	})

	return pool
}
