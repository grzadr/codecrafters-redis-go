package rheltypes

import (
	"iter"
	"maps"
	"sync"
	"time"
)

func currentTime() int64 {
	return time.Now().UnixMilli()
}

type RhelMapValue struct {
	Value      RhelType
	Expiration int64
}

func (v RhelMapValue) IsExpired() bool {
	return v.Expiration > 0 && currentTime() >= v.Expiration
}

type SafeMap struct {
	mu     sync.RWMutex
	data   map[string]RhelMapValue
	ticker *time.Ticker
	done   chan struct{}
}

func NewSafeMap(cleanupInterval time.Duration) *SafeMap {
	sm := &SafeMap{
		data: make(map[string]RhelMapValue),
		done: make(chan struct{}),
	}

	if cleanupInterval > 0 {
		sm.ticker = time.NewTicker(cleanupInterval)
		go sm.cleanupExpired()
	}

	return sm
}

func (sm *SafeMap) SetToExpire(key string, value RhelType, px int64) {
	sm.mu.Lock()
	if px > 0 {
		px += currentTime()
	}

	sm.data[key] = RhelMapValue{
		Value:      value,
		Expiration: px,
	}

	sm.mu.Unlock()
}

func (sm *SafeMap) Set(key string, value RhelType) {
	sm.SetToExpire(key, value, 0)
}

func (sm *SafeMap) SetString(key, value string, px int64) {
	rhelValue := NewBulkString(value)

	sm.SetToExpire(key, rhelValue, px)
}

func (sm *SafeMap) Get(key string) (value RhelType, found bool) {
	valueRaw, found := sm.getValue(key)

	if found && valueRaw.IsExpired() && sm.deleteExpired(key) {
		return nil, false
	}

	return valueRaw.Value, found
}

func (sm *SafeMap) Delete(key string) bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	_, exists := sm.data[key]
	if exists {
		delete(sm.data, key)
	}

	return exists
}

func (sm *SafeMap) Close() {
	if sm == nil {
		panic("map is nil")
	}

	if sm.ticker != nil {
		sm.ticker.Stop()
	}

	close(sm.done)
}

func (sm *SafeMap) Keys() iter.Seq[string] {
	return maps.Keys(sm.data)
}

func (sm *SafeMap) getValue(key string) (value RhelMapValue, found bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, found = sm.data[key]

	return
}

func (sm *SafeMap) deleteExpired(key string) (deleted bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	v, found := sm.data[key]

	if deleted = found && v.IsExpired(); deleted {
		delete(sm.data, key)
	}

	return
}

func (sm *SafeMap) cleanupExpired() {
	for {
		select {
		case <-sm.ticker.C:
			sm.scanAndDelete()
		case <-sm.done:
			return
		}
	}
}

func (sm *SafeMap) scanAndDelete() {
	for key := range sm.data {
		go sm.deleteExpired(key)
	}
}
