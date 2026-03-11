package components

type ParentInterface interface {
	GetChildren() []PageInterface
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
