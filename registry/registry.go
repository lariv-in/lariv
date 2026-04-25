package registry

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"sort"
	"strings"
	"sync"

	"github.com/lariv-in/lago/getters"
)

func NewRegistry[T any]() *Registry[T] {
	return &Registry[T]{
		unpatchedItems: map[string]RegistryItem[T]{},
		patches:        map[string][]func(T) T{},
		items:          map[string]RegistryItem[T]{},
		itemsList:      make(map[RegistrySorter[T]]*registryItems[T]),
		isBuilt:        false,
	}
}

type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

// PairsFromMap converts a map into a stable slice of pairs sorted by key.
func PairsFromMap[K cmp.Ordered, V any](m map[K]V) []Pair[K, V] {
	pairs := make([]Pair[K, V], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, Pair[K, V]{
			Key:   k,
			Value: v,
		})
	}
	slices.SortFunc(pairs, func(a, b Pair[K, V]) int {
		return cmp.Compare(a.Key, b.Key)
	})
	return pairs
}

// PairFromMap returns the pair for key when it exists in m.
func PairFromMap[K comparable, V any](key K, m map[K]V) (Pair[K, V], bool) {
	v, ok := m[key]
	if !ok {
		return Pair[K, V]{}, false
	}
	return Pair[K, V]{
		Key:   key,
		Value: v,
	}, true
}

// MapFromPairs builds a map from pairs. If the same key appears more than once, the last occurrence wins.
func MapFromPairs[K comparable, V any](pairs []Pair[K, V]) map[K]V {
	m := make(map[K]V, len(pairs))
	for _, p := range pairs {
		m[p.Key] = p.Value
	}
	return m
}

// PairFromPairs returns the first pair whose Key equals key.
func PairFromPairs[K comparable, V any](key K, pairs []Pair[K, V]) (Pair[K, V], bool) {
	for _, p := range pairs {
		if p.Key == key {
			return p, true
		}
	}
	return Pair[K, V]{}, false
}

// KeysFromPairs returns each pair's Key in slice order (e.g. for generators or tests).
func KeysFromPairs[K comparable, V any](pairs []Pair[K, V]) []K {
	out := make([]K, len(pairs))
	for i, p := range pairs {
		out[i] = p.Key
	}
	return out
}

// PairValueFromKey maps a stored string key to its value using [PairFromPairs].
// Empty key returns "". Unknown keys return the key unchanged (same fallback as form selects in Caveats).
func PairValueFromKey(keyGetter getters.Getter[string], pairs []Pair[string, string]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := keyGetter(ctx)
		if err != nil {
			return "", err
		}
		if s == "" {
			return "", nil
		}
		if p, ok := PairFromPairs(s, pairs); ok {
			return p.Value, nil
		}
		return s, nil
	}
}

// PairFromGetter returns a getter for components.InputSelect current value: a
// Pair with Key = stored value and Value = label. Empty key returns a zero pair.
// Unknown keys return {Key: s, Value: s}, matching plugin-local *PairGetter helpers.
func PairFromGetter(keyGetter getters.Getter[string], pairs []Pair[string, string]) getters.Getter[Pair[string, string]] {
	return func(ctx context.Context) (Pair[string, string], error) {
		s, err := keyGetter(ctx)
		if err != nil || s == "" {
			return Pair[string, string]{}, nil
		}
		if p, ok := PairFromPairs(s, pairs); ok {
			return p, nil
		}
		return Pair[string, string]{Key: s, Value: s}, nil
	}
}

func (p Pair[K, V]) ToKVJson() string {
	b, err := json.Marshal(map[K]V{p.Key: p.Value})
	if err != nil {
		slog.Error("Could not marshal Pair to JSON", "error", err)
		return ""
	}
	return string(b)
}

// RegistryItem is the stored value plus registration order for custom [RegistrySorter]s.
// Use a stable zero-value sorter (e.g. [AlphabeticalByKey], [RegisterOrder]) as the map key for [Registry.AllStable] caching.
type RegistryItem[T any] struct {
	Order int
	Item  T
}

