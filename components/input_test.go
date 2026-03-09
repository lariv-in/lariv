package components

import (
	"testing"
)

func TestInputImplementations(t *testing.T) {
	// Compile-time checks to ensure that the input components implement the InputInterface
	var _ InputInterface = InputCheckbox{}
	var _ InputInterface = InputEmail{}
	var _ InputInterface = InputPassword{}
	var _ InputInterface = InputPhone{}
	var _ InputInterface = InputText{}
}
