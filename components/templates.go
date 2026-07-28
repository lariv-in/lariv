package components

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"sync"
)

//go:embed templates/*.html
var templateFS embed.FS

var (
	templatesOnce sync.Once
	templates     *template.Template
	templatesErr  error
)

func loadTemplates() (*template.Template, error) {
	templatesOnce.Do(func() {
		t := template.New("").Funcs(template.FuncMap{
			"attrs":    renderAttrs,
			"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		})
		templates, templatesErr = t.ParseFS(templateFS, "templates/*.html")
	})
	return templates, templatesErr
}

// Execute renders the named embedded template with data to w.
func Execute(w io.Writer, name string, data any) error {
	t, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("components: load templates: %w", err)
	}
	if t.Lookup(name) == nil {
		return fmt.Errorf("components: template %q not found", name)
	}
	return t.ExecuteTemplate(w, name, data)
}

// RenderChildren renders each child and concatenates the HTML.
func RenderChildren(cat Catalog, ctx context.Context, children []PageInterface) (template.HTML, error) {
	if len(children) == 0 {
		return "", nil
	}
	var b bytes.Buffer
	for _, child := range children {
		if err := Render(child, cat, ctx, &b); err != nil {
			return "", err
		}
	}
	return template.HTML(b.String()), nil
}

// RenderAll renders each child to a separate HTML fragment.
func RenderAll(cat Catalog, ctx context.Context, children []PageInterface) ([]template.HTML, error) {
	out := make([]template.HTML, 0, len(children))
	for _, child := range children {
		h, err := RenderHTML(child, cat, ctx)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, nil
}
