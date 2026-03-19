package components

import (
	"context"
	"mime/multipart"
)

type InputInterface interface {
	PageInterface
	Parse(any, context.Context) (any, error)
	GetName() string
}

// MultipartInputInterface allows an input to receive uploaded file parts from a
// multipart form instead of only text values.
type MultipartInputInterface interface {
	InputInterface
	ParseMultipart([]*multipart.FileHeader, context.Context) (any, error)
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
