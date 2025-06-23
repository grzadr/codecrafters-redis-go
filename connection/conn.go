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
	defaultTcpReadBuffer      = 64 * 1024
	defaultReadTimeout        = 50 * time.Millisecond
)

type ConnectionPool struct {
	connections  []net.Conn
	mutex        sync.Mutex
	acknowledged map[string]int
}

func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		connections:  make([]net.Conn, 0, defaultConnectionCapacity),
		acknowledged: make(map[string]int, defaultConnectionCapacity),
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
	p.acknowledged[conn.RemoteAddr().String()] = 0
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
	p.acknowledged = make(map[string]int, defaultConnectionCapacity)
}

func (p *ConnectionPool) Ack(addr string, ack int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.acknowledged[addr] = ack
}

func (p *ConnectionPool) Resend(cmd, ackCmd []byte) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	log.Printf("resending %q\n", cmd)

	p.ResetAck()

	for _, conn := range p.connections {
		if _, err := conn.Write(cmd); err != nil {
			return fmt.Errorf("error resending cmd %q: %w", cmd, err)
		}

		if _, err := conn.Write(ackCmd); err != nil {
			return fmt.Errorf("error sending %q: %w", ackCmd, err)
		}
	}

	log.Printf("completed resending %q\n", cmd)

	return nil
}

// func (p *ConnectionPool) NumConnections() int {
// 	p.mutex.Lock()
// 	defer p.mutex.Unlock()

// 	return len(p.connections)
// }

func (p *ConnectionPool) NumAcknowledged() int {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return len(p.acknowledged)
}

// func (p *ConnectionPool) readWithTimeout(
// 	conn net.Conn,
// 	buf []byte,
// 	timeout time.Duration,
// ) (int, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), timeout)
// 	defer cancel()

// 	type result struct {
// 		n   int
// 		err error
// 	}

// 	resultCh := make(chan result, 1)

// 	go func() {
// 		n, err := conn.Read(buf)
// 		resultCh <- result{n: n, err: err}
// 	}()

// 	select {
// 	case res := <-resultCh:
// 		return res.n, res.err
// 	case <-ctx.Done():
// 		return 0, nil
// 	}
// }

func GetConnectionPool() *ConnectionPool {
	poolOnce.Do(func() {
		pool = NewConnectionPool()
	})

	return pool
}
