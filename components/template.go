package components

import (
	"context"
	"html/template"
	"io"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
)

// TemplateComponent is a component that can be used to render a
// [html/template.Template]
type TemplateComponent struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// Template is the template that will be rendered
	Template template.Template
	// TemplateName is the name of the sub template that needs to be rendered, empty means render the full template
	TemplateName string
	// TemplateContext returns the context that will be available inside of the template
	TemplateContext getters.Getter[any]
}

// GetKey returns the unique key identifier for this EscapedString component.
func (e TemplateComponent) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this EscapedString.
func (e TemplateComponent) GetRoles() []string {
	return e.Roles
}

// Build returns the content wrapped in a Text Node
func (e TemplateComponent) Build(ctx context.Context) Node {
	return NodeFunc(func(w io.Writer) error {
		var data any
		if e.TemplateContext != nil {
			data_, err := e.TemplateContext(ctx)
			data = data_
			if err != nil {
				return err
			}
		}
		if e.TemplateName != "" {
			return e.Template.ExecuteTemplate(w, e.TemplateName, data)
		}
		return e.Template.Execute(w, data)
	})
}
