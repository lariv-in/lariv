package components

import (
	"log/slog"
)

// ParentInterface represents a component that houses one or more child sub-components (PageInterface).
type ParentInterface interface {
	PageInterface
	// GetChildren returns the slice of nested sub-components.
	GetChildren() []PageInterface
}

// MutableParentInterface represents a component whose child sub-components can be dynamically replaced or reordered.
type MutableParentInterface interface {
	ParentInterface
	// SetChildren replaces the slice of nested sub-components.
	SetChildren([]PageInterface)
}

// FindChildren recursively traverses a ParentInterface component tree and retrieves all components of type T.
func FindChildren[T PageInterface](p ParentInterface) []T {
	children := []T{}
	for _, child := range p.GetChildren() {
		if needle, isNeedle := child.(T); isNeedle {
			children = append(children, needle)
		}
		if parent, isParent := child.(ParentInterface); isParent {
			children = append(children, FindChildren[T](parent)...)
		}
	}
	return children
}

// ReplaceChild recursively searches a MutableParentInterface hierarchy for a child matching key,
// and replaces it with the output of the replacement callback.
func ReplaceChild[T PageInterface](p MutableParentInterface, key string, replacement func(T) T) {
	children := p.GetChildren()
	for i, child := range children {
		if needle, isNeedle := child.(T); isNeedle && child.GetKey() == key {
			children[i] = replacement(needle)
		} else if parent, isParent := child.(MutableParentInterface); isParent {
			ReplaceChild(parent, key, replacement)
		}
	}
	p.SetChildren(children)
}

// InsertChildBefore recursively searches a MutableParentInterface hierarchy for a child matching key,
// and inserts a replacement component before it in the child list.
func InsertChildBefore[T, V PageInterface](p MutableParentInterface, key string, replacement func(T) V) {
	children := p.GetChildren()
	result := make([]PageInterface, 0, len(children)+1)
	for _, child := range children {
		if needle, isNeedle := child.(T); isNeedle && child.GetKey() == key {
			result = append(result, replacement(needle))
			result = append(result, child)
		} else {
			if parent, isParent := child.(MutableParentInterface); isParent {
				InsertChildBefore(parent, key, replacement)
			}
			result = append(result, child)
		}
	}
	p.SetChildren(result)
}

// InsertChildAfter recursively searches a MutableParentInterface hierarchy for a child matching key,
// and inserts a replacement component after it in the child list. It logs an error if the target key is not found.
func InsertChildAfter[T, V PageInterface](p MutableParentInterface, key string, replacement func(T) V) bool {
	return insertChildAfter[T, V](p, key, replacement, true)
}

func insertChildAfter[T, V PageInterface](p MutableParentInterface, key string, replacement func(T) V, logMiss bool) bool {
	children := p.GetChildren()
	result := make([]PageInterface, 0, len(children)+1)
	targetFound := false
	for _, child := range children {
		if needle, isNeedle := child.(T); isNeedle && child.GetKey() == key {
			result = append(result, child)
			result = append(result, replacement(needle))
			targetFound = true
		} else {
			if parent, isParent := child.(MutableParentInterface); isParent {
				childTargetFound := insertChildAfter(parent, key, replacement, false)
				if childTargetFound {
					targetFound = childTargetFound
				}
			}
			result = append(result, child)
		}
	}

	if !targetFound && logMiss {
		slog.Error("Target not found", "parent", p, "key", key)
	}

	p.SetChildren(result)
	return targetFound
}

// RemoveChild recursively searches a MutableParentInterface hierarchy for a child matching key,
// and removes it from the child list.
func RemoveChild[T PageInterface](p MutableParentInterface, key string) bool {
	children := p.GetChildren()
	result := make([]PageInterface, 0, len(children))
	targetFound := false
	for _, child := range children {
		if _, isNeedle := child.(T); isNeedle && child.GetKey() == key {
			targetFound = true
		} else {
			if parent, isParent := child.(MutableParentInterface); isParent {
				childTargetFound := RemoveChild[T](parent, key)
				if childTargetFound {
					targetFound = childTargetFound
				}
			}
			result = append(result, child)
		}
	}

	p.SetChildren(result)
	return targetFound
}
