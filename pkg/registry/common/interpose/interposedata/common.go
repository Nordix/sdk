package interposedata

import (
	"sync"

	"github.com/networkservicemesh/api/pkg/api/registry"
)

// Map helds all information about interpose endpoints
type Map struct {
	sync.RWMutex
	internal map[string]*registry.NetworkServiceEndpoint
}

// NewMap creates a new interpose endpoint data Map
func NewMap() *Map {
	return &Map{
		internal: make(map[string]*registry.NetworkServiceEndpoint),
	}
}

// Load find the map entry by key and returns it
func (m *Map) Load(key string) (value *registry.NetworkServiceEndpoint, ok bool) {
	m.RLock()
	result, ok := m.internal[key]
	m.RUnlock()
	return result, ok
}

// Store add an entry to the map by key
func (m *Map) Store(key string, value *registry.NetworkServiceEndpoint) {
	m.Lock()
	m.internal[key] = value
	m.Unlock()
}

// Delete removes map entry by key
func (m *Map) Delete(key string) {
	m.Lock()
	delete(m.internal, key)
	m.Unlock()
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *Map) Range(f func(key string, value *registry.NetworkServiceEndpoint, i ...interface{}) bool, args ...interface{}) {
	m.Lock()
	for k, v := range m.internal {
		if r := f(k, v, args...); !r {
			break
		}
	}
	m.Unlock()
}
