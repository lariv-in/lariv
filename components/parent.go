package components

type ParentInterface interface {
	PageInterface
	GetChildren() []PageInterface
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

func ReplaceChild[T PageInterface](p ParentInterface, key string, replacement func(T) T) {
	children := p.GetChildren()
	for i, child := range children {
		if needle, isNeedle := child.(T); isNeedle && child.GetKey() == key {
			children[i] = replacement(needle)
		} else if parent, isParent := child.(ParentInterface); isParent {
			ReplaceChild[T](parent, key, replacement)
		}
	}
	p.SetChildren(children)
}

func InsertChildBefore[T PageInterface](p ParentInterface, key string, replacement func(T) T) {
	children := p.GetChildren()
	result := make([]PageInterface, 0, len(children)+1)
	for _, child := range children {
		if needle, isNeedle := child.(T); isNeedle && child.GetKey() == key {
			result = append(result, replacement(needle))
			result = append(result, child)
		} else {
			if parent, isParent := child.(ParentInterface); isParent {
				InsertChildBefore[T](parent, key, replacement)
			}
			result = append(result, child)
		}
	}
	p.SetChildren(result)
}

func InsertChildAfter[T PageInterface](p ParentInterface, key string, replacement func(T) T) {
	children := p.GetChildren()
	result := make([]PageInterface, 0, len(children)+1)
	for _, child := range children {
		if needle, isNeedle := child.(T); isNeedle && child.GetKey() == key {
			result = append(result, child)
			result = append(result, replacement(needle))
		} else {
			if parent, isParent := child.(ParentInterface); isParent {
				InsertChildAfter[T](parent, key, replacement)
			}
			result = append(result, child)
		}
	}
	p.SetChildren(result)
}
