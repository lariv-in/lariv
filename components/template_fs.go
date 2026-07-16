package components

import (
	"context"
	"html/template"
	"io"
	"io/fs"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
)

// TemplateComponent is a component that can be used to render a
// [html/template.Template]
type TemplateFSComponent struct {
	// Page embeds common component properties like Key and Roles.
	Page

	template *template.Template
	// TemplatePatterns are the patterns that will be used for getting the templates from Filesystem
	TemplatePatterns []string
	// TemplateName is the name of the sub template that needs to be rendered, empty means render the full template
	TemplateName string
	// Filesystem is the filesystem that will be used for looking up the file
	Filesystem fs.FS
	// TemplateContext returns the context that will be available inside of the template
	TemplateContext getters.Getter[any]
}

// GetKey returns the unique key identifier for this EscapedString component.
func (e TemplateFSComponent) GetKey() string {
	return e.Key
}

// GetRoles returns the authorized roles required to view this EscapedString.
func (e TemplateFSComponent) GetRoles() []string {
	return e.Roles
}

func (e *TemplateFSComponent) CompileTemplate() error {
	template, err := template.ParseFS(e.Filesystem, e.TemplatePatterns...)
	if err != nil {
		return err
	}
	e.template = template
	return nil
}

// Build returns the rendered content of template as is
func (e TemplateFSComponent) Build(ctx context.Context) Node {
	return NodeFunc(func(w io.Writer) error {
		if e.template == nil {
			err := e.CompileTemplate()
			if err != nil {
				return err
			}
		}
		var data any
		if e.TemplateContext != nil {
			data_, err := e.TemplateContext(ctx)
			data = data_
			if err != nil {
				return err
			}
		}
		if e.TemplateName != "" {
			return e.template.ExecuteTemplate(w, e.TemplateName, data)
		}
		return e.template.Execute(w, data)
	})
}
