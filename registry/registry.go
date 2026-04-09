package registry

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/lariv-in/lago/getters"
)

func NewRegistry[T any]() Registry[T] {
	return Registry[T]{
		unpatchedItems: map[string]T{},
		patches:        map[string][]func(T) T{},
		items:          map[string]T{},
		isBuilt:        false,
		itemsList:      []Pair[string, T]{},
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

func (p Pair[K, V]) ToKVJson() string {
	b, err := json.Marshal(map[K]V{p.Key: p.Value})
	if err != nil {
		slog.Error("Could not marshal Pair to JSON", "error", err)
		return ""
	}
	return string(b)
}

type Registry[T any] struct {
	unpatchedItems map[string]T
	patches        map[string][]func(T) T
	items          map[string]T
	itemsList      []Pair[string, T]
	isBuilt        bool
	isBuilding     bool
}

func (r *Registry[T]) Register(name string, unpatchedItem T) error {
	_, isPresent := r.unpatchedItems[name]
	if isPresent {
		return fmt.Errorf("Entry with name %s is already present in the registry %#v, consider patching it instead", name, *r)
	}
	r.unpatchedItems[name] = unpatchedItem
	r.isBuilt = false
	return nil
}

func (r *Registry[T]) Patch(name string, patcher func(T) T) {
	if len(r.patches[name]) == 0 {
		r.patches[name] = []func(T) T{patcher}
	} else {
		r.patches[name] = append(r.patches[name], patcher)
	}
	r.isBuilt = false
}

func (r *Registry[T]) Build() {
	if r.isBuilding {
		return
	}
	r.isBuilding = true
	defer func() { r.isBuilding = false }()

	items := maps.Clone(r.unpatchedItems)
	patches := maps.Clone(r.patches)

	for k := range r.patches {
		_, isItemPresent := items[k]
		if !isItemPresent {
			continue
		}
		p := patches[k]
		delete(patches, k)
		for _, patcher := range p {
			items[k] = patcher(items[k])
		}
		// Fold applied patches into the base so a later Build still sees the
		// transformed value; drop them from the pending patch list.
		r.unpatchedItems[k] = items[k]
	}

	r.patches = patches

	maps.Copy(r.items, items)

	if len(r.patches) > 0 {
		slog.Warn("The following patches were not applied since no corresponding keys were found in the registry", "registry", *r, "patches", r.patches)
	}

	r.itemsList = []Pair[string, T]{}
	for k, v := range items {
		r.itemsList = append(r.itemsList, Pair[string, T]{
			Key:   k,
			Value: v,
		})
	}

	slices.SortFunc(r.itemsList, func(a, b Pair[string, T]) int {
		return strings.Compare(a.Key, b.Key)
	})

	r.isBuilt = true
}

func (r *Registry[T]) Get(name string) (T, bool) {
	var zero T

	if !r.isBuilt {
		// Avoid infinite recursion when patches or getters try to resolve
		// registry entries while a build is already in progress.
		if r.isBuilding {
			return zero, false
		}
		r.Build()
	}
	v, isPresent := r.items[name]
	return v, isPresent
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
	if !r.isBuilt {
		r.Build()
	}
	return r.items
}

func (r *Registry[T]) AllStable() *[]Pair[string, T] {
	if !r.isBuilt {
		r.Build()
	}
	return &r.itemsList
}
