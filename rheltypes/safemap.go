package rheltypes

import (
	"sync"
	"time"
)

type RhelMapValue struct {
	Value      RhelType
	Expiration int64
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
	sm.mu.RLock()

	valueRaw, found := sm.data[key]
	if found && valueRaw.Expiration > 0 &&
		time.Now().UnixMilli() >= valueRaw.Expiration {
		sm.mu.RUnlock()
		sm.Delete(key)

		value = nil
		found = false
	} else {
		value = valueRaw.Value
		sm.mu.RUnlock()
	}

	return
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
