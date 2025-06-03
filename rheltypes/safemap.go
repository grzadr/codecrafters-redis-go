package rheltypes

import (
	"sync"
)

type SafeMap struct {
	mu   sync.RWMutex
	data map[string]RhelType
}

var (
	instance *SafeMap
	once     sync.Once
)

func GetSageMapInstance() *SafeMap {
	once.Do(func() {
		instance = &SafeMap{
			data: make(map[string]RhelType, 1024),
		}
	})
	return instance
}

func (sm *SafeMap) Set(key string, value RhelType) {
	sm.mu.Lock()
	sm.data[key] = value
	sm.mu.Unlock()
}

func (sm *SafeMap) Get(key string) (RhelType, bool) {
	sm.mu.RLock()
	value, exists := sm.data[key]
	sm.mu.RUnlock()
	return value, exists
}

func (sm *SafeMap) Delete(key string) bool {
	sm.mu.Lock()
	_, exists := sm.data[key]
	if exists {
		delete(sm.data, key)
	}
	sm.mu.Unlock()
	return exists
}
