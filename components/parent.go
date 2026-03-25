package components

import (
	"log/slog"
)

type ParentInterface interface {
	PageInterface
	GetChildren() []PageInterface
}

type MutableParentInterface interface {
	ParentInterface
	SetChildren([]PageInterface)
}

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