// RegistrySorter orders entries for [Registry.AllStable]. Compare returns <0 if a sorts before b,
// 0 if equal for ordering purposes, >0 if a sorts after b (same contract as [strings.Compare] / [cmp.Compare]).
type RegistrySorter[T any] interface {
	Compare(a, b Pair[string, RegistryItem[T]]) int
}

// AlphabeticalByKey sorts by registry name (string key), using [strings.Compare].
type AlphabeticalByKey[T any] struct{}

func (AlphabeticalByKey[T]) Compare(a, b Pair[string, RegistryItem[T]]) int {
	return strings.Compare(a.Key, b.Key)
}

// RegisterOrder sorts by [RegistryItem.Order] (the order entries were registered).
type RegisterOrder[T any] struct{}

func (RegisterOrder[T]) Compare(a, b Pair[string, RegistryItem[T]]) int {
	return cmp.Compare(a.Value.Order, b.Value.Order)
}

type registryItems[T any] struct {
	isBuilt bool
	items   []Pair[string, T]
}

type Registry[T any] struct {
	unpatchedItems map[string]RegistryItem[T]
	patches        map[string][]func(T) T
	items          map[string]RegistryItem[T]
	itemsList      map[RegistrySorter[T]]*registryItems[T]
	isBuilt        bool
	isBuilding     bool
	mu             sync.RWMutex
	cond           *sync.Cond
}

func (r *Registry[T]) Register(name string, unpatchedItem T) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.initCondLocked()
	for r.isBuilding {
		r.cond.Wait()
	}

	_, isPresent := r.unpatchedItems[name]
	if isPresent {
		return fmt.Errorf("entry with name %s is already present in the registry, consider patching it instead", name)
	}
	r.unpatchedItems[name] = RegistryItem[T]{Order: len(r.unpatchedItems), Item: unpatchedItem}
	r.isBuilt = false
	return nil
}

func (r *Registry[T]) Patch(name string, patcher func(T) T) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.initCondLocked()
	for r.isBuilding {
		r.cond.Wait()
	}

	if len(r.patches[name]) == 0 {
		r.patches[name] = []func(T) T{patcher}
	} else {
		r.patches[name] = append(r.patches[name], patcher)
	}
	r.isBuilt = false
}

func (r *Registry[T]) initCondLocked() {
	if r.cond == nil {
		r.cond = sync.NewCond(&r.mu)
	}
}

func (r *Registry[T]) buildLocked() {
	if r.isBuilt || r.isBuilding {
		return
	}
	r.initCondLocked()
	r.isBuilding = true

	baseItems := maps.Clone(r.unpatchedItems)
	basePatches := maps.Clone(r.patches)
	r.mu.Unlock()

	items := maps.Clone(baseItems)
	patches := maps.Clone(basePatches)

	for k := range basePatches {
		_, isItemPresent := items[k]
		if !isItemPresent {
			continue
		}
		p := patches[k]
		delete(patches, k)
		for _, patcher := range p {
			ri := items[k]
			ri.Item = patcher(ri.Item)
			items[k] = ri
		}
	}

	r.mu.Lock()
	defer r.cond.Broadcast()
	defer func() { r.isBuilding = false }()

	// Fold applied patches into the base so a later build still sees the
	// transformed value; drop them from the pending patch list.
	for k := range basePatches {
		if item, ok := items[k]; ok {
			r.unpatchedItems[k] = item
		}
	}
	r.patches = patches

	clear(r.items)
	for k, v := range items {
		r.items[k] = v
	}

	if len(r.patches) > 0 {
		slog.Warn("The following patches were not applied since no corresponding keys were found in the registry", "patches", r.patches)
	}

	r.itemsList = make(map[RegistrySorter[T]]*registryItems[T])

	r.isBuilt = true
}

