package components

import (
	"context"
	"mime/multipart"
)

// InputInterface defines the standard behavior for all user input fields within a form.
// It supports retrieving the input field name identifier and parsing raw POST request values.
type InputInterface interface {
	PageInterface
	// Parse cleans and parses the input values from a form request, returning the parsed value or validation error.
	Parse(value any, ctx context.Context) (any, error)
	// GetName returns the HTML form element's name attribute (e.g., "email" or "username").
	GetName() string
}

// MultipartInputInterface extends InputInterface to allow file-upload inputs to retrieve and process
// file headers from multipart form request contexts.
type MultipartInputInterface interface {
	InputInterface
	// ParseMultipart parses uploaded file headers from a multipart form, returning a parsed result or error.
	ParseMultipart(headers []*multipart.FileHeader, ctx context.Context) (any, error)
}

// FindInputs performs a recursive post-order traversal through a parent component's children
// to locate and return all nested components that implement the [InputInterface].
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
