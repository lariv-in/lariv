package components

import (
	"context"
	"html/template"
	"io"
)

// ClientData wraps child components inside a div element containing Alpine.js state attributes.
// It initializes client-side reactive variables using Alpine's x-data and x-init.
type ClientData struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Data specifies the raw JavaScript expression/object literal to bind to Alpine's x-data.
	// Defaults to "{}" if empty.
	Data string
	// Init specifies an optional raw JavaScript expression to evaluate on Alpine's x-init.
	Init string
	// Children represents the child components enclosed within the state container.
	Children []PageInterface
}

// Build compiles the ClientData component into a div decorated with x-data and x-init attributes.
func (e ClientData) Build(cat Catalog, ctx context.Context, w io.Writer) error {
	data := e.Data
	if data == "" {
		data = "{}"
	}
	children, err := RenderChildren(cat, ctx, e.Children)
	if err != nil {
		return err
	}
	return Execute(w, "client_data", struct {
		Data     string
		Init     string
		Children template.HTML
	}{Data: data, Init: e.Init, Children: children})
}

// GetKey returns the unique key identifier for this ClientData component.
func (e ClientData) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this ClientData.
func (e ClientData) GetRoles() []string {
	return e.Roles
}

// GetChildren returns the slice of child components managed by this ClientData wrapper.
func (e ClientData) GetChildren() []PageInterface {
	return e.Children
}

// SetChildren overwrites the child components managed by this ClientData wrapper.
func (e *ClientData) SetChildren(children []PageInterface) {
	e.Children = children
}
