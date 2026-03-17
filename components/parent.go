package components

import "fmt"

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
	children := p.GetChildren()
	result := make([]PageInterface, 0, len(children)+1)
	targetFound := false
	for _, child := range children {
		if needle, isNeedle := child.(T); isNeedle && child.GetKey() == key {
			result = append(result, child)
			result = append(result, replacement(needle))
			targetFound = true
			break
		} else {
			if parent, isParent := child.(MutableParentInterface); isParent {
				targetFound = InsertChildAfter(parent, key, replacement)
				if targetFound {
					return targetFound
				}
			}
			result = append(result, child)
		}
	}

	if !targetFound {
		fmt.Println("Target not found")
	}

	p.SetChildren(result)
	return targetFound
}
