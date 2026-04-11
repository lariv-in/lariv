package components

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

type ButtonPost struct {
	Page
	Label       string
	URL         getters.Getter[string]
	Icon        string
	IconClasses string
	Classes     string
	// Attr is merged onto the submit button (e.g. Disabled(), extra classes).
	Attr getters.Getter[Node]
}

func (e ButtonPost) GetKey() string {
	return e.Key
}

func (e ButtonPost) GetRoles() []string {
	return e.Roles
}

func (e ButtonPost) Build(ctx context.Context) Node {
	url := ""
	if e.URL != nil {
		if v, err := e.URL(ctx); err == nil {
			url = v
		}
	}

	content := Group{}
	if e.Icon != "" {
		content = append(content, Render(Icon{Name: e.Icon, Classes: e.IconClasses}, ctx))
	}
	if e.Label != "" {
		content = append(content, Text(e.Label))
	}

	buttonClasses := "btn " + e.Classes
	if e.Icon != "" && e.Label != "" {
		buttonClasses += " inline-flex items-center gap-2"
	}

	return Form(
		Action(url), Method(http.MethodPost),
		// Use htmx boost so the POST is handled via HTMX without a
		// full-page navigation; the response (e.g. updated detail view
		// showing "Generating..." state) will be swapped in-place.
		Attr("hx-boost", "true"),
		Attr("@click.stop", ""),
		Button(
			Type("submit"),
			Class(buttonClasses),
			Iff(e.Attr != nil, func() Node {
				n, err := e.Attr(ctx)
				if err != nil {
					slog.Error("ButtonPost Attr getter failed", "error", err, "key", e.Key)
					return Raw("")
				}
				if n == nil {
					return Raw("")
				}
				return n
			}),
			content,
		),
	)
}
