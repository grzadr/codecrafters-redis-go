package rheltypes

import (
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
	mu   sync.RWMutex
	data map[string]RhelMapValue
}

var (
	instance *SafeMap
	once     sync.Once
)

const DEFAULT_MAP_CAPACITY = 1024

func GetSageMapInstance() *SafeMap {
	once.Do(func() {
		instance = &SafeMap{
			data: make(map[string]RhelMapValue, DEFAULT_MAP_CAPACITY),
		}
	})

	return instance
}

func (sm *SafeMap) Set(key string, value RhelType) {
	sm.mu.Lock()
	sm.data[key] = RhelMapValue{Value: value}
	sm.mu.Unlock()
}

func (sm *SafeMap) SetToExpire(key string, value RhelType, px int64) {
	sm.mu.Lock()
	d := time.Duration(px) * time.Millisecond
	sm.data[key] = RhelMapValue{
		Value:      value,
		Expiration: time.Now().Add(d).UnixMilli(),
	}
	sm.mu.Unlock()
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

func (sm *SafeMap) deleteExpired(key string) (deleted bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	v, found := sm.data[key]

	if found && v.IsExpired() {
		delete(sm.data, key)

		deleted = true
	}

	return
}

func (sm *SafeMap) getValue(key string) (value RhelMapValue, found bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	value, found = sm.data[key]

	return
}
