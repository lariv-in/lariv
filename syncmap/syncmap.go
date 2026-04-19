package syncmap

import "sync"

// SyncMap is a typed map[K]V guarded by a [sync.RWMutex]. Method names and shapes match [sync.Map].
// The zero SyncMap is empty and ready for use. A SyncMap must not be copied after first use.
type SyncMap[K comparable, V any] struct {
	mu      sync.RWMutex
	entries map[K]V
}

func (m *SyncMap[K, V]) ensure() {
	if m.entries == nil {
		m.entries = make(map[K]V)
	}
}

// Load returns the value stored in the map for a key, or the zero value if no value is present.
// The ok result reports whether an item was found.
func (m *SyncMap[K, V]) Load(key K) (value V, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.entries == nil {
		return value, false
	}
	value, ok = m.entries[key]
	return value, ok
}

// Store sets the value for a key.
func (m *SyncMap[K, V]) Store(key K, value V) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure()
	m.entries[key] = value
}

// Delete removes the value for a key.
func (m *SyncMap[K, V]) Delete(key K) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entries == nil {
		return
	}
	delete(m.entries, key)
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (m *SyncMap[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entries == nil {
		return value, false
	}
	v, ok := m.entries[key]
	if !ok {
		return value, false
	}
	delete(m.entries, key)
	return v, true
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *SyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure()
	if v, ok := m.entries[key]; ok {
		return v, true
	}
	m.entries[key] = value
	return value, false
}

// Swap swaps the value for a key and returns the previous value if any.
// The loaded result reports whether the key was present.
func (m *SyncMap[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ensure()
	if v, ok := m.entries[key]; ok {
		previous = v
		loaded = true
	}
	m.entries[key] = value
	return previous, loaded
}

// CompareAndSwap swaps the old and new values for key if the value stored in the map is equal to old.
// The old value must be of a comparable type (same rules as [sync.Map.CompareAndSwap]).
func (m *SyncMap[K, V]) CompareAndSwap(key K, old, new any) (swapped bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entries == nil {
		return false
	}
	cur, present := m.entries[key]
	if !present {
		return false
	}
	oldV, okOld := old.(V)
	if !okOld {
		return false
	}
	newV, okNew := new.(V)
	if !okNew {
		return false
	}
	var a, b any = cur, oldV
	if a != b {
		return false
	}
	m.entries[key] = newV
	return true
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
// The old value must be of a comparable type (same rules as [sync.Map.CompareAndDelete]).
func (m *SyncMap[K, V]) CompareAndDelete(key K, old any) (deleted bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entries == nil {
		return false
	}
	cur, present := m.entries[key]
	if !present {
		return false
	}
	oldV, okOld := old.(V)
	if !okOld {
		return false
	}
	var a, b any = cur, oldV
	if a != b {
		return false
	}
	delete(m.entries, key)
	return true
}

// Range calls f sequentially for each key and value in the map. If f returns false, range stops the iteration.
//
// Range takes a snapshot of the map; f is not called while the lock is held (so f must not call back into this SyncMap on the same goroutine expecting progress).
func (m *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	m.mu.RLock()
	if len(m.entries) == 0 {
		m.mu.RUnlock()
		return
	}
	keys := make([]K, 0, len(m.entries))
	vals := make([]V, 0, len(m.entries))
	for k, v := range m.entries {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	m.mu.RUnlock()
	for i := range keys {
		if !f(keys[i], vals[i]) {
			return
		}
	}
}

// Clear deletes all the entries, resulting in an empty map.
func (m *SyncMap[K, V]) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.entries != nil {
		clear(m.entries)
	}
}
