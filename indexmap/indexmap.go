package indexmap

import (
	"iter"
	"sync"
)

// IndexMap is a thread-safe map that maintains the order of keys.
type IndexMap[k comparable, v comparable] struct {
	mu     sync.RWMutex
	keys   []k
	values map[k]v
}

// NewIndexMap creates and returns a new IndexMap instance.
func NewIndexMap[k comparable, v comparable]() *IndexMap[k, v] {
	return &IndexMap[k, v]{}
}

// Set stores a value for key.
// If key is new, the insertion order is preserved.
func (m *IndexMap[k, v]) Set(key k, value v) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.values == nil {
		m.values = make(map[k]v)
	}
	if _, ok := m.values[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

// Get returns the value for key and whether it exists.
func (m *IndexMap[k, v]) Get(key k) (v, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	val, ok := m.values[key]
	return val, ok
}

// Index returns the key/value pair at index n in insertion order.
func (m *IndexMap[k, v]) Index(n int) (k, v, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var zeroK k
	var zeroV v

	if n < 0 || n >= len(m.keys) {
		return zeroK, zeroV, false
	}
	key := m.keys[n]
	value, ok := m.values[key]
	if !ok {
		return zeroK, zeroV, false
	}
	return key, value, true
}

// All returns an iterator of key/value pairs in insertion order.
func (m *IndexMap[k, v]) All() iter.Seq2[k, v] {
	type item struct {
		key   k
		value v
	}

	m.mu.RLock()
	items := make([]item, 0, len(m.keys))
	for _, k := range m.keys {
		v, ok := m.values[k]
		if !ok {
			continue
		}
		items = append(items, item{key: k, value: v})
	}
	m.mu.RUnlock()

	return func(yield func(k, v) bool) {
		for _, it := range items {
			if !yield(it.key, it.value) {
				return
			}
		}
	}
}

// Keys returns an iterator of keys in insertion order.
func (m *IndexMap[k, v]) Keys() iter.Seq[k] {
	m.mu.RLock()
	keys := make([]k, len(m.keys))
	copy(keys, m.keys)
	m.mu.RUnlock()

	return func(yield func(k) bool) {
		for _, k := range keys {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns an iterator of values in insertion order.
func (m *IndexMap[k, v]) Values() iter.Seq[v] {
	m.mu.RLock()
	values := make([]v, 0, len(m.keys))
	for _, k := range m.keys {
		v, ok := m.values[k]
		if !ok {
			continue
		}
		values = append(values, v)
	}
	m.mu.RUnlock()

	return func(yield func(v) bool) {
		for _, v := range values {
			if !yield(v) {
				return
			}
		}
	}
}

// SetValues sets values for keys in insertion order.
// If there are more values than keys, the extra values are ignored.
// If there are fewer values than keys, the remaining keys are set to the zero value of v.
func (m *IndexMap[k, v]) SetValues(slice []v) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.values == nil {
		m.values = make(map[k]v)
	}
	for i, v := range slice {
		var zeroK k
		if i < len(m.keys) {
			zeroK = m.keys[i]
		} else {
			zeroK = *new(k)
			m.keys = append(m.keys, zeroK)
		}
		m.values[zeroK] = v
	}
}

func (m *IndexMap[k, v]) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.keys)
}