func (r *Registry[T]) Get(name string) (T, bool) {
	var zero T

	r.mu.RLock()
	if r.isBuilt {
		ri, isPresent := r.items[name]
		r.mu.RUnlock()
		if !isPresent {
			return zero, false
		}
		return ri.Item, true
	}
	// Avoid recursion deadlocks when patches/getters re-enter the same registry
	// while a build is in progress.
	if r.isBuilding {
		r.mu.RUnlock()
		return zero, false
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.isBuilt {
		r.buildLocked()
	}
	ri, isPresent := r.items[name]
	if !isPresent {
		return zero, false
	}
	return ri.Item, true
}

func (r *Registry[T]) Getter(name string) getters.Getter[T] {
	var zero T
	return func(ctx context.Context) (T, error) {
		if v, isPresent := r.Get(name); isPresent {
			return v, nil
		}
		return zero, fmt.Errorf("Couldn't find the value for %s in the registry", name)
	}
}

func (r *Registry[T]) All() map[string]T {
	r.mu.RLock()
	if r.isBuilt {
		out := make(map[string]T, len(r.items))
		for k, ri := range r.items {
			out[k] = ri.Item
		}
		r.mu.RUnlock()
		return out
	}
	if r.isBuilding {
		r.mu.RUnlock()
		return map[string]T{}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.isBuilt {
		r.buildLocked()
	}
	out := make(map[string]T, len(r.items))
	for k, ri := range r.items {
		out[k] = ri.Item
	}
	return out
}

// AllStable returns an internal cached slice ordered by sorter. Pass a stable sorter value
// (e.g. [AlphabeticalByKey] zero value) so cache lookups match.
// Treat the returned slice as immutable and do not mutate it.
func (r *Registry[T]) AllStable(sorter RegistrySorter[T]) *[]Pair[string, T] {
	r.mu.RLock()
	if r.isBuilt {
		ent := r.itemsList[sorter]
		r.mu.RUnlock()
		if ent != nil && ent.isBuilt {
			return &ent.items
		}
		r.mu.Lock()
		defer r.mu.Unlock()
		if !r.isBuilt {
			r.buildLocked()
		}
	} else {
		if r.isBuilding {
			r.mu.RUnlock()
			return new([]Pair[string, T]{})
		}
		r.mu.RUnlock()
		r.mu.Lock()
		defer r.mu.Unlock()
		if !r.isBuilt {
			r.buildLocked()
		}
	}

	if !r.isBuilt {
		return new([]Pair[string, T]{})
	}
	ent := r.itemsList[sorter]
	if ent == nil {
		ent = &registryItems[T]{}
		r.itemsList[sorter] = ent
	}
	if !ent.isBuilt {
		buf := make([]Pair[string, RegistryItem[T]], 0, len(r.items))
		for k, v := range r.items {
			buf = append(buf, Pair[string, RegistryItem[T]]{Key: k, Value: v})
		}
		sort.SliceStable(buf, func(i, j int) bool {
			return sorter.Compare(buf[i], buf[j]) < 0
		})
		out := make([]Pair[string, T], len(buf))
		for i := range buf {
			out[i] = Pair[string, T]{Key: buf[i].Key, Value: buf[i].Value.Item}
		}
		ent.items = out
		ent.isBuilt = true
	}
	return &ent.items
}

func PairGetter[K comparable, V any](keyGetter getters.Getter[K], mapGetter getters.Getter[map[K]V]) getters.Getter[Pair[K, V]] {
	return func(ctx context.Context) (Pair[K, V], error) {
		key, err := keyGetter(ctx)
		if err != nil {
			return Pair[K, V]{}, err
		}
		values, err := mapGetter(ctx)
		if err != nil {
			return Pair[K, V]{}, err
		}
		value, ok := values[key]
		if !ok {
			return Pair[K, V]{}, fmt.Errorf("key %v not found in map", key)
		}
		return Pair[K, V]{Key: key, Value: value}, nil
	}
}
