package collector

import (
	"github.com/ThoronicLLC/collector/internal/app/manager"
	"sync"
)

type instanceManagerMap struct {
	sync.RWMutex
	internalMap map[string]*manager.Manager
}

func NewInstanceManagerMap() *instanceManagerMap {
	return &instanceManagerMap{
		RWMutex:     sync.RWMutex{},
		internalMap: make(map[string]*manager.Manager),
	}
}

func (m *instanceManagerMap) Get(key string) (*manager.Manager, bool) {
	m.RLock()
	defer m.RUnlock()
	value, exists := m.internalMap[key]
	return value, exists
}

func (m *instanceManagerMap) Set(key string, value *manager.Manager) {
	m.Lock()
	m.internalMap[key] = value
	m.Unlock()
}

func (m *instanceManagerMap) Delete(key string) {
	m.Lock()
	delete(m.internalMap, key)
	m.Unlock()
}

func (m *instanceManagerMap) ListKeys() []string {
	m.RLock()
	defer m.RUnlock()
	keyList := make([]string, 0)
	for k := range m.internalMap {
		keyList = append(keyList, k)
	}
	return keyList
}

func (m *instanceManagerMap) List() []*manager.Manager {
	m.RLock()
	defer m.RUnlock()
	valueList := make([]*manager.Manager, 0)
	for _, v := range m.internalMap {
		valueList = append(valueList, v)
	}
	return valueList
}
