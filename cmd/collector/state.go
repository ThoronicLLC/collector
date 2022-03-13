package main

import "sync"

type StateMap struct {
	sync.RWMutex
	internal map[string][]byte
}

func NewStateMap() *StateMap {
	return &StateMap{
		internal: make(map[string][]byte),
	}
}

func (sm *StateMap) Load(key string) (value []byte, ok bool) {
	sm.RLock()
	result, ok := sm.internal[key]
	sm.RUnlock()
	return result, ok
}

func (sm *StateMap) Delete(key string) {
	sm.Lock()
	delete(sm.internal, key)
	sm.Unlock()
}

func (sm *StateMap) Store(key string, value []byte) {
	sm.Lock()
	sm.internal[key] = value
	sm.Unlock()
}
