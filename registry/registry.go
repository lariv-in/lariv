package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"

	"github.com/lariv-in/getters"
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
	}

	maps.Copy(r.items, items)

	if len(patches) > 0 {
		slog.Warn("The following patches were not applied since no corresponding keys were found in the registry", "registry", *r, "patches", patches)
	}

	r.itemsList = []Pair[string, T]{}
	for k, v := range items {
		r.itemsList = append(r.itemsList, Pair[string, T]{
			Key:   k,
			Value: v,
		})
	}

	slices.SortFunc(r.itemsList, func(a Pair[string, T], b Pair[string, T]) int {
		return strings.Compare(a.Key, b.Key)
	})

	r.isBuilt = true
}

func (r *Registry[T]) Get(name string) (T, bool) {
	if !r.isBuilt {
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
