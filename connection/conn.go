package connection

import (
	"fmt"
	"net"
	"sync"
)

const defaultConnectionCapacity = 16

type ConnectionPool struct {
	connections []net.Conn
	mutex       sync.Mutex
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections: make([]net.Conn, 0, defaultConnectionCapacity),
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
}

func (p *ConnectionPool) CloseAlls() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, conn := range p.connections {
		conn.Close()
	}

	p = NewConnectionPool()
}

func (p *ConnectionPool) Resend(cmd []byte, errCh chan error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, conn := range p.connections {
		if _, err := conn.Write(cmd); err != nil {
			errCh <- fmt.Errorf("error resending data: %w", err)

			break
		}
	}
}

func GetConnectionPool() *ConnectionPool {
	poolOnce.Do(func() {
		pool = NewConnectionPool()
	})

	return pool
}
