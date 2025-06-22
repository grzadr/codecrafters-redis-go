package connection

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	defaultConnectionCapacity = 16
	defaultTcpBuffer          = 64 * 1024
	defaultReadTimeout        = 100 * time.Millisecond
)

type ConnectionPool struct {
	connections  []net.Conn
	mutex        sync.Mutex
	acknowledged int
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

func (p *ConnectionPool) Resend(cmd []byte, errCh chan error, respond bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, conn := range p.connections {
		if _, err := conn.Write(cmd); err != nil {
			errCh <- fmt.Errorf("error resending data: %w", err)

			break
		}
	}

	if !respond {
		return
	}

	buf := make([]byte, defaultConnectionCapacity)

	received := make(map[string]struct{}, p.NumConnections())

	log.Println("waiting for replicas response")

	for _, conn := range p.connections {
		addr := conn.RemoteAddr().String()
		if _, ok := received[addr]; ok {
			continue
		}

		if err := conn.SetReadDeadline(time.Now().Add(defaultReadTimeout)); err != nil {
			errCh <- fmt.Errorf("error setting read timeout: %w", err)
		}

		_, err := conn.Read(buf)
		if err != nil {
			errCh <- fmt.Errorf("error reading replica response: %w", err)

			break
		}

		received[addr] = struct{}{}
	}

	p.acknowledged = len(received)
}

func (p *ConnectionPool) NumConnections() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return len(p.connections)
}

func (p *ConnectionPool) NumAcknowledged() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.acknowledged
}

func GetConnectionPool() *ConnectionPool {
	poolOnce.Do(func() {
		pool = NewConnectionPool()
	})

	return pool
}
