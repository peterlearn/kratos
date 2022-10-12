package paladin

import (
	"strings"
	"sync/atomic"
)

// KeyNamed key naming to lower case.
func KeyNamed(key string) string {
	return strings.ToLower(key)
}

// Map is config map, key(filename) -> value(file).
type Map struct {
	values atomic.Value
}

// Store sets the value of the Value to values map.
func (m *Map) Store(values map[string]*Value) {
	m.values.Store(values)
}

// Load returns the value set by the most recent Store.
func (m *Map) Load() map[string]*Value {
	src,ok := m.values.Load().(map[string]*Value)
	if ok {
		return src
	}
	return nil
}

// Exist check if values map exist a key.
func (m *Map) Exist(key string) bool {
	mm := m.Load()
	if mm == nil {
		return false
	}
	_, ok := mm[KeyNamed(key)]
	return ok
}

// Get return get value by key.
func (m *Map) Get(key string) *Value {
	mm := m.Load()
	if mm == nil {
		return nil
	}
	v, ok := mm[KeyNamed(key)]
	if ok {
		return v
	}
	return nil
}

// Keys return map keys.
func (m *Map) Keys() []string {
	values := m.Load()
	if values == nil {
		return []string{}
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
