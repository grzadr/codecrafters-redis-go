package pubsub

import (
	"context"
	"iter"
	"maps"
	"slices"
	"sync"
	"time"
)

const defaultStreamCapacity = 64

// Message represents a published message.
//
//	type Message struct {
//		Topic   string
//		Payload any
//	}
type (
	Message     any
	doneChannel chan struct{}
)

// Subscription represents a single Subscription to a stream.
type Subscription struct {
	id       int
	Messages chan Message
	Done     doneChannel
	once     sync.Once
	unsub    chan int
}

func newSubscription(id int, unsub chan int) *Subscription {
	return &Subscription{
		id:       id,
		Messages: make(chan Message, 1),
		Done:     make(doneChannel),
		unsub:    unsub,
	}
}

func (sub *Subscription) Close() {
	sub.once.Do(func() {
		close(sub.Messages)
		sub.unsub <- sub.id
		close(sub.Done)
	})
}

// // subscribeReq represents a subscription request.
// type subscribeReq struct {
// 	id string
// 	ch chan Message
// }

// stream manages subscribers for a single topic.
type stream struct {
	subscribers map[int]*Subscription
	lastId      int
	lock        sync.Mutex
	msg         chan Message
	sub         chan *Subscription
	unsub       chan int
	sendFirst   bool          // send only to first subscriber
	done        chan struct{} // sends message when done
	quit        chan struct{} // receives signal to quit
}

func newStream(quit chan struct{}, sendFirst bool) *stream {
	return &stream{
		subscribers: make(map[int]*Subscription, defaultStreamCapacity),
		sub:         make(chan *Subscription, 1),
		msg:         make(chan Message, 1),
		unsub:       make(chan int, 1),
		done:        make(chan struct{}),
		quit:        quit,
		sendFirst:   sendFirst,
	}
}

func (s *stream) delete(id int) {
	// if sub, found := s.subscribers[id]; found {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.subscribers, id)
	// }
}

func (s *stream) subscribe() *Subscription {
	s.lock.Lock()
	defer s.lock.Unlock()

	sub := newSubscription(s.lastId, s.unsub)
	s.lastId++
	s.subscribers[sub.id] = sub

	return sub
}

func (s *stream) iterSubscriptions() iter.Seq[*Subscription] {
	if s.sendFirst {
		return func(yield func(*Subscription) bool) {
			// keys := slices.Collect()
			for key := range slices.Sorted(maps.Keys(s.subscribers)) {
				if !yield(s.subscribers[key]) {
					return
				}
			}
		}
	} else {
		return maps.Values(s.subscribers)
	}
}

func (s *stream) publishMsg(msg Message) {
	s.lock.Lock()
	defer s.lock.Unlock()

loop:
	for sub := range s.iterSubscriptions() {
		select {
		case sub.Messages <- msg:
			if s.sendFirst {
				break loop
			}
		case <-sub.Done:
		default:
			sub.Close()
		}
	}
}

func (s *stream) clean() {
	s.lock.Lock()
	defer s.lock.Unlock()

	var closed sync.WaitGroup

	for _, sub := range s.subscribers {
		closed.Add(1)

		go func(s *Subscription) {
			defer closed.Done()
			s.Close() // Wait for stream completion
		}(sub)
	}

	closed.Wait()
}

func (s *stream) close() {
	s.lock.Lock()
	defer s.lock.Unlock()
	close(s.msg)
	close(s.sub)
	close(s.unsub)

	close(s.done)
}

func (s *stream) shouldClose() bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	return len(s.subscribers) == 0
}

// run is the main event loop for a stream.
func (s *stream) run() {
	defer s.close()

	for {
		select {
		// case sub := <-s.subCh:
		// 	s.subscribe(sub)
		case id := <-s.unsub:
			s.delete(id)

			if s.shouldClose() {
				return
			}
		case msg := <-s.msg:
			s.publishMsg(msg)
		case <-s.quit:
			s.clean()
		}
	}
}

