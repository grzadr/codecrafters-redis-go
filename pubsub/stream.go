package pubsub

import (
	"sync"
)

const defaultStreamManagerCapacity = 1024

var (
	manager     *StreamManager
	managerOnce sync.Once
)

type Message struct {
	Topic   string
	Payload any
}

type subscriber struct {
	ch   chan Message
	quit chan struct{}
}

type stream struct {
	subscribers map[string]*subscriber
	publish     chan Message
	subscribe   chan subscribeReq
	unsubscribe chan string
	quit        chan struct{}
}

type subscribeReq struct {
	id string
	ch chan Message
}

type StreamManager struct {
	streams map[string]*stream
	mu      sync.RWMutex
}

func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*stream, defaultStreamManagerCapacity),
	}
}

func GetStreamManager() *StreamManager {
	managerOnce.Do(func() {
		manager = NewStreamManager()
	})

	return manager
}
