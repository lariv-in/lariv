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
			for _, needle := range FindChildren[T](parent) {
				children = append(children, needle)
			}
		}
	}
	return children
}
