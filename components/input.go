package components

type InputInterface interface {
	Parse(string) (any, error)
	GetName() string
}

func FindInputs(p ParentInterface) []InputInterface {
	inputs := []InputInterface{}
	for _, child := range p.GetChildren() {
		if input, isInput := child.(InputInterface); isInput {
			inputs = append(inputs, input)
		}
		if parent, isParent := child.(ParentInterface); isParent {
			for _, input := range FindInputs(parent) {
				inputs = append(inputs, input)
			}
		}
	}
	return inputs
}
