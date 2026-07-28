package p_llm_assistant

import (
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
		templates, templatesErr = template.New("").ParseFS(templateFS, "templates/*.html")
	})
	return templates, templatesErr
}

func executeTemplate(w io.Writer, name string, data any) error {
	t, err := loadTemplates()
	if err != nil {
		return fmt.Errorf("p_llm_assistant: load templates: %w", err)
	}
	if t.Lookup(name) == nil {
		return fmt.Errorf("p_llm_assistant: template %q not found", name)
	}
	return t.ExecuteTemplate(w, name, data)
}