// StreamManager is the main broker managing all streams.
type StreamManager struct {
	streams map[string]*stream
	quit    chan struct{}
	mu      sync.RWMutex
}

// newStreamManager creates a new PubSub instance.
func newStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*stream),
		quit:    make(chan struct{}),
	}
}

// Subscribe creates a subscription to a stream.
func (m *StreamManager) Subscribe(
	streamName string,
	sendFirst bool,
) *Subscription {
	m.mu.Lock()
	defer m.mu.Unlock()

	st, exists := m.streams[streamName]
	if !exists {
		st = newStream(m.quit, sendFirst)
		m.streams[streamName] = st

		go st.run()
	}

	return st.subscribe()
}

// Publish sends a message to all subscribers of a stream.
func (m *StreamManager) Publish(streamName string, msg any) {
	m.mu.RLock()
	st, exists := m.streams[streamName]
	m.mu.RUnlock()

	if !exists {
		return
	}

	select {
	case st.msg <- Message(msg):
	default:
	}
}

// // Unsubscribe removes a subscriber from a stream.
// func (ps *PubSub) Unsubscribe(streamName, subscriberID string) {
// 	ps.mu.RLock()
// 	st, exists := ps.streams[streamName]
// 	ps.mu.RUnlock()

// 	if !exists {
// 		return
// 	}

// 	select {
// 	case st.unsubscribe <- subscriberID:
// 	case <-st.quit:
// 	}
// }

// CloseStream closes a specific stream.
// func (m *StreamManager) CloseStream(streamName string) {
// 	m.mu.Lock()
// 	defer m.mu.Unlock()

// 	m.closeStreamLocked(streamName)
// }

// Close closes all streams.
func (m *StreamManager) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	close(m.quit)

	var closed sync.WaitGroup

	for _, s := range m.streams {
		closed.Add(1)

		go func(s *stream) {
			defer closed.Done()
			<-s.done // Wait for stream completion
		}(s)
	}

	closed.Wait()

	m.streams = nil
}

// closeStreamLocked closes a stream (assumes lock is held).
// func (m *StreamManager) closeStreamLocked(streamName string) {
// 	if st, exists := m.streams[streamName]; exists {
// 		close(st.closed)
// 		delete(m.streams, streamName)
// 	}
// }

func (m *StreamManager) run() {
	for {
		deletionList := make([]string, 0, len(m.streams))

		for name, stream := range m.streams {
			select {
			case <-stream.done:
				deletionList = append(deletionList, name)
			default:
			}
		}

		m.mu.Lock()

		for _, id := range deletionList {
			delete(m.streams, id)
		}

		m.mu.Unlock()
	}
}

var (
	streamManager     *StreamManager
	streamManagerOnce sync.Once
)

func GetStreamManager() *StreamManager {
	streamManagerOnce.Do(func() {
		streamManager = newStreamManager()
		go streamManager.run()
	})

	return streamManager
}

func CreateContextFromTimeout(
	timeoutMS int,
) (context.Context, context.CancelFunc) {
	switch {
	case timeoutMS == 0:
		// Infinite timeout with cancellation capability
		return context.WithCancel(context.Background())
	case timeoutMS > 0:
		// Explicit timeout duration
		duration := time.Duration(timeoutMS) * time.Millisecond

		return context.WithTimeout(context.Background(), duration)
	default:
		// Invalid negative values default to no timeout
		return context.Background(), func() {}
	}
}

func ReadLast(key string, timeout int) (any, error) {
	sub := GetStreamManager().Subscribe(key, true)
	defer sub.Close()

	ctx, cancel := CreateContextFromTimeout(timeout)
	defer cancel()

	for {
		select {
		case msg := <-sub.Messages:
			return msg, nil

		case <-ctx.Done():
			return nil, nil
		}
	}
}
